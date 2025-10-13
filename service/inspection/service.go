package inspection

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"inspection-service/cluster/analyzer"
	"inspection-service/cluster/brigade"
	"inspection-service/cluster/file"
	"inspection-service/cluster/subscriber"
	"inspection-service/cluster/task"
	"inspection-service/config"
	"inspection-service/database/inspection"
	"io"
	"path/filepath"
	"strings"
	"time"

	"github.com/lukasjarosch/go-docx"
	"github.com/shopspring/decimal"
	"github.com/sunshineOfficial/golib/goctx"
	"github.com/sunshineOfficial/golib/gokafka"
	"github.com/sunshineOfficial/golib/golog"
	"github.com/sunshineOfficial/golib/gotime"
)

const kafkaSubscribeTimeout = 2 * time.Minute

type Service struct {
	inspectionRepository *inspection.Postgres
	inspectionPublisher  *Publisher
	analyzerClient       *analyzer.Client
	subscriberClient     *subscriber.Client
	fileClient           *file.Client
	taskClient           *task.Client
	brigadeClient        *brigade.Client
	templates            config.Templates
}

func NewService(inspectionRepository *inspection.Postgres, inspectionPublisher *Publisher, analyzerClient *analyzer.Client,
	subscriberClient *subscriber.Client, fileClient *file.Client, taskClient *task.Client, brigadeClient *brigade.Client, templates config.Templates) *Service {
	return &Service{
		inspectionRepository: inspectionRepository,
		inspectionPublisher:  inspectionPublisher,
		analyzerClient:       analyzerClient,
		subscriberClient:     subscriberClient,
		fileClient:           fileClient,
		taskClient:           taskClient,
		brigadeClient:        brigadeClient,
		templates:            templates,
	}
}

func (s *Service) GetByTaskID(ctx goctx.Context, taskID int) (Inspection, error) {
	dbIns, err := s.inspectionRepository.GetByTaskID(ctx, taskID)
	if err != nil {
		return Inspection{}, fmt.Errorf("get inspection by task id: %w", err)
	}

	return MapFromDB(dbIns), nil
}

func (s *Service) AttachPhoto(ctx goctx.Context, log golog.Logger, request AttachPhotoRequest) (Attachment, error) {
	if request.Type != AttachmentTypeDevicePhoto && request.Type != AttachmentTypeSealPhoto {
		return Attachment{}, fmt.Errorf("invalid attachment type: %d", request.Type)
	}

	f, err := request.FileHeader.Open()
	if err != nil {
		return Attachment{}, fmt.Errorf("open file: %w", err)
	}

	defer func() {
		if closeErr := f.Close(); closeErr != nil {
			log.Errorf("close file: %v", closeErr)
		}
	}()

	var fileBuffer bytes.Buffer
	_, err = io.Copy(&fileBuffer, f)
	if err != nil {
		return Attachment{}, fmt.Errorf("copy file: %w", err)
	}

	processedImage, err := s.analyzerClient.ProcessImage(ctx, request.FileHeader.Filename, bytes.NewReader(fileBuffer.Bytes()))
	if err != nil {
		return Attachment{}, fmt.Errorf("process image: %w", err)
	}

	if processedImage.IsBlurred || processedImage.HasError {
		return Attachment{}, ErrBlurredPhoto
	}

	var object subscriber.ObjectExtended
	switch request.Type {
	case AttachmentTypeDevicePhoto:
		object, err = s.subscriberClient.GetObjectExtendedByDevice(ctx, request.DeviceID)
	case AttachmentTypeSealPhoto:
		object, err = s.subscriberClient.GetObjectExtendedBySeal(ctx, request.SealID)
	default:
		return Attachment{}, fmt.Errorf("invalid attachment type: %d", request.Type)
	}

	if err != nil {
		return Attachment{}, fmt.Errorf("get object extended: %w", err)
	}

	number, err := attachmentNumber(request, object)
	if err != nil {
		return Attachment{}, fmt.Errorf("get attachment number: %w", err)
	}

	fileName := fmt.Sprintf(
		"%s - %s №%s от %s%s",
		object.Address,
		attachmentName(request.Type),
		number,
		gotime.MoscowNow().Format("02.01.2006 15.04.05"),
		filepath.Ext(request.FileHeader.Filename),
	)

	uploadedFile, err := s.fileClient.Upload(ctx, fileName, bytes.NewReader(fileBuffer.Bytes()))
	if err != nil {
		return Attachment{}, fmt.Errorf("upload file: %w", err)
	}

	dbAttachment, err := s.inspectionRepository.AddAttachment(ctx, inspection.Attachment{
		InspectionID: request.InspectionID,
		Type:         int(request.Type),
		FileID:       uploadedFile.ID,
	})
	if err != nil {
		return Attachment{}, fmt.Errorf("add attachment: %w", err)
	}

	return MapAttachmentFromDB(dbAttachment), nil
}

