package s3

import (
	"fmt"
	"time"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	Endpoint     string        `envconfig:"ENDPOINT" required:"true"`
	Region       string        `envconfig:"REGION" required:"true"`
	Bucket       string        `envconfig:"BUCKET" required:"true"`
	AccessKeyID  string        `envconfig:"ACCESS_KEY_ID" required:"true"`
	SecretKey    string        `envconfig:"SECRET_ACCESS_KEY" required:"true"`
	UsePathStyle bool          `envconfig:"USE_PATH_STYLE" default:"true"`
	PresignTTL   time.Duration `envconfig:"PRESIGN_TTL" default:"15m"`
}

func NewConfig() (Config, error) {
	var config Config
	if err := envconfig.Process("S3", &config); err != nil {
		return Config{}, fmt.Errorf("process s3 envconfig: %w", err)
	}

	return config, nil
}
