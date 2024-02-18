package internal

import (
	"errors"
	"os"

	"github.com/caarlos0/env/v6"
	"github.com/joho/godotenv"
)

type Config struct {
	Env         string `env:"ENV" envDefault:"development"`
	LogLevel    string `env:"LOG_LEVEL" envDefault:"info"`
	HTTPAddr    string `env:"HTTP_ADDR" envDefault:":3000"`
	PostgresDSN string `env:"POSTGRES_DSN"`
}

func LoadConfig() (Config, error) {
	cfg := Config{}

	if err := godotenv.Load(); err != nil && !errors.Is(err, os.ErrNotExist) {
		return cfg, err
	}

	if err := env.Parse(&cfg); err != nil {
		return cfg, err
	}

	return cfg, nil
}
