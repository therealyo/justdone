package usecase

import "github.com/therealyo/justdone/domain"

type Orders struct{}

func (o *Orders) GetOrder(id string) (*domain.Order, error) {
	return nil, nil
}

func (o *Orders) GetOrders(filter *domain.OrderFilter) ([]domain.Order, error) {
	return nil, nil
}
