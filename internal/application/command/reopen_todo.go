package command

import (
	"context"

	"github.com/pivaldi/mmw/todo/internal/application/dto"
	"github.com/pivaldi/mmw/todo/internal/application/ports"
	domain "github.com/pivaldi/mmw/todo/internal/domain/todo"
)

// ReopenTodoCommand handles reopening completed or cancelled todos
type ReopenTodoCommand struct {
	repository      ports.TodoRepository
	eventDispatcher ports.EventDispatcher
}

// NewReopenTodoCommand creates a new ReopenTodoCommand
func NewReopenTodoCommand(
	repository ports.TodoRepository,
	eventDispatcher ports.EventDispatcher,
) *ReopenTodoCommand {
	return &ReopenTodoCommand{
		repository:      repository,
		eventDispatcher: eventDispatcher,
	}
}

// Execute reopens a completed or cancelled todo
func (c *ReopenTodoCommand) Execute(
	ctx context.Context,
	id string,
) (*dto.TodoResponse, error) {
	return executeStatusChange(
		ctx,
		id,
		c.repository,
		c.eventDispatcher,
		(*domain.Todo).Reopen,
		"reopening todo",
	)
}
