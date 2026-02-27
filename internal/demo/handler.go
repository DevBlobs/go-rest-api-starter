package demo

import (
	"log/slog"
	"net/http"

	"github.com/boilerplate-api/go-rest-api-starter/internal/auth"
	"github.com/labstack/echo/v4"
)

type Handler struct{}

func NewHandler() *Handler {
	return &Handler{}
}

func (h *Handler) RegisterRoutes(g *echo.Group) {
	g.GET("/items", h.ListItems, auth.RequireScope("read:items")) // TODO RequireScope smells strange
	g.POST("/items", h.CreateItem, auth.RequireScope("write:items"))
	g.GET("/admin/stats", h.AdminStats, auth.RequireScope("admin"))
}

func (h *Handler) ListItems(c echo.Context) error {
	principal, _ := auth.GetPrincipal(c)
	slog.Info("serving mock items", "principal", principal.Subject, "type", principal.Type)

	return c.JSON(http.StatusOK, map[string]any{
		"items": []map[string]string{
			{"id": "1", "name": "Mock Item 1"},
			{"id": "2", "name": "Mock Item 2"},
		},
		"principal": map[string]any{
			"type":   principal.Type,
			"org_id": principal.OrgID,
		},
	})
}

func (h *Handler) CreateItem(c echo.Context) error {
	principal, _ := auth.GetPrincipal(c)
	slog.Info("creating mock item", "principal", principal.Subject, "type", principal.Type)

	return c.JSON(http.StatusCreated, map[string]any{
		"id":      "3",
		"name":    "New Mock Item",
		"message": "Item created (mock)",
		"principal": map[string]any{
			"type":   principal.Type,
			"org_id": principal.OrgID,
		},
	})
}

func (h *Handler) AdminStats(c echo.Context) error {
	principal, _ := auth.GetPrincipal(c)
	slog.Info("serving admin stats", "principal", principal.Subject, "type", principal.Type)

	return c.JSON(http.StatusOK, map[string]any{
		"total_users": 42,
		"total_items": 1337,
		"principal": map[string]any{
			"type":   principal.Type,
			"org_id": principal.OrgID,
		},
	})
}
