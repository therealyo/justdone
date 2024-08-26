package domain

import (
	"errors"
	"time"
)

type Order struct {
	OrderID   string
	UserID    string
	Status    OrderStatus
	IsFinal   bool
	Events    []OrderEvent
	LastEvent *OrderEvent
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (o *Order) ProcessEvent(event OrderEvent) error {
	if o.IsFinal {
		return errors.New("order is already in final status")
	}

	if !IsValidTransition(o.Status, event.OrderStatus) {
		return errors.New("invalid status transition")
	}

	o.Status = event.OrderStatus
	o.Events = append(o.Events, event)
	o.LastEvent = &event
	o.UpdatedAt = event.UpdatedAt

	if o.CreatedAt.IsZero() {
		o.CreatedAt = event.CreatedAt
	}

	if IsFinalStatus(event.OrderStatus) {
		o.IsFinal = true
	} else if event.OrderStatus == Chinazes {
		// Start a goroutine to check if the order becomes final after 30 seconds
		go func(o *Order) {
			time.Sleep(30 * time.Second)
			if o.Status == Chinazes {
				o.IsFinal = true
			}
		}(o)
	}

	return nil
}
