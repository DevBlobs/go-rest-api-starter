package demo

type Module struct {
	Handler *Handler
}

func New() *Module {
	return &Module{Handler: NewHandler()}
}
