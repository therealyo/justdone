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

func (status OrderStatus) IsFinal() bool {
	return status.isCancel() || status.isRefund()
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

func (os OrderStatus) Value() int {
	return map[OrderStatus]int{
		CoolOrderCreated:       1,
		SbuVerificationPending: 2,
		ConfirmedByMayor:       3,
		Chinazes:               4,
		GiveMyMoneyBack:        5,
		ChangedMyMind:          6,
		Failed:                 7,
	}[os]
}

func ParseOrderStatus(status string) (OrderStatus, error) {
	if val, ok := orderStatusStrings[status]; ok {
		return val, nil
	}
	return "", errors.New("invalid order status")
}
