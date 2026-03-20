package command

import (
	"context"
	"fmt"

	"github.com/pivaldi/mmw/todo/internal/application/authctx"
	"github.com/pivaldi/mmw/todo/internal/application/ports"
	domain "github.com/pivaldi/mmw/todo/internal/domain/todo"
)

// DeleteTodoCommand handles deleting todos
type DeleteTodoCommand struct {
	repository      ports.TodoRepository
	eventDispatcher ports.EventDispatcher
}

// NewDeleteTodoCommand creates a new DeleteTodoCommand
func NewDeleteTodoCommand(
	repository ports.TodoRepository,
	eventDispatcher ports.EventDispatcher,
) *DeleteTodoCommand {
	return &DeleteTodoCommand{
		repository:      repository,
		eventDispatcher: eventDispatcher,
	}
}

// Execute deletes a todo
func (c *DeleteTodoCommand) Execute(
	ctx context.Context,
	id string,
) error {
	userID, err := authctx.UserIDFromContext(ctx)
	if err != nil {
		return fmt.Errorf("delete todo: %w", err)
	}

	// Parse and validate ID
	todoID, err := domain.ParseTodoID(id)
	if err != nil {
		return fmt.Errorf("invalid todo ID: %w", err)
	}

	// Delete from repository
	if err := c.repository.Delete(ctx, todoID, userID); err != nil {
		return fmt.Errorf("deleting todo: %w", err)
	}

	// Create and dispatch deleted event
	deletedEvent := domain.NewTodoDeletedEvent(todoID)
	if err := c.eventDispatcher.Dispatch(ctx, []domain.DomainEvent{deletedEvent}); err != nil {
		return fmt.Errorf("dispatching events: %w", err)
	}

	return nil
}