func attachmentName(t AttachmentType) string {
	switch t {
	case AttachmentTypeDevicePhoto:
		return "прибор учета"
	case AttachmentTypeSealPhoto:
		return "пломба"
	default:
		return ""
	}
}

func attachmentNumber(request AttachPhotoRequest, object subscriber.ObjectExtended) (string, error) {
	switch request.Type {
	case AttachmentTypeDevicePhoto:
		for _, device := range object.Devices {
			if device.ID == request.DeviceID {
				return device.Number, nil
			}
		}

		return "", fmt.Errorf("device %d not found", request.DeviceID)

	case AttachmentTypeSealPhoto:
		for _, device := range object.Devices {
			for _, seal := range device.Seals {
				if seal.ID == request.SealID {
					return seal.Number, nil
				}
			}
		}

		return "", fmt.Errorf("seal %d not found", request.SealID)

	default:
		return "", fmt.Errorf("invalid attachment type: %d", request.Type)
	}
}

func (s *Service) FinishInspection(ctx goctx.Context, log golog.Logger, request FinishInspectionRequest) (file.File, error) {
	dbIns, err := s.inspectionRepository.GetByID(ctx, request.ID)
	if err != nil {
		return file.File{}, fmt.Errorf("get inspection by id: %w", err)
	}

	tsk, err := s.taskClient.GetTaskByID(ctx, dbIns.TaskID)
	if err != nil {
		return file.File{}, fmt.Errorf("get task by id: %w", err)
	}

	if tsk.BrigadeID == nil {
		return file.File{}, fmt.Errorf("task %d has no brigade", dbIns.TaskID)
	}

	brig, err := s.brigadeClient.GetBrigadeByID(ctx, *tsk.BrigadeID)
	if err != nil {
		return file.File{}, fmt.Errorf("get brigade by id: %w", err)
	}

	object, err := s.subscriberClient.GetObjectExtendedByID(ctx, tsk.ObjectID)
	if err != nil {
		return file.File{}, fmt.Errorf("get object extended: %w", err)
	}

	if len(object.Devices) == 0 {
		return file.File{}, errors.New("no devices found")
	}

	var buf *bytes.Buffer
	switch request.Type {
	case TypeLimitation, TypeResumption:
		buf, err = s.generateUniversalAct(request, brig, object)
	case TypeVerification, TypeUnauthorizedConnection:
		devices, dErr := s.inspectionRepository.GetPreviousDeviceInspections(ctx, object.Devices[0].ID, request.ID)
		if dErr != nil {
			return file.File{}, fmt.Errorf("get device inspections: %w", dErr)
		}

		buf, err = s.generateControlAct(request, brig, object, devices)
	default:
		return file.File{}, fmt.Errorf("invalid inspection type: %d", request.Type)
	}

	if err != nil {
		return file.File{}, fmt.Errorf("generate act: %w", err)
	}

	actType := "о введении ограничения и возобновления"
	if request.Type == TypeVerification || request.Type == TypeUnauthorizedConnection {
		actType = "контроля"
	}

	actName := fmt.Sprintf(
		"Акт %s №%d от %s (%s).docx",
		actType,
		request.ID,
		gotime.MoscowNow().Format("02.01.2006"),
		object.Address,
	)

	uploadedFile, err := s.fileClient.Upload(ctx, actName, buf)
	if err != nil {
		return file.File{}, fmt.Errorf("upload file: %w", err)
	}

	_, err = s.inspectionRepository.AddAttachment(ctx, inspection.Attachment{
		InspectionID: dbIns.ID,
		Type:         int(AttachmentTypeAct),
		FileID:       uploadedFile.ID,
	})
	if err != nil {
		return file.File{}, fmt.Errorf("add attachment: %w", err)
	}

	devices := make([]inspection.InspectedDevice, 0, len(request.InspectedDevices))
	seals := make([]inspection.InspectedSeal, 0, len(request.InspectedDevices))
	for _, device := range request.InspectedDevices {
		devices = append(devices, inspection.InspectedDevice{
			DeviceID:     device.DeviceID,
			InspectionID: dbIns.ID,
			Value:        device.Value,
			Consumption:  device.Consumption,
		})

		for _, seal := range device.InspectedSeals {
			seals = append(seals, inspection.InspectedSeal{
				SealID:       seal.SealID,
				InspectionID: dbIns.ID,
				IsBroken:     seal.IsBroken,
			})
		}
	}

	err = s.inspectionRepository.AddDevices(ctx, devices)
	if err != nil {
		return file.File{}, fmt.Errorf("add devices: %w", err)
	}

	err = s.inspectionRepository.AddSeals(ctx, seals)
	if err != nil {
		return file.File{}, fmt.Errorf("add seals: %w", err)
	}

	var (
		insType    = int(request.Type)
		resolution = int(request.Resolution)
		methodBy   = int(request.MethodBy)
		reasonType = int(request.ReasonType)
	)
	dbIns, err = s.inspectionRepository.FinishInspection(ctx, inspection.Inspection{
		ID:                      dbIns.ID,
		TaskID:                  dbIns.TaskID,
		Type:                    &insType,
		Resolution:              &resolution,
		LimitReason:             request.LimitReason,
		Method:                  &request.Method,
		MethodBy:                &methodBy,
		ReasonType:              &reasonType,
		ReasonDescription:       request.ReasonDescription,
		IsRestrictionChecked:    &request.IsRestrictionChecked,
		IsViolationDetected:     &request.IsViolationDetected,
		IsExpenseAvailable:      &request.IsExpenseAvailable,
		ViolationDescription:    request.ViolationDescription,
		IsUnauthorizedConsumers: &request.IsUnauthorizedConsumers,
		UnauthorizedDescription: request.UnauthorizedDescription,
		UnauthorizedExplanation: request.UnauthorizedExplanation,
		EnergyActionAt:          &request.EnergyActionAt,
	})
	if err != nil {
		return file.File{}, fmt.Errorf("finish inspection: %w", err)
	}

	go s.inspectionPublisher.Publish(ctx, log, EventTypeFinish, MapFromDB(dbIns))

	return uploadedFile, nil
}

