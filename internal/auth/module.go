package auth

import (
	"context"

	"github.com/DevBlobs/go-rest-api-starter/internal/clients/workos"
	"github.com/DevBlobs/go-rest-api-starter/internal/users"
)

type UserService interface {
	GetOrCreateUser(ctx context.Context, actor *users.ProviderUser) (*users.User, error)
}

type Module struct {
	Handler    *Handler
	Middleware *Middleware
}

type ModuleConfig struct {
	Provider        workos.Client
	RedirectURL     string
	BaseURL         string
	Domain          string
	StateSecret     string
	WorkOSClientID  string
	WorkOSIssuers   []string
	WorkOSNamespace string
	UserService     UserService
}

func New(cfg ModuleConfig) (*Module, error) {
	provider := NewProvider(cfg.Provider)
	svc := NewService(provider, cfg.RedirectURL, cfg.UserService)

	jwtValidator, err := NewJWTValidator(svc.JWKS())
	if err != nil {
		return nil, err
	}

	cookieCfg := CookieCfg{Domain: cfg.Domain}
	handler := &Handler{
		Svc:         svc,
		Cookie:      cookieCfg,
		BaseURL:     cfg.BaseURL,
		StateSecret: cfg.StateSecret,
	}
	middleware := &Middleware{
		Svc:            svc,
		Validator:      jwtValidator,
		Cookie:         cookieCfg,
		ClientID:       cfg.WorkOSClientID,
		AllowedIssuers: cfg.WorkOSIssuers,
		Namespace:      cfg.WorkOSNamespace,
	}

	return &Module{
		Handler:    handler,
		Middleware: middleware,
	}, nil
}
