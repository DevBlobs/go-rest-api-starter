package items

import (
	"context"
	"database/sql"

	"github.com/boilerplate-api/go-rest-api-starter/internal/clients/postgres"
)

type Repository interface {
	Create(ctx context.Context, item Item) (Item, error)
	List(ctx context.Context) ([]Item, error)
	GetByID(ctx context.Context, id string) (Item, error)
	Update(ctx context.Context, item Item) (Item, error)
	Delete(ctx context.Context, id string) error
}

type repositoryImpl struct {
	pool *postgres.Client
}

func NewRepository(pool *postgres.Client) Repository {
	return &repositoryImpl{pool: pool}
}

func (r *repositoryImpl) Create(ctx context.Context, item Item) (Item, error) {
	const q = `
		INSERT INTO items (name, arrival_date)
		VALUES ($1, $2)
		RETURNING id::text, name, arrival_date, created_at
	`
	var out Item
	err := r.pool.QueryRow(ctx, q, item.Name, item.ArrivalDate).Scan(&out.ID, &out.Name, &out.ArrivalDate, &out.CreatedAt)
	if err != nil {
		return Item{}, err
	}
	return out, nil
}

func (r *repositoryImpl) List(ctx context.Context) ([]Item, error) {
	const q = `
		SELECT id::text, name, arrival_date, created_at
		FROM items
		ORDER BY created_at DESC
	`
	rows, err := r.pool.Query(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []Item
	for rows.Next() {
		var item Item
		if err := rows.Scan(&item.ID, &item.Name, &item.ArrivalDate, &item.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

func (r *repositoryImpl) GetByID(ctx context.Context, id string) (Item, error) {
	const q = `
		SELECT id::text, name, arrival_date, created_at
		FROM items
		WHERE id = $1
		LIMIT 1
	`
	var item Item
	err := r.pool.QueryRow(ctx, q, id).Scan(&item.ID, &item.Name, &item.ArrivalDate, &item.CreatedAt)
	if err != nil {
		return Item{}, err
	}
	return item, nil
}

func (r *repositoryImpl) Update(ctx context.Context, item Item) (Item, error) {
	const q = `
		UPDATE items
		SET name = $2, arrival_date = $3
		WHERE id = $1
		RETURNING id::text, name, arrival_date, created_at
	`
	var out Item
	err := r.pool.QueryRow(ctx, q, item.ID, item.Name, item.ArrivalDate).Scan(&out.ID, &out.Name, &out.ArrivalDate, &out.CreatedAt)
	if err != nil {
		return Item{}, err
	}
	return out, nil
}

func (r *repositoryImpl) Delete(ctx context.Context, id string) error {
	const q = `DELETE FROM items WHERE id = $1`
	result, err := r.pool.Exec(ctx, q, id)
	if err != nil {
		return err
	}
	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}
	return nil
}
