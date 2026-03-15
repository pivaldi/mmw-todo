package query

import (
	"context"
	"fmt"

	"github.com/pivaldi/mmw/todo/internal/application/dto"
	"github.com/pivaldi/mmw/todo/internal/application/ports"
	domain "github.com/pivaldi/mmw/todo/internal/domain/todo"
)

// GetTodoQuery handles retrieving a single todo by ID
type GetTodoQuery struct {
	repository ports.TodoRepository
}

// NewGetTodoQuery creates a new GetTodoQuery
func NewGetTodoQuery(repository ports.TodoRepository) *GetTodoQuery {
	return &GetTodoQuery{repository: repository}
}

// Execute retrieves a todo by ID
func (q *GetTodoQuery) Execute(
	ctx context.Context,
	id string,
) (*dto.TodoResponse, error) {
	// Parse and validate ID
	todoID, err := domain.ParseTodoID(id)
	if err != nil {
		return nil, fmt.Errorf("invalid todo ID: %w", err)
	}

	// Retrieve from repository
	todo, err := q.repository.FindByID(ctx, todoID)
	if err != nil {
		return nil, fmt.Errorf("finding todo: %w", err)
	}

	// Map to response DTO
	return dto.MapTodoToResponse(todo), nil
}
