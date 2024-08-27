package inmemory

import (
	"sync"

	"github.com/therealyo/justdone/domain"
)

var _ domain.ProcessedEvents = new(InMemoryProcessedEvents)

type InMemoryProcessedEvents struct {
	mu     sync.Mutex
	events map[string]bool
}

func (c *InMemoryProcessedEvents) Add(eventID string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.events[eventID] = true
}

func (c *InMemoryProcessedEvents) Contains(eventID string) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.events[eventID]
}

func (c *InMemoryProcessedEvents) Remove(eventID string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.events, eventID)
}

func NewInMemoryProcessedEvents() *InMemoryProcessedEvents {
	return &InMemoryProcessedEvents{
		events: make(map[string]bool),
	}
}
