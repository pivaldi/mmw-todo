package domain

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// Todo is the aggregate root for the todo domain
// It enforces all business rules and maintains consistency
type Todo struct {
	id          TodoID
	title       TaskTitle
	description string
	status      TaskStatus
	priority    Priority
	dueDate     *DueDate
	createdAt   time.Time
	updatedAt   time.Time
	completedAt *time.Time
	events      []DomainEvent
	userID      uuid.UUID
}

// NewTodo creates a new Todo aggregate with validation
func NewTodo(title TaskTitle, description string, priority Priority, dueDate *DueDate, userID uuid.UUID) *Todo {
	id := NewTodoID()
	now := time.Now()

	todo := &Todo{
		id:          id,
		title:       title,
		description: description,
		status:      TaskStatusPending,
		priority:    priority,
		dueDate:     dueDate,
		createdAt:   now,
		updatedAt:   now,
		events:      []DomainEvent{},
		userID:      userID,
	}

	// Emit TodoCreated event
	todo.addEvent(NewTodoCreatedEvent(id, title, description, priority, dueDate))

	return todo
}

// Snapshot returns the Memento for this Todo — a plain-data representation
// of the aggregate's current state, suitable for persistence.
// events is not included; it is runtime state only.
func (t *Todo) Snapshot() TodoSnapshot {
	snap := TodoSnapshot{
		ID:          uuid.MustParse(string(t.id)), // TodoID is type string; MustParse is safe — ID is always valid UUID
		Title:       t.title.String(),
		Description: t.description,
		Status:      t.status.String(),
		Priority:    t.priority.String(),
		CreatedAt:   t.createdAt,
		UpdatedAt:   t.updatedAt,
		CompletedAt: t.completedAt,
		UserID:      t.userID,
	}
	if t.dueDate != nil {
		d := t.dueDate.Time()
		snap.DueDate = &d
	}

	return snap
}

// ReconstituteTodo reconstitutes a Todo from stored data (used by repository).
// Panics if the snapshot contains values that violate basic type invariants —
// this should never happen since the DB is the authoritative source of truth.
func ReconstituteTodo(snap *TodoSnapshot) *Todo {
	title, err := NewTaskTitle(snap.Title)
	if err != nil {
		panic(fmt.Sprintf("ReconstituteTodo: invalid title from DB: %v", err))
	}

	status, err := ParseTaskStatus(snap.Status)
	if err != nil {
		panic(fmt.Sprintf("ReconstituteTodo: invalid status from DB: %v", err))
	}

	priority, err := ParsePriority(snap.Priority)
	if err != nil {
		panic(fmt.Sprintf("ReconstituteTodo: invalid priority from DB: %v", err))
	}

	var dueDate *DueDate
	if snap.DueDate != nil {
		dd := DueDate{value: *snap.DueDate}
		dueDate = &dd
	}

	return &Todo{
		id:          TodoID(snap.ID.String()),
		title:       title,
		description: snap.Description,
		status:      status,
		priority:    priority,
		dueDate:     dueDate,
		createdAt:   snap.CreatedAt,
		updatedAt:   snap.UpdatedAt,
		completedAt: snap.CompletedAt,
		events:      []DomainEvent{},
		userID:      snap.UserID,
	}
}

// Getters

// ID returns the todo ID
func (t *Todo) ID() TodoID {
	return t.id
}

// Title returns the todo title
func (t *Todo) Title() TaskTitle {
	return t.title
}

// Description returns the todo description
func (t *Todo) Description() string {
	return t.description
}

// Status returns the todo status
func (t *Todo) Status() TaskStatus {
	return t.status
}

// Priority returns the todo priority
func (t *Todo) Priority() Priority {
	return t.priority
}

// DueDate returns the optional due date
func (t *Todo) DueDate() *DueDate {
	return t.dueDate
}

// CreatedAt returns when the todo was created
func (t *Todo) CreatedAt() time.Time {
	return t.createdAt
}

// UpdatedAt returns when the todo was last updated
func (t *Todo) UpdatedAt() time.Time {
	return t.updatedAt
}

// CompletedAt returns when the todo was completed (nil if not completed)
func (t *Todo) CompletedAt() *time.Time {
	return t.completedAt
}

// UserID returns the ID of the user who owns this todo
func (t *Todo) UserID() uuid.UUID {
	return t.userID
}

// Events returns the unpublished domain events
func (t *Todo) Events() []DomainEvent {
	return t.events
}

// ClearEvents clears the unpublished events (called after publishing)
func (t *Todo) ClearEvents() {
	t.events = []DomainEvent{}
}

// Business methods

func (t *Todo) Update(title, description *string, priority *Priority, dueDate *time.Time, status *TaskStatus) error {
	if title != nil {
		taskTitle, err := NewTaskTitle(*title)
		if err != nil {
			return err
		}

		t.title = taskTitle
	}

	if description != nil {
		t.description = *description
	}

	if priority != nil {
		t.priority = *priority
	}

	if dueDate != nil {
		var ldueDate *DueDate
		if !dueDate.Equal(time.Time{}) {
			dd, err := NewDueDate(*dueDate)
			if err != nil {
				return err
			}
			ldueDate = &dd
		}

		t.dueDate = ldueDate
	}

	if status != nil {
		t.status = *status
	}

	t.addEvent(NewTodoUpdatedEvent(t))

	return nil
}

