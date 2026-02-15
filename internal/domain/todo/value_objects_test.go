package domain

import (
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
)

// TestNewTodoID tests TodoID creation
func TestNewTodoID(t *testing.T) {
	id := NewTodoID()

	if id.IsEmpty() {
		t.Error("NewTodoID() should not return empty ID")
	}

	// Should be valid UUID
	if _, err := uuid.Parse(id.String()); err != nil {
		t.Errorf("NewTodoID() should return valid UUID, got error: %v", err)
	}
}

// TestParseTodoID tests TodoID parsing and validation
func TestParseTodoID(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name:    "valid UUID",
			input:   "a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11",
			wantErr: false,
		},
		{
			name:    "empty string",
			input:   "",
			wantErr: true,
		},
		{
			name:    "invalid UUID format",
			input:   "not-a-uuid",
			wantErr: true,
		},
		{
			name:    "partial UUID",
			input:   "a0eebc99",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id, err := ParseTodoID(tt.input)

			if tt.wantErr {
				if err == nil {
					t.Error("ParseTodoID() expected error but got nil")
				}
				if !id.IsEmpty() {
					t.Error("ParseTodoID() should return empty ID on error")
				}
			} else {
				if err != nil {
					t.Errorf("ParseTodoID() unexpected error: %v", err)
				}
				if id.String() != tt.input {
					t.Errorf("ParseTodoID() = %v, want %v", id.String(), tt.input)
				}
			}
		})
	}
}

// TestTaskTitle tests TaskTitle validation
func TestNewTaskTitle(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{
			name:    "valid title",
			input:   "Buy groceries",
			want:    "Buy groceries",
			wantErr: false,
		},
		{
			name:    "title with leading and trailing spaces",
			input:   "  Buy groceries  ",
			want:    "Buy groceries",
			wantErr: false,
		},
		{
			name:    "single character",
			input:   "A",
			want:    "A",
			wantErr: false,
		},
		{
			name:    "200 characters (max)",
			input:   strings.Repeat("a", 200),
			want:    strings.Repeat("a", 200),
			wantErr: false,
		},
		{
			name:    "empty string",
			input:   "",
			wantErr: true,
		},
		{
			name:    "only whitespace",
			input:   "   ",
			wantErr: true,
		},
		{
			name:    "201 characters (exceeds max)",
			input:   strings.Repeat("a", 201),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			title, err := NewTaskTitle(tt.input)

			if tt.wantErr {
				if err == nil {
					t.Error("NewTaskTitle() expected error but got nil")
				}
			} else {
				if err != nil {
					t.Errorf("NewTaskTitle() unexpected error: %v", err)
				}
				if title.String() != tt.want {
					t.Errorf("NewTaskTitle() = %q, want %q", title.String(), tt.want)
				}
			}
		})
	}
}

// TestTaskStatus tests TaskStatus validation and methods
func TestNewTaskStatus(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    TaskStatus
		wantErr bool
	}{
		{
			name:    "pending lowercase",
			input:   "pending",
			want:    StatusPending,
			wantErr: false,
		},
		{
			name:    "pending uppercase",
			input:   "PENDING",
			want:    StatusPending,
			wantErr: false,
		},
		{
			name:    "in_progress",
			input:   "in_progress",
			want:    StatusInProgress,
			wantErr: false,
		},
		{
			name:    "completed",
			input:   "completed",
			want:    StatusCompleted,
			wantErr: false,
		},
		{
			name:    "cancelled",
			input:   "cancelled",
			want:    StatusCancelled,
			wantErr: false,
		},
		{
			name:    "invalid status",
			input:   "invalid",
			wantErr: true,
		},
		{
			name:    "empty string",
			input:   "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status, err := NewTaskStatus(tt.input)

			if tt.wantErr {
				if err == nil {
					t.Error("NewTaskStatus() expected error but got nil")
				}
			} else {
				if err != nil {
					t.Errorf("NewTaskStatus() unexpected error: %v", err)
				}
				if status != tt.want {
					t.Errorf("NewTaskStatus() = %v, want %v", status, tt.want)
				}
			}
		})
	}
}

