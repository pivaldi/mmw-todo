# Development Guide

This guide explains how to set up, run, and develop the Todo API application.

## Prerequisites

### Option 1: Manual Installation

- **Go 1.22+** - [Download](https://go.dev/dl/)
- **Docker & Docker Compose** - [Download](https://www.docker.com/get-started)
- **PostgreSQL 18** (for local development without Docker)
- **Buf CLI** (for protobuf generation) - `go install github.com/bufbuild/buf/cmd/buf@latest`
- **migrate** (for database migrations) - `go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest`

### Option 2: Using Mise (Recommended)

[Mise](https://mise.jdx.dev/) automatically manages tool versions and provides a better task runner:

```bash
# Install mise
curl https://mise.jdx.dev/install.sh | sh

# Activate in your shell
echo 'eval "$(mise activate bash)"' >> ~/.bashrc  # or zsh, fish, etc.

# All tools (Go, buf, migrate, etc.) will be auto-installed when needed
```

See [MISE.md](MISE.md) for detailed mise usage.

## Quick Start with Docker Compose

The easiest way to run the full stack:

```bash
# Start all services (PostgreSQL + API)
docker-compose -f deployments/docker-compose.yml up -d

# View logs
docker-compose -f deployments/docker-compose.yml logs -f

# Stop all services
docker-compose -f deployments/docker-compose.yml down
```

The API will be available at `http://localhost:8090`

### With pgAdmin (optional)

```bash
# Start with pgAdmin for database management
docker-compose -f deployments/docker-compose.yml --profile tools up -d

# Access pgAdmin at http://localhost:5050
# Email: admin@todoapp.com
# Password: admin
```

## Local Development (without Docker)

### 1. Start PostgreSQL

```bash
# Using Docker
docker run --name todoapp-postgres \
  -e POSTGRES_DB=todoapp \
  -e POSTGRES_USER=postgres \
  -e POSTGRES_PASSWORD=postgres \
  -p 5432:5435 \
  -d postgres:18-alpine

# Or use a local PostgreSQL installation
```

### 2. Run Database Migrations

```bash
export DATABASE_URL="postgres://postgres:postgres@localhost:5435/todoapp?sslmode=disable"

# Run migrations up
make db-migrate-up

# Or manually
migrate -path ./scripts/migrations -database "$DATABASE_URL" up
```

### 3. Install Dependencies

```bash
# Download Go modules
go mod download

# Install development tools
make install-tools
```

### 4. Generate Protobuf Code

```bash
# Generate Go code from .proto files
make generate

# Or manually
buf generate
```

### 5. Run the Application

```bash
# Using make
make run

# Or directly
go run ./cmd/todo/main.go

# With custom configuration
DATABASE_URL="..." PORT=9000 go run ./cmd/todo/main.go
```

The API will be available at `http://localhost:8090` (or your custom port)

## Configuration

The application is configured via environment variables:

| Variable | Description | Default |
|----------|-------------|---------|
| `DATABASE_URL` | PostgreSQL connection string | `postgres://postgres:postgres@localhost:5435/todoapp?sslmode=disable` |
| `PORT` | HTTP server port | `8090` |
| `ENVIRONMENT` | Environment (development/production) | `development` |

## Testing

### Unit Tests

```bash
# Run all unit tests
make test

# Run with coverage
make test-coverage

# Run specific package tests
go test -v ./internal/domain/todo/...
go test -v ./internal/application/...
go test -v ./internal/adapters/handler/connect/...
```

### Integration Tests

Integration tests require Docker for testcontainers:

```bash
# Run integration tests
make test-integration

# Or directly
go test -tags=integration -v ./internal/adapters/repository/postgres/...
```

### Test Coverage

```bash
# Generate coverage report
make test-coverage

# Open coverage.html in browser to view detailed coverage
```

Current coverage:
- **Domain**: 90.5%
- **Application**: 64.6%
- **Handlers**: 100% (mocked tests)

## Project Structure

```
├── api/                      # API definitions (protobuf)
│   └── todo/v1/
│       └── todo.proto
├── cmd/                      # Application entrypoints
│   └── todo/
│       └── main.go
├── internal/                 # Private application code
│   ├── domain/              # Domain layer (business logic)
│   │   ├── todo.go          # Todo aggregate
│   │   ├── value_objects.go # Value objects
│   │   ├── events.go        # Domain events
│   │   └── errors.go        # Domain errors
│   ├── application/         # Application layer (use cases)
│   │   ├── todo_service.go  # Application service
│   │   └── dto.go           # Data transfer objects
│   ├── ports/               # Port interfaces
│   │   ├── repository.go    # Repository port
│   │   └── events.go        # Event dispatcher port
│   └── adapters/            # Adapter implementations
│       ├── handler/
│       │   └── connect/     # Connect/gRPC handlers
│       ├── repository/
│       │   └── postgres/    # PostgreSQL repository
│       └── events/          # Event dispatcher implementations
├── gen/                     # Generated code (from protobuf)
├── scripts/                 # Scripts and migrations
│   └── migrations/
├── build/                   # Build configurations
│   └── package/
│       └── Dockerfile
├── deployments/             # Deployment configurations
│   └── docker-compose.yml
└── docs/                    # Documentation
```

## Architecture

This project demonstrates **Domain-Driven Design (DDD)** and **Hexagonal Architecture**:

### Layers

1. **Domain Layer** (`internal/domain/todo/`)
   - Pure business logic
   - No external dependencies
   - Aggregates, value objects, domain events
   - Self-validating entities

2. **Application Layer** (`internal/application/`)
   - Use case orchestration
   - Thin layer coordinating domain objects
   - Transaction boundaries
   - Event dispatching

3. **Ports Layer** (`internal/ports/`)
   - Interfaces defining boundaries
   - Repository interfaces (secondary ports)
   - Service interfaces (primary ports)

4. **Adapters Layer** (`internal/adapters/`)
   - Infrastructure implementations
   - Connect/gRPC handlers (driving adapters)
   - PostgreSQL repository (driven adapter)
   - Event dispatchers (driven adapter)

### Dependency Direction

Dependencies flow inward:
```
Adapters → Application → Domain
         ↘   Ports    ↗
```

The domain layer has **zero dependencies** on infrastructure.

## API Usage

### Using Connect Protocol (HTTP)

```bash
# Create a todo
curl -X POST http://localhost:8090/todo.v1.TodoService/CreateTodo \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Buy groceries",
    "description": "Milk, eggs, bread",
    "priority": "PRIORITY_MEDIUM"
  }'

# Get a todo
curl http://localhost:8090/todo.v1.TodoService/GetTodo?id=<uuid>

# List todos
curl http://localhost:8090/todo.v1.TodoService/ListTodos

# Update a todo
curl -X POST http://localhost:8090/todo.v1.TodoService/UpdateTodo \
  -H "Content-Type: application/json" \
  -d '{
    "id": "<uuid>",
    "title": "Updated title"
  }'

# Complete a todo
curl -X POST http://localhost:8090/todo.v1.TodoService/CompleteTodo \
  -H "Content-Type: application/json" \
  -d '{"id": "<uuid>"}'

# Delete a todo
curl -X POST http://localhost:8090/todo.v1.TodoService/DeleteTodo \
  -H "Content-Type: application/json" \
  -d '{"id": "<uuid>"}'
```

### Using gRPC

The same endpoints support native gRPC and gRPC-Web protocols automatically via Connect.

## Database Management

### Migrations

```bash
# Create a new migration
migrate create -ext sql -dir scripts/migrations -seq <migration_name>

# Run migrations
make db-migrate-up
DB_URL="..." make db-migrate-up

# Rollback migrations
make db-migrate-down

# Check migration status
migrate -path ./scripts/migrations -database "$DATABASE_URL" version
```

### Direct Database Access

```bash
# Using psql
psql postgres://postgres:postgres@localhost:5435/todoapp

# Or via Docker
docker exec -it todoapp-postgres psql -U postgres -d todoapp
```

## Code Quality

### Formatting

```bash
# Format all Go code
make fmt

# Format protobuf files
make buf-format
```

### Linting

```bash
# Lint Go code
make lint

# Lint protobuf files
make buf-lint
```

### All Checks

```bash
# Run all checks (format, vet, lint)
make check
```

## Building

### Local Build

```bash
# Build binary
make build

# Binary will be in bin/todo
./bin/todo
```

### Docker Build

```bash
# Build Docker image
docker build -f build/package/Dockerfile -t todo:latest .

# Run container
docker run -p 8090:8090 \
  -e DATABASE_URL="..." \
  todo:latest
```

## Troubleshooting

### Database Connection Issues

```bash
# Check if PostgreSQL is running
docker ps | grep postgres

# Check connection
pg_isready -h localhost -p 5435 -U postgres

# View PostgreSQL logs
docker logs todoapp-postgres
```

### Port Already in Use

```bash
# Find process using port 8090
lsof -i :8090

# Change port
PORT=9000 go run ./cmd/todo/main.go
```

### Migration Errors

```bash
# Force to specific version (use with caution!)
migrate -path ./scripts/migrations -database "$DATABASE_URL" force <version>

# Check dirty state
migrate -path ./scripts/migrations -database "$DATABASE_URL" version
```

## Contributing

1. Create a feature branch
2. Make changes following the existing patterns
3. Run tests and checks: `make test && make check`
4. Submit a pull request

## Resources

- [Connect Documentation](https://connectrpc.com/docs/)
- [Buf Documentation](https://buf.build/docs/)
- [Domain-Driven Design](https://martinfowler.com/bliki/DomainDrivenDesign.html)
- [Hexagonal Architecture](https://alistair.cockburn.us/hexagonal-architecture/)
- [Go Project Layout](https://github.com/golang-standards/project-layout)