// UpdateTitle updates the todo title with validation
func (t *Todo) UpdateTitle(newTitle TaskTitle) error {
	if t.status.IsCompleted() {
		return ErrCannotModifyCompleted
	}

	t.title = newTitle
	t.updatedAt = time.Now()
	t.addEvent(NewTodoUpdatedEvent(t))

	return nil
}

// UpdateDescription updates the todo description
func (t *Todo) UpdateDescription(newDescription string) error {
	if t.status.IsCompleted() {
		return ErrCannotModifyCompleted
	}

	t.description = newDescription
	t.updatedAt = time.Now()
	t.addEvent(NewTodoUpdatedEvent(t))

	return nil
}

// UpdatePriority updates the todo priority
func (t *Todo) UpdatePriority(newPriority Priority) error {
	if t.status.IsCompleted() {
		return ErrCannotModifyCompleted
	}

	t.priority = newPriority
	t.updatedAt = time.Now()
	t.addEvent(NewTodoUpdatedEvent(t))

	return nil
}

// UpdateDueDate updates the due date
func (t *Todo) UpdateDueDate(newDueDate *DueDate) error {
	if t.status.IsCompleted() {
		return ErrCannotModifyCompleted
	}

	t.dueDate = newDueDate
	t.updatedAt = time.Now()
	t.addEvent(NewTodoUpdatedEvent(t))

	return nil
}

// UpdateStatus updates the status with transition validation
func (t *Todo) UpdateStatus(newStatus TaskStatus) error {
	if !t.status.CanTransitionTo(newStatus) {
		return ErrInvalidStatusTransition
	}

	t.status = newStatus
	t.updatedAt = time.Now()
	t.addEvent(NewTodoUpdatedEvent(t))

	return nil
}

// Complete marks the todo as completed
func (t *Todo) Complete() error {
	if t.status.IsCancelled() {
		return ErrCannotCompleteCancelled
	}

	if t.status.IsCompleted() {
		return nil // Already completed, idempotent
	}

	t.status = TaskStatusCompleted
	now := time.Now()
	t.completedAt = &now
	t.updatedAt = now

	t.addEvent(NewTodoCompletedEvent(t.id, now))

	return nil
}

// Reopen reopens a completed or cancelled todo back to pending
func (t *Todo) Reopen() error {
	if !t.status.IsCompleted() && !t.status.IsCancelled() {
		return nil // Already open, idempotent
	}

	previousStatus := t.status
	t.status = TaskStatusPending
	t.completedAt = nil
	t.updatedAt = time.Now()

	t.addEvent(NewTodoReopenedEvent(t.id, previousStatus))

	return nil
}

// Cancel marks the todo as cancelled
func (t *Todo) Cancel() error {
	if t.status.IsCompleted() {
		return ErrCannotModifyCompleted
	}

	if t.status.IsCancelled() {
		return nil // Already cancelled, idempotent
	}

	t.status = TaskStatusCancelled
	t.updatedAt = time.Now()
	t.addEvent(NewTodoUpdatedEvent(t))

	return nil
}

// MarkInProgress marks the todo as in progress
func (t *Todo) MarkInProgress() error {
	if t.status.IsCompleted() {
		return ErrCannotModifyCompleted
	}

	if t.status == TaskStatusInProgress {
		return nil // Already in progress, idempotent
	}

	t.status = TaskStatusInProgress
	t.updatedAt = time.Now()
	t.addEvent(NewTodoUpdatedEvent(t))

	return nil
}

// IsDue checks if the todo has a due date and it has passed
func (t *Todo) IsDue() bool {
	if t.dueDate == nil {
		return false
	}

	return t.dueDate.IsPast()
}

// IsDueSoon checks if the todo is due within the specified duration
func (t *Todo) IsDueSoon(within time.Duration) bool {
	if t.dueDate == nil {
		return false
	}

	return t.dueDate.IsApproaching(within)
}

// Private methods

// addEvent adds a domain event to the unpublished events list
func (t *Todo) addEvent(event DomainEvent) {
	t.events = append(t.events, event)
}

// MarshalJSON implements json.Marshaler so that Todo serializes its private fields.
func (t *Todo) MarshalJSON() ([]byte, error) {
	type todoJSON struct {
		ID          string     `json:"id"`
		Title       string     `json:"title"`
		Description string     `json:"description"`
		Status      string     `json:"status"`
		Priority    string     `json:"priority"`
		DueDate     *time.Time `json:"dueDate,omitempty"`
		CreatedAt   time.Time  `json:"createdAt"`
		UpdatedAt   time.Time  `json:"updatedAt"`
		CompletedAt *time.Time `json:"completedAt,omitempty"`
		UserID      uuid.UUID  `json:"userId"`
	}

	var dueDate *time.Time
	if t.dueDate != nil {
		d := t.dueDate.Time()
		dueDate = &d
	}

	bs, err := json.Marshal(todoJSON{
		ID:          t.id.String(),
		Title:       t.title.String(),
		Description: t.description,
		Status:      t.status.String(),
		Priority:    t.priority.String(),
		DueDate:     dueDate,
		CreatedAt:   t.createdAt,
		UpdatedAt:   t.updatedAt,
		CompletedAt: t.completedAt,
		UserID:      t.userID,
	})
	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}

	return bs, nil
}
