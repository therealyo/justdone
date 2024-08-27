package domain

import "errors"

type OrderStatus string

const (
	CoolOrderCreated       OrderStatus = "cool_order_created"
	SbuVerificationPending OrderStatus = "sbu_verification_pending"
	ConfirmedByMayor       OrderStatus = "confirmed_by_mayor"
	Chinazes               OrderStatus = "chinazes"
	ChangedMyMind          OrderStatus = "changed_my_mind"
	Failed                 OrderStatus = "failed"
	GiveMyMoneyBack        OrderStatus = "give_my_money_back"
)

func (status OrderStatus) isCancel() bool {
	return status == ChangedMyMind || status == Failed
}

func (status OrderStatus) isRefund() bool {
	return status == GiveMyMoneyBack
}

var orderStatusStrings = map[string]OrderStatus{
	"cool_order_created":       CoolOrderCreated,
	"sbu_verification_pending": SbuVerificationPending,
	"confirmed_by_mayor":       ConfirmedByMayor,
	"chinazes":                 Chinazes,
	"changed_my_mind":          ChangedMyMind,
	"failed":                   Failed,
	"give_my_money_back":       GiveMyMoneyBack,
}

func (os OrderStatus) String() string {
	return string(os)
}

func ParseOrderStatus(status string) (OrderStatus, error) {
	if val, ok := orderStatusStrings[status]; ok {
		return val, nil
	}
	return "", errors.New("invalid order status")
}
