package domain

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

// Helper functions for tests

func createValidTodo(t *testing.T) *Todo {
	t.Helper()
	title, _ := NewTaskTitle("Test Todo")

	return New(title, "Test description", PriorityMedium, nil, uuid.Nil)
}

func createTodoWithStatus(t *testing.T, status TaskStatus) *Todo {
	t.Helper()
	todo := createValidTodo(t)
	// Directly set status for testing (bypass business rules)
	todo.status = status
	if status == TaskStatusCompleted {
		now := time.Now()
		todo.completedAt = &now
	}
	todo.ClearEvents() // Clear creation event for cleaner testing
	return todo
}

// TestNewTodo tests todo creation
func TestNewTodo(t *testing.T) {
	title, _ := NewTaskTitle("Buy groceries")
	description := "Milk, eggs, bread"
	priority := PriorityHigh
	futureDate := time.Now().Add(24 * time.Hour)
	dueDate, _ := NewDueDate(futureDate)

	todo := New(title, description, priority, &dueDate, uuid.Nil)

	// Verify initial state
	if todo.ID().IsEmpty() {
		t.Error("NewTodo() should generate a valid ID")
	}
	if todo.Title().String() != "Buy groceries" {
		t.Errorf("Title = %v, want %v", todo.Title().String(), "Buy groceries")
	}
	if todo.Description() != description {
		t.Errorf("Description = %v, want %v", todo.Description(), description)
	}
	if todo.Status() != TaskStatusPending {
		t.Errorf("Status = %v, want %v", todo.Status(), TaskStatusPending)
	}
	if todo.Priority() != priority {
		t.Errorf("Priority = %v, want %v", todo.Priority(), priority)
	}
	if todo.DueDate() == nil {
		t.Error("DueDate should not be nil")
	}
	if todo.CompletedAt() != nil {
		t.Error("CompletedAt should be nil for new todo")
	}

	// Verify TodoCreated event was emitted
	events := todo.Events()
	if len(events) != 1 {
		t.Errorf("Expected 1 event, got %d", len(events))
	}
	if events[0].EventType() != "todo.created" {
		t.Errorf("Expected TodoCreated event, got %s", events[0].EventType())
	}
}

// TestTodo_Complete tests completing a todo
func TestTodo_Complete(t *testing.T) {
	tests := []struct {
		name          string
		initialStatus TaskStatus
		wantErr       bool
		wantEvent     bool
	}{
		{
			name:          "complete pending todo",
			initialStatus: TaskStatusPending,
			wantErr:       false,
			wantEvent:     true,
		},
		{
			name:          "complete in_progress todo",
			initialStatus: TaskStatusInProgress,
			wantErr:       false,
			wantEvent:     true,
		},
		{
			name:          "complete already completed todo (idempotent)",
			initialStatus: TaskStatusCompleted,
			wantErr:       false,
			wantEvent:     false,
		},
		{
			name:          "complete cancelled todo",
			initialStatus: TaskStatusCancelled,
			wantErr:       true,
			wantEvent:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			todo := createTodoWithStatus(t, tt.initialStatus)

			err := todo.Complete()

			if tt.wantErr {
				if err == nil {
					t.Error("Complete() expected error but got nil")
				}
			} else {
				if err != nil {
					t.Errorf("Complete() unexpected error: %v", err)
				}
				if todo.Status() != TaskStatusCompleted {
					t.Errorf("Status = %v, want %v", todo.Status(), TaskStatusCompleted)
				}
				if todo.CompletedAt() == nil {
					t.Error("CompletedAt should be set after completing")
				}

				if tt.wantEvent {
					events := todo.Events()
					if len(events) == 0 {
						t.Error("Expected todo.completed event")
					}
					if events[0].EventType() != "todo.completed" {
						t.Errorf("Expected todo.completed event, got %s", events[0].EventType())
					}
				}
			}
		})
	}
}

