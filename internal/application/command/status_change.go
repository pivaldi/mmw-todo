package command

import (
	"context"

	"github.com/google/uuid"
	"github.com/rotisserie/eris"

	"github.com/pivaldi/mmw-todo/internal/application/authctx"
	"github.com/pivaldi/mmw-todo/internal/application/dto"
	"github.com/pivaldi/mmw-todo/internal/application/ports"
	domain "github.com/pivaldi/mmw-todo/internal/domain/todo"
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

	return executeStatusChange(ctx, id, userID, c.repository, c.eventDispatcher, c.action, c.actionLabel)
}

// executeStatusChange handles the common flow: retrieve todo -> execute action -> update -> dispatch events.
//
//nolint:revive // Because it's an helper
func executeStatusChange(
	ctx context.Context,
	id string,
	userID uuid.UUID,
	repository ports.TodoRepository,
	eventDispatcher ports.EventDispatcher,
	action func(*domain.Todo) error,
	actionName string,
) (*dto.TodoResponse, error) {
	todoID, err := domain.ParseTodoID(id)
	if err != nil {
		return nil, eris.Wrap(err, "invalid todo ID")
	}

	todo, err := repository.FindByID(ctx, todoID, userID)
	if err != nil {
		return nil, eris.Wrap(err, "finding todo")
	}

	if err := action(todo); err != nil {
		return nil, eris.Wrapf(err, "%s", actionName)
	}

	if err := repository.Update(ctx, todo); err != nil {
		return nil, eris.Wrap(err, "updating todo")
	}

	if err := eventDispatcher.Dispatch(ctx, todo.Events()); err != nil {
		return nil, eris.Wrap(err, "dispatching events")
	}

	todo.ClearEvents()

	return dto.MapTodoToResponse(todo), nil
}
