package inmemory

import (
	"sync"

	"github.com/therealyo/justdone/domain"
)

var _ domain.ProcessedEvents = new(ProcessedEvents)

type ProcessedEvents struct {
	mu     sync.Mutex
	events map[string]bool
}

func (c *ProcessedEvents) Add(eventID string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.events[eventID] = true
}

func (c *ProcessedEvents) Contains(eventID string) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.events[eventID]
}

func (c *ProcessedEvents) Remove(eventID string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.events, eventID)
}

func NewProcessedEvents() *ProcessedEvents {
	return &ProcessedEvents{
		events: make(map[string]bool),
	}
}
