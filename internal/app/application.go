package app

import (
	"time"

	"github.com/therealyo/justdone/config"
	"github.com/therealyo/justdone/domain"
	"github.com/therealyo/justdone/infrastructure/database/postgres"
	"github.com/therealyo/justdone/internal/inmemory"
	"github.com/therealyo/justdone/internal/sse"
	"github.com/therealyo/justdone/internal/usecase"
)

type Application struct {
	Orders   usecase.Orders
	Events   usecase.Events
	Notifier domain.OrderObserver
}

func New(config *config.Config) (*Application, error) {
	db, err := postgres.New(config.Postgres.ConnectionString)
	if err != nil {
		return nil, err
	}

	sseNotifier := sse.NewSSENotifier()
	orders := postgres.NewOrderRepository(db)

	orderProcessor := domain.NewOrderProcessor(
		orders,
		postgres.NewEventRepository(db),
		sseNotifier,
		inmemory.NewProcessedEvents(),
		30*time.Second,
	)

	return &Application{
		Orders:   usecase.NewOrders(orders),
		Events:   usecase.NewEvents(orderProcessor),
		Notifier: sseNotifier,
	}, nil
}
