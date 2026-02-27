package auth

import (
	"context"
	"time"

	wk "github.com/boilerplate-api/go-rest-api-starter/internal/clients/workos"
)

type Provider interface {
	AuthURL(redirectURI, state string, opts Opts) (string, error)
	ExchangeCode(ctx context.Context, code, redirectURI string) (*Tokens, error)
	Refresh(ctx context.Context, refreshToken string) (*Tokens, error)
	JWKSURL() string
	LogoutURL(sessionID, returnTo string) string
	GetUser(ctx context.Context, providerID string) (*ProviderUser, error)
}

type provider struct{ c wk.Client }

func NewProvider(c wk.Client) Provider { return &provider{c: c} }

func (p *provider) AuthURL(redirectURI, state string, opts Opts) (string, error) {
	return p.c.AuthURL(redirectURI, state, wk.AuthOpts{
		OrganizationID: opts.OrganizationID,
		ConnectionID:   opts.ConnectionID,
	})
}

func (p *provider) ExchangeCode(ctx context.Context, code, redirectURI string) (*Tokens, error) {
	t, err := p.c.ExchangeCode(ctx, code, redirectURI)
	if err != nil {
		return nil, err
	}
	return &Tokens{AccessToken: t.AccessToken, RefreshToken: t.RefreshToken, ExpiresIn: t.ExpiresIn}, nil
}

func (p *provider) Refresh(ctx context.Context, rt string) (*Tokens, error) {
	t, err := p.c.Refresh(ctx, rt)
	if err != nil {
		return nil, err
	}
	return &Tokens{AccessToken: t.AccessToken, RefreshToken: t.RefreshToken, ExpiresIn: t.ExpiresIn}, nil
}

func (p *provider) JWKSURL() string { return p.c.JWKSURL() }

func (p *provider) LogoutURL(sessionID, returnTo string) string {
	return p.c.LogoutURL(sessionID, returnTo)
}

func (p *provider) GetUser(ctx context.Context, providerID string) (*ProviderUser, error) {
	user, err := p.c.GetUser(ctx, providerID)
	if err != nil {
		return nil, err
	}
	var createdAt time.Time
	if user.CreatedAt != nil {
		if t, err := time.Parse(time.RFC3339, *user.CreatedAt); err == nil {
			createdAt = t
		}
	}
	return &ProviderUser{
		ID:        user.ID,
		Email:     user.Email,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		CreatedAt: createdAt,
	}, nil
}
