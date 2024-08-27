package usecase

import (
	"fmt"
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
	sseClients               map[string][]chan domain.OrderEvent
	sseMutex                 sync.RWMutex
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

func (e *Event) RegisterSSEClient(orderID string, ch chan domain.OrderEvent) {
	e.sseMutex.Lock()
	defer e.sseMutex.Unlock()
	fmt.Println("Registering SSE client for orderID:", orderID)
	fmt.Println("SSE clients:", e.sseClients)
	if e.sseClients == nil {
		e.sseClients = make(map[string][]chan domain.OrderEvent)
	}
	e.sseClients[orderID] = append(e.sseClients[orderID], ch)
}

func (e *Event) UnregisterSSEClient(orderID string, ch chan domain.OrderEvent) {
	e.sseMutex.Lock()
	defer e.sseMutex.Unlock()
	clients := e.sseClients[orderID]
	for i, client := range clients {
		if client == ch {
			e.sseClients[orderID] = append(clients[:i], clients[i+1:]...)
			close(ch)
			break
		}
	}
}

func (e *Event) notifySSEClients(orderID string, event domain.OrderEvent) {
	e.sseMutex.RLock()
	defer e.sseMutex.RUnlock()
	for _, ch := range e.sseClients[orderID] {
		ch <- event
	}
}

func (e *Event) Create(event *domain.OrderEvent) error {
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
			Status:    domain.CoolOrderCreated,
			CreatedAt: event.CreatedAt,
			UpdatedAt: event.UpdatedAt,
		}
	} else if order.IsFinal {
		return domain.ErrOrderAlreadyFinal
	}

	// Instead of immediately updating the order, we'll check for valid transitions
	if err := e.processOrderEvents(order); err != nil {
		return err
	}

	return nil
}

func (e *Event) processOrderEvents(order *domain.Order) error {
	events, err := e.orderRepo.GetEvents(order.OrderID)
	if err != nil {
		return err
	}

	for _, event := range events {
		if domain.IsValidTransition(order.Status, event.OrderStatus) {
			order.Status = event.OrderStatus
			order.UpdatedAt = event.UpdatedAt
			order.IsFinal = event.IsFinal()

			if err := e.orderRepo.Update(order); err != nil {
				return err
			}

			e.notifySSEClients(order.OrderID, event)
		}
	}

	return nil
}

// }

// func (e *Event) Create(event *domain.OrderEvent) error {
// 	existingEvent, err := e.eventRepo.Get(event.EventID)
// 	if err == nil && existingEvent != nil {
// 		return domain.ErrEventConflict
// 	}

// 	if e.currentlyProcessedEvents.Contains(event.EventID) {
// 		return domain.ErrEventConflict
// 	}

// 	e.currentlyProcessedEvents.Add(event.EventID)
// 	defer e.currentlyProcessedEvents.Remove(event.EventID)

// 	if err := e.eventRepo.Create(event); err != nil {
// 		return errors.Wrap(err, "handle event")
// 	}

// 	order, err := e.orderRepo.Get(event.OrderID)
// 	if err != nil {
// 		return err
// 	}

// 	if order == nil {
// 		order = &domain.Order{
// 			OrderID:   event.OrderID,
// 			UserID:    event.UserID,
// 			Status:    event.OrderStatus,
// 			CreatedAt: event.CreatedAt,
// 			UpdatedAt: event.UpdatedAt,
// 		}
// 	} else {
// 		if order.IsFinal {
// 			return domain.ErrOrderAlreadyFinal
// 		}
// 		order.Status = event.OrderStatus
// 		order.UpdatedAt = event.UpdatedAt
// 	}

// 	order.IsFinal = event.IsFinal()

// 	err = e.orderRepo.Save(order)
// 	if err != nil {
// 		return err
// 	}

// 	return nil

// }

func New(eventRepo EventRepository, orderRepo OrderRepository) *Event {
	return &Event{
		eventRepo: eventRepo,
		orderRepo: orderRepo,
		currentlyProcessedEvents: CurrentlyProcessedEvents{
			events: make(map[string]bool),
		},
		sseClients: make(map[string][]chan domain.OrderEvent),
	}
}
