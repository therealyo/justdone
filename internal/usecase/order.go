package usecase

import "github.com/therealyo/justdone/domain"

type OrderRepository interface {
	Save(order *domain.Order) error
	Get(id string) (*domain.Order, error)
	Update(order *domain.Order) error
}

type Order struct{}

// func (o *Order) HandleEvent(event *domain.OrderEvent) error {

// 	return nil
// }

func (o *Order) GetOrder(id string) (*domain.Order, error) {
	return nil, nil
}

func (o *Order) GetOrders() ([]domain.Order, error) {
	return nil, nil
}

// func (o *Order) Update(order *domain.Order) error {
// 	return nil
// }
