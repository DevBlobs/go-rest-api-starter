package items

import (
	"github.com/DevBlobs/go-rest-api-starter/internal/clients/postgres"
	"github.com/DevBlobs/go-rest-api-starter/internal/platform/validator"
)

type Module struct {
	Handler *Handler
}

func New(db *postgres.Client, vld validator.Validator) *Module {
	repo := NewRepository(db)
	svc := NewService(repo)
	handler := NewHandler(svc, vld)
	return &Module{Handler: handler}
}
