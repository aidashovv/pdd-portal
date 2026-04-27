package auth

import (
	"fmt"
	"time"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	AccessSecret  string        `envconfig:"ACCESS_SECRET" required:"true"`
	RefreshSecret string        `envconfig:"REFRESH_SECRET" required:"true"`
	AccessTTL     time.Duration `envconfig:"ACCESS_TTL" required:"true"`
	RefreshTTL    time.Duration `envconfig:"REFRESH_TTL" required:"true"`
}

func NewConfig() (Config, error) {
	var config Config
	if err := envconfig.Process("JWT", &config); err != nil {
		return Config{}, fmt.Errorf("process jwt envconfig: %w", err)
	}

	return config, nil
}

func LoadConfig() Config {
	config, err := NewConfig()
	if err != nil {
		panic(fmt.Errorf("load jwt config: %w", err))
	}

	return config
}
