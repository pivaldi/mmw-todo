//go:generate go-enum --marshal --sql --values
package domain

import (
	"strings"
	"time"

	"github.com/google/uuid"
)

// ENUM(pending, in_progress, completed, cancelled)
type TaskStatus string

// TodoID is a unique identifier for a Todo aggregate
type TodoID string

// NewTodoID creates a new unique TodoID
func NewTodoID() TodoID {
	return TodoID(uuid.New().String())
}

// ParseTodoID parses a string into a TodoID with validation
func ParseTodoID(id string) (TodoID, error) {
	if id == "" {
		return "", ErrInvalidID
	}
	// Validate UUID format
	if _, err := uuid.Parse(id); err != nil {
		return "", ErrInvalidID
	}

	return TodoID(id), nil
}

// String returns the string representation of TodoID
func (id TodoID) String() string {
	return string(id)
}

// IsEmpty checks if the TodoID is empty
func (id TodoID) IsEmpty() bool {
	return string(id) == ""
}

// TaskTitle represents a validated todo title
type TaskTitle struct {
	value string
}

// NewTaskTitle creates a new TaskTitle with validation
func NewTaskTitle(title string) (TaskTitle, error) {
	// Trim whitespace
	trimmed := strings.TrimSpace(title)

	// Validate length
	if trimmed == "" {
		return TaskTitle{}, NewValidationError("title", "cannot be empty")
	}
	if len(trimmed) > 200 {
		return TaskTitle{}, NewValidationError("title", "cannot exceed 200 characters")
	}

	return TaskTitle{value: trimmed}, nil
}

// String returns the string value of the title
func (t TaskTitle) String() string {
	return t.value
}

// NewTaskStatus creates a TaskStatus from a string with validation
func NewTaskStatus(status string) (TaskStatus, error) {
	return ParseTaskStatus(status)
}

// IsCompleted checks if the status is completed
func (s TaskStatus) IsCompleted() bool {
	return s == TaskStatusCompleted
}

// IsCancelled checks if the status is cancelled
func (s TaskStatus) IsCancelled() bool {
	return s == TaskStatusCancelled
}

// CanTransitionTo checks if transition to new status is valid
func (s TaskStatus) CanTransitionTo(newStatus TaskStatus) bool {
	// Completed tasks can only be reopened to pending
	if s == TaskStatusCompleted && newStatus != TaskStatusPending {
		return false
	}

	// Cancelled tasks can be reopened to pending
	if s == TaskStatusCancelled && newStatus == TaskStatusCompleted {
		return false
	}

	return true
}

// ENUM(low, medium, high, urgent)
type Priority string

// NewPriority creates a Priority from a string with validation
func NewPriority(priority string) (Priority, error) {
	return ParsePriority(priority)
}

// DefaultPriority returns the default priority (Medium)
func DefaultPriority() Priority {
	return PriorityMedium
}

// DueDate represents a validated due date that must be in the future
type DueDate struct {
	value time.Time
}

// NewDueDate creates a new DueDate with validation
func NewDueDate(date time.Time) (DueDate, error) {
	// Due date must be in the future
	if !date.After(time.Now()) {
		return DueDate{}, ErrInvalidDueDate
	}

	return DueDate{value: date}, nil
}

// Time returns the time.Time value
func (d DueDate) Time() time.Time {
	return d.value
}

// IsApproaching checks if due date is within the given duration
func (d DueDate) IsApproaching(within time.Duration) bool {
	return time.Until(d.value) <= within
}

// IsPast checks if the due date has passed
func (d DueDate) IsPast() bool {
	return time.Now().After(d.value)
}
