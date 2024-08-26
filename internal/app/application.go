package app

import (
	"github.com/therealyo/justdone/config"
	"github.com/therealyo/justdone/internal/usecase"
)

type Application struct {
	Orders *usecase.Order
	Events *usecase.Event
}

func New(config *config.Config) (*Application, error) {
	return &Application{
		Orders: &usecase.Order{},
		Events: &usecase.Event{},
	}, nil
}
