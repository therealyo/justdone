package domain

import (
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/pkg/errors"
)

type OrderRepository interface {
	Get(orderID string) (*Order, error)
	GetMany(filter *OrderFilter) ([]Order, error)
	Save(order *Order) error
}

type EventRepository interface {
	Get(eventID string) (*OrderEvent, error)
	Create(event OrderEvent) error
	Update(event OrderEvent) error
	Delete(eventID string) error
}

type OrderEventsSubscriber struct {
	EventChan  chan OrderEvent
	Disconnect chan bool
	Timeout    time.Duration
}

func NewOrderEventsSubscriber(timeout time.Duration) OrderEventsSubscriber {
	return OrderEventsSubscriber{
		EventChan:  make(chan OrderEvent, 1),
		Disconnect: make(chan bool),
		Timeout:    timeout,
	}
}

type OrderObserver interface {
	RegisterClient(orderID string, client OrderEventsSubscriber)
	UnregisterClient(orderID string, client OrderEventsSubscriber)
	AddProcessedEvent(orderID string, event OrderEvent)
	Notify(order *Order, event OrderEvent)
}

type ProcessedEvents interface {
	Add(eventID string)
	Contains(eventID string) bool
	Remove(eventID string)
}

// OrderProcessor handles incoming order events, ensures correct event sequencing,
// and manages the order lifecycle.
type OrderProcessor struct {
	orderRepo       OrderRepository
	eventRepo       EventRepository
	observer        OrderObserver
	processing      ProcessedEvents
	finalizeTimeout time.Duration

	mu sync.Mutex
}

// HandleEvent processes an incoming OrderEvent, handling deduplication,
// order creation, event sequencing, and client notifications.
func (op *OrderProcessor) HandleEvent(event OrderEvent) error {
	// Check if the event has already been processed
	if op.isEventAlreadyProcessed(event.EventID) {
		return ErrEventConflict
	}

	// Mark the event as being processed
	op.processing.Add(event.EventID)
	defer op.processing.Remove(event.EventID)

	// Process the event
	if err := op.processEvent(event); err != nil {
		// Handle domain errors
		if IsDomainError(err) {
			return err
		}
		// Delete the event if processing failed
		if deleteErr := op.eventRepo.Delete(event.EventID); deleteErr != nil {
			return errors.Wrap(deleteErr, "delete event")
		}
		return errors.Wrap(err, "process event")
	}

	return nil
}

// processEvent handles the core logic of event processing, including
// order updates, event sequencing, and finalization.
func (op *OrderProcessor) processEvent(event OrderEvent) error {
	op.mu.Lock()
	defer op.mu.Unlock()

	// Retrieve the order
	order, err := op.orderRepo.Get(event.OrderID)
	if err != nil {
		return errors.Wrap(err, "retrieve order")
	}

	// Create a new order if it doesn't exist
	if order == nil {
		if event.OrderStatus == CoolOrderCreated {
			order = &Order{
				OrderID:   event.OrderID,
				UserID:    event.UserID,
				Status:    CoolOrderCreated,
				CreatedAt: event.CreatedAt,
				UpdatedAt: event.UpdatedAt,
			}

			if err := op.orderRepo.Save(order); err != nil {
				return errors.Wrap(err, "save new order")
			}
		} else {
			return ErrOrderNotFound
		}
	}

	// Check if the order is already final
	if order.IsFinal {
		return ErrOrderAlreadyFinal
	}

	// Save the event
	if err := op.eventRepo.Create(event); err != nil {
		return errors.Wrap(err, "save event")
	}

	// Append and sort events
	order.Events = append(order.Events, event)
	sort.SliceStable(order.Events, func(i, j int) bool {
		return order.Events[i].CreatedAt.Before(order.Events[j].CreatedAt)
	})

	// Handle cancellation events
	if event.OrderStatus.isCancel() {
		order.IsFinal = true
		order.Status = event.OrderStatus
		order.LastEvent = &event
		order.UpdatedAt = event.UpdatedAt

		if err := op.orderRepo.Save(order); err != nil {
			return errors.Wrap(err, "save order")
		}
		op.observer.Notify(order, event)
		return nil
	}

	// Process the last event
	lastEvent := order.Events[len(order.Events)-1]
	if order.isValidSequence() {
		order.Status = lastEvent.OrderStatus
		order.LastEvent = &lastEvent
		order.UpdatedAt = lastEvent.UpdatedAt

		// Start finalization timer for Chinazes status
		if lastEvent.OrderStatus == Chinazes {
			go op.waitAndFinalize(order, lastEvent)
		}

		// Mark order as final for refund status
		if lastEvent.OrderStatus.isRefund() {
			order.IsFinal = true
		}

		// Save the updated order
		if err := op.orderRepo.Save(order); err != nil {
			return errors.Wrap(err, "save order")
		}

		// Notify observers
		op.observer.Notify(order, lastEvent)
	}
	return nil
}

// waitAndFinalize starts a timer to finalize an order after receiving
// the Chinazes status, marking it as complete if no further events are received.
func (op *OrderProcessor) waitAndFinalize(order *Order, lastEvent OrderEvent) {
	time.Sleep(op.finalizeTimeout)

	op.mu.Lock()
	defer op.mu.Unlock()

	// Retrieve the latest order state
	finalOrder, err := op.orderRepo.Get(order.OrderID)
	if err != nil || finalOrder == nil {
		fmt.Println("order not found")
		return
	}

	// Finalize the order if still in Chinazes status
	if finalOrder.Status == Chinazes && !finalOrder.IsFinal {
		finalOrder.IsFinal = true
		if err := op.orderRepo.Save(finalOrder); err != nil {
			fmt.Println("error saving order")
			return
		}

		// Update the event to finalized state
		updatedEvent := lastEvent.Finalize()
		if err := op.eventRepo.Update(*updatedEvent); err != nil {
			fmt.Println("error updating event")
			return
		}

		// Notify observers of the finalized order
		op.observer.Notify(finalOrder, *updatedEvent)
	}
}

// isEventAlreadyProcessed checks if an event has already been processed
// or is currently being processed to avoid duplicates.
func (op *OrderProcessor) isEventAlreadyProcessed(eventID string) bool {
	if op.processing.Contains(eventID) {
		return true
	}

	existingEvent, err := op.eventRepo.Get(eventID)
	if err == nil && existingEvent != nil {
		return true
	}

	return false
}

func NewOrderProcessor(
	orderRepo OrderRepository,
	eventRepo EventRepository,
	observer OrderObserver,
	processing ProcessedEvents,
	finalizeTimeout time.Duration,
) *OrderProcessor {
	return &OrderProcessor{
		orderRepo:       orderRepo,
		eventRepo:       eventRepo,
		observer:        observer,
		processing:      processing,
		finalizeTimeout: finalizeTimeout,
	}
}
