package sse

import (
	"fmt"
	"sync"
	"time"

	"github.com/therealyo/justdone/domain"
)

type SSENotifier struct {
	mu         sync.Mutex
	clients    map[string][]domain.OrderEventsSubscriber
	eventCache map[string][]domain.OrderEvent
}

func NewSSENotifier() *SSENotifier {
	return &SSENotifier{
		clients:    make(map[string][]domain.OrderEventsSubscriber),
		eventCache: make(map[string][]domain.OrderEvent),
	}
}

func (n *SSENotifier) RegisterClient(orderID string, client domain.OrderEventsSubscriber) {
	n.mu.Lock()
	defer n.mu.Unlock()

	// Send any cached events to the new client

	for _, event := range n.eventCache[orderID] {
		client.EventChan <- event
	}

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
			// n.UnregisterClient(orderID, client)
			// n.mu.Lock()
			n.UnregisterClient(orderID, client)
			// n.mu.Unlock()

			return
		case <-client.Disconnect:
			fmt.Println("Timeout: Client disconnected", orderID)
			// n.UnregisterClient(orderID, client)
			// n.mu.Lock()
			fmt.Println("Unregistering client", orderID)
			n.UnregisterClient(orderID, client)
			// n.mu.Unlock()

			return
		case event, ok := <-client.EventChan:
			// Send event to client and reset the timeout
			if !ok {
				// If the channel is closed, we should return to avoid a panic
				return
			}

			timeout.Reset(client.Timeout)
			client.EventChan <- event
		}
	}
}

// func ()

func (n *SSENotifier) UnregisterClient(orderID string, client domain.OrderEventsSubscriber) {
	n.mu.Lock()
	defer n.mu.Unlock()

	fmt.Println("Unregistering client", orderID)

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
	}

	fmt.Println("Clients", n.clients)
}

func (n *SSENotifier) Notify(order *domain.Order, event domain.OrderEvent) {
	n.mu.Lock()
	defer n.mu.Unlock()

	fmt.Println("Notifying clients", order.OrderID, event)

	// Cache the event
	n.eventCache[order.OrderID] = append(n.eventCache[order.OrderID], event)

	// Notify all clients
	for _, client := range n.clients[order.OrderID] {
		select {
		case client.EventChan <- event:
		default:
		}
	}

	// If the order is in a final state, clean up the clients
	if order.IsFinal {
		n.cleanup(order.OrderID)
	}
}

func (n *SSENotifier) cleanup(orderID string) {
	for _, client := range n.clients[orderID] {
		close(client.EventChan)
		close(client.Disconnect)
	}
	delete(n.clients, orderID)
}

var _ domain.OrderObserver = new(SSENotifier)
