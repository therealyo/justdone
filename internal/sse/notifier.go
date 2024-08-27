package sse

import (
	"sync"
	"time"

	"github.com/therealyo/justdone/domain"
)

type SSENotifier struct {
	mu      sync.Mutex
	clients map[string][]domain.OrderEventsSubscriber
}

func NewSSENotifier() *SSENotifier {
	return &SSENotifier{
		clients: make(map[string][]domain.OrderEventsSubscriber),
	}
}

func (n *SSENotifier) RegisterClient(orderID string, client domain.OrderEventsSubscriber) {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.clients[orderID] = append(n.clients[orderID], client)

	go func() {
		select {
		case <-time.After(client.Timeout):
			client.Disconnect <- true
		case <-client.Disconnect:
			return
		}
	}()
}

func (n *SSENotifier) UnregisterClient(orderID string, client domain.OrderEventsSubscriber) {
	n.mu.Lock()
	defer n.mu.Unlock()

	clients := n.clients[orderID]
	for i, c := range clients {
		if c == client {
			n.clients[orderID] = append(clients[:i], clients[i+1:]...)
			close(client.EventChan)
			break
		}
	}
}

func (n *SSENotifier) Notify(order *domain.Order, event domain.OrderEvent) {
	n.mu.Lock()
	defer n.mu.Unlock()

	for _, client := range n.clients[order.OrderID] {
		select {
		case client.EventChan <- event:
		default:
		}
	}
}

var _ domain.OrderObserver = new(SSENotifier)
