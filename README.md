# Boilerplate REST API

A production-ready REST API boilerplate built with Go 1.26, Echo web framework, and PostgreSQL.

> This is a clean, frontend-agnostic REST API starter designed to work with **any HTTP client**. The complete API contract is defined in [OpenAPI 3.0 format](openapi/spec/openapi.yaml).

## Features

- **Authentication**: WorkOS integration with JWT (JWKS) validation
  - User authentication via authorization_code flow (browser-based)
  - Machine-to-machine (M2M) authentication via client_credentials flow
  - Scope-based authorization for fine-grained access control
- **Database**: PostgreSQL 17 with pgx/v5 driver and connection pooling
- **Architecture**: Clean layered design (Handler → Service → Repository)
- **Testing**: Integration tests with testcontainers, BDD-style with Ginkgo/Gomega
- **Migrations**: Database migrations via golang-migrate
- **API Docs**: OpenAPI 3.0 specification with YAML endpoint

## Quick Start

```bash
# Clone and start the development stack
git clone <your-repo-url>
cd go-rest-api-starter
make up

# Run migrations
make migrate

# Run tests
make test

# Build the binary
make build
```

## Project Structure

```
├── cmd/api/              # Application entry point
├── internal/
│   ├── app/              # Application bootstrap and DI wiring
│   ├── auth/             # WorkOS authentication (JWT validation, middleware)
│   ├── clients/          # External service clients
│   │   ├── postgres/     # PostgreSQL client wrapper
│   │   └── workos/       # WorkOS OAuth client
│   ├── demo/             # Demo endpoints showing scope enforcement
│   ├── items/            # Example domain (Item CRUD)
│   │   ├── model.go      # Data models
│   │   ├── repository.go # Data access layer
│   │   ├── service.go    # Business logic
│   │   └── handler.go    # HTTP handlers
│   ├── platform/         # Platform utilities (logger, validator)
│   └── users/            # User management (synced from WorkOS)
├── migrations/           # Database migrations
├── tests/                # Integration tests
│   ├── testsuite/        # Test harness and mocks
│   └── internal/         # Domain-specific tests
├── openapi/spec/         # OpenAPI specification
└── docker-compose.yml    # Development stack
```

## Architecture

### Layer Pattern

```
Handler (HTTP) → Service (business logic) → Repository (data access)
```

- **Handlers**: Echo route handlers that decode requests, call services, encode responses
- **Services**: Business logic, validation, orchestrate repositories
- **Repositories**: Database operations using pgx

### Authentication Flow

This API supports two authentication flows:

#### User Authentication (Authorization Code Flow)

Browser-based authentication for interactive users:

1. User visits `/api/v1/auth/login` → redirected to WorkOS
2. User authenticates with WorkOS
3. WorkOS redirects back to `/api/v1/auth/callback` with auth code
4. API exchanges code for JWT tokens via WorkOS client
5. Access token stored in httpOnly cookie
6. `auth.Middleware.RequireAuth()` validates JWT using JWKS
7. Protected routes extract `Principal` from JWT claims

```bash
# Example: Login via browser
open http://localhost:8080/api/v1/auth/login

# After callback, access protected endpoints (cookie sent automatically)
curl http://localhost:8080/api/v1/demo/items
```

#### M2M Authentication (Client Credentials Flow)

Service-to-service authentication for background jobs, APIs, and CLIs:

1. M2M application calls WorkOS `/oauth2/token` directly
2. WorkOS returns JWT access token with `org_id` claim
3. M2M application includes token in `Authorization: Bearer <token>` header
4. Backend validates JWT signature, iss, aud, and exp claims
5. Backend extracts `Principal` from JWT claims (type="service")
6. Scope-based authorization enforced via `auth.RequireScope()`

```bash
# Step 1: Obtain token from WorkOS
AUTHKIT_DOMAIN="your-app.authkit.app"
M2M_CLIENT_ID="client_xxx"
M2M_CLIENT_SECRET="sk_xxx"

ACCESS_TOKEN=$(curl -s "https://${AUTHKIT_DOMAIN}/oauth2/token" \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "grant_type=client_credentials" \
  -d "client_id=${M2M_CLIENT_ID}" \
  -d "client_secret=${M2M_CLIENT_SECRET}" \
  -d "scope=read:items write:items" | jq -r '.access_token')

# Step 2: Call API with Bearer token
curl -H "Authorization: Bearer ${ACCESS_TOKEN}" \
  http://localhost:8080/api/v1/demo/items
```

### Token Validation

The backend validates JWT tokens in this order:

1. **Signature** - Verified against WorkOS JWKS endpoint
2. **Expiration** - `exp` claim must be in the future
3. **Issuer** - `iss` claim must match configured issuer exactly
4. **Audience** - `aud` claim must match `ClientID` exactly (required for M2M)

