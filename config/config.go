package config

import (
	"fmt"

	"github.com/caarlos0/env/v8"
	"github.com/joho/godotenv"
)

type Config struct {
	Address      string `env:"ADDRESS"`
	SecretKey    string `env:"SECRET_KEY,unset"`
	PasswordCost int    `env:"PASSWORD_COST"`

	ConnectionString string `env:"CONNECTION_STRING"`

	StaticBase string `env:"STATIC_BASE"`
	Production bool   `env:"PRODUCTION"`
}

// New returns a new instance of Config.
func New() (*Config, error) {
	err := godotenv.Load()
	if err != nil {
		return nil, fmt.Errorf("failed to return Config: %v", err)
	}

	var cfg Config
	err = env.Parse(&cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to return Config: %v", err)
	}

	return &cfg, nil
}