func (s *Service) generateUniversalAct(request FinishInspectionRequest, brig brigade.Brigade, object subscriber.ObjectExtended) (*bytes.Buffer, error) {
	now := gotime.MoscowNow()

	isLimitation := "☒"
	isResumption := "☐"
	if request.Resolution == ResolutionResumed {
		isLimitation = "☐"
		isResumption = "☒"
	}

	haveAutomaton := "☐"
	noAutomaton := "☒"
	if object.HaveAutomaton {
		haveAutomaton = "☒"
		noAutomaton = "☐"
	}

	isIncomplete := "☒"
	isOtherReason := "☐"
	otherReason := ""
	if request.LimitReason != nil && len(*request.LimitReason) != 0 {
		isIncomplete = "☐"
		isOtherReason = "☒"
		otherReason = *request.LimitReason
	}

	isEnergyLimited := "☐"
	isEnergyStopped := "☐"
	isEnergyResumed := "☐"
	switch request.Resolution {
	case ResolutionLimited:
		isEnergyLimited = "☒"
	case ResolutionStopped:
		isEnergyStopped = "☒"
	case ResolutionResumed:
		isEnergyResumed = "☒"
	default:
		return nil, fmt.Errorf("invalid resolution: %d", request.Resolution)
	}

	energyDate := request.EnergyActionAt.In(gotime.Moscow)

	isByConsumer := "☐"
	isByInspector := "☒"
	if request.MethodBy == MethodByConsumer {
		isByConsumer = "☒"
		isByInspector = "☐"
	}

	device := object.Devices[0]
	inspectedDevice := request.InspectedDevices[0]

	isInside := "☐"
	isOutside := "☐"
	switch device.PlaceType {
	case subscriber.DevicePlaceFlat:
		isInside = "☒"
	case subscriber.DevicePlaceStairLanding:
		isOutside = "☒"
	default:
		return nil, fmt.Errorf("invalid device place: %d", device.PlaceType)
	}

	seals := make([]string, 0, len(inspectedDevice.InspectedSeals))
	for _, seal := range inspectedDevice.InspectedSeals {
		isBroken := "на месте"
		if seal.IsBroken {
			isBroken = "сорвана"
		}

		seals = append(seals, fmt.Sprintf("№%d - %s", seal.SealID, isBroken))
	}

	isConsumerLimited := "☐"
	isInspectorLimited := "☐"
	isNotIntroduced := "☐"
	switch request.ReasonType {
	case ReasonTypeNotIntroduced:
		isNotIntroduced = "☒"
	case ReasonTypeConsumerLimited:
		isConsumerLimited = "☒"
	case ReasonTypeInspectorLimited:
		isInspectorLimited = "☒"
	default:
		return nil, fmt.Errorf("invalid reason type: %d", request.ReasonType)
	}

	notIntroduced := ""
	if request.ReasonDescription != nil && len(*request.ReasonDescription) != 0 {
		notIntroduced = *request.ReasonDescription
	}

	if len(brig.Inspectors) != 2 {
		return nil, fmt.Errorf("invalid inspectors len: %d", len(brig.Inspectors))
	}

	firstInspector := brig.Inspectors[0]
	secondInspector := brig.Inspectors[1]

	placeholderMap := docx.PlaceholderMap{
		"act_number":               request.ID,
		"is_limitation":            isLimitation,
		"is_resumption":            isResumption,
		"act_day":                  now.Format("02"),
		"act_month":                russianMonth(now.Month()),
		"act_year":                 now.Year(),
		"act_hour":                 now.Format("15"),
		"act_minute":               now.Format("04"),
		"act_place":                object.Address,
		"consumer_fio":             fullFIO(object.Subscriber.Surname, object.Subscriber.Name, object.Subscriber.Patronymic),
		"address":                  object.Address,
		"have_automaton":           haveAutomaton,
		"no_automaton":             noAutomaton,
		"account_number":           object.Subscriber.AccountNumber,
		"is_incomplete_payment":    isIncomplete,
		"is_other_reason":          isOtherReason,
		"other_reason":             otherReason,
		"is_energy_limited":        isEnergyLimited,
		"is_energy_stopped":        isEnergyStopped,
		"is_energy_resumed":        isEnergyResumed,
		"energy_hour":              energyDate.Format("15"),
		"energy_minute":            energyDate.Format("04"),
		"energy_day":               energyDate.Format("02"),
		"energy_month":             russianMonth(energyDate.Month()),
		"energy_year":              energyDate.Year(),
		"is_by_consumer":           isByConsumer,
		"is_by_inspector":          isByInspector,
		"method":                   request.Method,
		"is_inside":                isInside,
		"is_outside":               isOutside,
		"other_place":              device.PlaceDescription,
		"device_type":              device.Type,
		"device_number":            device.Number,
		"device_value":             inspectedDevice.Value,
		"seals":                    strings.Join(seals, ", "),
		"is_consumer_limited":      isConsumerLimited,
		"is_inspector_limited":     isInspectorLimited,
		"is_not_introduced":        isNotIntroduced,
		"is_not_introduced_reason": notIntroduced,
		"inspector1_initials":      shortFIO(firstInspector.Surname, firstInspector.Name, firstInspector.Patronymic),
		"inspector2_initials":      shortFIO(secondInspector.Surname, secondInspector.Name, secondInspector.Patronymic),
	}

	buf, err := writeDocXTemplate(s.templates.Universal, placeholderMap)
	if err != nil {
		return nil, fmt.Errorf("writeDocXTemplate: %w", err)
	}

	return buf, nil
}

