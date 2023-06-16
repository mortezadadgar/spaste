package config

import (
	"fmt"

	"github.com/caarlos0/env/v8"
	"github.com/joho/godotenv"
)

type Config struct {
	Address string `env:"ADDRESS,required"`

	ConnectionString string `env:"CONNECTION_STRING,notEmpty"`

	StaticBase string `env:"STATIC_BASE,notEmpty"`
	Production bool   `env:"PRODUCTION"`

	AddressLength int64 `env:"ADDRESS_LENGTH,notEmpty"`
}

// New returns a new instance of Config.
func New() (Config, error) {
	err := godotenv.Load()
	if err != nil {
		return Config{}, fmt.Errorf("failed to load godotenv: %v", err)
	}

	var cfg Config
	err = env.Parse(&cfg)
	if err != nil {
		return Config{}, fmt.Errorf("failed to parse struct config: %v", err)
	}

	return cfg, nil
}
