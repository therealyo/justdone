package app

import (
	"time"

	"github.com/therealyo/justdone/config"
	"github.com/therealyo/justdone/domain"
	"github.com/therealyo/justdone/infrastructure/database/postgres"
	"github.com/therealyo/justdone/internal/console"
	"github.com/therealyo/justdone/internal/inmemory"
	"github.com/therealyo/justdone/internal/usecase"
)

type Application struct {
	Orders usecase.Orders
	Events usecase.Events
}

func New(config *config.Config) (*Application, error) {
	db, err := postgres.New(config.Postgres.ConnectionString)
	if err != nil {
		return nil, err
	}

	orderProcessor := domain.NewOrderProcessor(
		postgres.NewOrderRepository(db),
		postgres.NewEventRepository(db),
		console.NewConsoleNotifier(),
		inmemory.NewProcessedEvents(),
		30*time.Second,
	)

	return &Application{
		Orders: usecase.Orders{},
		Events: usecase.NewEvents(orderProcessor),
	}, nil
}
