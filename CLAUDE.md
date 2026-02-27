# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Boilerplate REST API is a **production-ready REST API starter** built with Go 1.26. It demonstrates clean architecture patterns with WorkOS authentication, PostgreSQL database, and comprehensive testing infrastructure.

> **IMPORTANT**: This is a pure REST API designed to work with **any HTTP client**. It is not tied to any specific frontend framework or technology.
>
> The complete API contract is defined in [OpenAPI 3.0 format](openapi/spec/openapi.yaml) and can be used to generate client SDKs in any language via [OpenAPI Generator](https://openapi-generator.tech).

**Tech Stack:**
- Go 1.26, Echo web framework
- PostgreSQL 17 with pgx/v5 driver
- WorkOS for authentication (JWT with JWKS)
- golang-migrate for database migrations
- Ginkgo/Gomega for BDD-style testing
- testcontainers for integration tests

## Common Commands

```bash
# Development (Docker Compose - recommended)
make up                  # Start API + Postgres + migrate

# Database
make migrate            # Run migrations up
make migrate-down       # Roll back last migration
make db-shell           # Open psql shell

# Code quality
make test               # Run all tests
make lint               # Run golangci-lint
make build              # Build API binary
```

## High-Level Architecture

### Application Bootstrap

The application follows a layered architecture with explicit dependency injection:

1. **`cmd/api/main.go`**: Entry point - initializes logger, loads `.env`, builds external dependencies
2. **`internal/app/app.go`**: `NewApp()` orchestrates all components:
   - Loads configs via `envconfig` (struct tags bind env vars)
   - Creates external clients (WorkOS)
   - Initializes auth (provider + JWT validator + middleware)
   - Creates DB connection pool (pgx)
   - Instantiates repositories, services, handlers
   - Registers HTTP routes

3. **`BuildExternalDeps()`**: Factory for external service clients - these are the only dependencies that interface with outside systems

### Layer Pattern

```
Handler (HTTP) -> Service (business logic) -> Repository (data access)
```

- **Handlers**: Echo route handlers, decode requests, call services, encode responses
- **Services**: Business logic, validation, orchestrate repositories and external clients
- **Repositories**: Database operations using pgx

### Domain Package Structure

Each domain (e.g., items) contains:
- `model.go` - Data models with JSON tags
- `repository.go` - Data access interface and implementation
- `service.go` - Business logic interface and implementation
- `handler.go` - HTTP handlers with route registration

### Authentication Flow

1. User redirected to WorkOS login
2. WorkOS redirects back with auth code
3. API exchanges code for JWT tokens via WorkOS client
4. `auth.Middleware.RequireAuth()` validates JWT using JWKS
5. Protected routes extract user from JWT `sub` claim

### Users and IDP Sync

Users synced on-demand from WorkOS:
- `users` table with `work_os_id UNIQUE` constraint
- On auth, extract WorkOS user ID from JWT
- Fetch by `work_os_id` - if not found, query IDP and insert
- Idempotent using `ON CONFLICT DO NOTHING`

## Environment Variables

All config loaded via `envconfig` using struct tags:

```go
type Config struct {
    Port int `envconfig:"PORT" default:"8080"`
    APIKey string `envconfig:"API_KEY_REQUIRED"`
}
```

Use `.env.example` as template. Required variables are validated at config load time.

## Code Style Guidelines

### 1. Constants for Magic Strings

**Rule:** Replace all magic strings with named constants.

**Clarification:** Prefer **typed string constants** for domain values (modes, statuses, types). Use untyped constants only for truly local, non-domain values.

```go
// Good - typed constants for domain values
type ImportMode string

const (
    ModeYardArrivalDate ImportMode = "yard-date"
    ModeS3Import        ImportMode = "s3"
)

// Bad - magic strings
switch *mode {
case "yard-date":
    // ...
}
```

### 2. Function Naming Convention

**Rule:** Use the pattern `<Verb><What>[From<Source>][To<Destination>][As<Format>][For<Purpose>][By<Method>][If<Condition>]`.

**Preferred Verbs (Go-idiomatic):** `Load`, `Parse`, `Init`, `Fetch`, `Import`, `Update`, `Save`, `Delete`, `Validate`, `Build`, `Run`, `Ensure`, `Must`

```go
// Good - clear, specific names
func buildChassisToProductMap(products []Product) map[string]string
func loadProductsFromCSV(path string) ([]Product, error)

// Bad - unclear purpose
func handle(data []byte) error
func get(id string) (*Item, error)  // unless truly side-effect-free
```

### 3. Map and Index Naming Consistency

**Rule:** Use consistent naming for maps and indexes.

- Use `indexBy<Key>` when lookup by key is the primary intent
- Use `map<From>To<To>` when expressing a relationship

```go
// Good - index for lookup
func indexProductsByChassis(products []Product) map[string]*Product

// Good - map expressing relationship
func mapChassisToProductID(products []Product) map[string]string
```

### 4. Helper Functions Over Comments

**Rule:** Avoid inline comments describing logic blocks - extract them into named helper functions instead.

```go
// Bad - comments explaining what code does
func processImage(data []byte) ([]byte, error) {
    // Decode image
    img, format, err := image.Decode(bytes.NewReader(data))
    // ...
}

// Good - helper functions with clear names
func processImage(data []byte) ([]byte, error) {
    img, _, err := s.decodeImage(data)
    // ...
}
```

### 5. Error Style

**Rule:** Follow Go error wrapping conventions.

- Always wrap errors with context using `%w`
- Error messages: lower-case, no trailing period

```go
// Good
func (s *service) decodeImage(data []byte) (image.Image, string, error) {
    img, format, err := image.Decode(bytes.NewReader(data))
    if err != nil {
        return nil, "", fmt.Errorf("decode image: %w", err)
    }
    return img, format, nil
}
```

### 6. Logging Responsibility

**Rule:** Log at service, batch, or command boundaries. Avoid logging inside low-level helpers unless unavoidable.

```go
// Good - log at service level
func (s *service) ImportPhotos(ctx context.Context) error {
    slog.Info("starting photo import", "mode", s.mode)
    // ... do work ...
    slog.Info("photo import completed", "success", successCount)
    return nil
}
```

### 7. When to Use Comments

Comments are appropriate for:
- **Package documentation** - Explain the purpose of a package
- **Exported function documentation** - Godoc comments for public APIs
- **Complex algorithms** - Explain non-obvious logic
- **Workarounds** - Explain why a non-standard approach was taken

## CI/CD

- **Lint**: `golangci-lint` on PRs (config in `.github/workflows/lint.yml`)
- **Test**: `make test` on PRs (Go 1.26)

## Database

- **Connection**: Supports `DATABASE_URL` or individual `DB_*` variables
- **Migrations**: `golang-migrate` via Docker Compose
- **Driver**: `pgx/v5` with connection pooling
