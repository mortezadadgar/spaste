package config

import (
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

func New() (*Config, error) {
	err := godotenv.Load()
	if err != nil {
		return nil, err
	}

	var cfg Config
	err = env.Parse(&cfg)
	if err != nil {
		return nil, err
	}

	return &cfg, nil
}
