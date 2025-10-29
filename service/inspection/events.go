package inspection

import (
	"context"
	"encoding/json"
	"fmt"
	"inspection-service/cluster/task"

	"github.com/sunshineOfficial/golib/goctx"
	"github.com/sunshineOfficial/golib/gokafka"
	"github.com/sunshineOfficial/golib/golog"
)

func (s *Service) SubscriberOnTaskEvent(mainCtx context.Context, log golog.Logger) gokafka.Subscriber {
	return func(message gokafka.Message, err error) {
		ctx, cancel := context.WithTimeout(mainCtx, kafkaSubscribeTimeout)
		defer cancel()

		if err != nil {
			log.Errorf("got error on task event: %v", err)
			return
		}

		var event task.Event
		err = json.Unmarshal(message.Value, &event)
		if err != nil {
			log.Errorf("failed to unmarshal task event: %v", err)
			return
		}

		switch event.Type {
		case task.EventTypeAdd:
			err = s.handleAddedTask(ctx, event.Task)
		case task.EventTypeStart:
			err = s.handleStartedTask(ctx, log, event.Task)
		case task.EventTypeFinish:
			err = s.handleFinishedTask(ctx, event.Task)
		default:
			err = fmt.Errorf("unknown event type: %v", event.Type)
		}

		if err != nil {
			log.Errorf("failed to handle task event (type = %d): %v", event.Type, err)
			return
		}
	}
}

func (s *Service) handleAddedTask(ctx context.Context, t task.Task) error {
	return nil
}

func (s *Service) handleStartedTask(ctx context.Context, log golog.Logger, t task.Task) error {
	if t.Status != task.StatusInWork {
		return fmt.Errorf("invalid task status: %v", t.Status)
	}

	ins, err := s.repository.StartInspection(ctx, t.ID)
	if err != nil {
		return fmt.Errorf("start inspection: %v", err)
	}

	go s.publisher.Publish(goctx.Wrap(ctx), log, EventTypeStart, ins)

	return nil
}

func (s *Service) handleFinishedTask(ctx context.Context, t task.Task) error {
	return nil
}