// TestTodo_Reopen tests reopening a todo
func TestTodo_Reopen(t *testing.T) {
	tests := []struct {
		name           string
		initialStatus  TaskStatus
		expectedStatus TaskStatus
		wantEvent      bool
	}{
		{
			name:           "reopen completed todo",
			initialStatus:  TaskStatusCompleted,
			expectedStatus: TaskStatusPending,
			wantEvent:      true,
		},
		{
			name:           "reopen cancelled todo",
			initialStatus:  TaskStatusCancelled,
			expectedStatus: TaskStatusPending,
			wantEvent:      true,
		},
		{
			name:           "reopen pending todo (idempotent)",
			initialStatus:  TaskStatusPending,
			expectedStatus: TaskStatusPending,
			wantEvent:      false,
		},
		{
			name:           "reopen in_progress todo (idempotent)",
			initialStatus:  TaskStatusInProgress,
			expectedStatus: TaskStatusInProgress,
			wantEvent:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			todo := createTodoWithStatus(t, tt.initialStatus)

			err := todo.Reopen()
			if err != nil {
				t.Errorf("Reopen() unexpected error: %v", err)
			}

			if todo.Status() != tt.expectedStatus {
				t.Errorf("Status = %v, want %v", todo.Status(), tt.expectedStatus)
			}

			if todo.CompletedAt() != nil && (tt.initialStatus == TaskStatusCompleted || tt.initialStatus == TaskStatusCancelled) {
				t.Error("CompletedAt should be nil after reopening completed/cancelled")
			}

			if tt.wantEvent {
				events := todo.Events()
				if len(events) == 0 {
					t.Error("Expected todo.reopened event")
				}
				if events[0].EventType() != "todo.reopened" {
					t.Errorf("Expected todo.reopened event, got %s", events[0].EventType())
				}
			}
		})
	}
}

// TestTodo_UpdateTitle tests updating the title
func TestTodo_UpdateTitle(t *testing.T) {
	tests := []struct {
		name          string
		initialStatus TaskStatus
		newTitle      string
		wantErr       bool
	}{
		{
			name:          "update pending todo title",
			initialStatus: TaskStatusPending,
			newTitle:      "Updated title",
			wantErr:       false,
		},
		{
			name:          "update completed todo title",
			initialStatus: TaskStatusCompleted,
			newTitle:      "Updated title",
			wantErr:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			todo := createTodoWithStatus(t, tt.initialStatus)
			newTitle, _ := NewTaskTitle(tt.newTitle)

			err := todo.UpdateTitle(newTitle)

			if tt.wantErr {
				if err == nil {
					t.Error("UpdateTitle() expected error but got nil")
				}
			} else {
				if err != nil {
					t.Errorf("UpdateTitle() unexpected error: %v", err)
				}
				if todo.Title().String() != tt.newTitle {
					t.Errorf("Title = %v, want %v", todo.Title().String(), tt.newTitle)
				}

				events := todo.Events()
				if len(events) == 0 {
					t.Error("Expected TodoUpdated event")
				}
			}
		})
	}
}

// TestTodo_UpdateDescription tests updating the description
func TestTodo_UpdateDescription(t *testing.T) {
	todo := createValidTodo(t)
	todo.ClearEvents()

	newDescription := "New description"
	err := todo.UpdateDescription(newDescription)
	if err != nil {
		t.Errorf("UpdateDescription() unexpected error: %v", err)
	}

	if todo.Description() != newDescription {
		t.Errorf("Description = %v, want %v", todo.Description(), newDescription)
	}

	// Verify event
	events := todo.Events()
	if len(events) != 1 {
		t.Errorf("Expected 1 event, got %d", len(events))
	}
}

// TestTodo_UpdatePriority tests updating the priority
func TestTodo_UpdatePriority(t *testing.T) {
	todo := createValidTodo(t)
	todo.ClearEvents()

	newPriority := PriorityUrgent
	err := todo.UpdatePriority(newPriority)
	if err != nil {
		t.Errorf("UpdatePriority() unexpected error: %v", err)
	}

	if todo.Priority() != newPriority {
		t.Errorf("Priority = %v, want %v", todo.Priority(), newPriority)
	}

	// Verify event
	events := todo.Events()
	if len(events) != 1 {
		t.Errorf("Expected 1 event, got %d", len(events))
	}
}

