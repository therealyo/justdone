package app

import (
	"github.com/therealyo/justdone/config"
	"github.com/therealyo/justdone/internal/usecase"
)

type Application struct {
	Orders usecase.Orders
	Events usecase.Events
}

func New(config *config.Config) (*Application, error) {
	return &Application{
		Orders: usecase.Orders{},
		Events: usecase.Events{},
	}, nil
}
