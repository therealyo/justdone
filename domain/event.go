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

func (e *OrderEvent) Finalize() *OrderEvent {
	e.IsFinal = true
	e.UpdatedAt = time.Now()
	return e
}

// func (e *OrderEvent) IsFinal() bool {
// 	return
// }
