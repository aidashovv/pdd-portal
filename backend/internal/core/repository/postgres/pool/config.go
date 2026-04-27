package pool

import (
	"fmt"
	"time"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	Host     string `envconfig:"HOST" required:"true"`
	Port     string `envconfig:"PORT" default:"5432"`
	User     string `envconfig:"USER" required:"true"`
	Password string `envconfig:"PASSWORD" required:"true"`
	Database string `envconfig:"DB" required:"true"`
	SSLMode  string `envconfig:"SSL_MODE" default:"disable"`

	QueryTimeout time.Duration `envconfig:"QUERY_TIMEOUT" required:"true"`

	MinConns int32 `envconfig:"MIN_CONNS" default:"1"`
	MaxConns int32 `envconfig:"MAX_CONNS" default:"10"`
}

func NewConfig() (Config, error) {
	var config Config
	if err := envconfig.Process("POSTGRES", &config); err != nil {
		return Config{}, fmt.Errorf("process postgres envconfig: %w", err)
	}

	return config, nil
}

func LoadConfig() Config {
	config, err := NewConfig()
	if err != nil {
		panic(fmt.Errorf("load postgres config: %w", err))
	}

	return config
}

func (c Config) DSN() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		c.User,
		c.Password,
		c.Host,
		c.Port,
		c.Database,
		c.SSLMode,
	)
}
