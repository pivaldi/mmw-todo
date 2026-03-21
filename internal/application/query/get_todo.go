package query

import (
	"context"

	"github.com/rotisserie/eris"

	"github.com/pivaldi/mmw-todo/internal/application/authctx"
	"github.com/pivaldi/mmw-todo/internal/application/dto"
	"github.com/pivaldi/mmw-todo/internal/application/ports"
	domain "github.com/pivaldi/mmw-todo/internal/domain/todo"
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
	userID, err := authctx.UserIDFromContext(ctx)
	if err != nil {
		return nil, eris.Wrap(err, "get todo")
	}

	// Parse and validate ID
	todoID, err := domain.ParseTodoID(id)
	if err != nil {
		return nil, eris.Wrap(err, "invalid todo ID")
	}

	// Retrieve from repository
	todo, err := q.repository.FindByID(ctx, todoID, userID)
	if err != nil {
		return nil, eris.Wrap(err, "finding todo")
	}

	// Map to response DTO
	return dto.MapTodoToResponse(todo), nil
}
