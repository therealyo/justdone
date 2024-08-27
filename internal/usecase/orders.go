package usecase

import (
	"github.com/therealyo/justdone/domain"
)

type Orders struct {
	orderRepo domain.OrderRepository
}

func NewOrders(orderRepo domain.OrderRepository) Orders {
	return Orders{orderRepo: orderRepo}
}

func (o *Orders) GetOrder(id string) (*domain.Order, error) {
	return o.orderRepo.Get(id)
}

func (o *Orders) GetOrders(filter *domain.OrderFilter) ([]domain.Order, error) {
	orders, err := o.orderRepo.GetMany(filter)
	if err != nil {
		return nil, err
	}
	return orders, nil
}
