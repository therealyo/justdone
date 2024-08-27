package usecase

import (
	"github.com/pkg/errors"
	"github.com/therealyo/justdone/domain"
)

type Events struct {
	processor *domain.OrderProcessor
}

func NewEvents(orderProcessor *domain.OrderProcessor) Events {
	return Events{
		processor: orderProcessor,
	}
}

func (e Events) Create(event *domain.OrderEvent) error {
	if err := e.processor.HandleEvent(*event); err != nil {
		if domain.IsDomainError(err) {
			return err
		}
		return errors.Wrap(err, "handle event")
	}

	return nil
}
