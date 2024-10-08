package sse

import (
	"fmt"
	"sync"
	"time"

	"github.com/therealyo/justdone/domain"
)

type SSENotifier struct {
	mu              sync.Mutex
	clients         map[string][]domain.OrderEventsSubscriber
	processedEvents map[string][]string
}

func NewSSENotifier() *SSENotifier {
	return &SSENotifier{
		clients:         make(map[string][]domain.OrderEventsSubscriber),
		processedEvents: make(map[string][]string),
	}
}

func (n *SSENotifier) AddProcessedEvent(orderId string, event domain.OrderEvent) {
	n.mu.Lock()
	defer n.mu.Unlock()

	n.processedEvents[orderId] = append(n.processedEvents[orderId], event.EventID)
}

func (n *SSENotifier) RegisterClient(orderID string, client domain.OrderEventsSubscriber) {
	n.mu.Lock()
	defer n.mu.Unlock()

	// Register the client
	n.clients[orderID] = append(n.clients[orderID], client)

	// Start the timeout handler
	go n.startTimeout(orderID, client)
}

func (n *SSENotifier) startTimeout(orderID string, client domain.OrderEventsSubscriber) {
	timeout := time.NewTimer(client.Timeout)

	for {
		select {
		case <-timeout.C:
			fmt.Println("Timeout: Client timeout", orderID)
			n.UnregisterClient(orderID, client)
			return
		case <-client.Disconnect:
			fmt.Println("Timeout: Client disconnected", orderID)
			n.UnregisterClient(orderID, client)
			return
		case event, ok := <-client.EventChan:
			if !ok {
				return
			}

			// Reset the timeout and propagate the event
			timeout.Reset(client.Timeout)
			client.EventChan <- event
		}
	}
}

func (n *SSENotifier) UnregisterClient(orderID string, client domain.OrderEventsSubscriber) {
	n.mu.Lock()
	defer n.mu.Unlock()

	clients := n.clients[orderID]
	for i, c := range clients {
		if c == client {
			// Remove client from the list
			n.clients[orderID] = append(clients[:i], clients[i+1:]...)
			// Close channels to signal the end of connection
			close(client.EventChan)
			close(client.Disconnect)
			break
		}
	}

	// Clean up if there are no more clients for the order
	if len(n.clients[orderID]) == 0 {
		delete(n.clients, orderID)
		delete(n.processedEvents, orderID)
	}
}

func (n *SSENotifier) Notify(order *domain.Order, event domain.OrderEvent) {
	n.mu.Lock()
	defer n.mu.Unlock()

	if event.IsFinal {
		for _, client := range n.clients[order.OrderID] {
			select {
			case client.EventChan <- event:
			default:
			}
		}

		delete(n.processedEvents, order.OrderID)
		return
	}

	for _, evt := range order.Events {
		if !n.isEventProcessed(order.OrderID, evt) {
			for _, client := range n.clients[order.OrderID] {
				select {
				case client.EventChan <- evt:
				default:
				}
			}

			n.processedEvents[order.OrderID] = append(n.processedEvents[order.OrderID], evt.EventID)
		}
	}
}

func (n *SSENotifier) isEventProcessed(orderID string, event domain.OrderEvent) bool {
	for _, cachedEventID := range n.processedEvents[orderID] {
		if cachedEventID == event.EventID {
			return true
		}
	}
	return false
}

var _ domain.OrderObserver = new(SSENotifier)
