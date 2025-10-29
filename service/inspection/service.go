package inspection

import (
	"bytes"
	"errors"
	"fmt"
	"inspection-service/cluster/file"
	"inspection-service/cluster/subscriber"
	"inspection-service/config"
	"io"
	"path/filepath"
	"time"

	"github.com/sunshineOfficial/golib/goctx"
	"github.com/sunshineOfficial/golib/golog"
	"github.com/sunshineOfficial/golib/gotime"
)

const kafkaSubscribeTimeout = 2 * time.Minute

type Service struct {
	repository        Repository
	publisher         *Publisher
	analyzerService   AnalyzerService
	subscriberService SubscriberService
	fileService       FileService
	taskService       TaskService
	brigadeService    BrigadeService
	templates         config.Templates
}

func NewService(repository Repository, publisher *Publisher, analyzerService AnalyzerService, subscriberService SubscriberService, fileService FileService,
	taskService TaskService, brigadeService BrigadeService, templates config.Templates) *Service {
	return &Service{
		repository:        repository,
		publisher:         publisher,
		analyzerService:   analyzerService,
		subscriberService: subscriberService,
		fileService:       fileService,
		taskService:       taskService,
		brigadeService:    brigadeService,
		templates:         templates,
	}
}

func (s *Service) GetAll(ctx goctx.Context) ([]Inspection, error) {
	inspections, err := s.repository.GetAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("get all inspections: %w", err)
	}

	return inspections, nil
}

func (s *Service) GetByTaskID(ctx goctx.Context, taskID int) (Inspection, error) {
	ins, err := s.repository.GetByTaskID(ctx, taskID)
	if err != nil {
		return Inspection{}, fmt.Errorf("get inspection by task id: %w", err)
	}

	return ins, nil
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

	processedImage, err := s.analyzerService.ProcessImage(ctx, request.FileHeader.Filename, bytes.NewReader(fileBuffer.Bytes()))
	if err != nil {
		return Attachment{}, fmt.Errorf("process image: %w", err)
	}

	if processedImage.IsBlurred || processedImage.HasError {
		return Attachment{}, ErrBlurredPhoto
	}

	var object subscriber.Object
	switch request.Type {
	case AttachmentTypeDevicePhoto:
		object, err = s.subscriberService.GetObjectByDeviceID(ctx, request.DeviceID)
	case AttachmentTypeSealPhoto:
		object, err = s.subscriberService.GetObjectBySealID(ctx, request.SealID)
	default:
		return Attachment{}, fmt.Errorf("invalid attachment type: %d", request.Type)
	}

	if err != nil {
		return Attachment{}, fmt.Errorf("get object: %w", err)
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

	uploadedFile, err := s.fileService.Upload(ctx, fileName, bytes.NewReader(fileBuffer.Bytes()))
	if err != nil {
		return Attachment{}, fmt.Errorf("upload file: %w", err)
	}

	attachment, err := s.repository.AddAttachment(ctx, request.InspectionID, uploadedFile.ID, request.Type)
	if err != nil {
		return Attachment{}, fmt.Errorf("add attachment: %w", err)
	}

	return attachment, nil
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

func attachmentNumber(request AttachPhotoRequest, object subscriber.Object) (string, error) {
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
	ins, err := s.repository.GetByID(ctx, request.ID)
	if err != nil {
		return file.File{}, fmt.Errorf("get inspection by id: %w", err)
	}

	tsk, err := s.taskService.GetTaskByID(ctx, ins.TaskID)
	if err != nil {
		return file.File{}, fmt.Errorf("get task by id: %w", err)
	}

	if tsk.BrigadeID == nil {
		return file.File{}, fmt.Errorf("task %d has no brigade", ins.TaskID)
	}

	brig, err := s.brigadeService.GetBrigadeByID(ctx, *tsk.BrigadeID)
	if err != nil {
		return file.File{}, fmt.Errorf("get brigade by id: %w", err)
	}

	contract, err := s.subscriberService.GetLastContractByObjectID(ctx, tsk.ObjectID)
	if err != nil {
		return file.File{}, fmt.Errorf("get contract by object id: %w", err)
	}

	if len(contract.Object.Devices) == 0 {
		return file.File{}, errors.New("no devices found")
	}

	var buf *bytes.Buffer
	switch request.Type {
	case TypeLimitation, TypeResumption:
		buf, err = s.generateUniversalAct(request, brig, contract)
	case TypeVerification, TypeUnauthorizedConnection:
		devices, dErr := s.repository.GetPreviousDeviceInspections(ctx, contract.Object.Devices[0].ID, request.ID)
		if dErr != nil {
			return file.File{}, fmt.Errorf("get device inspections: %w", dErr)
		}

		buf, err = s.generateControlAct(request, brig, contract, devices)
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
		gotime.MoscowNow().Format(gotime.DateOnlyNet),
		contract.Object.Address,
	)

	uploadedFile, err := s.fileService.Upload(ctx, actName, buf)
	if err != nil {
		return file.File{}, fmt.Errorf("upload file: %w", err)
	}

	_, err = s.repository.AddAttachment(ctx, ins.ID, uploadedFile.ID, AttachmentTypeAct)
	if err != nil {
		return file.File{}, fmt.Errorf("add attachment: %w", err)
	}

	err = s.repository.AddInspectedDevices(ctx, ins.ID, request.InspectedDevices)
	if err != nil {
		return file.File{}, fmt.Errorf("add inspected devices: %w", err)
	}

	ins, err = s.repository.FinishInspection(ctx, request)
	if err != nil {
		return file.File{}, fmt.Errorf("finish inspection: %w", err)
	}

	go s.publisher.Publish(ctx, log, EventTypeFinish, ins)

	return uploadedFile, nil
}
