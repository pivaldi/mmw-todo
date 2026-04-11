package command

import (
	"context"

	"github.com/rotisserie/eris"

	"github.com/piprim/mmw/pkg/platform/authctx"
	"github.com/pivaldi/mmw-todo/internal/application/dto"
	"github.com/pivaldi/mmw-todo/internal/application/ports"
	"github.com/pivaldi/mmw-todo/internal/domain"
)

// UpdateTodoCommand handles updating existing todos
type UpdateTodoCommand struct {
	repository      ports.TodoRepository
	eventDispatcher ports.EventDispatcher
}

// NewUpdateTodoCommand creates a new UpdateTodoCommand
func NewUpdateTodoCommand(
	repository ports.TodoRepository,
	eventDispatcher ports.EventDispatcher,
) *UpdateTodoCommand {
	return &UpdateTodoCommand{
		repository:      repository,
		eventDispatcher: eventDispatcher,
	}
}

// Execute updates an existing todo
func (c *UpdateTodoCommand) Execute(
	ctx context.Context,
	id string,
	req *dto.UpdateTodoRequest,
) (*dto.TodoResponse, error) {
	userID, err := authctx.UserIDFromContext(ctx)
	if err != nil {
		return nil, eris.Wrap(err, "update todo")
	}

	todoID, err := domain.ParseTodoID(id)
	if err != nil {
		return nil, eris.Wrap(err, "invalid todo ID")
	}

	todo, err := c.repository.FindByID(ctx, todoID, userID)
	if err != nil {
		return nil, eris.Wrap(err, "finding todo")
	}

	err = todo.Update(req.Title, req.Description, req.Priority, req.DueDate, req.Status)
	if err != nil {
		return nil, eris.Wrap(err, "failed to update")
	}

	if err := c.repository.Update(ctx, todo); err != nil {
		return nil, eris.Wrap(err, "updating todo")
	}

	if err := c.eventDispatcher.Dispatch(ctx, todo.Events()); err != nil {
		return nil, eris.Wrap(err, "dispatching events")
	}

	// Clear events after dispatching
	todo.ClearEvents()

	return dto.MapTodoToResponse(todo), nil
}