No prefix matching, no partial validation.

### Scope-Based Authorization

Endpoints are protected with scope requirements:

```go
// In handler registration
g.GET("/items", h.ListItems, auth.RequireScope("read:items"))
g.POST("/items", h.CreateItem, auth.RequireScope("write:items"))
g.GET("/admin/stats", h.AdminStats, auth.RequireScope("admin"))
```

- Scopes are extracted from the `scope` claim (space-separated string)
- `RequireScope()` middleware returns 403 if scope is missing
- User tokens typically have broad scopes (e.g., from WorkOS session)
- M2M tokens have limited, explicitly requested scopes

### Demo Protected Endpoints

The `/api/v1/demo/*` endpoints demonstrate scope enforcement:

| Endpoint | Method | Required Scope |
|----------|--------|----------------|
| `/api/v1/demo/items` | GET | `read:items` |
| `/api/v1/demo/items` | POST | `write:items` |
| `/api/v1/demo/admin/stats` | GET | `admin` |

```bash
# Test scope enforcement (M2M token with limited scope)
curl -H "Authorization: Bearer ${READ_ONLY_TOKEN}" \
  -X POST http://localhost:8080/api/v1/demo/items
# Returns: 403 Forbidden - insufficient scope
```

### Domain Pattern

Each new domain follows this structure:

```go
// internal/<domain>/model.go
type Item struct {
    ID        string    `json:"id"`
    Name      string    `json:"name"`
    CreatedAt time.Time `json:"createdAt"`
}

// internal/<domain>/repository.go
type Repository interface {
    Create(ctx context.Context, item Item) (Item, error)
    List(ctx context.Context) ([]Item, error)
    // ...
}

// internal/<domain>/service.go
type Service interface {
    CreateItem(ctx context.Context, req CreateItemRequest) (*Item, error)
    // ...
}

// internal/<domain>/handler.go
func (h *Handler) RegisterRoutes(g *echo.Group) {
    g.POST("/items", h.CreateItem)
    // ...
}
```

## Environment Variables

Create a `.env` file from `.env.example`:

```bash
# WorkOS Authentication
WORKOS_CLIENT_ID=your_workos_client_id
WORKOS_API_KEY=your_workos_api_key

# Auth Configuration
AUTH_REDIRECT_URL=http://localhost:8080/api/v1/auth/callback
AUTH_DOMAIN=localhost
AUTH_BASE_URL=http://localhost:3000
AUTH_ALLOWED_ORIGINS=http://localhost:3000

# Database
DB_HOST=localhost
DB_PORT=5432
DB_USER=boilerplate
DB_PASSWORD=boilerplate
DB_NAME=boilerplate
DB_SSLMODE=disable
```

## Available Commands

| Command | Description |
|---------|-------------|
| `make up` | Start Docker Compose stack (API + Postgres + migrate) |
| `make build` | Build the API binary |
| `make test` | Run all tests |
| `make lint` | Run golangci-lint |
| `make migrate` | Run database migrations |
| `make migrate-down` | Rollback last migration |
| `make db-shell` | Open psql shell |

## Example API Endpoints

### Authentication
- `GET /api/v1/auth/login` - Start WorkOS login flow
- `GET /api/v1/auth/callback` - OAuth callback
- `POST /api/v1/auth/logout` - Logout
- `GET /api/v1/auth/me` - Get current user (protected)

### Demo (Scope-Based Authorization)
- `GET /api/v1/demo/items` - List mock items (requires `read:items` scope)
- `POST /api/v1/demo/items` - Create mock item (requires `write:items` scope)
- `GET /api/v1/demo/admin/stats` - Get admin stats (requires `admin` scope)

### Items (Example Domain)
- `POST /api/v1/items` - Create item (protected)
- `GET /api/v1/items` - List items (protected)
- `GET /api/v1/items/{id}` - Get item by ID (protected)
- `DELETE /api/v1/items/{id}` - Delete item (protected)

### Documentation
- `GET /health` - Health check
- `GET /docs/openapi.yaml` - OpenAPI specification (YAML)

## Testing

Tests use Ginkgo/Gomega for BDD-style testing and testcontainers for database isolation.

```bash
# Run all tests
make test

# Run specific test suite
go test ./tests/internal/items -v

# Run tests with coverage
go test ./... -cover
```

## Creating a New Domain

1. Create domain directory: `internal/<domain>/`
2. Add `model.go`, `repository.go`, `service.go`, `handler.go`
3. Create migration in `migrations/`
4. Wire up in `internal/app/app.go`:
   - Add repository initialization
   - Add service initialization
   - Add handler and routes
5. Add integration tests in `tests/internal/<domain>/`

## License

MIT