func (s *Service) generateControlAct(request FinishInspectionRequest, brig brigade.Brigade, object subscriber.ObjectExtended, devices []inspection.InspectedDevice) (*bytes.Buffer, error) {
	now := gotime.MoscowNow()

	isVerification := "☒"
	isUnauthorizedConnection := "☐"
	if request.Type == TypeUnauthorizedConnection {
		isVerification = "☐"
		isUnauthorizedConnection = "☒"
	}

	haveAutomaton := "☐"
	noAutomaton := "☒"
	if object.HaveAutomaton {
		haveAutomaton = "☒"
		noAutomaton = "☐"
	}

	isIncomplete := "☒"
	isOtherReason := "☐"
	otherReason := ""
	if request.LimitReason != nil && len(*request.LimitReason) != 0 {
		isIncomplete = "☐"
		isOtherReason = "☒"
		otherReason = *request.LimitReason
	}

	isEnergyLimited := "☐"
	isEnergyStopped := "☐"
	switch request.Resolution {
	case ResolutionLimited:
		isEnergyLimited = "☒"
	case ResolutionStopped:
		isEnergyStopped = "☒"
	default:
		return nil, fmt.Errorf("invalid resolution: %d", request.Resolution)
	}

	energyDate := request.EnergyActionAt.In(gotime.Moscow)

	isByConsumer := "☐"
	isByInspector := "☒"
	if request.MethodBy == MethodByConsumer {
		isByConsumer = "☒"
		isByInspector = "☐"
	}

	isChecked := "☐"
	if request.IsRestrictionChecked {
		isChecked = "☒"
	}

	isViolationDetected := "☐"
	isViolationNotDetected := "☒"
	if request.IsViolationDetected {
		isViolationDetected = "☒"
		isViolationNotDetected = "☐"
	}

	isExpenseAvailable := "☐"
	if request.IsExpenseAvailable {
		isExpenseAvailable = "☒"
	}

	isOtherViolation := "☐"
	otherViolation := ""
	if request.ViolationDescription != nil && len(*request.ViolationDescription) != 0 {
		isOtherViolation = "☒"
		otherViolation = *request.ViolationDescription
	}

	isUnauthorizedConsumers := "☐"
	isNotUnauthorizedConsumers := "☒"
	if request.IsUnauthorizedConsumers {
		isUnauthorizedConsumers = "☒"
		isNotUnauthorizedConsumers = "☐"
	}

	unauthorizedDescription := ""
	if request.UnauthorizedDescription != nil && len(*request.UnauthorizedDescription) != 0 {
		unauthorizedDescription = *request.UnauthorizedDescription
	}

	device := object.Devices[0]
	inspectedDevice := request.InspectedDevices[0]

	isInside := "☐"
	isOutside := "☐"
	switch device.PlaceType {
	case subscriber.DevicePlaceFlat:
		isInside = "☒"
	case subscriber.DevicePlaceStairLanding:
		isOutside = "☒"
	default:
		return nil, fmt.Errorf("invalid device place: %d", device.PlaceType)
	}

	oldDevice := inspection.InspectedDevice{
		Value:     decimal.Zero,
		CreatedAt: now,
	}
	if len(devices) != 0 {
		oldDevice = devices[0]
		oldDevice.CreatedAt = oldDevice.CreatedAt.In(gotime.Moscow)
	}

	seals := make([]string, 0, len(inspectedDevice.InspectedSeals))
	for _, seal := range inspectedDevice.InspectedSeals {
		isBroken := "на месте"
		if seal.IsBroken {
			isBroken = "сорвана"
		}

		seals = append(seals, fmt.Sprintf("№%d - %s", seal.SealID, isBroken))
	}

	if len(brig.Inspectors) != 2 {
		return nil, fmt.Errorf("invalid inspectors len: %d", len(brig.Inspectors))
	}

	firstInspector := brig.Inspectors[0]
	secondInspector := brig.Inspectors[1]

	unauthorizedExplanation := ""
	if request.UnauthorizedExplanation != nil && len(*request.UnauthorizedExplanation) != 0 {
		unauthorizedExplanation = *request.UnauthorizedExplanation
	}

	placeholderMap := docx.PlaceholderMap{
		"act_number":                    request.ID,
		"is_verification":               isVerification,
		"is_unauthorized_connection":    isUnauthorizedConnection,
		"act_day":                       now.Format("02"),
		"act_month":                     russianMonth(now.Month()),
		"act_year":                      now.Year(),
		"act_hour":                      now.Format("15"),
		"act_minute":                    now.Format("04"),
		"act_place":                     object.Address,
		"consumer_fio":                  fullFIO(object.Subscriber.Surname, object.Subscriber.Name, object.Subscriber.Patronymic),
		"address":                       object.Address,
		"have_automaton":                haveAutomaton,
		"no_automaton":                  noAutomaton,
		"account_number":                object.Subscriber.AccountNumber,
		"consumer_phone":                object.Subscriber.PhoneNumber,
		"is_incomplete_payment":         isIncomplete,
		"is_other_reason":               isOtherReason,
		"other_reason":                  otherReason,
		"is_energy_limited":             isEnergyLimited,
		"is_energy_stopped":             isEnergyStopped,
		"energy_hour":                   energyDate.Format("15"),
		"energy_minute":                 energyDate.Format("04"),
		"energy_day":                    energyDate.Format("02"),
		"energy_month":                  russianMonth(energyDate.Month()),
		"energy_year":                   energyDate.Year(),
		"is_by_consumer":                isByConsumer,
		"is_by_inspector":               isByInspector,
		"is_checked":                    isChecked,
		"check_hour":                    now.Format("15"),
		"check_minute":                  now.Format("04"),
		"check_day":                     now.Format("02"),
		"check_month":                   russianMonth(now.Month()),
		"check_year":                    now.Year(),
		"is_violation_not_detected":     isViolationNotDetected,
		"is_violation_detected":         isViolationDetected,
		"is_expense_available":          isExpenseAvailable,
		"is_other_violation":            isOtherViolation,
		"other_violation":               otherViolation,
		"is_not_unauthorized_consumers": isNotUnauthorizedConsumers,
		"is_unauthorized_consumers":     isUnauthorizedConsumers,
		"unauthorized_description":      unauthorizedDescription,
		"is_inside":                     isInside,
		"is_outside":                    isOutside,
		"other_place":                   device.PlaceDescription,
		"device_type":                   device.Type,
		"device_number":                 device.Number,
		"device_value":                  inspectedDevice.Value,
		"old_value_day":                 oldDevice.CreatedAt.Format("02"),
		"old_value_month":               oldDevice.CreatedAt.Format("01"),
		"old_value_year":                oldDevice.CreatedAt.Year(),
		"old_device_value":              oldDevice.Value,
		"device_consumption":            inspectedDevice.Consumption,
		"seals":                         strings.Join(seals, ", "),
		"unauthorized_explanation":      unauthorizedExplanation,
		"inspector1_initials":           shortFIO(firstInspector.Surname, firstInspector.Name, firstInspector.Patronymic),
		"inspector2_initials":           shortFIO(secondInspector.Surname, secondInspector.Name, secondInspector.Patronymic),
	}

	buf, err := writeDocXTemplate(s.templates.Control, placeholderMap)
	if err != nil {
		return nil, fmt.Errorf("writeDocXTemplate: %w", err)
	}

	return buf, nil
}

