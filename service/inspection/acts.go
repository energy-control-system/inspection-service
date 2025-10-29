package inspection

import (
	"bytes"
	"fmt"
	"inspection-service/cluster/brigade"
	"inspection-service/cluster/subscriber"
	"strings"
	"time"

	"github.com/lukasjarosch/go-docx"
	"github.com/shopspring/decimal"
	"github.com/sunshineOfficial/golib/gotime"
)

func (s *Service) generateUniversalAct(request FinishInspectionRequest, brig brigade.Brigade, contract subscriber.Contract) (*bytes.Buffer, error) {
	now := gotime.MoscowNow()

	isLimitation := "☒"
	isResumption := "☐"
	if request.Resolution == ResolutionResumed {
		isLimitation = "☐"
		isResumption = "☒"
	}

	haveAutomaton := "☐"
	noAutomaton := "☒"
	if contract.Object.HaveAutomaton {
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

	device := contract.Object.Devices[0]
	inspectedDevice := request.InspectedDevices[0]

	isInside := "☐"
	isOutside := "☐"
	switch device.PlaceType {
	case subscriber.DevicePlaceFlat:
		isInside = "☒"
	case subscriber.DevicePlaceStairLanding:
		isOutside = "☒"
	case subscriber.DevicePlaceOther:
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
		"act_place":                contract.Object.Address,
		"consumer_fio":             fullFIO(contract.Subscriber.Surname, contract.Subscriber.Name, contract.Subscriber.Patronymic),
		"address":                  contract.Object.Address,
		"have_automaton":           haveAutomaton,
		"no_automaton":             noAutomaton,
		"account_number":           contract.Subscriber.AccountNumber,
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

func (s *Service) generateControlAct(request FinishInspectionRequest, brig brigade.Brigade, contract subscriber.Contract, devices []InspectedDevice) (*bytes.Buffer, error) {
	now := gotime.MoscowNow()

	isVerification := "☒"
	isUnauthorizedConnection := "☐"
	if request.Type == TypeUnauthorizedConnection {
		isVerification = "☐"
		isUnauthorizedConnection = "☒"
	}

	haveAutomaton := "☐"
	noAutomaton := "☒"
	if contract.Object.HaveAutomaton {
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

	device := contract.Object.Devices[0]
	inspectedDevice := request.InspectedDevices[0]

	isInside := "☐"
	isOutside := "☐"
	switch device.PlaceType {
	case subscriber.DevicePlaceFlat:
		isInside = "☒"
	case subscriber.DevicePlaceStairLanding:
		isOutside = "☒"
	case subscriber.DevicePlaceOther:
	default:
		return nil, fmt.Errorf("invalid device place: %d", device.PlaceType)
	}

	oldDevice := InspectedDevice{
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
		"act_place":                     contract.Object.Address,
		"consumer_fio":                  fullFIO(contract.Subscriber.Surname, contract.Subscriber.Name, contract.Subscriber.Patronymic),
		"address":                       contract.Object.Address,
		"have_automaton":                haveAutomaton,
		"no_automaton":                  noAutomaton,
		"account_number":                contract.Subscriber.AccountNumber,
		"consumer_phone":                contract.Subscriber.PhoneNumber,
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
