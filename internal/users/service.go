package users

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"time"
)

type ProviderUser struct {
	ID        string // The external user ID from the identity provider (e.g., WorkOS)
	Email     string
	FirstName *string
	LastName  *string
}

type Service interface {
	GetOrCreateUser(ctx context.Context, actor *ProviderUser) (*User, error)
}

type serviceImpl struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &serviceImpl{
		repo: repo,
	}
}

func (s serviceImpl) GetOrCreateUser(ctx context.Context, actor *ProviderUser) (*User, error) {
	user, err := s.repo.GetByExternalID(ctx, actor.ID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			slog.Info("user not found, creating", "external_id", actor.ID, "email", actor.Email)
			user, err := s.repo.Create(ctx, User{
				ExternalID: actor.ID,
				Email:      actor.Email,
				FirstName:  actor.FirstName,
				LastName:   actor.LastName,
				CreatedAt:  time.Now(),
			})
			if err != nil {
				slog.Error("create user error", "external_id", actor.ID, "error", err)
				return nil, errors.New("create user error")
			}
			slog.Info("user created", "external_id", user.ExternalID)
			return &user, nil
		}
		slog.Error("get user error", "external_id", actor.ID, "error", err)
		return nil, errors.New("get user error")
	}
	return &user, nil
}
