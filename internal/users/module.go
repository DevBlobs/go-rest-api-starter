package users

import (
	"github.com/DevBlobs/go-rest-api-starter/internal/clients/postgres"
)

type Module struct {
	Service Service
}

func New(db *postgres.Client) *Module {
	repo := NewRepository(db)
	svc := NewService(repo)
	return &Module{Service: svc}
}
