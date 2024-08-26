package domain

import (
	"time"
)

type OrderStatus string

const (
	CoolOrderCreated       OrderStatus = "cool_order_created"
	SbuVerificationPending OrderStatus = "sbu_verification_pending"
	ConfirmedByMayor       OrderStatus = "confirmed_by_mayor"
	ChangedMyMind          OrderStatus = "changed_my_mind"
	Failed                 OrderStatus = "failed"
	Chinazes               OrderStatus = "chinazes"
	GiveMyMoneyBack        OrderStatus = "give_my_money_back"
)

type OrderEvent struct {
	EventID     string      `json:"event_id"`
	OrderID     string      `json:"order_id"`
	UserID      string      `json:"user_id"`
	OrderStatus OrderStatus `json:"order_status"`
	CreatedAt   time.Time   `json:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at"`
}

func IsFinalStatus(status OrderStatus) bool {
	return status == ChangedMyMind || status == Failed || status == GiveMyMoneyBack
}

func IsValidTransition(currentStatus, newStatus OrderStatus) bool {
	if currentStatus == "" {
		return newStatus == CoolOrderCreated
	}

	validTransitions := map[OrderStatus][]OrderStatus{
		CoolOrderCreated:       {SbuVerificationPending, ChangedMyMind, Failed},
		SbuVerificationPending: {ConfirmedByMayor, ChangedMyMind, Failed},
		ConfirmedByMayor:       {Chinazes, ChangedMyMind, Failed},
		Chinazes:               {GiveMyMoneyBack},
	}

	if transitions, ok := validTransitions[currentStatus]; ok {
		for _, validStatus := range transitions {
			if newStatus == validStatus {
				return true
			}
		}
	}

	return false
}

func (e *OrderEvent) IsFinal() bool {
	return IsFinalStatus(e.OrderStatus)
}
