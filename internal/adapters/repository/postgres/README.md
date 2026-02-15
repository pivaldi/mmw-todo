# PostgreSQL Repository Adapter

This package provides a PostgreSQL implementation of the `TodoRepository` port.

## Implementation

The `PostgresTodoRepository` implements the repository pattern, handling the impedance mismatch between the domain model (with value objects and business logic) and the database schema (primitive types).

### Key Features

- Maps domain aggregates to/from database rows
- Returns domain errors (e.g., `ErrTodoNotFound`) not database errors
- Uses `pgx/v5` connection pool for efficient database access
- Supports filtering, pagination, and sorting
- Handles nullable fields properly (e.g., `DueDate`)

## Testing

### Integration Tests

The integration tests use [testcontainers-go](https://golang.testcontainers.org/) to spin up a real PostgreSQL database in a Docker container. This ensures tests run against actual database behavior.

**Requirements:**
- Docker must be installed and running
- Docker daemon must be accessible to the test runner

**Running Integration Tests:**

```bash
# Run with the integration build tag
go test -tags=integration -v ./internal/adapters/repository/postgres/...

# Or use the Makefile target
make test-integration
```

**What the tests cover:**
- All CRUD operations (Save, FindByID, FindAll, Update, Delete)
- Filter functionality (status, priority, limit, offset)
- Error cases (not found, invalid data)
- Domain aggregate reconstitution from database
- Concurrent operations
- Field preservation across save/load cycles

### Test Structure

1. **setupTestDB**: Creates PostgreSQL container and runs migrations
2. **runMigrations**: Executes migration files from `scripts/migrations/`
3. **Test cases**: Comprehensive coverage of all repository methods

### Why Integration Tests?

Unlike unit tests that mock the database, integration tests:
- Catch SQL syntax errors
- Verify migrations work correctly
- Test actual query performance
- Ensure database constraints are enforced
- Validate type mapping between Go and PostgreSQL

### Skipping Integration Tests

If Docker is not available, integration tests are automatically skipped due to the `//go:build integration` build tag. Regular unit tests will still run:

```bash
# Run only unit tests (no Docker required)
go test -v ./internal/adapters/repository/postgres/...
```

## Database Schema

The repository expects the following table structure (created by migrations):

```sql
CREATE TABLE IF NOT EXISTS todos (
    id UUID PRIMARY KEY,
    title VARCHAR(200) NOT NULL,
    description TEXT,
    status VARCHAR(20) NOT NULL,
    priority VARCHAR(20) NOT NULL,
    due_date TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL,

    CONSTRAINT valid_status CHECK (status IN ('pending', 'in_progress', 'completed', 'cancelled')),
    CONSTRAINT valid_priority CHECK (priority IN ('low', 'medium', 'high', 'urgent'))
);
```

## Usage Example

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/jackc/pgx/v5/pgxpool"
    "github.com/pivaldi/mmw/todo/internal/adapters/repository/postgres"
    "github.com/pivaldi/mmw/todo/internal/domain/todo"
)

func main() {
    // Create connection pool
    pool, err := pgxpool.New(context.Background(), "postgres://user:pass@localhost:5435/dbname")
    if err != nil {
        log.Fatal(err)
    }
    defer pool.Close()

    // Create repository
    repo := postgres.NewPostgresTodRepository(pool)

    // Create and save a todo
    title, _ := domain.NewTaskTitle("Buy groceries")
    todo := domain.NewTodo(title, "Milk, eggs, bread", domain.PriorityMedium, nil)

    if err := repo.Save(context.Background(), todo); err != nil {
        log.Fatal(err)
    }

    // Retrieve it
    retrieved, err := repo.FindByID(context.Background(), todo.ID())
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Todo: %s\n", retrieved.Title().String())
}
```

## Performance Considerations

- **Connection Pooling**: Uses `pgxpool` for efficient connection management
- **Prepared Statements**: Queries use parameterized statements (protection against SQL injection)
- **Indexes**: Database schema includes indexes on `status` and `due_date`
- **Batch Operations**: Consider adding batch insert/update methods for high-volume operations

## Future Enhancements

Potential improvements for production use:

1. **Query Builder**: Consider using a query builder library for complex dynamic queries
2. **Batch Operations**: Add `SaveBatch`, `UpdateBatch` methods
3. **Soft Deletes**: Implement soft delete with `deleted_at` timestamp
4. **Auditing**: Track who made changes (add `created_by`, `updated_by` fields)
5. **Full-Text Search**: Add full-text search on title and description
6. **Caching**: Add Redis caching layer for frequently accessed todos
7. **Read Replicas**: Support read-only replicas for list queries
