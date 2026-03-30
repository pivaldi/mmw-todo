package command

import (
	"context"

	pfevents "github.com/piprim/mmw/platform/events"
	deftodo "github.com/pivaldi/mmw-contracts/definitions/todo"
	"github.com/rotisserie/eris"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type DeleteUserTasksCommand struct {
	bus pfevents.SystemEventBus
	// ... repo, etc.
}

func NewDeleteUserTasksCommand(bus pfevents.SystemEventBus) *DeleteUserTasksCommand {
	return &DeleteUserTasksCommand{bus: bus}
}

func (c *DeleteUserTasksCommand) Execute(ctx context.Context, userID string) error {
	// TODO: Delete the user's tasks from the todo repository
	// ids := c.repo.DeleteUserTasks(ctx, userID)

	eventDTO := &deftodo.UserTasksDeletedEvent{
		UserId:    userID,
		DeletedAt: timestamppb.Now(),
		TaskIds:   []int64{1, 2, 3}, // TODO: use the real ids
	}

	// 3. Serialize the payload using protojson (produces camelCase JSON)
	payload, err := protojson.Marshal(eventDTO)
	if err != nil {
		return eris.Wrap(err, "serializing payload failed")
	}

	return eris.Wrap(c.bus.Publish(ctx, deftodo.TopicUserTasksDeleted, payload), "publishing events failed")
}
