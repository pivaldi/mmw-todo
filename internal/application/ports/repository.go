package ports

import (
	"context"

	"github.com/google/uuid"

	"github.com/pivaldi/mmw-todo/internal/domain"
)

// TodoRepository defines the interface for todo persistence operations
// This is a secondary port (driven) - needed by the application, implemented by adapters
type TodoRepository interface {
	// Save persists a new todo
	Save(ctx context.Context, todo *domain.Todo) error

	// FindByID retrieves a todo by its ID
	// Returns ErrTodoNotFound if not found
	FindByID(ctx context.Context, id domain.TodoID, userID uuid.UUID) (*domain.Todo, error)

	// FindAll retrieves todos matching the given filters
	FindAll(ctx context.Context, filters Filters) ([]*domain.Todo, error)

	// Update updates an existing todo
	Update(ctx context.Context, todo *domain.Todo) error

	// Delete removes a todo
	Delete(ctx context.Context, id domain.TodoID, userID uuid.UUID) error

	// Health repo function
	Health(ctx context.Context) (any, error)
}

// Filters represents query filters for finding todos
type Filters struct {
	Status   *domain.TaskStatus
	Priority *domain.Priority
	Limit    *int
	Offset   *int
	// UserID scopes results to a specific user.
	// nil = no user filter (reserved for future admin use; not reachable from current call paths).
	UserID *uuid.UUID
}

// UnitOfWork defines the contract for atomic operations
type UnitOfWork interface {
	// WithTransaction excute the fn in a transaction context allowin atomic operation.
	WithTransaction(ctx context.Context, fn func(ctx context.Context) error) error
}
