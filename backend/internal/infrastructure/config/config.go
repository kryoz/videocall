package config

import (
	"fmt"

	"github.com/caarlos0/env/v11"
)

type Config struct {
	Addr string `env:"ADDR" envDefault:":8080"`
	JWT
	RefreshToken
	Turn
	VAPID
	Storage
}

type Storage struct {
	Type string `env:"STORAGE_TYPE" envDefault:"memory"`

	DBHost      string `env:"DB_HOST" envDefault:"localhost"`
	DBPort      int    `env:"DB_PORT" envDefault:"3306"`
	DBName      string `env:"DB_NAME" envDefault:"videocall"`
	DBUser      string `env:"DB_USER" envDefault:"root"`
	DBPassword  string `env:"DB_PASSWORD" envDefault:""`
	DBMaxConns  int    `env:"DB_MAX_CONNECTIONS" envDefault:"25"`
	DBIdleConns int    `env:"DB_MAX_IDLE_CONNECTIONS" envDefault:"5"`
}

type JWT struct {
	Secret string `env:"JWT_SECRET" envDefault:"secret"`
	TTL    int64  `env:"JWT_TTL" envDefault:"7200"`
}

type RefreshToken struct {
	TTL int64 `env:"REFRESH_TOKEN_TTL" envDefault:"86400"`
}

type Turn struct {
	Secret string `env:"TURN_SECRET" envDefault:"turn_secret"`
	TTL    int64  `env:"TURN_TTL" envDefault:"7200"`
	Host   string `env:"TURN_HOST" envDefault:"localhost:3478"`
}

type VAPID struct {
	PublicKey  string `env:"VAPID_PUBLIC_KEY" envDefault:""`
	PrivateKey string `env:"VAPID_PRIVATE_KEY" envDefault:""`
}

func NewFromEnv() (*Config, error) {
	cfg, err := env.ParseAs[Config]()

	if err != nil {
		return nil, fmt.Errorf("parsing environment config: %w", err)
	}

	return &cfg, nil
}
