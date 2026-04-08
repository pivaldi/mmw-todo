package domain

import (
	"time"

	"github.com/google/uuid"
)

// Event type constants — semantic identifiers owned by the domain.
// Adapters are responsible for mapping these to transport-layer routing keys.
const (
	EventTypeCreated          = "todo.created"
	EventTypeUpdated          = "todo.updated"
	EventTypeCompleted        = "todo.completed"
	EventTypeReopened         = "todo.reopened"
	EventTypeDeleted          = "todo.deleted"
	EventTypeUserTasksDeleted = "todo.user_tasks_deleted"
)

// DomainEvent is the interface that all domain events must implement
type DomainEvent interface {
	// EventType returns the type of the event
	EventType() string
	// GetAggregateID returns the ID of the aggregate that emitted the event
	GetAggregateID() string
	// GetUserID returns the ID of the user associated with this event
	GetUserID() uuid.UUID
	// GetOccurredAt returns when the event occurred
	GetOccurredAt() time.Time
}

// BaseDomainEvent contains common fields for all domain events
type BaseDomainEvent struct {
	AggregateID string
	UserID      uuid.UUID
	OccurredAt  time.Time
}

// GetAggregateID returns the ID of the aggregate
func (e BaseDomainEvent) GetAggregateID() string {
	return e.AggregateID
}

// GetUserID returns the ID of the user
func (e BaseDomainEvent) GetUserID() uuid.UUID {
	return e.UserID
}

// GetOccurredAt returns when the event occurred
func (e BaseDomainEvent) GetOccurredAt() time.Time {
	return e.OccurredAt
}

// TodoCreated event is emitted when a new todo is created
type TodoCreated struct {
	BaseDomainEvent
	Title       string
	Description string
	Priority    string
	DueDate     *time.Time
}

// EventType returns the event type
func (*TodoCreated) EventType() string {
	return EventTypeCreated
}

// NewTodoCreatedEvent creates a new TodoCreated event
func NewTodoCreatedEvent(id TodoID, userID uuid.UUID, title TaskTitle, description string, priority Priority, dueDate *DueDate) *TodoCreated {
	var dueDatePtr *time.Time
	if dueDate != nil {
		t := dueDate.Time()
		dueDatePtr = &t
	}

	return &TodoCreated{
		BaseDomainEvent: BaseDomainEvent{
			AggregateID: id.String(),
			UserID:      userID,
			OccurredAt:  time.Now(),
		},
		Title:       title.String(),
		Description: description,
		Priority:    priority.String(),
		DueDate:     dueDatePtr,
	}
}

// TodoUpdated event is emitted when a todo is modified
type TodoUpdated struct {
	BaseDomainEvent
	Todo *Todo
}

// EventType returns the event type
func (*TodoUpdated) EventType() string {
	return EventTypeUpdated
}

// NewTodoUpdatedEvent creates a new TodoUpdated event
func NewTodoUpdatedEvent(todo *Todo) *TodoUpdated {
	return &TodoUpdated{
		BaseDomainEvent: BaseDomainEvent{
			AggregateID: todo.id.String(),
			UserID:      todo.userID,
			OccurredAt:  time.Now(),
		},
		Todo: todo,
	}
}

// TodoCompleted event is emitted when a todo is marked as completed
type TodoCompleted struct {
	BaseDomainEvent
	CompletedAt time.Time
}

// EventType returns the event type
func (*TodoCompleted) EventType() string {
	return EventTypeCompleted
}

// NewTodoCompletedEvent creates a new TodoCompleted event
func NewTodoCompletedEvent(id TodoID, userID uuid.UUID, completedAt time.Time) *TodoCompleted {
	return &TodoCompleted{
		BaseDomainEvent: BaseDomainEvent{
			AggregateID: id.String(),
			UserID:      userID,
			OccurredAt:  time.Now(),
		},
		CompletedAt: completedAt,
	}
}

// TodoReopened event is emitted when a completed todo is reopened
type TodoReopened struct {
	BaseDomainEvent
	PreviousStatus string
}

// EventType returns the event type
func (*TodoReopened) EventType() string {
	return EventTypeReopened
}

// NewTodoReopenedEvent creates a new TodoReopened event
func NewTodoReopenedEvent(id TodoID, userID uuid.UUID, previousStatus TaskStatus) *TodoReopened {
	return &TodoReopened{
		BaseDomainEvent: BaseDomainEvent{
			AggregateID: id.String(),
			UserID:      userID,
			OccurredAt:  time.Now(),
		},
		PreviousStatus: previousStatus.String(),
	}
}

// TodoDeleted event is emitted when a todo is deleted
type TodoDeleted struct {
	BaseDomainEvent
}

// EventType returns the event type
func (*TodoDeleted) EventType() string {
	return EventTypeDeleted
}

// NewTodoDeletedEvent creates a new TodoDeleted event
func NewTodoDeletedEvent(id TodoID, userID uuid.UUID) *TodoDeleted {
	return &TodoDeleted{
		BaseDomainEvent: BaseDomainEvent{
			AggregateID: id.String(),
			UserID:      userID,
			OccurredAt:  time.Now(),
		},
	}
}

// UserTasksDeleted event is emitted when all tasks of a user are deleted
type UserTasksDeleted struct {
	BaseDomainEvent
}

// EventType returns the event type
func (*UserTasksDeleted) EventType() string {
	return EventTypeUserTasksDeleted
}

// NewUserTasksDeletedEvent creates a new UserTasksDeleted event
func NewUserTasksDeletedEvent(userID string) *UserTasksDeleted {
	uID, _ := uuid.Parse(userID)
	return &UserTasksDeleted{
		BaseDomainEvent: BaseDomainEvent{
			AggregateID: userID,
			UserID:      uID,
			OccurredAt:  time.Now(),
		},
	}
}
