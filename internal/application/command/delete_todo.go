package command

import (
	"context"

	"github.com/rotisserie/eris"

	"github.com/piprim/mmw/pkg/platform/authctx"
	"github.com/pivaldi/mmw-todo/internal/application/ports"
	"github.com/pivaldi/mmw-todo/internal/domain"
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
		return eris.Wrap(err, "delete todo")
	}

	// Parse and validate ID
	todoID, err := domain.ParseTodoID(id)
	if err != nil {
		return eris.Wrap(err, "invalid todo ID")
	}

	// Delete from repository
	if err := c.repository.Delete(ctx, todoID, userID); err != nil {
		return eris.Wrap(err, "deleting todo")
	}

	// Create and dispatch deleted event
	deletedEvent := domain.NewTodoDeletedEvent(todoID, userID)

	if err := c.eventDispatcher.Dispatch(ctx, []domain.DomainEvent{deletedEvent}); err != nil {
		return eris.Wrap(err, "dispatching events")
	}

	return nil
}
