package inspection

import (
	"time"

	"github.com/shopspring/decimal"
)

type Status struct {
	ID   int    `db:"id"`
	Name string `db:"name"`
}

type Type struct {
	ID   int    `db:"id"`
	Name string `db:"name"`
}

type Resolution struct {
	ID   int    `db:"id"`
	Name string `db:"name"`
}

type MethodBy struct {
	ID   int    `db:"id"`
	Name string `db:"name"`
}

type ReasonType struct {
	ID   int    `db:"id"`
	Name string `db:"name"`
}

type Inspection struct {
	ID                      int        `db:"id"`
	TaskID                  int        `db:"task_id"`
	Status                  int        `db:"status"`
	Type                    *int       `db:"type"`
	Resolution              *int       `db:"resolution"`
	LimitReason             *string    `db:"limit_reason"`
	Method                  *string    `db:"method"`
	MethodBy                *int       `db:"method_by"`
	ReasonType              *int       `db:"reason_type"`
	ReasonDescription       *string    `db:"reason_description"`
	IsRestrictionChecked    *bool      `db:"is_restriction_checked"`
	IsViolationDetected     *bool      `db:"is_violation_detected"`
	IsExpenseAvailable      *bool      `db:"is_expense_available"`
	ViolationDescription    *string    `db:"violation_description"`
	IsUnauthorizedConsumers *bool      `db:"is_unauthorized_consumers"`
	UnauthorizedDescription *string    `db:"unauthorized_description"`
	UnauthorizedExplanation *string    `db:"unauthorized_explanation"`
	InspectAt               *time.Time `db:"inspect_at"`
	EnergyActionAt          *time.Time `db:"energy_action_at"`
	CreatedAt               time.Time  `db:"created_at"`
	UpdatedAt               time.Time  `db:"updated_at"`
}

type InspectedDevice struct {
	ID           int             `db:"id"`
	DeviceID     int             `db:"device_id"`
	InspectionID int             `db:"inspection_id"`
	Value        decimal.Decimal `db:"value"`
	Consumption  decimal.Decimal `db:"consumption"`
	CreatedAt    time.Time       `db:"created_at"`
}

type InspectedSeal struct {
	ID           int       `db:"id"`
	SealID       int       `db:"seal_id"`
	InspectionID int       `db:"inspection_id"`
	IsBroken     bool      `db:"is_broken"`
	CreatedAt    time.Time `db:"created_at"`
}

type AttachmentType struct {
	ID   int    `db:"id"`
	Name string `db:"name"`
}

type Attachment struct {
	ID           int       `db:"id"`
	InspectionID int       `db:"inspection_id"`
	Type         int       `db:"type"`
	FileID       int       `db:"file_id"`
	CreatedAt    time.Time `db:"created_at"`
}
