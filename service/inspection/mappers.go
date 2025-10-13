package inspection

import (
	"inspection-service/database/inspection"
)

func MapFromDB(i inspection.Inspection) Inspection {
	return Inspection{
		ID:                      i.ID,
		TaskID:                  i.TaskID,
		Status:                  Status(i.Status),
		Type:                    (*Type)(i.Type),
		Resolution:              (*Resolution)(i.Resolution),
		LimitReason:             i.LimitReason,
		Method:                  i.Method,
		MethodBy:                (*MethodBy)(i.MethodBy),
		ReasonType:              (*ReasonType)(i.ReasonType),
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

func MapAttachmentFromDB(a inspection.Attachment) Attachment {
	return Attachment{
		ID:           a.ID,
		InspectionID: a.InspectionID,
		Type:         AttachmentType(a.Type),
		FileID:       a.FileID,
		CreatedAt:    a.CreatedAt,
	}
}
