package auth

import "github.com/labstack/echo/v4"

type contextKey string

const PrincipalContextKey contextKey = "principal"

type Principal struct {
	Subject string
	Type    string
	Email   string
	OrgID   string
	Scopes  []string
	Claims  map[string]any
}

func GetPrincipal(c echo.Context) (*Principal, bool) {
	p, ok := c.Get(string(PrincipalContextKey)).(*Principal)
	return p, ok
}
