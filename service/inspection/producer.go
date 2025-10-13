package inspection

import (
	"context"
	"time"

	"github.com/sunshineOfficial/golib/goctx"
	"github.com/sunshineOfficial/golib/gokafka"
	"github.com/sunshineOfficial/golib/golog"
)

const kafkaProduceTimeout = 1 * time.Minute

type EventType int

const (
	EventTypeUnknown EventType = iota
	EventTypeStart
	EventTypeFinish
)

type Event struct {
	Type       EventType  `json:"Type"`
	Date       time.Time  `json:"Date"`
	UserID     int        `json:"UserID"`
	Inspection Inspection `json:"Inspection"`
}

type Publisher struct {
	baseContext context.Context
	producer    gokafka.Producer
}

func NewPublisher(baseContext context.Context, producer gokafka.Producer) *Publisher {
	return &Publisher{
		baseContext: baseContext,
		producer:    producer,
	}
}

func (p *Publisher) Publish(ctx goctx.Context, log golog.Logger, eventType EventType, ins Inspection) {
	event := Event{
		Type:       eventType,
		Date:       time.Now(),
		UserID:     ctx.Authorize.UserId,
		Inspection: ins,
	}

	message, err := gokafka.NewJSONMessage("", event)
	if err != nil {
		log.Errorf("failed to create json message for inspection event: %v", err)
		return
	}

	produceCtx, produceCtxCancel := context.WithTimeout(p.baseContext, kafkaProduceTimeout)
	defer produceCtxCancel()

	err = p.producer.Produce(produceCtx, message)
	if err != nil {
		log.Errorf("failed to produce message for inspection event: %v", err)
		return
	}
}
