package usecase

import "github.com/therealyo/justdone/domain"

type OrderRepository interface {
	Save(order *domain.Order) error
	Get(id string) (*domain.Order, error)
	GetEvents(id string) ([]domain.OrderEvent, error)
	Update(order *domain.Order) error
}

type Order struct{}

func (o *Order) GetOrder(id string) (*domain.Order, error) {
	return nil, nil
}

func (o *Order) GetOrders(filter *domain.OrderFilter) ([]domain.Order, error) {
	return nil, nil
}
