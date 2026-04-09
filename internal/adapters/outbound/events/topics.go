package events

import (
	tododef "github.com/pivaldi/mmw-contracts/go/application/todo"
	"github.com/pivaldi/mmw-todo/internal/domain"
)

// domainTopics maps semantic domain event types to Watermill routing keys.
// This is the single place where domain semantics are translated to transport concerns.
//
//nolint:gochecknoglobals // package-level lookup table, not mutable state
var domainTopics = map[string]string{
	domain.EventTypeCreated:          tododef.TopicUserTaskCreated,
	domain.EventTypeUpdated:          tododef.TopicUserTaskUpdated,
	domain.EventTypeCompleted:        tododef.TopicUserTaskCompleted,
	domain.EventTypeReopened:         tododef.TopicUserTaskReopened,
	domain.EventTypeDeleted:          tododef.TopicUserTaskDeleted,
	domain.EventTypeUserTasksDeleted: tododef.TopicUserTasksDeleted,
}
