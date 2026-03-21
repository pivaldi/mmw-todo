// services/authsvc/internal/application/command/delete_user.go
package command

import (
	"context"
	"encoding/json"
	"time"

	oglevnets "github.com/ovya/ogl/platform/events"
	deftodo "github.com/pivaldi/mmw-contracts/definitions/todo"
	"github.com/rotisserie/eris"
)

type DeleteUserTasksCommand struct {
	bus oglevnets.SystemEventBus
	// ... repo, etc.
}

func (c *DeleteUserTasksCommand) Execute(ctx context.Context, userID string) error {
	// TODO: Delete user from the Auth database
	// ids := c.repo.DeleteUserTasks(ctx, userID)

	eventDTO := deftodo.UserTasksDeletedEvent{
		UserID:    userID,
		DeletedAt: time.Now(),
		TasksIDS:  []int{1, 2, 3}, // TODO: use the real ids
	}

	// 3. Serialize the payload (JSON for now, Protobuf later)
	payload, err := json.Marshal(eventDTO)
	if err != nil {
		return eris.Wrap(err, "serializing payload failed")
	}

	return eris.Wrap(c.bus.Publish(ctx, deftodo.TopicUserTasksDeleted, payload), "publishing events failed")
}