// TestTodo_UpdateDueDate tests updating the due date
func TestTodo_UpdateDueDate(t *testing.T) {
	todo := createValidTodo(t)
	todo.ClearEvents()

	futureDate := time.Now().Add(48 * time.Hour)
	newDueDate, _ := NewDueDate(futureDate)

	err := todo.UpdateDueDate(&newDueDate)
	if err != nil {
		t.Errorf("UpdateDueDate() unexpected error: %v", err)
	}

	if todo.DueDate() == nil {
		t.Error("DueDate should not be nil")
	}

	// Verify event
	events := todo.Events()
	if len(events) != 1 {
		t.Errorf("Expected 1 event, got %d", len(events))
	}
}

// TestTodo_UpdateStatus tests status updates with validation
func TestTodo_UpdateStatus(t *testing.T) {
	tests := []struct {
		name    string
		from    TaskStatus
		to      TaskStatus
		wantErr bool
	}{
		{
			name:    "pending to in_progress",
			from:    TaskStatusPending,
			to:      TaskStatusInProgress,
			wantErr: false,
		},
		{
			name:    "in_progress to completed",
			from:    TaskStatusInProgress,
			to:      TaskStatusCompleted,
			wantErr: false,
		},
		{
			name:    "completed to in_progress (invalid)",
			from:    TaskStatusCompleted,
			to:      TaskStatusInProgress,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			todo := createTodoWithStatus(t, tt.from)

			err := todo.UpdateStatus(tt.to)

			if tt.wantErr {
				if err == nil {
					t.Error("UpdateStatus() expected error but got nil")
				}
			} else {
				if err != nil {
					t.Errorf("UpdateStatus() unexpected error: %v", err)
				}
				if todo.Status() != tt.to {
					t.Errorf("Status = %v, want %v", todo.Status(), tt.to)
				}
			}
		})
	}
}

// TestTodo_Cancel tests cancelling a todo
func TestTodo_Cancel(t *testing.T) {
	tests := []struct {
		name          string
		initialStatus TaskStatus
		wantErr       bool
	}{
		{
			name:          "cancel pending todo",
			initialStatus: TaskStatusPending,
			wantErr:       false,
		},
		{
			name:          "cancel in_progress todo",
			initialStatus: TaskStatusInProgress,
			wantErr:       false,
		},
		{
			name:          "cancel completed todo",
			initialStatus: TaskStatusCompleted,
			wantErr:       true,
		},
		{
			name:          "cancel already cancelled todo (idempotent)",
			initialStatus: TaskStatusCancelled,
			wantErr:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			todo := createTodoWithStatus(t, tt.initialStatus)

			err := todo.Cancel()

			if tt.wantErr {
				if err == nil {
					t.Error("Cancel() expected error but got nil")
				}
			} else {
				if err != nil {
					t.Errorf("Cancel() unexpected error: %v", err)
				}
				if todo.Status() != TaskStatusCancelled {
					t.Errorf("Status = %v, want %v", todo.Status(), TaskStatusCancelled)
				}
			}
		})
	}
}

// TestTodo_MarkInProgress tests marking todo as in progress
func TestTodo_MarkInProgress(t *testing.T) {
	todo := createValidTodo(t)
	todo.ClearEvents()

	err := todo.MarkInProgress()
	if err != nil {
		t.Errorf("MarkInProgress() unexpected error: %v", err)
	}

	if todo.Status() != TaskStatusInProgress {
		t.Errorf("Status = %v, want %v", todo.Status(), TaskStatusInProgress)
	}

	events := todo.Events()
	if len(events) != 1 {
		t.Errorf("Expected 1 event, got %d", len(events))
	}
}

