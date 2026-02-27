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

	// 2) Auth provider
	authProvider := auth.NewProvider(deps.WorkOS)

	// 3) HTTP server
	e := newEcho(authCfg.AllowedOrigins)
	registerHealthAndDocs(e)

	api := e.Group("/api/v1")

	// 4) DB
	pgPool, err := postgres.New(ctx, pgCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create postgres client: %w", err)
	}

	// 5) Repositories
	itemsRepo := items.NewRepository(pgPool)
	usersRepo := users.NewRepository(pgPool)

	// 6) Services
	itemsService := items.NewService(itemsRepo)
	usersService := users.NewService(usersRepo)

	// 7) Auth
	authService := auth.NewService(authProvider, authCfg.RedirectURL, usersService)
	jwtValidator, err := auth.NewJWTValidator(authService.JWKS())
	if err != nil {
		return nil, fmt.Errorf("failed to create JWT validator: %w", err)
	}

	cookieCfg := auth.CookieCfg{Domain: authCfg.Domain}

	authHandler := &auth.Handler{
		Svc:         authService,
		Cookie:      cookieCfg,
		BaseURL:     authCfg.BaseURL,
		StateSecret: authCfg.StateSecret,
	}

	authMiddleware := &auth.Middleware{
		Svc:            authService,
		Validator:      jwtValidator,
		Cookie:         cookieCfg,
		ClientID:       deps.WorkOSCfg.ClientID,
		AllowedIssuers: deps.WorkOSCfg.Issuers,
		Namespace:      deps.WorkOSCfg.Namespace,
	}

	registerAuth(api, authHandler)

	// 8) Handlers
	vld := validator.New()
	protected := api.Group("", authMiddleware.RequireAuth)

	registerProtectedAuth(protected, authHandler)

	itemsHandler := items.NewHandler(itemsService, vld)
	itemsHandler.RegisterRoutes(protected)

	demoHandler := demo.NewHandler()
	demoHandler.RegisterRoutes(protected.Group("/demo"))

	// 9) App assembly
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
