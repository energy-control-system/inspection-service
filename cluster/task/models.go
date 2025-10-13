package task

import "time"

type Status int

const (
	StatusUnknown Status = iota
	StatusPlanned
	StatusInWork
	StatusDone
)

type Task struct {
	ID          int        `json:"ID"`
	BrigadeID   *int       `json:"BrigadeID"`
	ObjectID    int        `json:"ObjectID"`
	PlanVisitAt *time.Time `json:"PlanVisitAt"`
	Status      Status     `json:"Status"`
	Comment     *string    `json:"Comment"`
	StartedAt   *time.Time `json:"StartedAt"`
	FinishedAt  *time.Time `json:"FinishedAt"`
	CreatedAt   time.Time  `json:"CreatedAt"`
	UpdatedAt   time.Time  `json:"UpdatedAt"`
}
