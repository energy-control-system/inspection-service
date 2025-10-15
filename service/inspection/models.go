package inspection

import (
	"mime/multipart"
	"time"

	"github.com/shopspring/decimal"
)

type Status int

const (
	StatusUnknown Status = iota
	StatusInWork
	StatusDone
)

type Type int

const (
	TypeUnknown Type = iota
	TypeLimitation
	TypeResumption
	TypeVerification
	TypeUnauthorizedConnection
)

type Resolution int

const (
	ResolutionUnknown Resolution = iota
	ResolutionLimited
	ResolutionStopped
	ResolutionResumed
)

type MethodBy int

const (
	MethodByUnknown MethodBy = iota
	MethodByConsumer
	MethodByInspector
)

type ReasonType int

const (
	ReasonTypeUnknown ReasonType = iota
	ReasonTypeNotIntroduced
	ReasonTypeConsumerLimited
	ReasonTypeInspectorLimited
	ReasonTypeResumed
)

type Inspection struct {
	ID                      int         `json:"ID"`
	TaskID                  int         `json:"TaskID"`
	Status                  Status      `json:"Status"`
	Type                    *Type       `json:"Type,omitempty"`
	Resolution              *Resolution `json:"Resolution,omitempty"`
	LimitReason             *string     `json:"LimitReason,omitempty"`
	Method                  *string     `json:"Method,omitempty"`
	MethodBy                *MethodBy   `json:"MethodBy,omitempty"`
	ReasonType              *ReasonType `json:"ReasonType,omitempty"`
	ReasonDescription       *string     `json:"ReasonDescription,omitempty"`
	IsRestrictionChecked    *bool       `json:"IsRestrictionChecked,omitempty"`
	IsViolationDetected     *bool       `json:"IsViolationDetected,omitempty"`
	IsExpenseAvailable      *bool       `json:"IsExpenseAvailable,omitempty"`
	ViolationDescription    *string     `json:"ViolationDescription,omitempty"`
	IsUnauthorizedConsumers *bool       `json:"IsUnauthorizedConsumers,omitempty"`
	UnauthorizedDescription *string     `json:"UnauthorizedDescription,omitempty"`
	UnauthorizedExplanation *string     `json:"UnauthorizedExplanation,omitempty"`
	InspectAt               *time.Time  `json:"InspectAt,omitempty"`
	EnergyActionAt          *time.Time  `json:"EnergyActionAt,omitempty"`
	CreatedAt               time.Time   `json:"CreatedAt"`
	UpdatedAt               time.Time   `json:"UpdatedAt"`
}

type AttachmentType int

const (
	AttachmentTypeUnknown AttachmentType = iota
	AttachmentTypeDevicePhoto
	AttachmentTypeSealPhoto
	AttachmentTypeAct
)

type Attachment struct {
	ID           int            `json:"ID"`
	InspectionID int            `json:"InspectionID"`
	Type         AttachmentType `json:"Type"`
	FileID       int            `json:"FileID"`
	CreatedAt    time.Time      `json:"CreatedAt"`
}

type InspectedDevice struct {
	ID           int             `json:"id"`
	DeviceID     int             `json:"device_id"`
	InspectionID int             `json:"inspection_id"`
	Value        decimal.Decimal `json:"value"`
	Consumption  decimal.Decimal `json:"consumption"`
	CreatedAt    time.Time       `json:"created_at"`
}

type AttachPhotoRequest struct {
	InspectionID int
	Type         AttachmentType
	DeviceID     int
	SealID       int
	FileHeader   *multipart.FileHeader
}

type FinishInspectionRequest struct {
	ID                      int                      `json:"ID"`
	Type                    Type                     `json:"Type"`
	Resolution              Resolution               `json:"Resolution"`
	LimitReason             *string                  `json:"LimitReason"`
	Method                  string                   `json:"Method"`
	MethodBy                MethodBy                 `json:"MethodBy"`
	ReasonType              ReasonType               `json:"ReasonType"`
	ReasonDescription       *string                  `json:"ReasonDescription"`
	IsRestrictionChecked    bool                     `json:"IsRestrictionChecked"`
	IsViolationDetected     bool                     `json:"IsViolationDetected"`
	IsExpenseAvailable      bool                     `json:"IsExpenseAvailable"`
	ViolationDescription    *string                  `json:"ViolationDescription"`
	IsUnauthorizedConsumers bool                     `json:"IsUnauthorizedConsumers"`
	UnauthorizedDescription *string                  `json:"UnauthorizedDescription"`
	UnauthorizedExplanation *string                  `json:"UnauthorizedExplanation"`
	EnergyActionAt          time.Time                `json:"EnergyActionAt"`
	InspectedDevices        []InspectedDeviceRequest `json:"InspectedDevices"`
}

type InspectedDeviceRequest struct {
	DeviceID       int                    `json:"DeviceID"`
	Value          decimal.Decimal        `json:"Value"`
	Consumption    decimal.Decimal        `json:"Consumption"`
	InspectedSeals []InspectedSealRequest `json:"InspectedSeals"`
}

type InspectedSealRequest struct {
	SealID   int  `json:"SealID"`
	IsBroken bool `json:"IsBroken"`
}
