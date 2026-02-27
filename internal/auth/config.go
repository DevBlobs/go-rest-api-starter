package auth

import (
	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	RedirectURL    string   `envconfig:"AUTH_REDIRECT_URL" required:"true"`
	Domain         string   `envconfig:"AUTH_DOMAIN" default:"localhost"`
	BaseURL        string   `envconfig:"AUTH_BASE_URL" default:"http://localhost:5173"`
	AllowedOrigins []string `envconfig:"AUTH_ALLOWED_ORIGINS" default:"http://localhost:5173"`
	StateSecret    string   `envconfig:"AUTH_STATE_SECRET" required:"true"`
}

func LoadFromEnv() (*Config, error) {
	var cfg Config
	if err := envconfig.Process("", &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
