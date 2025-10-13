package brigade

import "time"

type Inspector struct {
	ID          int       `json:"ID"`
	Surname     string    `json:"Surname"`
	Name        string    `json:"Name"`
	Patronymic  string    `json:"Patronymic"`
	PhoneNumber string    `json:"PhoneNumber"`
	Email       string    `json:"Email"`
	AssignedAt  time.Time `json:"AssignedAt"`
	CreatedAt   time.Time `json:"CreatedAt"`
	UpdatedAt   time.Time `json:"UpdatedAt"`
}

type Status int

const (
	StatusUnknown Status = iota
	StatusIdle
	StatusOnTask
	StatusArchived
)

type Brigade struct {
	ID         int         `json:"ID"`
	Status     Status      `json:"Status"`
	Inspectors []Inspector `json:"Inspectors"`
	CreatedAt  time.Time   `json:"CreatedAt"`
	UpdatedAt  time.Time   `json:"UpdatedAt"`
}
