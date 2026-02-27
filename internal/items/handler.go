package items

import (
	"net/http"

	"github.com/boilerplate-api/go-rest-api-starter/internal/platform/validator"
	"github.com/labstack/echo/v4"
)

type Handler struct {
	svc       Service
	validator validator.Validator
}

func NewHandler(svc Service) *Handler {
	return &Handler{
		svc:       svc,
		validator: validator.NewValidator(),
	}
}

func (h *Handler) RegisterRoutes(g *echo.Group) {
	g.POST("/items", h.CreateItem)
	g.GET("/items", h.ListItems)
	g.GET("/items/:id", h.GetItem)
	g.PATCH("/items/:id", h.UpdateItem)
	g.DELETE("/items/:id", h.DeleteItem)
}

func (h *Handler) CreateItem(c echo.Context) error {
	var req CreateItemRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}
	if err := h.validator.Struct(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	item, err := h.svc.CreateItem(c.Request().Context(), req)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to create item")
	}
	return c.JSON(http.StatusCreated, item)
}

func (h *Handler) ListItems(c echo.Context) error {
	items, err := h.svc.ListItems(c.Request().Context())
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to list items")
	}
	return c.JSON(http.StatusOK, items)
}

func (h *Handler) GetItem(c echo.Context) error {
	id := c.Param("id")
	if id == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "missing id")
	}

	item, err := h.svc.GetItem(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get item")
	}
	if item == nil {
		return echo.NewHTTPError(http.StatusNotFound, "item not found")
	}
	return c.JSON(http.StatusOK, item)
}

func (h *Handler) UpdateItem(c echo.Context) error {
	id := c.Param("id")
	if id == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "missing id")
	}

	var req UpdateItemRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}
	if err := h.validator.Struct(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	item, err := h.svc.UpdateItem(c.Request().Context(), id, req)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to update item")
	}
	if item == nil {
		return echo.NewHTTPError(http.StatusNotFound, "item not found")
	}
	return c.JSON(http.StatusOK, item)
}

func (h *Handler) DeleteItem(c echo.Context) error {
	id := c.Param("id")
	if id == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "missing id")
	}

	if err := h.svc.DeleteItem(c.Request().Context(), id); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to delete item")
	}
	return c.NoContent(http.StatusNoContent)
}
