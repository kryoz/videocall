package config

import (
	"fmt"

	"github.com/caarlos0/env/v11"
)

type Config struct {
	Addr string `env:"ADDR" envDefault:":8080"`
	JWT
	Turn
}

type JWT struct {
	Secret string `env:"JWT_SECRET" envDefault:"secret"`
	TTL    int64  `env:"JWT_TTL" envDefault:"7200"`
}

type Turn struct {
	Secret string `env:"TURN_SECRET" envDefault:"turn_secret"`
	TTL    int64  `env:"TURN_TTL" envDefault:"7200"`
	Host   string `env:"TURN_HOST" envDefault:"localhost:3478"`
}

func NewFromEnv() (*Config, error) {
	cfg, err := env.ParseAs[Config]()

	if err != nil {
		return nil, fmt.Errorf("parsing environment config: %w", err)
	}

	return &cfg, nil
}
