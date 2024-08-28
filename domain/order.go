package domain

import (
	"time"

	"github.com/therealyo/justdone/pkg/array"
)

type Order struct {
	OrderID   string       `json:"order_id"`
	UserID    string       `json:"user_id"`
	Status    OrderStatus  `json:"status"`
	IsFinal   bool         `json:"is_final"`
	Events    []OrderEvent `json:"-"`
	LastEvent *OrderEvent  `json:"-"`
	CreatedAt time.Time    `json:"created_at"`
	UpdatedAt time.Time    `json:"updated_at"`
}

func (o *Order) isValidSequence() bool {
	requiredSequence := []OrderStatus{
		CoolOrderCreated,
		SbuVerificationPending,
		ConfirmedByMayor,
		Chinazes,
		GiveMyMoneyBack,
	}

	currentSequence := []OrderStatus{}

	for _, event := range o.Events {
		currentSequence = append(currentSequence, event.OrderStatus)
	}

	return array.IsSubArray(currentSequence, requiredSequence)
}

type OrderFilter struct {
	Status    []OrderStatus
	UserID    string
	IsFinal   *bool
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

func WithIsFinal(isFinal *bool) FilterOption {
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
