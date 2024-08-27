package domain

import "time"

type OrderEvent struct {
	EventID     string      `json:"event_id"`
	OrderID     string      `json:"order_id"`
	UserID      string      `json:"user_id"`
	OrderStatus OrderStatus `json:"order_status"`
	CreatedAt   time.Time   `json:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at"`
	IsFinal     bool        `json:"is_final"`
}
