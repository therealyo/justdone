package domain_test

import (
	"sync"
	"testing"
	"time"

	"github.com/therealyo/justdone/domain"
	"github.com/therealyo/justdone/internal/console"
	"github.com/therealyo/justdone/internal/inmemory"
)

type InMemoryStorageOrders struct {
	orders map[string]*domain.Order
	mu     sync.Mutex
}

type InMemoryStorageEvents struct {
	events map[string]domain.OrderEvent
	mu     sync.Mutex
}

func NewInMemoryOrders() *InMemoryStorageOrders {
	return &InMemoryStorageOrders{
		orders: make(map[string]*domain.Order),
	}
}

func NewInMemoryEvents() *InMemoryStorageEvents {
	return &InMemoryStorageEvents{
		events: make(map[string]domain.OrderEvent),
	}
}

func (s *InMemoryStorageOrders) Get(orderID string) (*domain.Order, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if order, exists := s.orders[orderID]; exists {
		return order, nil
	}
	return nil, nil
}

func (s *InMemoryStorageOrders) Save(order *domain.Order) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.orders[order.OrderID] = order
	return nil
}

func (s *InMemoryStorageEvents) Create(event domain.OrderEvent) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.events[event.EventID]; exists {
		return domain.ErrEventConflict
	}
	s.events[event.EventID] = event
	return nil
}

func (s *InMemoryStorageEvents) Get(eventID string) (*domain.OrderEvent, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if event, exists := s.events[eventID]; exists {
		return &event, nil
	}
	return nil, nil
}

func (s *InMemoryStorageEvents) Delete(eventID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.events, eventID)
	return nil
}

func sendEvent(event domain.OrderEvent) domain.OrderEvent {
	time.Sleep(500 * time.Millisecond)
	return event
}

