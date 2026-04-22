package app

import (
	"context"
	"errors"
	"fmt"
	"github.com/DevBlobs/go-rest-api-starter/internal/platform/validator"
	"log/slog"
	"net/http"

	"github.com/DevBlobs/go-rest-api-starter/internal/auth"
	"github.com/DevBlobs/go-rest-api-starter/internal/clients/postgres"
	"github.com/DevBlobs/go-rest-api-starter/internal/clients/workos"
	"github.com/DevBlobs/go-rest-api-starter/internal/demo"
	"github.com/DevBlobs/go-rest-api-starter/internal/items"
	"github.com/DevBlobs/go-rest-api-starter/internal/shared/tz"
	"github.com/DevBlobs/go-rest-api-starter/internal/users"
	"github.com/DevBlobs/go-rest-api-starter/openapi/spec"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	httpSwagger "github.com/swaggo/http-swagger"
)

type App struct {
	Echo  *echo.Echo
	Close func(ctx context.Context) error
}

type ExternalDeps struct {
	WorkOS    workos.Client
	WorkOSCfg *workos.Config
}

func newEcho(allowedOrigins []string) *echo.Echo {
	e := echo.New()
	e.Use(middleware.Recover())
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins:     allowedOrigins,
		AllowMethods:     []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete, http.MethodOptions},
		AllowHeaders:     []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, echo.HeaderAuthorization},
		AllowCredentials: true,
	}))
	e.Use(requestLoggerMiddleware())
	return e
}

func requestLoggerMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			req := c.Request()
			res := c.Response()

			err := next(c)

			slog.Info("request",
				"method", req.Method,
				"path", req.URL.Path,
				"status", res.Status,
			)
			return err
		}
	}
}

func registerHealthAndDocs(e *echo.Echo) {
	e.GET("/health", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
	})

	e.GET("/docs/openapi.yaml", func(c echo.Context) error {
		return c.Blob(http.StatusOK, "application/yaml", spec.OpenAPIYAML)
	})

	e.GET("/docs/*", echo.WrapHandler(
		httpSwagger.Handler(
			httpSwagger.URL("/docs/openapi.yaml"),
		),
	))
}

func registerAuth(api *echo.Group, authHandler *auth.Handler) {
	authGroup := api.Group("/auth")
	authHandler.RegisterPublic(authGroup)
}

func registerProtectedAuth(api *echo.Group, authHandler *auth.Handler) {
	authGroup := api.Group("/auth")
	authHandler.RegisterProtected(authGroup)
}

func NewApp(ctx context.Context, deps *ExternalDeps) (*App, error) {
	if deps == nil {
		return nil, errors.New("missing external dependencies")
	}

	// 1) load configs from env
	pgCfg, err := postgres.LoadFromEnv()
	if err != nil {
		return nil, fmt.Errorf("failed to load postgres config: %w", err)
	}

	authCfg, err := auth.LoadFromEnv()
	if err != nil {
		return nil, fmt.Errorf("failed to load auth config: %w", err)
	}

	tzCfg, err := tz.LoadFromEnv()
	if err != nil {
		return nil, fmt.Errorf("failed to load timezone config: %w", err)
	}

	if err := tz.Init(tzCfg); err != nil {
		return nil, fmt.Errorf("failed to init timezone: %w", err)
	}

	// 2) HTTP server
	e := newEcho(authCfg.AllowedOrigins)
	registerHealthAndDocs(e)

	api := e.Group("/api/v1")

	// 3) DB
	pgPool, err := postgres.New(ctx, pgCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create postgres client: %w", err)
	}

	// 4) Initialize modules
	vld := validator.New()

	usersModule := users.New(pgPool)

	authModule, err := auth.New(auth.ModuleConfig{
		Provider:        deps.WorkOS,
		RedirectURL:     authCfg.RedirectURL,
		BaseURL:         authCfg.BaseURL,
		Domain:          authCfg.Domain,
		StateSecret:     authCfg.StateSecret,
		WorkOSClientID:  deps.WorkOSCfg.ClientID,
		WorkOSIssuers:   deps.WorkOSCfg.Issuers,
		WorkOSNamespace: deps.WorkOSCfg.Namespace,
		UserService:     usersModule.Service,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create auth module: %w", err)
	}

	itemsModule := items.New(pgPool, vld)
	demoModule := demo.New()

	// 5) Register routes
	registerAuth(api, authModule.Handler)

	protected := api.Group("", authModule.Middleware.RequireAuth)
	registerProtectedAuth(protected, authModule.Handler)

	itemsModule.Handler.RegisterRoutes(protected)
	demoModule.Handler.RegisterRoutes(protected.Group("/demo"))

	// 6) App assembly
	app := &App{
		Echo: e,
		Close: func(ctx context.Context) error {
			pgPool.Close()
			return nil
		},
	}

	return app, nil
}

func BuildExternalDeps() (*ExternalDeps, error) {
	workOSCfg, err := workos.LoadFromEnv()
	if err != nil {
		return nil, fmt.Errorf("failed to load workos config: %w", err)
	}
	workOSClient := workos.New(workOSCfg)

	return &ExternalDeps{
		WorkOS:    workOSClient,
		WorkOSCfg: workOSCfg,
	}, nil
}
