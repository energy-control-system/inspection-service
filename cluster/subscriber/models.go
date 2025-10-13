package subscriber

import "time"

type Status int

const (
	StatusUnknown Status = iota
	StatusActive
	StatusViolator
	StatusArchived
)

type Subscriber struct {
	ID            int       `json:"ID"`
	AccountNumber string    `json:"AccountNumber"`
	Surname       string    `json:"Surname"`
	Name          string    `json:"Name"`
	Patronymic    string    `json:"Patronymic"`
	PhoneNumber   string    `json:"PhoneNumber"`
	Email         string    `json:"Email"`
	INN           string    `json:"INN"`
	BirthDate     string    `json:"BirthDate"`
	Status        Status    `json:"Status"`
	CreatedAt     time.Time `json:"CreatedAt"`
	UpdatedAt     time.Time `json:"UpdatedAt"`
}

type ObjectExtended struct {
	ID            int              `json:"ID"`
	Address       string           `json:"Address"`
	HaveAutomaton bool             `json:"HaveAutomaton"`
	CreatedAt     time.Time        `json:"CreatedAt"`
	UpdatedAt     time.Time        `json:"UpdatedAt"`
	Subscriber    Subscriber       `json:"Subscriber"`
	Devices       []DeviceExtended `json:"Devices"`
}

type DevicePlaceType int

const (
	DevicePlaceUnknown DevicePlaceType = iota
	DevicePlaceOther
	DevicePlaceFlat
	DevicePlaceStairLanding
)

type DeviceExtended struct {
	ID               int             `json:"ID"`
	ObjectID         int             `json:"ObjectID"`
	Type             string          `json:"Type"`
	Number           string          `json:"Number"`
	PlaceType        DevicePlaceType `json:"PlaceType"`
	PlaceDescription string          `json:"PlaceDescription"`
	CreatedAt        time.Time       `json:"CreatedAt"`
	UpdatedAt        time.Time       `json:"UpdatedAt"`
	Seals            []Seal          `json:"Seals"`
}

type Seal struct {
	ID        int       `json:"ID"`
	DeviceID  int       `json:"DeviceID"`
	Number    string    `json:"Number"`
	Place     string    `json:"Place"`
	CreatedAt time.Time `json:"CreatedAt"`
	UpdatedAt time.Time `json:"UpdatedAt"`
}
