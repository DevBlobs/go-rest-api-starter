// testsuite/testapp.go
package testsuite

import (
	"context"
	"errors"
	"fmt"
	pgclient "github.com/boilerplate-api/go-rest-api-starter/internal/clients/postgres"
	"github.com/boilerplate-api/go-rest-api-starter/internal/clients/workos"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"sync"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"

	appPkg "github.com/boilerplate-api/go-rest-api-starter/internal/app"
)

var (
	once    sync.Once
	echoApp *echo.Echo
)

func projectRoot() string {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		log.Fatal("failed to get current file path")
	}

	root := filepath.Dir(filepath.Dir(filepath.Dir(filename)))
	abs, err := filepath.Abs(root)
	if err != nil {
		log.Fatalf("failed to get absolute project root: %v", err)
	}
	return abs
}

func GetTestApp() *echo.Echo {
	once.Do(func() {
		ctx := context.Background()

		dbName := "testdb"
		dbUser := "user"
		dbPassword := "password"

		postgresContainer, err := postgres.Run(ctx,
			"postgres:16-alpine",
			postgres.WithDatabase(dbName),
			postgres.WithUsername(dbUser),
			postgres.WithPassword(dbPassword),
			postgres.BasicWaitStrategies(),
		)
		if err != nil {
			log.Fatalf("failed to start postgres container: %v", err)
		}

		defer func() {
			if err != nil {
				_ = testcontainers.TerminateContainer(postgresContainer)
			}
		}()

		host, _ := postgresContainer.Host(ctx)
		port, _ := postgresContainer.MappedPort(ctx, "5432/tcp")

		root := projectRoot()

		envPath := filepath.Join(root, ".env.test")
		if err := godotenv.Load(envPath); err != nil && !os.IsNotExist(err) {
			log.Printf("warning: could not load .env.test: %v", err)
		}

		setEnv("DB_HOST", host)
		setEnv("DB_PORT", port.Port())
		setEnv("DB_USER", dbUser)
		setEnv("DB_PASSWORD", dbPassword)
		setEnv("DB_NAME", dbName)
		setEnv("DB_SSLMODE", "disable")

		migrationsPath := filepath.Join(root, "migrations")
		dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", dbUser, dbPassword, host, port.Port(), dbName)

		migrateSource := "file://" + migrationsPath
		runMigrations(migrateSource, dsn)

		// Use test mocks for all external dependencies
		wk := buildExternalTestDeps()

		app, err := appPkg.NewApp(ctx, &appPkg.ExternalDeps{
			WorkOS: wk,
			WorkOSCfg: &workos.Config{
				ClientID:  "test_client_id",
				Issuers:   []string{"https://api.workos.com/user_management/client_test_client_id"},
				Namespace: "https://api.workos.com/claims/user",
			},
		})
		if err != nil {
			log.Fatalf("failed to initialize test app: %v", err)
		}
		echoApp = app.Echo
	})

	return echoApp
}

func setEnv(key, value string) {
	if err := os.Setenv(key, value); err != nil {
		log.Fatalf("failed to set env %s=%s: %v", key, value, err)
	}
}

func runMigrations(sourceURL, dsn string) {
	m, err := migrate.New(sourceURL, dsn)
	if err != nil {
		log.Fatalf("failed to create migrator: %v", err)
	}
	defer m.Close()

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		log.Fatalf("failed to apply migrations: %v", err)
	}
}

// GetPGConfig returns the postgres config from environment
func GetPGConfig() (*pgclient.Config, error) {
	return pgclient.LoadFromEnv()
}

// NewPGClient creates a new postgres client
func NewPGClient(ctx context.Context, cfg *pgclient.Config) (*pgclient.Client, error) {
	return pgclient.New(ctx, cfg)
}
