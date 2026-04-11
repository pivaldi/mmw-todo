package query

import (
	"context"

	"github.com/rotisserie/eris"

	"github.com/piprim/mmw/pkg/platform/authctx"
	"github.com/pivaldi/mmw-todo/internal/application/dto"
	"github.com/pivaldi/mmw-todo/internal/application/ports"
	"github.com/pivaldi/mmw-todo/internal/domain"
)

// ListTodosQuery handles retrieving a list of todos with filters
type ListTodosQuery struct {
	repository ports.TodoRepository
}

// NewListTodosQuery creates a new ListTodosQuery
func NewListTodosQuery(repository ports.TodoRepository) *ListTodosQuery {
	return &ListTodosQuery{repository: repository}
}

// Execute retrieves todos with optional filters
func (q *ListTodosQuery) Execute(
	ctx context.Context,
	filters *dto.ListFilters,
) (*dto.ListTodosResponse, error) {
	userID, err := authctx.UserIDFromContext(ctx)
	if err != nil {
		return nil, eris.Wrap(err, "list todos")
	}

	// Convert application filters to repository filters
	repoFilters := ports.Filters{
		Limit:  filters.Limit,
		Offset: filters.Offset,
		UserID: &userID,
	}

	if filters.Status != nil {
		repoFilters.Status = filters.Status
	}

	if filters.Priority != nil {
		// Validate priority enum
		if !filters.Priority.IsValid() {
			return nil, eris.Wrap(domain.ErrInvalidPriority, "invalid priority filter")
		}
		repoFilters.Priority = filters.Priority
	}

	// Retrieve todos from repository
	todos, err := q.repository.FindAll(ctx, repoFilters)
	if err != nil {
		return nil, eris.Wrap(err, "finding todos")
	}

	// Map to response DTOs
	return &dto.ListTodosResponse{
		Todos:      dto.MapTodosToResponse(todos),
		TotalCount: len(todos),
	}, nil
}
