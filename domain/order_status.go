package domain

import "errors"

type OrderStatus int

const (
	CoolOrderCreated OrderStatus = iota + 1
	SbuVerificationPending
	ConfirmedByMayor
	Chinazes
	ChangedMyMind
	Failed
	GiveMyMoneyBack
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
	for k, v := range orderStatusStrings {
		if v == os {
			return k
		}
	}
	return "unknown"
}

func ParseOrderStatus(status string) (OrderStatus, error) {
	if val, ok := orderStatusStrings[status]; ok {
		return val, nil
	}
	return 0, errors.New("invalid order status")
}
