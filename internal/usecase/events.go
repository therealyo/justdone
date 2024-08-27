package usecase

import (
	"github.com/pkg/errors"
	"github.com/therealyo/justdone/domain"
)

type Events struct {
	processor *domain.OrderProcessor
}

func NewEvents(orderProcessor *domain.OrderProcessor) *Events {
	return &Events{
		processor: orderProcessor,
	}
}

func (e *Events) Create(event *domain.OrderEvent) error {
	if err := e.processor.HandleEvent(*event); err != nil {
		if errors.Is(err, domain.ErrOrderAlreadyFinal) {
			return err
		}
		if errors.Is(err, domain.ErrEventConflict) {
			return err
		}
		return errors.Wrap(err, "failed to handle event")
	}

	return nil
}
