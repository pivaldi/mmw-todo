package domain

import (
	"testing"
	"time"
)

// Helper functions for tests

func createValidTodo(t *testing.T) *Todo {
	t.Helper()
	title, _ := NewTaskTitle("Test Todo")
	return NewTodo(title, "Test description", PriorityMedium, nil)
}

func createTodoWithStatus(t *testing.T, status TaskStatus) *Todo {
	t.Helper()
	todo := createValidTodo(t)
	// Directly set status for testing (bypass business rules)
	todo.status = status
	if status == StatusCompleted {
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

	todo := NewTodo(title, description, priority, &dueDate)

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
	if todo.Status() != StatusPending {
		t.Errorf("Status = %v, want %v", todo.Status(), StatusPending)
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
	if events[0].EventType() != "TodoCreated" {
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
			initialStatus: StatusPending,
			wantErr:       false,
			wantEvent:     true,
		},
		{
			name:          "complete in_progress todo",
			initialStatus: StatusInProgress,
			wantErr:       false,
			wantEvent:     true,
		},
		{
			name:          "complete already completed todo (idempotent)",
			initialStatus: StatusCompleted,
			wantErr:       false,
			wantEvent:     false,
		},
		{
			name:          "complete cancelled todo",
			initialStatus: StatusCancelled,
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
				if todo.Status() != StatusCompleted {
					t.Errorf("Status = %v, want %v", todo.Status(), StatusCompleted)
				}
				if todo.CompletedAt() == nil {
					t.Error("CompletedAt should be set after completing")
				}

				if tt.wantEvent {
					events := todo.Events()
					if len(events) == 0 {
						t.Error("Expected TodoCompleted event")
					}
					if events[0].EventType() != "TodoCompleted" {
						t.Errorf("Expected TodoCompleted event, got %s", events[0].EventType())
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
			initialStatus:  StatusCompleted,
			expectedStatus: StatusPending,
			wantEvent:      true,
		},
		{
			name:           "reopen cancelled todo",
			initialStatus:  StatusCancelled,
			expectedStatus: StatusPending,
			wantEvent:      true,
		},
		{
			name:           "reopen pending todo (idempotent)",
			initialStatus:  StatusPending,
			expectedStatus: StatusPending,
			wantEvent:      false,
		},
		{
			name:           "reopen in_progress todo (idempotent)",
			initialStatus:  StatusInProgress,
			expectedStatus: StatusInProgress,
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

			if todo.CompletedAt() != nil && (tt.initialStatus == StatusCompleted || tt.initialStatus == StatusCancelled) {
				t.Error("CompletedAt should be nil after reopening completed/cancelled")
			}

			if tt.wantEvent {
				events := todo.Events()
				if len(events) == 0 {
					t.Error("Expected TodoReopened event")
				}
				if events[0].EventType() != "TodoReopened" {
					t.Errorf("Expected TodoReopened event, got %s", events[0].EventType())
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
			initialStatus: StatusPending,
			newTitle:      "Updated title",
			wantErr:       false,
		},
		{
			name:          "update completed todo title",
			initialStatus: StatusCompleted,
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
			from:    StatusPending,
			to:      StatusInProgress,
			wantErr: false,
		},
		{
			name:    "in_progress to completed",
			from:    StatusInProgress,
			to:      StatusCompleted,
			wantErr: false,
		},
		{
			name:    "completed to in_progress (invalid)",
			from:    StatusCompleted,
			to:      StatusInProgress,
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
			initialStatus: StatusPending,
			wantErr:       false,
		},
		{
			name:          "cancel in_progress todo",
			initialStatus: StatusInProgress,
			wantErr:       false,
		},
		{
			name:          "cancel completed todo",
			initialStatus: StatusCompleted,
			wantErr:       true,
		},
		{
			name:          "cancel already cancelled todo (idempotent)",
			initialStatus: StatusCancelled,
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
				if todo.Status() != StatusCancelled {
					t.Errorf("Status = %v, want %v", todo.Status(), StatusCancelled)
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

	if todo.Status() != StatusInProgress {
		t.Errorf("Status = %v, want %v", todo.Status(), StatusInProgress)
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
			todo := NewTodo(title, "", PriorityMedium, tt.dueDate)

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
			todo := NewTodo(title, "", PriorityMedium, tt.dueDate)

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

// TestReconstituteTodo tests reconstituting a todo from stored data
func TestReconstituteTodo(t *testing.T) {
	id, _ := ParseTodoID("a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11")
	title, _ := NewTaskTitle("Reconstituted Todo")
	description := "From database"
	status := StatusInProgress
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
