package tz

import (
	"fmt"
	"time"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	Timezone string `envconfig:"COMPANY_TIMEZONE" required:"true"`
}

var location *time.Location

func LoadFromEnv() (*Config, error) {
	var cfg Config
	if err := envconfig.Process("", &cfg); err != nil {
		return nil, fmt.Errorf("failed to load timezone config: %w", err)
	}
	return &cfg, nil
}

func Init(cfg *Config) error {
	var err error
	location, err = time.LoadLocation(cfg.Timezone)
	if err != nil {
		return fmt.Errorf("failed to load timezone %s: %w", cfg.Timezone, err)
	}
	return nil
}

func ParseDate(s string) (time.Time, error) {
	return time.ParseInLocation("20060102", s, location)
}
