package domain

import (
	"time"

	tododef "github.com/pivaldi/mmw-contracts/definitions/todo"
)

// DomainEvent is the interface that all domain events must implement
type DomainEvent interface {
	// EventType returns the type of the event
	EventType() string
	// GetAggregateID returns the ID of the aggregate that emitted the event
	GetAggregateID() string
	// GetOccurredAt returns when the event occurred
	GetOccurredAt() time.Time
}

// BaseDomainEvent contains common fields for all domain events
type BaseDomainEvent struct {
	AggregateID string
	OccurredAt  time.Time
}

// GetAggregateID returns the ID of the aggregate
func (e BaseDomainEvent) GetAggregateID() string {
	return e.AggregateID
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
	return tododef.TopicUserTaskCreated
}

// NewTodoCreatedEvent creates a new TodoCreated event
func NewTodoCreatedEvent(id TodoID, title TaskTitle, description string, priority Priority, dueDate *DueDate) *TodoCreated {
	var dueDatePtr *time.Time
	if dueDate != nil {
		t := dueDate.Time()
		dueDatePtr = &t
	}

	return &TodoCreated{
		BaseDomainEvent: BaseDomainEvent{
			AggregateID: id.String(),
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
	return tododef.TopicUserTaskUpdated
}

// NewTodoUpdatedEvent creates a new TodoUpdated event
func NewTodoUpdatedEvent(todo *Todo) *TodoUpdated {
	return &TodoUpdated{
		BaseDomainEvent: BaseDomainEvent{
			AggregateID: todo.id.String(),
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
	return tododef.TopicUserTaskCompleted
}

// NewTodoCompletedEvent creates a new TodoCompleted event
func NewTodoCompletedEvent(id TodoID, completedAt time.Time) *TodoCompleted {
	return &TodoCompleted{
		BaseDomainEvent: BaseDomainEvent{
			AggregateID: id.String(),
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
	return tododef.TopicUserTaskReopened
}

// NewTodoReopenedEvent creates a new TodoReopened event
func NewTodoReopenedEvent(id TodoID, previousStatus TaskStatus) *TodoReopened {
	return &TodoReopened{
		BaseDomainEvent: BaseDomainEvent{
			AggregateID: id.String(),
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
	return tododef.TopicUserTaskDeleted
}

// NewTodoDeletedEvent creates a new TodoDeleted event
func NewTodoDeletedEvent(id TodoID) *TodoDeleted {
	return &TodoDeleted{
		BaseDomainEvent: BaseDomainEvent{
			AggregateID: id.String(),
			OccurredAt:  time.Now(),
		},
	}
}
