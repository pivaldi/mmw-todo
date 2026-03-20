package command

import (
	"context"
	"fmt"

	"github.com/pivaldi/mmw/todo/internal/application/authctx"
	"github.com/pivaldi/mmw/todo/internal/application/dto"
	"github.com/pivaldi/mmw/todo/internal/application/ports"
	domain "github.com/pivaldi/mmw/todo/internal/domain/todo"
)

// CompleteTodoCommand handles marking todos as completed
type CompleteTodoCommand struct {
	repository      ports.TodoRepository
	eventDispatcher ports.EventDispatcher
}

// NewCompleteTodoCommand creates a new CompleteTodoCommand
func NewCompleteTodoCommand(
	repository ports.TodoRepository,
	eventDispatcher ports.EventDispatcher,
) *CompleteTodoCommand {
	return &CompleteTodoCommand{
		repository:      repository,
		eventDispatcher: eventDispatcher,
	}
}

// Execute marks a todo as completed
func (c *CompleteTodoCommand) Execute(
	ctx context.Context,
	id string,
) (*dto.TodoResponse, error) {
	userID, err := authctx.UserIDFromContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("complete todo: %w", err)
	}

	return executeStatusChange(
		ctx,
		id,
		userID,
		c.repository,
		c.eventDispatcher,
		(*domain.Todo).Complete,
		"completing todo",
	)
}