func russianMonth(month time.Month) string {
	switch month {
	case time.January:
		return "января"
	case time.February:
		return "февраля"
	case time.March:
		return "марта"
	case time.April:
		return "апреля"
	case time.May:
		return "мая"
	case time.June:
		return "июня"
	case time.July:
		return "июля"
	case time.August:
		return "августа"
	case time.September:
		return "сентября"
	case time.October:
		return "октября"
	case time.November:
		return "ноября"
	case time.December:
		return "декабря"
	}

	return ""
}

func fullFIO(surname, name, patronymic string) string {
	result := fmt.Sprintf("%s %s", surname, name)

	if len(patronymic) > 0 {
		result = fmt.Sprintf("%s %s", result, patronymic)
	}

	return result
}

func shortFIO(surname, name, patronymic string) string {
	result := fmt.Sprintf("%s %s.", surname, string([]rune(name)[0]))

	if len(patronymic) > 0 {
		result = fmt.Sprintf("%s%s.", result, string([]rune(patronymic)[0]))
	}

	return result
}

func writeDocXTemplate(path string, placeholderMap docx.PlaceholderMap) (*bytes.Buffer, error) {
	doc, err := docx.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open template: %w", err)
	}

	defer doc.Close()

	if err = doc.ReplaceAll(placeholderMap); err != nil {
		return nil, fmt.Errorf("replace all: %w", err)
	}

	var buf bytes.Buffer
	if err = doc.Write(&buf); err != nil {
		return nil, fmt.Errorf("write template: %w", err)
	}

	return &buf, nil
}