func TestEnforcesCorrectSequence(t *testing.T) {
	storageOrders := NewInMemoryOrders()
	storageEvents := NewInMemoryEvents()
	notifier := console.NewConsoleNotifier()
	processedEvents := inmemory.NewProcessedEvents()
	processor := domain.NewOrderProcessor(storageOrders, storageEvents, notifier, processedEvents, 5*time.Second)

	event1 := domain.OrderEvent{
		EventID:     "event1",
		OrderID:     "order1",
		UserID:      "user1",
		OrderStatus: domain.CoolOrderCreated,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	event2 := domain.OrderEvent{
		EventID:     "event2",
		OrderID:     "order1",
		UserID:      "user1",
		OrderStatus: domain.ConfirmedByMayor,
		CreatedAt:   time.Now().Add(2 * time.Minute),
		UpdatedAt:   time.Now().Add(2 * time.Minute),
	}

	event3 := domain.OrderEvent{
		EventID:     "event3",
		OrderID:     "order1",
		UserID:      "user1",
		OrderStatus: domain.SbuVerificationPending,
		CreatedAt:   time.Now().Add(1 * time.Minute),
		UpdatedAt:   time.Now().Add(1 * time.Minute),
	}

	// Test processing the first event
	if err := processor.HandleEvent(sendEvent(event1)); err != nil {
		t.Fatalf("Failed to process event1: %v", err)
	}

	// Verify that the order was created and has the correct status
	order, err := storageOrders.Get(event1.OrderID)
	if err != nil || order == nil {
		t.Fatalf("Failed to retrieve order after event1: %v", err)
	}

	if order.Status != domain.CoolOrderCreated {
		t.Errorf("Expected order status to be CoolOrderCreated, got %v", order.Status)
	}

	// Test processing the second event (ConfirmedByMayor) out of sequence
	if err := processor.HandleEvent(sendEvent(event2)); err != nil {
		t.Fatalf("Failed to process event2: %v", err)
	}

	// Verify that the order status was not updated
	order, err = storageOrders.Get(event2.OrderID)
	if err != nil || order == nil {
		t.Fatalf("Failed to retrieve order after event2: %v", err)
	}

	// Verify that the order status was not updated
	if order.Status != domain.CoolOrderCreated {
		t.Errorf("Expected order status to remain CoolOrderCreated, got %v", order.Status)
	}

	// Now process the correct event (SbuVerificationPending)
	if err := processor.HandleEvent(sendEvent(event3)); err != nil {
		t.Fatalf("Failed to process event3 (SbuVerificationPending): %v", err)
	}

	order, err = storageOrders.Get(event3.OrderID)
	if err != nil || order == nil {
		t.Fatalf("Failed to retrieve order after event3: %v", err)
	}

	// Verify that the order status was updated
	if order.Status != domain.ConfirmedByMayor {
		t.Errorf("Expected order status to be ConfirmedByMayor, got %v", order.Status)
	}

	// Error when processing ConfirmedByMayor once more
	if err := processor.HandleEvent(sendEvent(event2)); err == nil {
		t.Fatalf("Expected error when processing ConfirmedByMayor once more, but got none")
	}

}

func TestInvalidInitialEvent(t *testing.T) {
	storageOrders := NewInMemoryOrders()
	storageEvents := NewInMemoryEvents()
	notifier := console.NewConsoleNotifier()
	processedEvents := inmemory.NewProcessedEvents()
	processor := domain.NewOrderProcessor(storageOrders, storageEvents, notifier, processedEvents, 5*time.Second)

	event := domain.OrderEvent{
		EventID:     "event1",
		OrderID:     "order1",
		UserID:      "user1",
		OrderStatus: domain.ConfirmedByMayor,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := processor.HandleEvent(sendEvent(event)); err == nil {
		if err != domain.ErrOrderNotFound {
			t.Fatalf("Expected error when processing an out-of-sequence initial event to be ErrOrderNotFound, but got %v", err)
		}
		t.Fatalf("Expected error when processing an out-of-sequence initial event, but got none")
	}

	order, err := storageOrders.Get(event.OrderID)
	if err != nil {
		t.Fatalf("Failed to retrieve order after processing invalid initial event: %v", err)
	}

	if order != nil {
		t.Errorf("Expected no order to be created, but got one with status: %v", order.Status)
	}

}

func TestConcurrentEventProcessing(t *testing.T) {
	storageOrders := NewInMemoryOrders()
	storageEvents := NewInMemoryEvents()
	notifier := console.NewConsoleNotifier()
	processedEvents := inmemory.NewProcessedEvents()
	processor := domain.NewOrderProcessor(storageOrders, storageEvents, notifier, processedEvents, 5*time.Second)

	initialEvent := domain.OrderEvent{
		EventID:     "initialEvent",
		OrderID:     "order1",
		UserID:      "user1",
		OrderStatus: domain.CoolOrderCreated,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := processor.HandleEvent(initialEvent); err != nil {
		t.Fatalf("Failed to process initialEvent: %v", err)
	}

	sbuVerificationEvent := domain.OrderEvent{
		EventID:     "event2",
		OrderID:     "order1",
		UserID:      "user1",
		OrderStatus: domain.SbuVerificationPending,
		CreatedAt:   time.Now().Add(1 * time.Minute),
		UpdatedAt:   time.Now().Add(1 * time.Minute),
	}

	confirmedByMayorEvent := domain.OrderEvent{
		EventID:     "event3",
		OrderID:     "order1",
		UserID:      "user1",
		OrderStatus: domain.ConfirmedByMayor,
		CreatedAt:   time.Now().Add(2 * time.Minute),
		UpdatedAt:   time.Now().Add(2 * time.Minute),
	}

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		if err := processor.HandleEvent(sbuVerificationEvent); err != nil {
			t.Logf("Failed to process SbuVerificationPending event: %v", err)
		} else {
			t.Log("Successfully processed SbuVerificationPending event")
		}
	}()

	go func() {
		defer wg.Done()
		if err := processor.HandleEvent(confirmedByMayorEvent); err != nil {
			t.Logf("Failed to process ConfirmedByMayor event: %v", err)
		} else {
			t.Log("Successfully processed ConfirmedByMayor event")
		}
	}()

	wg.Wait()

	order, err := storageOrders.Get("order1")
	if err != nil {
		t.Fatalf("Failed to retrieve order after concurrent processing: %v", err)
	}

	if order.Status != domain.ConfirmedByMayor {
		t.Errorf("Expected order status to be ConfirmedByMayor, got %v", order.Status)
	}

	if len(order.Events) != 3 {
		t.Errorf("Expected 3 events to be processed, but got %d", len(order.Events))
	}

	expectedSequence := []domain.OrderStatus{
		domain.CoolOrderCreated,
		domain.SbuVerificationPending,
		domain.ConfirmedByMayor,
	}
	for i, event := range order.Events {
		if event.OrderStatus != expectedSequence[i] {
			t.Errorf("Expected event %d to have status %v, got %v", i, expectedSequence[i], event.OrderStatus)
		}
	}
}

func TestCancelingEventInBetween(t *testing.T) {
	storageOrders := NewInMemoryOrders()
	storageEvents := NewInMemoryEvents()
	notifier := console.NewConsoleNotifier()
	processedEvents := inmemory.NewProcessedEvents()
	processor := domain.NewOrderProcessor(storageOrders, storageEvents, notifier, processedEvents, 5*time.Second)

	event1 := domain.OrderEvent{
		EventID:     "event1",
		OrderID:     "order1",
		UserID:      "user1",
		OrderStatus: domain.CoolOrderCreated,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	event2 := domain.OrderEvent{
		EventID:     "event2",
		OrderID:     "order1",
		UserID:      "user1",
		OrderStatus: domain.SbuVerificationPending,
		CreatedAt:   time.Now().Add(1 * time.Minute),
		UpdatedAt:   time.Now().Add(1 * time.Minute),
	}

	event3 := domain.OrderEvent{
		EventID:     "event3",
		OrderID:     "order1",
		UserID:      "user1",
		OrderStatus: domain.ChangedMyMind,
		CreatedAt:   time.Now().Add(2 * time.Minute),
		UpdatedAt:   time.Now().Add(2 * time.Minute),
	}

	// Process events
	if err := processor.HandleEvent(sendEvent(event1)); err != nil {
		t.Fatalf("Failed to process event1: %v", err)
	}

	if err := processor.HandleEvent(sendEvent(event2)); err != nil {
		t.Fatalf("Failed to process event2: %v", err)
	}

	if err := processor.HandleEvent(sendEvent(event3)); err != nil {
		t.Fatalf("Failed to process event3: %v", err)
	}

	// Verify that the order was finalized with ChangedMyMind status
	order, err := storageOrders.Get(event3.OrderID)
	if err != nil || order == nil {
		t.Fatalf("Failed to retrieve order after ChangedMyMind: %v", err)
	}

	if order.Status != domain.ChangedMyMind {
		t.Errorf("Expected order status to be ChangedMyMind, got %v", order.Status)
	}

	if !order.IsFinal {
		t.Errorf("Expected order to be marked as final after ChangedMyMind")
	}
}

func TestGiveMyMoneyBack(t *testing.T) {
	storageOrders := NewInMemoryOrders()
	storageEvents := NewInMemoryEvents()
	notifier := console.NewConsoleNotifier()
	processedEvents := inmemory.NewProcessedEvents()
	processor := domain.NewOrderProcessor(storageOrders, storageEvents, notifier, processedEvents, 5*time.Second)

	event1 := domain.OrderEvent{
		EventID:     "event1",
		OrderID:     "order1",
		UserID:      "user1",
		OrderStatus: domain.CoolOrderCreated,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	event2 := domain.OrderEvent{
		EventID:     "event2",
		OrderID:     "order1",
		UserID:      "user1",
		OrderStatus: domain.SbuVerificationPending,
		CreatedAt:   time.Now().Add(1 * time.Minute),
		UpdatedAt:   time.Now().Add(1 * time.Minute),
	}

	event3 := domain.OrderEvent{
		EventID:     "event3",
		OrderID:     "order1",
		UserID:      "user1",
		OrderStatus: domain.ConfirmedByMayor,
		CreatedAt:   time.Now().Add(2 * time.Minute),
		UpdatedAt:   time.Now().Add(2 * time.Minute),
	}

	event4 := domain.OrderEvent{
		EventID:     "event4",
		OrderID:     "order1",
		UserID:      "user1",
		OrderStatus: domain.Chinazes,
		CreatedAt:   time.Now().Add(3 * time.Minute),
		UpdatedAt:   time.Now().Add(3 * time.Minute),
	}

	event5 := domain.OrderEvent{
		EventID:     "event5",
		OrderID:     "order1",
		UserID:      "user1",
		OrderStatus: domain.GiveMyMoneyBack,
		CreatedAt:   time.Now().Add(4 * time.Minute),
		UpdatedAt:   time.Now().Add(4 * time.Minute),
	}

	// Process events
	if err := processor.HandleEvent(sendEvent(event1)); err != nil {
		t.Fatalf("Failed to process event1: %v", err)
	}

	if err := processor.HandleEvent(sendEvent(event2)); err != nil {
		t.Fatalf("Failed to process event2: %v", err)
	}

	if err := processor.HandleEvent(sendEvent(event3)); err != nil {
		t.Fatalf("Failed to process event3: %v", err)
	}

	if err := processor.HandleEvent(sendEvent(event4)); err != nil {
		t.Fatalf("Failed to process event4: %v", err)
	}

	if err := processor.HandleEvent(sendEvent(event5)); err != nil {
		t.Fatalf("Failed to process event5: %v", err)
	}

	// Verify that the order was finalized with GiveMyMoneyBack status
	order, err := storageOrders.Get(event5.OrderID)
	if err != nil || order == nil {
		t.Fatalf("Failed to retrieve order after GiveMyMoneyBack: %v", err)
	}

	if order.Status != domain.GiveMyMoneyBack {
		t.Errorf("Expected order status to be GiveMyMoneyBack, got %v", order.Status)
	}

	if !order.IsFinal {
		t.Errorf("Expected order to be marked as final after GiveMyMoneyBack")
	}
}

func TestChinazesFinalization(t *testing.T) {
	storageOrders := NewInMemoryOrders()
	storageEvents := NewInMemoryEvents()
	notifier := console.NewConsoleNotifier()
	processedEvents := inmemory.NewProcessedEvents()
	processor := domain.NewOrderProcessor(storageOrders, storageEvents, notifier, processedEvents, 5*time.Second)

	event1 := domain.OrderEvent{
		EventID:     "event1",
		OrderID:     "order1",
		UserID:      "user1",
		OrderStatus: domain.CoolOrderCreated,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	event2 := domain.OrderEvent{
		EventID:     "event2",
		OrderID:     "order1",
		UserID:      "user1",
		OrderStatus: domain.SbuVerificationPending,
		CreatedAt:   time.Now().Add(1 * time.Minute),
		UpdatedAt:   time.Now().Add(1 * time.Minute),
	}

	event3 := domain.OrderEvent{
		EventID:     "event3",
		OrderID:     "order1",
		UserID:      "user1",
		OrderStatus: domain.ConfirmedByMayor,
		CreatedAt:   time.Now().Add(2 * time.Minute),
		UpdatedAt:   time.Now().Add(2 * time.Minute),
	}

	event4 := domain.OrderEvent{
		EventID:     "event4",
		OrderID:     "order1",
		UserID:      "user1",
		OrderStatus: domain.Chinazes,
		CreatedAt:   time.Now().Add(3 * time.Minute),
		UpdatedAt:   time.Now().Add(3 * time.Minute),
	}

	// Process events
	if err := processor.HandleEvent(sendEvent(event1)); err != nil {
		t.Fatalf("Failed to process event1: %v", err)
	}

	if err := processor.HandleEvent(sendEvent(event2)); err != nil {
		t.Fatalf("Failed to process event2: %v", err)
	}

	if err := processor.HandleEvent(sendEvent(event3)); err != nil {
		t.Fatalf("Failed to process event3: %v", err)
	}

	if err := processor.HandleEvent(sendEvent(event4)); err != nil {
		t.Fatalf("Failed to process event4: %v", err)
	}

	// Wait duration of 10 seconds and check if the order is finalized
	time.Sleep(10 * time.Second)

	order, err := storageOrders.Get(event4.OrderID)
	if err != nil || order == nil {
		t.Fatalf("Failed to retrieve order after Chinazes: %v", err)
	}

	if !order.IsFinal {
		t.Errorf("Expected order to be marked as final after 30 seconds, but it wasn't")
	}

	if order.Status != domain.Chinazes {
		t.Errorf("Expected order status to remain Chinazes, got %v", order.Status)
	}
}

func TestConcurrentProcessing(t *testing.T) {
	storageOrders := NewInMemoryOrders()
	storageEvents := NewInMemoryEvents()
	notifier := console.NewConsoleNotifier()
	processedEvents := inmemory.NewProcessedEvents()
	processor := domain.NewOrderProcessor(storageOrders, storageEvents, notifier, processedEvents, 5*time.Second)

	event := domain.OrderEvent{
		EventID:     "event1",
		OrderID:     "order1",
		UserID:      "user1",
		OrderStatus: domain.CoolOrderCreated,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	var wg sync.WaitGroup
	wg.Add(2)

	// Process the same event concurrently in two goroutines
	go func() {
		defer wg.Done()
		if err := processor.HandleEvent(event); err != nil {
			t.Logf("First goroutine: failed to process event: %v", err)
		} else {
			t.Log("First goroutine: successfully processed event")
		}
	}()

	go func() {
		defer wg.Done()
		if err := processor.HandleEvent(event); err != nil {
			t.Logf("Second goroutine: failed to process event: %v", err)
		} else {
			t.Log("Second goroutine: successfully processed event")
		}
	}()

	wg.Wait()

	// Verify that the event was processed exactly once
	order, err := storageOrders.Get(event.OrderID)
	if err != nil || order == nil {
		t.Fatalf("Failed to retrieve order after concurrent processing: %v", err)
	}

	if order.Status != domain.CoolOrderCreated {
		t.Errorf("Expected order status to be CoolOrderCreated, got %v", order.Status)
	}

	if len(order.Events) != 1 {
		t.Errorf("Expected 1 event to be processed, but got %d", len(order.Events))
	}
}
