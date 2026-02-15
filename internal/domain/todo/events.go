package domain

import "time"

// DomainEvent is the interface that all domain events must implement
type DomainEvent interface {
	// EventType returns the type of the event
	EventType() string
	// AggregateID returns the ID of the aggregate that emitted the event
	AggregateID() string
	// OccurredAt returns when the event occurred
	OccurredAt() time.Time
}

// BaseDomainEvent contains common fields for all domain events
type BaseDomainEvent struct {
	aggregateID string
	occurredAt  time.Time
}

// AggregateID returns the ID of the aggregate
func (e BaseDomainEvent) AggregateID() string {
	return e.aggregateID
}

// OccurredAt returns when the event occurred
func (e BaseDomainEvent) OccurredAt() time.Time {
	return e.occurredAt
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
func (e TodoCreated) EventType() string {
	return "TodoCreated"
}

// NewTodoCreatedEvent creates a new TodoCreated event
func NewTodoCreatedEvent(id TodoID, title TaskTitle, description string, priority Priority, dueDate *DueDate) TodoCreated {
	var dueDatePtr *time.Time
	if dueDate != nil {
		t := dueDate.Time()
		dueDatePtr = &t
	}

	return TodoCreated{
		BaseDomainEvent: BaseDomainEvent{
			aggregateID: id.String(),
			occurredAt:  time.Now(),
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
	Title       *string
	Description *string
	Priority    *string
	DueDate     *time.Time
	Status      *string
}

// EventType returns the event type
func (e TodoUpdated) EventType() string {
	return "TodoUpdated"
}

// NewTodoUpdatedEvent creates a new TodoUpdated event
func NewTodoUpdatedEvent(id TodoID) TodoUpdated {
	return TodoUpdated{
		BaseDomainEvent: BaseDomainEvent{
			aggregateID: id.String(),
			occurredAt:  time.Now(),
		},
	}
}

// TodoCompleted event is emitted when a todo is marked as completed
type TodoCompleted struct {
	BaseDomainEvent
	CompletedAt time.Time
}

// EventType returns the event type
func (e TodoCompleted) EventType() string {
	return "TodoCompleted"
}

// NewTodoCompletedEvent creates a new TodoCompleted event
func NewTodoCompletedEvent(id TodoID, completedAt time.Time) TodoCompleted {
	return TodoCompleted{
		BaseDomainEvent: BaseDomainEvent{
			aggregateID: id.String(),
			occurredAt:  time.Now(),
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
func (e TodoReopened) EventType() string {
	return "TodoReopened"
}

// NewTodoReopenedEvent creates a new TodoReopened event
func NewTodoReopenedEvent(id TodoID, previousStatus TaskStatus) TodoReopened {
	return TodoReopened{
		BaseDomainEvent: BaseDomainEvent{
			aggregateID: id.String(),
			occurredAt:  time.Now(),
		},
		PreviousStatus: previousStatus.String(),
	}
}

// TodoDeleted event is emitted when a todo is deleted
type TodoDeleted struct {
	BaseDomainEvent
}

// EventType returns the event type
func (e TodoDeleted) EventType() string {
	return "TodoDeleted"
}

// NewTodoDeletedEvent creates a new TodoDeleted event
func NewTodoDeletedEvent(id TodoID) TodoDeleted {
	return TodoDeleted{
		BaseDomainEvent: BaseDomainEvent{
			aggregateID: id.String(),
			occurredAt:  time.Now(),
		},
	}
}
