package command

import (
	"context"

	"github.com/rotisserie/eris"

	"github.com/piprim/mmw/pkg/platform/authctx"
	"github.com/pivaldi/mmw-todo/internal/application/dto"
	"github.com/pivaldi/mmw-todo/internal/application/ports"
	"github.com/pivaldi/mmw-todo/internal/domain"
)

// TodoStatusChangeCommand handles todo status transitions (complete, reopen, etc.).
type TodoStatusChangeCommand struct {
	repository      ports.TodoRepository
	eventDispatcher ports.EventDispatcher
	action          func(*domain.Todo) error
	actionLabel     string
	errorPrefix     string
}

// NewCompleteTodoCommand creates a command that marks a todo as completed.
func NewCompleteTodoCommand(
	repository ports.TodoRepository,
	eventDispatcher ports.EventDispatcher,
) *TodoStatusChangeCommand {
	return &TodoStatusChangeCommand{
		repository:      repository,
		eventDispatcher: eventDispatcher,
		action:          (*domain.Todo).Complete,
		actionLabel:     "completing todo",
		errorPrefix:     "complete todo",
	}
}

// NewReopenTodoCommand creates a command that reopens a completed or cancelled todo.
func NewReopenTodoCommand(
	repository ports.TodoRepository,
	eventDispatcher ports.EventDispatcher,
) *TodoStatusChangeCommand {
	return &TodoStatusChangeCommand{
		repository:      repository,
		eventDispatcher: eventDispatcher,
		action:          (*domain.Todo).Reopen,
		actionLabel:     "reopening todo",
		errorPrefix:     "reopen todo",
	}
}

// Execute runs the status change for the given todo ID.
func (c *TodoStatusChangeCommand) Execute(ctx context.Context, id string) (*dto.TodoResponse, error) {
	userID, err := authctx.UserIDFromContext(ctx)
	if err != nil {
		return nil, eris.Wrapf(err, "%s", c.errorPrefix)
	}

	todoID, err := domain.ParseTodoID(id)
	if err != nil {
		return nil, eris.Wrap(err, "invalid todo ID")
	}

	todo, err := c.repository.FindByID(ctx, todoID, userID)
	if err != nil {
		return nil, eris.Wrap(err, "finding todo")
	}

	if err := c.action(todo); err != nil {
		return nil, eris.Wrapf(err, "%s", c.actionLabel)
	}

	if err := c.repository.Update(ctx, todo); err != nil {
		return nil, eris.Wrap(err, "updating todo")
	}

	if err := c.eventDispatcher.Dispatch(ctx, todo.Events()); err != nil {
		return nil, eris.Wrap(err, "dispatching events")
	}

	todo.ClearEvents()

	return dto.MapTodoToResponse(todo), nil
}
