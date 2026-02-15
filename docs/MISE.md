# Using Mise with Todo API

This project supports [mise](https://mise.jdx.dev/) as an alternative to Make for task running and tool version management.

## Why Mise?

- **Unified tool management**: Automatically installs and manages Go versions
- **Better task syntax**: TOML-based configuration is more readable
- **Environment management**: Built-in support for `.env` files
- **Cross-platform**: Works consistently across Linux, macOS, and Windows
- **Task dependencies**: Clear dependency declarations between tasks

## Installation

### macOS/Linux

```bash
# Using curl
curl https://mise.jdx.dev/install.sh | sh

# Using Homebrew
brew install mise

# Activate mise in your shell
echo 'eval "$(mise activate bash)"' >> ~/.bashrc  # for bash
echo 'eval "$(mise activate zsh)"' >> ~/.zshrc    # for zsh
```

### Verification

```bash
mise --version
```

## Quick Start

### 1. Install Tools

Mise will automatically install the correct Go version when you enter the project directory:

```bash
mise install  # Installs lates Go as specified in mise.toml and all needed Go tools
```

### 2. Configure Environment

```bash
# Copy example environment file
cp .env.example .env

# Edit .env with your configuration
emacs .env
```

Mise will automatically load variables from `.env` when running tasks.

### 3. View Available Tasks

```bash
# List all available tasks
mise tasks

# Or using the help task
mise run help
```

## Common Tasks

### Development

```bash
# Generate protobuf code
mise run generate

# Run the Go application
mise run run
```

### Building

```bash
# Build for current platform
mise run build

# Build for Linux
mise run build-linux

# Build for macOS (both architectures)
mise run build-darwin

# Clean build artifacts
mise run clean
```

### Testing

```bash
# Run unit tests
mise run test

# Run tests with coverage report
mise run test-coverage

# Run integration tests (requires Docker)
mise run test-integration

# Run all tests
mise run test-all
```

### Code Quality

```bash
# Format code
mise run fmt

# Run linter
mise run lint

# Run go vet
mise run vet

# Run all checks
mise run check
```

### Database

```bash
# Run migrations up
mise run db-migrate-up

# Run migrations down
mise run db-migrate-down

# Force specific migration version
VERSION=5 mise run db-migrate-force

# Seed database
mise run db-seed

# Reset database (down, up, seed)
mise run db-reset
```

### Docker

```bash
# Build Docker image
mise run docker-build

# Start services
mise run docker-up

# Stop services
mise run docker-down

# View logs
mise run docker-logs
```

## Task Dependencies

Mise automatically handles task dependencies. For example:

```bash
# Running 'build' automatically runs 'generate' first
mise run build

# Running 'test-coverage' automatically runs 'test' first
mise run test-coverage

# Running 'check' automatically runs fmt, vet, lint, and buf-lint
mise run check
```

## Environment Variables

Mise loads environment variables in this order:

1. System environment
2. `.env` file (if present)
3. Variables defined in `mise.toml`

### Predefined Variables

The following variables are set in `mise.toml`:

- `BINARY_NAME=todo`
- `BINARY_PATH=./cmd/todo`
- `BUILD_DIR=bin`

### Custom Variables

Add custom variables to `.env`:

```bash
# .env
DATABASE_URL=postgres://localhost:5435/mydb
PORT=9000
ENVIRONMENT=production
```

## Comparing with Make

If you're familiar with Make, here's how the commands translate:

| Make | Mise |
|------|------|
| `make help` | `mise tasks` or `mise run help` |
| `make build` | `mise run build` |
| `make test` | `mise run test` |
| `make clean` | `mise run clean` |
| `make run` | `mise run run` |
| `make docker-up` | `mise run docker-up` |

## Advanced Features

### Watch Mode

You can use mise with watch tools for automatic rebuilds:

```bash
# Using mise with entr for auto-reload
ls **/*.go | entr -r mise run run

# Or use the provided run-dev task with air
mise run run-dev
```

### Task Aliases

The file `mise.toml` define some shell aliases:

```toml
[shell_alias]
dev = "mise run run"
build = "mise run build"
deploy = "mise run deploy"
mr = "mise run"
mt = "mise tasks"
```

### Custom Tasks

Add your own tasks to `mise.toml`:

```toml
[tasks.my-task]
description = "My custom task"
run = """
echo "Running custom task"
go run ./scripts/my-script.go
"""
```

### Parallel Tasks

Mise can run independent tasks in parallel:

```bash
# This will run lint and test in parallel
mise run lint & mise run test & wait
```

## Troubleshooting

### Go Version Issues

```bash
# Check current Go version
go version

# Reinstall Go via mise
mise install go@latest

# Use specific Go version
mise use go@1.24
```

### Environment Variables Not Loading

```bash
# Check if .env exists
ls -la .env

# Verify mise is loading .env
mise env

# Debug environment
mise run --verbose task-name
```

### Task Not Found

```bash
# List all tasks
mise tasks

# Check mise.toml syntax
mise doctor
```

## Resources

- [Mise Documentation](https://mise.jdx.dev/)
- [Mise GitHub](https://github.com/jdx/mise)
- [Task Runner Guide](https://mise.jdx.dev/tasks/)
- [Configuration Reference](https://mise.jdx.dev/configuration.html)
