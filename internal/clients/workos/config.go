package workos

import (
	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	APIKey    string   `envconfig:"WORKOS_API_KEY" required:"true"`
	ClientID  string   `envconfig:"WORKOS_CLIENT_ID" required:"true"`
	Issuers   []string `envconfig:"WORKOS_ISSUERS" required:"true"`
	Namespace string   `envconfig:"WORKOS_NAMESPACE" default:"https://api.workos.com/claims/user"`
}

func LoadFromEnv() (*Config, error) {
	var cfg Config
	if err := envconfig.Process("", &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
