package console

import (
	"fmt"

	"github.com/therealyo/justdone/domain"
)

type ConsoleNotifier struct{}

// RegisterClient implements domain.OrderObserver.
func (c *ConsoleNotifier) RegisterClient(orderID string, client domain.OrderEventsSubscriber) {
	panic("unimplemented")
}

// UnregisterClient implements domain.OrderObserver.
func (c *ConsoleNotifier) UnregisterClient(orderID string, client domain.OrderEventsSubscriber) {
	panic("unimplemented")
}

func (c *ConsoleNotifier) Notify(order *domain.Order, event domain.OrderEvent) {
	fmt.Printf("Order %s has been updated with status %s, isFinal: %t\n", order.OrderID, event.OrderStatus, order.IsFinal)
}

func (c *ConsoleNotifier) AddProcessedEvent(orderID string, event domain.OrderEvent) {
	fmt.Printf("Order %s has been updated with status %s, isFinal: %t\n", orderID, event.OrderStatus, event.IsFinal)
}

func NewConsoleNotifier() *ConsoleNotifier {
	return &ConsoleNotifier{}
}

var _ domain.OrderObserver = new(ConsoleNotifier)
