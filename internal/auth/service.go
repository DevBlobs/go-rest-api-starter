package auth

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/DevBlobs/go-rest-api-starter/internal/users"
)

type Opts struct {
	OrganizationID string
	ConnectionID   string
}

type Tokens struct {
	AccessToken  string
	RefreshToken string
	ExpiresIn    int
}

type ProviderUser struct {
	ID        string
	Email     string
	FirstName *string
	LastName  *string
	CreatedAt time.Time
}

type Service interface {
	LoginURL(state string, opts Opts) (string, error)
	Exchange(ctx context.Context, code string) (*Tokens, error)
	Refresh(ctx context.Context, rt string) (*Tokens, error)
	JWKS() string
	LogoutURL(sessionID, returnTo string) string
	GetActor(ctx context.Context, sub string) (*ProviderUser, error)
}

type serviceImpl struct {
	prov        Provider
	redirectURI string
	usersSvc    users.Service
}

func NewService(prov Provider, redirectURI string, usersSvc users.Service) Service {
	return &serviceImpl{prov: prov, redirectURI: redirectURI, usersSvc: usersSvc}
}

func (s *serviceImpl) LoginURL(state string, opts Opts) (string, error) {
	return s.prov.AuthURL(s.redirectURI, state, opts)
}

func (s *serviceImpl) Exchange(ctx context.Context, code string) (*Tokens, error) {
	return s.prov.ExchangeCode(ctx, code, s.redirectURI)
}

func (s *serviceImpl) Refresh(ctx context.Context, rt string) (*Tokens, error) {
	return s.prov.Refresh(ctx, rt)
}

func (s *serviceImpl) JWKS() string { return s.prov.JWKSURL() }

func (s *serviceImpl) LogoutURL(sessionID, returnTo string) string {
	return s.prov.LogoutURL(sessionID, returnTo)
}

func (s *serviceImpl) GetActor(ctx context.Context, sub string) (*ProviderUser, error) {
	if strings.TrimSpace(sub) == "" {
		return nil, errors.New("missing sub")
	}

	u, err := s.prov.GetUser(ctx, sub)
	if err != nil {
		return nil, err
	}

	providerUser := &users.ProviderUser{
		ID:        u.ID,
		Email:     u.Email,
		FirstName: u.FirstName,
		LastName:  u.LastName,
	}

	_, err = s.usersSvc.GetOrCreateUser(ctx, providerUser)
	if err != nil {
		return nil, fmt.Errorf("sync user to database: %w", err)
	}

	return u, nil
}
