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
	Save(order *Order) error
}

type EventRepository interface {
	Get(eventID string) (*OrderEvent, error)
	Create(event OrderEvent) error
	Delete(eventID string) error
}

type OrderObserver interface {
	Notify(order *Order, event OrderEvent)
}

type ProcessedEvents interface {
	Add(eventID string)
	Contains(eventID string) bool
	Remove(eventID string)
}

type OrderProcessor struct {
	orderRepo       OrderRepository
	eventRepo       EventRepository
	observer        OrderObserver
	processing      ProcessedEvents
	finalizeTimeout time.Duration

	mu sync.Mutex
}

func (op *OrderProcessor) HandleEvent(event OrderEvent) error {
	if op.isEventAlreadyProcessed(event.EventID) {
		return ErrEventConflict
	}

	op.processing.Add(event.EventID)
	defer op.processing.Remove(event.EventID)

	if err := op.processEvent(event); err != nil {
		if IsDomainError(err) {
			return err
		}
		if deleteErr := op.eventRepo.Delete(event.EventID); deleteErr != nil {
			return errors.Wrap(deleteErr, "delete event")
		}
		return errors.Wrap(err, "process event")
	}

	return nil
}

func (op *OrderProcessor) processEvent(event OrderEvent) error {
	op.mu.Lock()
	defer op.mu.Unlock()

	order, err := op.orderRepo.Get(event.OrderID)
	if err != nil {
		return errors.Wrap(err, "retrieve order")
	}

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

	if order.IsFinal {
		return ErrOrderAlreadyFinal
	}

	if err := op.eventRepo.Create(event); err != nil {
		return errors.Wrap(err, "save event")
	}

	order.Events = append(order.Events, event)
	sort.SliceStable(order.Events, func(i, j int) bool {
		return order.Events[i].CreatedAt.Before(order.Events[j].CreatedAt)
	})

	lastEvent := order.Events[len(order.Events)-1]
	if lastEvent.OrderStatus.isCancel() {
		order.IsFinal = true
		order.Status = lastEvent.OrderStatus
		order.LastEvent = &lastEvent
		order.UpdatedAt = lastEvent.UpdatedAt

		if err := op.orderRepo.Save(order); err != nil {
			return errors.Wrap(err, "save order")
		}
		op.observer.Notify(order, lastEvent)
		return nil
	}

	if order.isValidSequence() {
		order.Status = lastEvent.OrderStatus
		order.LastEvent = &lastEvent
		order.UpdatedAt = lastEvent.UpdatedAt

		if lastEvent.OrderStatus == Chinazes {
			go op.waitAndFinalize(order, lastEvent)
		}

		if lastEvent.OrderStatus.isRefund() {
			order.IsFinal = true
		}

		if err := op.orderRepo.Save(order); err != nil {
			return errors.Wrap(err, "save order")
		}

		op.observer.Notify(order, lastEvent)
	}
	return nil
}

func (op *OrderProcessor) waitAndFinalize(order *Order, lastEvent OrderEvent) {
	time.Sleep(op.finalizeTimeout)

	op.mu.Lock()
	defer op.mu.Unlock()

	finalOrder, err := op.orderRepo.Get(order.OrderID)
	if err != nil || finalOrder == nil {
		fmt.Println("order not found")
		return
	}

	if finalOrder.Status == Chinazes && !finalOrder.IsFinal {
		finalOrder.IsFinal = true
		if err := op.orderRepo.Save(finalOrder); err != nil {
			fmt.Println("error saving order")
			return
		}
		op.observer.Notify(finalOrder, lastEvent)
	}
}

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
) OrderProcessor {
	return OrderProcessor{
		orderRepo:       orderRepo,
		eventRepo:       eventRepo,
		observer:        observer,
		processing:      processing,
		finalizeTimeout: finalizeTimeout,
	}
}
