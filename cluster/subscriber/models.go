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
	BirthDate     time.Time `json:"BirthDate"`
	Status        Status    `json:"Status"`
	Passport      Passport  `json:"Passport"`
	CreatedAt     time.Time `json:"CreatedAt"`
	UpdatedAt     time.Time `json:"UpdatedAt"`
}

type Passport struct {
	ID        int    `json:"ID"`
	Series    string `json:"Series"`
	Number    string `json:"Number"`
	IssuedBy  string `json:"IssuedBy"`
	IssueDate string `json:"IssueDate"`
}

type Object struct {
	ID            int       `json:"ID"`
	Address       string    `json:"Address"`
	HaveAutomaton bool      `json:"HaveAutomaton"`
	CreatedAt     time.Time `json:"CreatedAt"`
	UpdatedAt     time.Time `json:"UpdatedAt"`
	Devices       []Device  `json:"Devices"`
}

type DevicePlaceType int

const (
	DevicePlaceUnknown DevicePlaceType = iota
	DevicePlaceOther
	DevicePlaceFlat
	DevicePlaceStairLanding
)

type Device struct {
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

type Contract struct {
	ID         int        `json:"ID"`
	Number     string     `json:"Number"`
	Subscriber Subscriber `json:"Subscriber"`
	Object     Object     `json:"Object"`
	SignDate   string     `json:"SignDate"`
	CreatedAt  time.Time  `json:"CreatedAt"`
	UpdatedAt  time.Time  `json:"UpdatedAt"`
}
