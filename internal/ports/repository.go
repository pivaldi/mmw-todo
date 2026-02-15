package ports

import (
	"context"

	domain "github.com/pivaldi/mmw/todo/internal/domain/todo"
)

// TodoRepository defines the interface for todo persistence operations
// This is a secondary port (driven) - needed by the application, implemented by adapters
type TodoRepository interface {
	// Save persists a new todo
	Save(ctx context.Context, todo *domain.Todo) error

	// FindByID retrieves a todo by its ID
	// Returns ErrTodoNotFound if not found
	FindByID(ctx context.Context, id domain.TodoID) (*domain.Todo, error)

	// FindAll retrieves todos matching the given filters
	FindAll(ctx context.Context, filters Filters) ([]*domain.Todo, error)

	// Update updates an existing todo
	Update(ctx context.Context, todo *domain.Todo) error

	// Delete removes a todo
	Delete(ctx context.Context, id domain.TodoID) error
}

// Filters represents query filters for finding todos
type Filters struct {
	Status   *domain.TaskStatus
	Priority *domain.Priority
	Limit    *int
	Offset   *int
}
