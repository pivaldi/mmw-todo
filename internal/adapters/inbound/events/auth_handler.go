package events

import (
	"github.com/ThreeDotsLabs/watermill/message"
	defauth "github.com/pivaldi/mmw-contracts/definitions/auth"
	"github.com/pivaldi/mmw-todo/internal/application/command"
	"github.com/rotisserie/eris"
	"google.golang.org/protobuf/encoding/protojson"
)

type AuthEventHandler struct {
	deleteUserTasksCmd *command.DeleteUserTasksCommand
}

func NewAuthEventHandler(cmd *command.DeleteUserTasksCommand) *AuthEventHandler {
	return &AuthEventHandler{deleteUserTasksCmd: cmd}
}

// HandleUserDeleted is wired to Watermill's subscriber for auth.TopicUserDeleted
func (h *AuthEventHandler) HandleUserDeleted(msg *message.Message) error {
	var event defauth.UserDeletedEvent

	// 1. Unmarshal the payload using the proto-generated contract type.
	// DiscardUnknown tolerates additive schema changes without breaking the consumer.
	opts := protojson.UnmarshalOptions{DiscardUnknown: true}
	if err := opts.Unmarshal(msg.Payload, &event); err != nil {
		return eris.Wrap(err, "fail to unmarshal payload") // Or send to a Dead Letter Queue!
	}

	// 2. Call the Todo internal application layer
	// Notice we only pass the raw data (event.UserId) into the Todo domain,
	// preventing the contract struct from leaking into business logic.
	return eris.Wrap(
		h.deleteUserTasksCmd.Execute(msg.Context(), event.UserId),
		"fail to delete user tasks")
}
