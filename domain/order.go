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

type OrderFilter struct {
	Status    []OrderStatus
	UserID    string
	IsFinal   bool
	Limit     int
	Offset    int
	SortBy    string
	SortOrder string
}

type FilterOption func(*OrderFilter)

func WithStatus(statuses ...OrderStatus) FilterOption {
	return func(f *OrderFilter) {
		f.Status = append(f.Status, statuses...)
	}
}

func WithUserID(userID string) FilterOption {
	return func(f *OrderFilter) {
		f.UserID = userID
	}
}

func WithLimit(limit int) FilterOption {
	return func(f *OrderFilter) {
		f.Limit = limit
	}
}

func WithOffset(offset int) FilterOption {
	return func(f *OrderFilter) {
		f.Offset = offset
	}
}

func WithIsFinal(isFinal bool) FilterOption {
	return func(f *OrderFilter) {
		f.IsFinal = isFinal
	}
}

func WithSortBy(sortBy string) FilterOption {
	return func(f *OrderFilter) {
		f.SortBy = sortBy
	}
}

func WithSortOrder(sortOrder string) FilterOption {
	return func(f *OrderFilter) {
		f.SortOrder = sortOrder
	}
}

func NewOrderFilter(options ...FilterOption) *OrderFilter {
	filter := &OrderFilter{
		Limit:     10,
		Offset:    0,
		SortBy:    "created_at",
		SortOrder: "desc",
	}

	for _, option := range options {
		option(filter)
	}

	return filter
}
