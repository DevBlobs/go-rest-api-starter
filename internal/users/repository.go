package users

import (
	"context"
	"strings"
	"time"

	"github.com/boilerplate-api/go-rest-api-starter/internal/clients/postgres"
	"github.com/jackc/pgx/v5"
)

type User struct {
	ExternalID string
	Email      string
	FirstName  *string
	LastName   *string
	CreatedAt  time.Time
}

type Repository interface {
	GetByExternalID(ctx context.Context, externalID string) (User, error)
	GetByEmail(ctx context.Context, email string) (User, error)
	Create(ctx context.Context, u User) (User, error)
}

type repositoryImpl struct {
	pool *postgres.Client
}

func NewRepository(pool *postgres.Client) Repository {
	return &repositoryImpl{pool: pool}
}

func (r *repositoryImpl) GetByExternalID(ctx context.Context, externalID string) (User, error) {
	const q = `
        SELECT external_id, email, first_name, last_name, created_at
        FROM users
        WHERE external_id = $1
    `
	var u User
	err := r.pool.QueryRow(ctx, q, externalID).Scan(&u.ExternalID, &u.Email, &u.FirstName, &u.LastName, &u.CreatedAt)
	if err != nil {
		return User{}, err
	}
	return u, nil
}

func (r *repositoryImpl) GetByEmail(ctx context.Context, email string) (User, error) {
	const q = `
        SELECT external_id, email, first_name, last_name, created_at
        FROM users
        WHERE email = $1
    `
	var u User
	err := r.pool.QueryRow(ctx, q, email).Scan(&u.ExternalID, &u.Email, &u.FirstName, &u.LastName, &u.CreatedAt)
	if err != nil {
		return User{}, err
	}
	return u, nil
}

func (r *repositoryImpl) Create(ctx context.Context, u User) (User, error) {
	// Normalize inputs
	u.ExternalID = strings.TrimSpace(u.ExternalID)
	u.Email = strings.TrimSpace(strings.ToLower(u.Email))

	const q = `
        INSERT INTO users (external_id, email, first_name, last_name, created_at)
        VALUES ($1, $2, $3, $4, $5)
        ON CONFLICT (external_id) DO UPDATE
            SET email = EXCLUDED.email,
                first_name = EXCLUDED.first_name,
                last_name = EXCLUDED.last_name
        RETURNING external_id, email, first_name, last_name, created_at
    `

	var out User
	err := r.pool.QueryRow(ctx, q, u.ExternalID, u.Email, u.FirstName, u.LastName, u.CreatedAt).Scan(
		&out.ExternalID, &out.Email, &out.FirstName, &out.LastName, &out.CreatedAt,
	)
	if err == nil {
		return out, nil
	}

	// If no row was returned due to conflict, try fetch existing
	if err == pgx.ErrNoRows {
		if u.ExternalID != "" {
			if got, e := r.GetByExternalID(ctx, u.ExternalID); e == nil {
				return got, nil
			}
		}
		if u.Email != "" {
			if got, e := r.GetByEmail(ctx, u.Email); e == nil {
				return got, nil
			}
		}
	}
	return User{}, err
}
