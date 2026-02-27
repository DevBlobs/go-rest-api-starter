package items

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"

	"github.com/boilerplate-api/go-rest-api-starter/internal/shared/tz"
)

type Service interface {
	CreateItem(ctx context.Context, req CreateItemRequest) (*Item, error)
	ListItems(ctx context.Context) ([]Item, error)
	GetItem(ctx context.Context, id string) (*Item, error)
	UpdateItem(ctx context.Context, id string, req UpdateItemRequest) (*Item, error)
	DeleteItem(ctx context.Context, id string) error
}

type serviceImpl struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &serviceImpl{repo: repo}
}

func (s *serviceImpl) CreateItem(ctx context.Context, req CreateItemRequest) (*Item, error) {
	item := req.ToItem()

	if req.ArrivalDate != "" {
		parsedDate, err := tz.ParseDate(req.ArrivalDate)
		if err != nil {
			slog.Error("parse arrival date error:", err)
			return nil, fmt.Errorf("parse arrival date: %w", err)
		}
		item.ArrivalDate = &parsedDate
	}

	item, err := s.repo.Create(ctx, item)
	if err != nil {
		slog.Error("create item error:", err)
		return nil, fmt.Errorf("create item: %w", err)
	}
	return &item, nil
}

func (s *serviceImpl) ListItems(ctx context.Context) ([]Item, error) {
	items, err := s.repo.List(ctx)
	if err != nil {
		slog.Error("list items error:", err)
		return nil, fmt.Errorf("list items: %w", err)
	}
	return items, nil
}

func (s *serviceImpl) GetItem(ctx context.Context, id string) (*Item, error) {
	item, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		slog.Error("get item error:", err)
		return nil, fmt.Errorf("get item: %w", err)
	}
	return &item, nil
}

func (s *serviceImpl) UpdateItem(ctx context.Context, id string, req UpdateItemRequest) (*Item, error) {
	item, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		slog.Error("update item error:", err)
		return nil, fmt.Errorf("update item: %w", err)
	}

	item.Name = req.Name

	if req.ArrivalDate != "" {
		parsedDate, err := tz.ParseDate(req.ArrivalDate)
		if err != nil {
			slog.Error("parse arrival date error:", err)
			return nil, fmt.Errorf("parse arrival date: %w", err)
		}
		item.ArrivalDate = &parsedDate
	}

	updated, err := s.repo.Update(ctx, item)
	if err != nil {
		slog.Error("update item error:", err)
		return nil, fmt.Errorf("update item: %w", err)
	}
	return &updated, nil
}

func (s *serviceImpl) DeleteItem(ctx context.Context, id string) error {
	err := s.repo.Delete(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil
		}
		slog.Error("delete item error:", err)
		return fmt.Errorf("delete item: %w", err)
	}
	return nil
}
