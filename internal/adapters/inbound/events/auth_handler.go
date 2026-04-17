package events

import (
	"context"

	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/google/uuid"
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
		userID, err := uuid.Parse(e.UserId)
		if err != nil {
			return eris.Wrap(err, "invalid user ID in UserDeletedEvent")
		}

		return eris.Wrap(cmd.Execute(ctx, userID), "fail to delete user tasks")
	})
}
