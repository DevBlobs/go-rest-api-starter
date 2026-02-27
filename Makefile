.PHONY: up lint lint-dir test build migrate migrate-down migrate-force migrate-create db-shell help

up:
	docker compose up

# Run the linter on the entire project
lint:
	golangci-lint run ./...

# Run the linter on a specific directory
lint-dir:
	golangci-lint run $(DIR)/...

# Run all tests
test:
	go test -v ./...

# Build the API binary
build:
	go build -o api ./cmd/api

DB_URL=postgres://boilerplate:boilerplate@db:5432/boilerplate?sslmode=disable

migrate:
	docker compose run --rm migrate

migrate-down:
	docker compose run --rm migrate \
		-path /migrations \
		-database "$(DB_URL)" \
		down 1

migrate-version:
	docker compose run --rm migrate version

db-shell:
	docker compose exec db psql -U boilerplate -d boilerplate

# Help
help:
	@echo "Available targets:"
	@echo "  make up             - Start the docker-compose stack (API + Postgres + migrate)"
	@echo "  make lint           - Run the linter on the entire project"
	@echo "  make lint-dir DIR=path/to/dir - Run the linter on a specific directory"
	@echo "  make test           - Run all tests"
	@echo "  make build          - Build the API binary (./api)"
	@echo "  make migrate        - Run 'migrate up' using the migrate service"
	@echo "  make migrate-down   - Run 'migrate down 1' using the migrate service"
	@echo "  make migrate-version - Show current migration version"
	@echo "  make db-shell       - Open psql shell inside docker Postgres"
	@echo "  make help           - Show this help message"
