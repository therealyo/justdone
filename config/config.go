package config

import "github.com/caarlos0/env/v9"

type Config struct {
	App struct {
		Port int `env:"PORT" envDefault:"8080"`
	}

	Postgres struct {
		ConnectionString string `env:"POSTGRES_URL"`
	}
}

func New() (*Config, error) {
	cfg := Config{}
	if err := env.ParseWithOptions(&cfg, env.Options{RequiredIfNoDef: true}); err != nil {
		return nil, err
	}
	return &cfg, nil
}
