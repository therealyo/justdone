package usecase

import (
	"sync"

	"github.com/pkg/errors"
	"github.com/therealyo/justdone/domain"
)

type EventRepository interface {
	Create(event *domain.OrderEvent) error
	Get(id string) (*domain.OrderEvent, error)
}

type Event struct {
	eventRepo                EventRepository
	orderRepo                OrderRepository
	currentlyProcessedEvents CurrentlyProcessedEvents
}

type CurrentlyProcessedEvents struct {
	mu     sync.Mutex
	events map[string]bool
}

func (c *CurrentlyProcessedEvents) Add(eventID string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.events[eventID] = true
}

func (c *CurrentlyProcessedEvents) Contains(eventID string) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.events[eventID]
}

func (c *CurrentlyProcessedEvents) Remove(eventID string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.events, eventID)
}

func (e *Event) Handle(event *domain.OrderEvent) error {
	existingEvent, err := e.eventRepo.Get(event.EventID)
	if err == nil && existingEvent != nil {
		return domain.ErrEventConflict
	}

	if e.currentlyProcessedEvents.Contains(event.EventID) {
		return domain.ErrEventConflict
	}

	e.currentlyProcessedEvents.Add(event.EventID)
	defer e.currentlyProcessedEvents.Remove(event.EventID)

	if err := e.eventRepo.Create(event); err != nil {
		return errors.Wrap(err, "handle event")
	}

	order, err := e.orderRepo.Get(event.OrderID)
	if err != nil {
		return err
	}

	if order == nil {
		order = &domain.Order{
			OrderID:   event.OrderID,
			UserID:    event.UserID,
			Status:    event.OrderStatus,
			CreatedAt: event.CreatedAt,
			UpdatedAt: event.UpdatedAt,
		}
	} else {
		if order.IsFinal {
			return domain.ErrOrderAlreadyFinal
		}
		order.Status = event.OrderStatus
		order.UpdatedAt = event.UpdatedAt
	}

	order.IsFinal = event.IsFinal()

	err = e.orderRepo.Save(order)
	if err != nil {
		return err
	}

	return nil

}

func New(eventRepo EventRepository, orderRepo OrderRepository) *Event {
	return &Event{
		eventRepo: eventRepo,
		orderRepo: orderRepo,
		currentlyProcessedEvents: CurrentlyProcessedEvents{
			events: make(map[string]bool),
		},
	}
}
