package redis

import (
	"fmt"
	"time"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	Addr     string        `envconfig:"ADDR" required:"true"`
	Password string        `envconfig:"PASSWORD"`
	DB       int           `envconfig:"DB" default:"0"`
	TTL      time.Duration `envconfig:"TTL" default:"1m"`
}

func NewConfig() (Config, error) {
	var config Config
	if err := envconfig.Process("REDIS", &config); err != nil {
		return Config{}, fmt.Errorf("process redis envconfig: %w", err)
	}

	return config, nil
}
