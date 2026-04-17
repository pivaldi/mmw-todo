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

// newCreatedEvent builds the TodoCreated event for this aggregate.
func (t *Todo) newCreatedEvent() *TodoCreated {
	var dueDatePtr *time.Time
	if t.dueDate != nil {
		d := t.dueDate.Time()
		dueDatePtr = &d
	}

	return &TodoCreated{
		BaseDomainEvent: BaseDomainEvent{
			AggregateID: t.id.String(),
			UserID:      t.userID,
			OccurredAt:  time.Now(),
		},
		Title:       t.title.String(),
		Description: t.description,
		Priority:    t.priority.String(),
		DueDate:     dueDatePtr,
	}
}

// TodoUpdated event is emitted when a todo is modified
type TodoUpdated struct {
	BaseDomainEvent
	Title       string
	Description string
	Priority    string
	Status      string
	DueDate     *time.Time
}

// EventType returns the event type
func (*TodoUpdated) EventType() string {
	return EventTypeUpdated
}

// newUpdatedEvent builds the TodoUpdated event for this aggregate.
func (t *Todo) newUpdatedEvent() *TodoUpdated {
	var dueDatePtr *time.Time
	if t.dueDate != nil {
		d := t.dueDate.Time()
		dueDatePtr = &d
	}

	return &TodoUpdated{
		BaseDomainEvent: BaseDomainEvent{
			AggregateID: t.id.String(),
			UserID:      t.userID,
			OccurredAt:  time.Now(),
		},
		Title:       t.title.String(),
		Description: t.description,
		Priority:    t.priority.String(),
		Status:      t.status.String(),
		DueDate:     dueDatePtr,
	}
}

// TodoCompleted event is emitted when a todo is marked as completed
type TodoCompleted struct {
	BaseDomainEvent
	Title       string
	Description string
	Priority    string
	CompletedAt time.Time
}

// EventType returns the event type
func (*TodoCompleted) EventType() string {
	return EventTypeCompleted
}

// newCompletedEvent builds the TodoCompleted event for this aggregate.
func (t *Todo) newCompletedEvent() *TodoCompleted {
	var completedAt time.Time
	if t.completedAt != nil {
		completedAt = *t.completedAt
	}

	return &TodoCompleted{
		BaseDomainEvent: BaseDomainEvent{
			AggregateID: t.id.String(),
			UserID:      t.userID,
			OccurredAt:  time.Now(),
		},
		Title:       t.title.String(),
		Description: t.description,
		Priority:    t.priority.String(),
		CompletedAt: completedAt,
	}
}

// TodoReopened event is emitted when a completed todo is reopened
type TodoReopened struct {
	BaseDomainEvent
	Title          string
	PreviousStatus string
}

// EventType returns the event type
func (*TodoReopened) EventType() string {
	return EventTypeReopened
}

// newReopenedEvent builds the TodoReopened event for this aggregate.
// previousStatus is passed explicitly as it is the local state captured at the moment of reopening.
func (t *Todo) newReopenedEvent(previousStatus TaskStatus) *TodoReopened {
	return &TodoReopened{
		BaseDomainEvent: BaseDomainEvent{
			AggregateID: t.id.String(),
			UserID:      t.userID,
			OccurredAt:  time.Now(),
		},
		Title:          t.title.String(),
		PreviousStatus: previousStatus.String(),
	}
}

// TodoDeleted event is emitted when a todo is deleted
type TodoDeleted struct {
	BaseDomainEvent
	Title string
}

// EventType returns the event type
func (*TodoDeleted) EventType() string {
	return EventTypeDeleted
}

// newDeletedEvent builds the TodoDeleted event for this aggregate.
func (t *Todo) newDeletedEvent() *TodoDeleted {
	return &TodoDeleted{
		BaseDomainEvent: BaseDomainEvent{
			AggregateID: t.id.String(),
			UserID:      t.userID,
			OccurredAt:  time.Now(),
		},
		Title: t.title.String(),
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

// NewUserTasksDeletedEvent creates a new UserTasksDeleted event.
func NewUserTasksDeletedEvent(userID uuid.UUID) *UserTasksDeleted {
	return &UserTasksDeleted{
		BaseDomainEvent: BaseDomainEvent{
			AggregateID: userID.String(),
			UserID:      userID,
			OccurredAt:  time.Now(),
		},
	}
}
