package command

import (
	"context"

	"github.com/pivaldi/mmw-todo/internal/application/ports"
	"github.com/pivaldi/mmw-todo/internal/domain"
)

type DeleteUserTasksCommand struct {
	dispatcher ports.EventDispatcher
	// ... repo, etc.
}

func NewDeleteUserTasksCommand(dispatcher ports.EventDispatcher) *DeleteUserTasksCommand {
	return &DeleteUserTasksCommand{dispatcher: dispatcher}
}

func (c *DeleteUserTasksCommand) Execute(ctx context.Context, userID string) error {
	// TODO: Delete the user's tasks from the todo repository
	// ids := c.repo.DeleteUserTasks(ctx, userID)

	event := domain.NewUserTasksDeletedEvent(userID)

	return c.dispatcher.Dispatch(ctx, []domain.DomainEvent{event})
}