// TestTaskStatus_IsCompleted tests IsCompleted method
func TestTaskStatus_IsCompleted(t *testing.T) {
	tests := []struct {
		status TaskStatus
		want   bool
	}{
		{StatusCompleted, true},
		{StatusPending, false},
		{StatusInProgress, false},
		{StatusCancelled, false},
	}

	for _, tt := range tests {
		t.Run(tt.status.String(), func(t *testing.T) {
			if got := tt.status.IsCompleted(); got != tt.want {
				t.Errorf("IsCompleted() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestTaskStatus_IsCancelled tests IsCancelled method
func TestTaskStatus_IsCancelled(t *testing.T) {
	tests := []struct {
		status TaskStatus
		want   bool
	}{
		{StatusCancelled, true},
		{StatusPending, false},
		{StatusInProgress, false},
		{StatusCompleted, false},
	}

	for _, tt := range tests {
		t.Run(tt.status.String(), func(t *testing.T) {
			if got := tt.status.IsCancelled(); got != tt.want {
				t.Errorf("IsCancelled() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestTaskStatus_CanTransitionTo tests status transition rules
func TestTaskStatus_CanTransitionTo(t *testing.T) {
	tests := []struct {
		name      string
		from      TaskStatus
		to        TaskStatus
		wantValid bool
	}{
		{
			name:      "pending to in_progress",
			from:      StatusPending,
			to:        StatusInProgress,
			wantValid: true,
		},
		{
			name:      "pending to completed",
			from:      StatusPending,
			to:        StatusCompleted,
			wantValid: true,
		},
		{
			name:      "in_progress to completed",
			from:      StatusInProgress,
			to:        StatusCompleted,
			wantValid: true,
		},
		{
			name:      "completed to pending (reopen)",
			from:      StatusCompleted,
			to:        StatusPending,
			wantValid: true,
		},
		{
			name:      "completed to in_progress (invalid)",
			from:      StatusCompleted,
			to:        StatusInProgress,
			wantValid: false,
		},
		{
			name:      "completed to cancelled (invalid)",
			from:      StatusCompleted,
			to:        StatusCancelled,
			wantValid: false,
		},
		{
			name:      "cancelled to pending (reopen)",
			from:      StatusCancelled,
			to:        StatusPending,
			wantValid: true,
		},
		{
			name:      "cancelled to completed (invalid)",
			from:      StatusCancelled,
			to:        StatusCompleted,
			wantValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.from.CanTransitionTo(tt.to); got != tt.wantValid {
				t.Errorf("CanTransitionTo() = %v, want %v", got, tt.wantValid)
			}
		})
	}
}

// TestPriority tests Priority validation
func TestNewPriority(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    Priority
		wantErr bool
	}{
		{
			name:    "low lowercase",
			input:   "low",
			want:    PriorityLow,
			wantErr: false,
		},
		{
			name:    "low uppercase",
			input:   "LOW",
			want:    PriorityLow,
			wantErr: false,
		},
		{
			name:    "medium",
			input:   "medium",
			want:    PriorityMedium,
			wantErr: false,
		},
		{
			name:    "high",
			input:   "high",
			want:    PriorityHigh,
			wantErr: false,
		},
		{
			name:    "urgent",
			input:   "urgent",
			want:    PriorityUrgent,
			wantErr: false,
		},
		{
			name:    "invalid priority",
			input:   "critical",
			wantErr: true,
		},
		{
			name:    "empty string",
			input:   "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			priority, err := NewPriority(tt.input)

			if tt.wantErr {
				if err == nil {
					t.Error("NewPriority() expected error but got nil")
				}
			} else {
				if err != nil {
					t.Errorf("NewPriority() unexpected error: %v", err)
				}
				if priority != tt.want {
					t.Errorf("NewPriority() = %v, want %v", priority, tt.want)
				}
			}
		})
	}
}

// TestDefaultPriority tests default priority
func TestDefaultPriority(t *testing.T) {
	if got := DefaultPriority(); got != PriorityMedium {
		t.Errorf("DefaultPriority() = %v, want %v", got, PriorityMedium)
	}
}

// TestDueDate tests DueDate validation
func TestNewDueDate(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name    string
		input   time.Time
		wantErr bool
	}{
		{
			name:    "future date (1 hour)",
			input:   now.Add(1 * time.Hour),
			wantErr: false,
		},
		{
			name:    "future date (1 day)",
			input:   now.Add(24 * time.Hour),
			wantErr: false,
		},
		{
			name:    "past date",
			input:   now.Add(-1 * time.Hour),
			wantErr: true,
		},
		{
			name:    "current time (edge case)",
			input:   now,
			wantErr: true, // Must be AFTER now, not equal
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dueDate, err := NewDueDate(tt.input)

			if tt.wantErr {
				if err == nil {
					t.Error("NewDueDate() expected error but got nil")
				}
			} else {
				if err != nil {
					t.Errorf("NewDueDate() unexpected error: %v", err)
				}
				if !dueDate.Time().Equal(tt.input) {
					t.Errorf("NewDueDate().Time() = %v, want %v", dueDate.Time(), tt.input)
				}
			}
		})
	}
}

// TestDueDate_IsApproaching tests IsApproaching method
func TestDueDate_IsApproaching(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name   string
		date   time.Time
		within time.Duration
		want   bool
	}{
		{
			name:   "due in 1 hour, checking within 2 hours",
			date:   now.Add(1 * time.Hour),
			within: 2 * time.Hour,
			want:   true,
		},
		{
			name:   "due in 3 hours, checking within 2 hours",
			date:   now.Add(3 * time.Hour),
			within: 2 * time.Hour,
			want:   false,
		},
		{
			name:   "due in exactly 2 hours, checking within 2 hours",
			date:   now.Add(2 * time.Hour),
			within: 2 * time.Hour,
			want:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dueDate, _ := NewDueDate(tt.date)
			if got := dueDate.IsApproaching(tt.within); got != tt.want {
				t.Errorf("IsApproaching() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestDueDate_IsPast tests IsPast method
func TestDueDate_IsPast(t *testing.T) {
	now := time.Now()

	// Create a due date in the future
	futureDate, _ := NewDueDate(now.Add(1 * time.Hour))

	if futureDate.IsPast() {
		t.Error("IsPast() should return false for future date")
	}

	// Note: We can't easily test past dates because NewDueDate won't allow them
	// This is correct behavior - once created, a DueDate is in the future
	// It only becomes past as time progresses
}
