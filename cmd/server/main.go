package main

import (
	"fmt"
	"log"

	_ "github.com/therealyo/justdone/docs"

	"github.com/therealyo/justdone/config"
	"github.com/therealyo/justdone/internal/api/http"
	"github.com/therealyo/justdone/internal/app"
)

func main() {
	config, err := config.New()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	app, err := app.New(config)
	if err != nil {
		log.Fatalf("failed to create app: %v", err)
	}

	server, err := http.NewServer(app).Setup()
	if err != nil {
		log.Fatalf("failed to setup server: %v", err)
	}
	if err := server.Run(fmt.Sprintf(":%d", config.App.Port)); err != nil {
		log.Fatalf("failed to run server: %v", err)
	}
}