// TestTodo_IsDue tests checking if todo is due
func TestTodo_IsDue(t *testing.T) {
	tests := []struct {
		name    string
		dueDate *DueDate
		want    bool
	}{
		{
			name:    "no due date",
			dueDate: nil,
			want:    false,
		},
		{
			name: "future due date",
			dueDate: func() *DueDate {
				d, _ := NewDueDate(time.Now().Add(24 * time.Hour))
				return &d
			}(),
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			title, _ := NewTaskTitle("Test")
			todo := New(title, "", PriorityMedium, tt.dueDate, uuid.Nil)

			if got := todo.IsDue(); got != tt.want {
				t.Errorf("IsDue() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestTodo_IsDueSoon tests checking if todo is due soon
func TestTodo_IsDueSoon(t *testing.T) {
	tests := []struct {
		name    string
		dueDate *DueDate
		within  time.Duration
		want    bool
	}{
		{
			name:    "no due date",
			dueDate: nil,
			within:  24 * time.Hour,
			want:    false,
		},
		{
			name: "due in 1 hour, checking within 2 hours",
			dueDate: func() *DueDate {
				d, _ := NewDueDate(time.Now().Add(1 * time.Hour))
				return &d
			}(),
			within: 2 * time.Hour,
			want:   true,
		},
		{
			name: "due in 3 hours, checking within 2 hours",
			dueDate: func() *DueDate {
				d, _ := NewDueDate(time.Now().Add(3 * time.Hour))
				return &d
			}(),
			within: 2 * time.Hour,
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			title, _ := NewTaskTitle("Test")
			todo := New(title, "", PriorityMedium, tt.dueDate, uuid.Nil)

			if got := todo.IsDueSoon(tt.within); got != tt.want {
				t.Errorf("IsDueSoon() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestTodo_ClearEvents tests clearing events
func TestTodo_ClearEvents(t *testing.T) {
	todo := createValidTodo(t)

	if len(todo.Events()) == 0 {
		t.Error("New todo should have TodoCreated event")
	}

	todo.ClearEvents()

	if len(todo.Events()) != 0 {
		t.Errorf("After ClearEvents(), expected 0 events, got %d", len(todo.Events()))
	}
}

// In package domain — no import qualifier needed for domain symbols.
func TestNewTodo_StoresUserID(t *testing.T) {
	userID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	title, _ := NewTaskTitle("Test")
	todo := New(title, "", PriorityMedium, nil, userID)
	assert.Equal(t, userID, todo.UserID())
}

func TestReconstituteTodo_StoresUserID(t *testing.T) {
	userID := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	id := NewTodoID()
	title, _ := NewTaskTitle("Test")
	todo := ReconstituteTodo(id, title, "", TaskStatusPending, PriorityMedium, nil, time.Now(), time.Now(), nil, userID)
	assert.Equal(t, userID, todo.UserID())
}

// TestReconstituteTodo tests reconstituting a todo from stored data
func TestReconstituteTodo(t *testing.T) {
	id, _ := ParseTodoID("a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11")
	title, _ := NewTaskTitle("Reconstituted Todo")
	description := "From database"
	status := TaskStatusInProgress
	priority := PriorityHigh
	createdAt := time.Now().Add(-24 * time.Hour)
	updatedAt := time.Now()

	todo := ReconstituteTodo(
		id,
		title,
		description,
		status,
		priority,
		nil,
		createdAt,
		updatedAt,
		nil,
		uuid.Nil,
	)

	// Verify all fields
	if todo.ID() != id {
		t.Errorf("ID = %v, want %v", todo.ID(), id)
	}
	if todo.Title().String() != title.String() {
		t.Errorf("Title = %v, want %v", todo.Title().String(), title.String())
	}
	if todo.Description() != description {
		t.Errorf("Description = %v, want %v", todo.Description(), description)
	}
	if todo.Status() != status {
		t.Errorf("Status = %v, want %v", todo.Status(), status)
	}
	if todo.Priority() != priority {
		t.Errorf("Priority = %v, want %v", todo.Priority(), priority)
	}

	// Reconstituted todos should not have events
	if len(todo.Events()) != 0 {
		t.Errorf("Reconstituted todo should have 0 events, got %d", len(todo.Events()))
	}
}