func (s *Service) SubscriberOnTaskEvent(mainCtx context.Context, log golog.Logger) gokafka.Subscriber {
	return func(message gokafka.Message, err error) {
		ctx, cancel := context.WithTimeout(mainCtx, kafkaSubscribeTimeout)
		defer cancel()

		if err != nil {
			log.Errorf("got error on task event: %v", err)
			return
		}

		var event task.Event
		err = json.Unmarshal(message.Value, &event)
		if err != nil {
			log.Errorf("failed to unmarshal task event: %v", err)
			return
		}

		switch event.Type {
		case task.EventTypeAdd:
			err = s.handleAddedTask(ctx, event.Task)
		case task.EventTypeStart:
			err = s.handleStartedTask(ctx, log, event.Task)
		case task.EventTypeFinish:
			err = s.handleFinishedTask(ctx, event.Task)
		default:
			err = fmt.Errorf("unknown event type: %v", event.Type)
		}

		if err != nil {
			log.Errorf("failed to handle task event (type = %d): %v", event.Type, err)
			return
		}
	}
}

func (s *Service) handleAddedTask(ctx context.Context, t task.Task) error {
	return nil
}

func (s *Service) handleStartedTask(ctx context.Context, log golog.Logger, t task.Task) error {
	if t.Status != task.StatusInWork {
		return fmt.Errorf("invalid task status: %v", t.Status)
	}

	dbIns, err := s.inspectionRepository.StartInspection(ctx, t.ID)
	if err != nil {
		return fmt.Errorf("start inspection: %v", err)
	}

	go s.inspectionPublisher.Publish(goctx.Wrap(ctx), log, EventTypeStart, MapFromDB(dbIns))

	return nil
}

func (s *Service) handleFinishedTask(ctx context.Context, t task.Task) error {
	return nil
}
