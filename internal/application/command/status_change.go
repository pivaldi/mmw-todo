package command

import (
	"context"
	"fmt"

	"github.com/pivaldi/mmw/todo/internal/application/dto"
	"github.com/pivaldi/mmw/todo/internal/application/ports"
	domain "github.com/pivaldi/mmw/todo/internal/domain/todo"
)

// executeStatusChange is a helper to avoid code duplication for status change operations
// It handles the common flow of: retrieve todo -> execute action -> update -> dispatch events
func executeStatusChange(
	ctx context.Context,
	id string,
	repository ports.TodoRepository,
	eventDispatcher ports.EventDispatcher,
	action func(*domain.Todo) error,
	actionName string,
) (*dto.TodoResponse, error) {
	// Parse and validate ID
	todoID, err := domain.ParseTodoID(id)
	if err != nil {
		return nil, fmt.Errorf("invalid todo ID: %w", err)
	}

	// Retrieve existing todo
	todo, err := repository.FindByID(ctx, todoID)
	if err != nil {
		return nil, fmt.Errorf("finding todo: %w", err)
	}

	// Execute the action
	if err := action(todo); err != nil {
		return nil, fmt.Errorf("%s: %w", actionName, err)
	}

	// Persist changes
	if err := repository.Update(ctx, todo); err != nil {
		return nil, fmt.Errorf("updating todo: %w", err)
	}

	// Dispatch domain events
	if err := eventDispatcher.Dispatch(ctx, todo.Events()); err != nil {
		return nil, fmt.Errorf("dispatching events: %w", err)
	}

	// Clear events after dispatching
	todo.ClearEvents()

	// Map to response DTO
	return dto.MapTodoToResponse(todo), nil
}
