package config

import (
	"fmt"
	"time"

	"github.com/caarlos0/env/v11"
)

type Config struct {
	Addr string `env:"ADDR" envDefault:":8080"`
	JWT
	RefreshToken
	Turn
	VAPID
	Storage
	RoomConfig
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
	Secret string        `env:"JWT_SECRET" envDefault:"secret"`
	TTL    time.Duration `env:"JWT_TTL" envDefault:"2h"`
}

type RefreshToken struct {
	TTL time.Duration `env:"REFRESH_TOKEN_TTL" envDefault:"24h"`
}

type Turn struct {
	Secret string        `env:"TURN_SECRET" envDefault:"turn_secret"`
	TTL    time.Duration `env:"TURN_TTL" envDefault:"2h"`
	Host   string        `env:"TURN_HOST" envDefault:"localhost:3478"`
}

type VAPID struct {
	PublicKey  string `env:"VAPID_PUBLIC_KEY" envDefault:""`
	PrivateKey string `env:"VAPID_PRIVATE_KEY" envDefault:""`
}

type RoomConfig struct {
	TTL           time.Duration `env:"ROOM_TTL" envDefault:"4h"`
	CleanInterval time.Duration `env:"ROOM_CLEAN_INTERVAL" envDefault:"60s"`
}

func NewFromEnv() (*Config, error) {
	cfg, err := env.ParseAs[Config]()

	if err != nil {
		return nil, fmt.Errorf("parsing environment config: %w", err)
	}

	return &cfg, nil
}
