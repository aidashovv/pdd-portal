package server

import (
	"fmt"
	"time"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	Port            string        `envconfig:"PORT" required:"true"`
	ShutdownTimeout time.Duration `envconfig:"SHUTDOWN_TIMEOUT" required:"true"`
}

func NewConfig() (Config, error) {
	var config Config
	if err := envconfig.Process("HTTP", &config); err != nil {
		return Config{}, fmt.Errorf("process http envconfig: %w", err)
	}

	return config, nil
}

func LoadConfig() Config {
	config, err := NewConfig()
	if err != nil {
		panic(fmt.Errorf("load http config: %w", err))
	}

	return config
}

func (c Config) Addr() string {
	return ":" + c.Port
}
