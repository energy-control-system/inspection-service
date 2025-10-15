package inspection

import (
	"inspection-service/service/inspection"
)

func MapFromDB(i Inspection) inspection.Inspection {
	return inspection.Inspection{
		ID:                      i.ID,
		TaskID:                  i.TaskID,
		Status:                  inspection.Status(i.Status),
		Type:                    (*inspection.Type)(i.Type),
		Resolution:              (*inspection.Resolution)(i.Resolution),
		LimitReason:             i.LimitReason,
		Method:                  i.Method,
		MethodBy:                (*inspection.MethodBy)(i.MethodBy),
		ReasonType:              (*inspection.ReasonType)(i.ReasonType),
		ReasonDescription:       i.ReasonDescription,
		IsRestrictionChecked:    i.IsRestrictionChecked,
		IsViolationDetected:     i.IsViolationDetected,
		IsExpenseAvailable:      i.IsExpenseAvailable,
		ViolationDescription:    i.ViolationDescription,
		IsUnauthorizedConsumers: i.IsUnauthorizedConsumers,
		UnauthorizedDescription: i.UnauthorizedDescription,
		UnauthorizedExplanation: i.UnauthorizedExplanation,
		InspectAt:               i.InspectAt,
		EnergyActionAt:          i.EnergyActionAt,
		CreatedAt:               i.CreatedAt,
		UpdatedAt:               i.UpdatedAt,
	}
}

func MapFinishInspectionRequestToDB(r inspection.FinishInspectionRequest) FinishInspectionRequest {
	return FinishInspectionRequest{
		ID:                      r.ID,
		Type:                    int(r.Type),
		Resolution:              int(r.Resolution),
		LimitReason:             r.LimitReason,
		Method:                  r.Method,
		MethodBy:                int(r.MethodBy),
		ReasonType:              int(r.ReasonType),
		ReasonDescription:       r.ReasonDescription,
		IsRestrictionChecked:    r.IsRestrictionChecked,
		IsViolationDetected:     r.IsViolationDetected,
		IsExpenseAvailable:      r.IsExpenseAvailable,
		ViolationDescription:    r.ViolationDescription,
		IsUnauthorizedConsumers: r.IsUnauthorizedConsumers,
		UnauthorizedDescription: r.UnauthorizedDescription,
		UnauthorizedExplanation: r.UnauthorizedExplanation,
		EnergyActionAt:          r.EnergyActionAt,
	}
}

func MapAttachmentFromDB(a Attachment) inspection.Attachment {
	return inspection.Attachment{
		ID:           a.ID,
		InspectionID: a.InspectionID,
		Type:         inspection.AttachmentType(a.Type),
		FileID:       a.FileID,
		CreatedAt:    a.CreatedAt,
	}
}

func MapInspectedDeviceFromDB(d InspectedDevice) inspection.InspectedDevice {
	return inspection.InspectedDevice{
		ID:           d.ID,
		DeviceID:     d.DeviceID,
		InspectionID: d.InspectionID,
		Value:        d.Value,
		Consumption:  d.Consumption,
		CreatedAt:    d.CreatedAt,
	}
}

func MapInspectedDevicesSliceFromDB(devices []InspectedDevice) []inspection.InspectedDevice {
	result := make([]inspection.InspectedDevice, 0, len(devices))
	for _, device := range devices {
		result = append(result, MapInspectedDeviceFromDB(device))
	}

	return result
}

func MapInspectedSealRequestToDB(r inspection.InspectedSealRequest, inspectionID int) InspectedSeal {
	return InspectedSeal{
		SealID:       r.SealID,
		InspectionID: inspectionID,
		IsBroken:     r.IsBroken,
	}
}

func MapInspectedSealRequestsSliceToDB(requests []inspection.InspectedSealRequest, inspectionID int) []InspectedSeal {
	result := make([]InspectedSeal, 0, len(requests))
	for _, request := range requests {
		result = append(result, MapInspectedSealRequestToDB(request, inspectionID))
	}

	return result
}

func MapInspectedDeviceRequestToDB(r inspection.InspectedDeviceRequest, inspectionID int) (InspectedDevice, []InspectedSeal) {
	return InspectedDevice{
		DeviceID:     r.DeviceID,
		InspectionID: inspectionID,
		Value:        r.Value,
		Consumption:  r.Consumption,
	}, MapInspectedSealRequestsSliceToDB(r.InspectedSeals, inspectionID)
}

func MapInspectedDeviceRequestsSliceToDB(requests []inspection.InspectedDeviceRequest, inspectionID int) ([]InspectedDevice, []InspectedSeal) {
	var (
		devices = make([]InspectedDevice, 0, len(requests))
		seals   = make([]InspectedSeal, 0, len(requests))
	)

	for _, request := range requests {
		device, sls := MapInspectedDeviceRequestToDB(request, inspectionID)
		devices = append(devices, device)
		seals = append(seals, sls...)
	}

	return devices, seals
}
