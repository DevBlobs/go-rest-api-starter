package postgres

import (
	"time"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	// If DSN is provided, it will be used as-is. Example: postgresql://user:pass@host/db
	DSN             string        `envconfig:"DATABASE_URL"`
	Host            string        `envconfig:"DB_HOST"`
	Port            int           `envconfig:"DB_PORT" default:"5432"`
	User            string        `envconfig:"DB_USER"`
	Password        string        `envconfig:"DB_PASSWORD"`
	Database        string        `envconfig:"DB_NAME"`
	SSLMode         string        `envconfig:"DB_SSLMODE" default:"disable"`
	MaxConns        int32         `envconfig:"DB_MAX_CONNS" default:"10"`
	MinConns        int32         `envconfig:"DB_MIN_CONNS" default:"0"`
	MaxConnLifetime time.Duration `envconfig:"DB_MAX_CONN_LIFETIME" default:"30m"`
	ConnTimeout     time.Duration `envconfig:"DB_CONN_TIMEOUT" default:"5s"`
}

func LoadFromEnv() (*Config, error) {
	var cfg Config
	if err := envconfig.Process("", &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
