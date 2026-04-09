package events

import (
	"context"

	"github.com/ThreeDotsLabs/watermill/message"
	pfevents "github.com/piprim/mmw/pkg/platform/events"
	defauth "github.com/pivaldi/mmw-contracts/go/application/auth"
	"github.com/pivaldi/mmw-todo/internal/application/command"
	"github.com/rotisserie/eris"
)

// HandleUserDeleted returns a Watermill handler that unmarshals a
// defauth.UserDeletedEvent from the message payload and calls the
// DeleteUserTasksCommand. Wire it to defauth.TopicUserDeleted using AddNoPublisherHandler.
func HandleUserDeleted(cmd *command.DeleteUserTasksCommand) func(*message.Message) error {
	return pfevents.Handle(func(ctx context.Context, e *defauth.UserDeletedEvent) error {
		return eris.Wrap(cmd.Execute(ctx, e.UserId), "fail to delete user tasks")
	})
}
