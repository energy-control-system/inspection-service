package task

import "time"

type EventType int

const (
	EventTypeUnknown EventType = iota
	EventTypeAdd
	EventTypeStart
	EventTypeFinish
)

type Event struct {
	Type   EventType `json:"Type"`
	Date   time.Time `json:"Date"`
	UserID int       `json:"UserID"`
	Task   Task      `json:"Task"`
}
