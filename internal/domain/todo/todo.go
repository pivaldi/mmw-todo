package domain

import "time"

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
}

// NewTodo creates a new Todo aggregate with validation
func NewTodo(title TaskTitle, description string, priority Priority, dueDate *DueDate) *Todo {
	id := NewTodoID()
	now := time.Now()

	todo := &Todo{
		id:          id,
		title:       title,
		description: description,
		status:      StatusPending,
		priority:    priority,
		dueDate:     dueDate,
		createdAt:   now,
		updatedAt:   now,
		events:      []DomainEvent{},
	}

	// Emit TodoCreated event
	todo.addEvent(NewTodoCreatedEvent(id, title, description, priority, dueDate))

	return todo
}

// ReconstituteTodo reconstitutes a Todo from stored data (used by repository)
func ReconstituteTodo(
	id TodoID,
	title TaskTitle,
	description string,
	status TaskStatus,
	priority Priority,
	dueDate *DueDate,
	createdAt, updatedAt time.Time,
	completedAt *time.Time,
) *Todo {
	return &Todo{
		id:          id,
		title:       title,
		description: description,
		status:      status,
		priority:    priority,
		dueDate:     dueDate,
		createdAt:   createdAt,
		updatedAt:   updatedAt,
		completedAt: completedAt,
		events:      []DomainEvent{},
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

// Events returns the unpublished domain events
func (t *Todo) Events() []DomainEvent {
	return t.events
}

// ClearEvents clears the unpublished events (called after publishing)
func (t *Todo) ClearEvents() {
	t.events = []DomainEvent{}
}

// Business methods

// UpdateTitle updates the todo title with validation
func (t *Todo) UpdateTitle(newTitle TaskTitle) error {
	if t.status.IsCompleted() {
		return ErrCannotModifyCompleted
	}

	t.title = newTitle
	t.updatedAt = time.Now()
	t.addEvent(NewTodoUpdatedEvent(t.id))

	return nil
}

// UpdateDescription updates the todo description
func (t *Todo) UpdateDescription(newDescription string) error {
	if t.status.IsCompleted() {
		return ErrCannotModifyCompleted
	}

	t.description = newDescription
	t.updatedAt = time.Now()
	t.addEvent(NewTodoUpdatedEvent(t.id))

	return nil
}

// UpdatePriority updates the todo priority
func (t *Todo) UpdatePriority(newPriority Priority) error {
	if t.status.IsCompleted() {
		return ErrCannotModifyCompleted
	}

	t.priority = newPriority
	t.updatedAt = time.Now()
	t.addEvent(NewTodoUpdatedEvent(t.id))

	return nil
}

// UpdateDueDate updates the due date
func (t *Todo) UpdateDueDate(newDueDate *DueDate) error {
	if t.status.IsCompleted() {
		return ErrCannotModifyCompleted
	}

	t.dueDate = newDueDate
	t.updatedAt = time.Now()
	t.addEvent(NewTodoUpdatedEvent(t.id))

	return nil
}

// UpdateStatus updates the status with transition validation
func (t *Todo) UpdateStatus(newStatus TaskStatus) error {
	if !t.status.CanTransitionTo(newStatus) {
		return NewBusinessRuleError(
			"status_transition",
			"cannot transition from "+t.status.String()+" to "+newStatus.String(),
		)
	}

	t.status = newStatus
	t.updatedAt = time.Now()
	t.addEvent(NewTodoUpdatedEvent(t.id))

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

	t.status = StatusCompleted
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
	t.status = StatusPending
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

	t.status = StatusCancelled
	t.updatedAt = time.Now()
	t.addEvent(NewTodoUpdatedEvent(t.id))

	return nil
}

// MarkInProgress marks the todo as in progress
func (t *Todo) MarkInProgress() error {
	if t.status.IsCompleted() {
		return ErrCannotModifyCompleted
	}

	if t.status == StatusInProgress {
		return nil // Already in progress, idempotent
	}

	t.status = StatusInProgress
	t.updatedAt = time.Now()
	t.addEvent(NewTodoUpdatedEvent(t.id))

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
