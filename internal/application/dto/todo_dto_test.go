package dto

import (
	"testing"
	"time"

	"github.com/google/uuid"

	domain "github.com/pivaldi/mmw-todo/internal/domain/todo"
)

// TestMapTodoToResponse tests mapping a domain Todo to TodoResponse DTO
func TestMapTodoToResponse(t *testing.T) {
	// Create a domain todo
	title, _ := domain.NewTaskTitle("Test Todo")
	futureDate := time.Now().Add(24 * time.Hour)
	dueDate, _ := domain.NewDueDate(futureDate)

	todo := domain.NewTodo(title, "Test description", domain.PriorityHigh, &dueDate, uuid.Nil)

	// Map to response
	response := MapTodoToResponse(todo)

	// Verify all fields are mapped correctly
	if response.ID != todo.ID().String() {
		t.Errorf("ID = %v, want %v", response.ID, todo.ID().String())
	}

	if response.Title != "Test Todo" {
		t.Errorf("Title = %v, want %v", response.Title, "Test Todo")
	}

	if response.Description != "Test description" {
		t.Errorf("Description = %v, want %v", response.Description, "Test description")
	}

	if response.Status != domain.TaskStatusPending {
		t.Errorf("Status = %v, want %v", response.Status, domain.TaskStatusPending)
	}

	if response.Priority != domain.PriorityHigh {
		t.Errorf("Priority = %v, want %v", response.Priority, domain.PriorityHigh)
	}

	if response.DueDate == nil {
		t.Error("DueDate should not be nil")
	} else if !response.DueDate.Equal(futureDate) {
		t.Errorf("DueDate = %v, want %v", response.DueDate, futureDate)
	}

	if response.CreatedAt.IsZero() {
		t.Error("CreatedAt should not be zero")
	}

	if response.UpdatedAt.IsZero() {
		t.Error("UpdatedAt should not be zero")
	}
}

// TestMapTodoToResponse_WithoutDueDate tests mapping when DueDate is nil
func TestMapTodoToResponse_WithoutDueDate(t *testing.T) {
	title, _ := domain.NewTaskTitle("Todo without due date")
	todo := domain.NewTodo(title, "No deadline", domain.PriorityLow, nil, uuid.Nil)

	response := MapTodoToResponse(todo)

	if response.DueDate != nil {
		t.Errorf("DueDate should be nil, got %v", response.DueDate)
	}

	// Verify other fields still mapped correctly
	if response.Title != "Todo without due date" {
		t.Errorf("Title = %v, want %v", response.Title, "Todo without due date")
	}

	if response.Description != "No deadline" {
		t.Errorf("Description = %v, want %v", response.Description, "No deadline")
	}

	if response.Priority != domain.PriorityLow {
		t.Errorf("Priority = %v, want %v", response.Priority, domain.PriorityLow)
	}
}

// TestMapTodoToResponse_WithCompletedStatus tests mapping a completed todo
func TestMapTodoToResponse_WithCompletedStatus(t *testing.T) {
	title, _ := domain.NewTaskTitle("Completed Todo")
	todo := domain.NewTodo(title, "Done", domain.PriorityMedium, nil, uuid.Nil)

	// Complete the todo
	_ = todo.Complete()
	todo.ClearEvents() // Clear events for clean testing

	response := MapTodoToResponse(todo)

	if response.Status != domain.TaskStatusCompleted {
		t.Errorf("Status = %v, want %v", response.Status, domain.TaskStatusCompleted)
	}

	// CompletedAt is tracked in domain but not exposed in DTO
	// Verify the DTO doesn't break with completed status
	if response.ID == "" {
		t.Error("ID should not be empty")
	}
}

// TestMapTodosToResponse tests mapping multiple todos
func TestMapTodosToResponse(t *testing.T) {
	// Create multiple domain todos
	title1, _ := domain.NewTaskTitle("Todo 1")
	title2, _ := domain.NewTaskTitle("Todo 2")
	title3, _ := domain.NewTaskTitle("Todo 3")

	todos := []*domain.Todo{
		domain.NewTodo(title1, "First", domain.PriorityLow, nil, uuid.Nil),
		domain.NewTodo(title2, "Second", domain.PriorityMedium, nil, uuid.Nil),
		domain.NewTodo(title3, "Third", domain.PriorityHigh, nil, uuid.Nil),
	}

	// Map to responses
	responses := MapTodosToResponse(todos)

	// Verify count
	if len(responses) != 3 {
		t.Errorf("Expected 3 responses, got %d", len(responses))
	}

	// Verify each mapping
	expectedTitles := []string{"Todo 1", "Todo 2", "Todo 3"}
	expectedPriorities := []domain.Priority{domain.PriorityLow, domain.PriorityMedium, domain.PriorityHigh}

	for i, response := range responses {
		if response.Title != expectedTitles[i] {
			t.Errorf("Response[%d].Title = %v, want %v", i, response.Title, expectedTitles[i])
		}

		if response.Priority != expectedPriorities[i] {
			t.Errorf("Response[%d].Priority = %v, want %v", i, response.Priority, expectedPriorities[i])
		}

		if response.Status != domain.TaskStatusPending {
			t.Errorf("Response[%d].Status = %v, want %v", i, response.Status, domain.TaskStatusPending)
		}

		if response.ID == "" {
			t.Errorf("Response[%d].ID should not be empty", i)
		}
	}
}

// TestMapTodosToResponse_EmptySlice tests mapping an empty slice
func TestMapTodosToResponse_EmptySlice(t *testing.T) {
	todos := []*domain.Todo{}

	responses := MapTodosToResponse(todos)

	if len(responses) != 0 {
		t.Errorf("Expected 0 responses, got %d", len(responses))
	}

	if responses == nil {
		t.Error("Responses should not be nil, should be empty slice")
	}
}

// TestMapTodosToResponse_WithMixedDueDates tests mapping todos with and without due dates
func TestMapTodosToResponse_WithMixedDueDates(t *testing.T) {
	futureDate := time.Now().Add(48 * time.Hour)
	dueDate, _ := domain.NewDueDate(futureDate)

	title1, _ := domain.NewTaskTitle("With due date")
	title2, _ := domain.NewTaskTitle("Without due date")

	todos := []*domain.Todo{
		domain.NewTodo(title1, "Has deadline", domain.PriorityUrgent, &dueDate, uuid.Nil),
		domain.NewTodo(title2, "No deadline", domain.PriorityLow, nil, uuid.Nil),
	}

	responses := MapTodosToResponse(todos)

	// First todo should have due date
	if responses[0].DueDate == nil {
		t.Error("First todo should have due date")
	}

	// Second todo should not have due date
	if responses[1].DueDate != nil {
		t.Error("Second todo should not have due date")
	}
}

// TestMapTodoToResponse_PreservesAllPriorities tests all priority levels are mapped
func TestMapTodoToResponse_PreservesAllPriorities(t *testing.T) {
	priorities := []domain.Priority{
		domain.PriorityLow,
		domain.PriorityMedium,
		domain.PriorityHigh,
		domain.PriorityUrgent,
	}

	for _, priority := range priorities {
		t.Run(priority.String(), func(t *testing.T) {
			title, _ := domain.NewTaskTitle("Test")
			todo := domain.NewTodo(title, "Test", priority, nil, uuid.Nil)

			response := MapTodoToResponse(todo)

			if response.Priority != priority {
				t.Errorf("Priority = %v, want %v", response.Priority, priority)
			}
		})
	}
}

// TestMapTodoToResponse_PreservesAllStatuses tests all status values are mapped
func TestMapTodoToResponse_PreservesAllStatuses(t *testing.T) {
	tests := []struct {
		name   string
		status domain.TaskStatus
		setup  func(*domain.Todo)
	}{
		{
			name:   "pending",
			status: domain.TaskStatusPending,
			setup:  func(t *domain.Todo) {}, // Already pending
		},
		{
			name:   "in_progress",
			status: domain.TaskStatusInProgress,
			setup:  func(t *domain.Todo) { _ = t.MarkInProgress() },
		},
		{
			name:   "completed",
			status: domain.TaskStatusCompleted,
			setup:  func(t *domain.Todo) { _ = t.Complete() },
		},
		{
			name:   "cancelled",
			status: domain.TaskStatusCancelled,
			setup:  func(t *domain.Todo) { _ = t.Cancel() },
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			title, _ := domain.NewTaskTitle("Test")
			todo := domain.NewTodo(title, "Test", domain.PriorityMedium, nil, uuid.Nil)
			tt.setup(todo)
			todo.ClearEvents()

			response := MapTodoToResponse(todo)

			if response.Status != tt.status {
				t.Errorf("Status = %v, want %v", response.Status, tt.status)
			}
		})
	}
}

// TestMapTodoToResponse_TimestampsArePreserved tests CreatedAt and UpdatedAt are mapped
func TestMapTodoToResponse_TimestampsArePreserved(t *testing.T) {
	title, _ := domain.NewTaskTitle("Test Todo")
	todo := domain.NewTodo(title, "Test", domain.PriorityMedium, nil, uuid.Nil)

	// Get original timestamps
	originalCreatedAt := todo.CreatedAt()
	originalUpdatedAt := todo.UpdatedAt()

	response := MapTodoToResponse(todo)

	if !response.CreatedAt.Equal(originalCreatedAt) {
		t.Errorf("CreatedAt = %v, want %v", response.CreatedAt, originalCreatedAt)
	}

	if !response.UpdatedAt.Equal(originalUpdatedAt) {
		t.Errorf("UpdatedAt = %v, want %v", response.UpdatedAt, originalUpdatedAt)
	}
}

// TestMapTodoToResponse_DescriptionCanBeEmpty tests empty description is handled
func TestMapTodoToResponse_DescriptionCanBeEmpty(t *testing.T) {
	title, _ := domain.NewTaskTitle("Todo with no description")
	todo := domain.NewTodo(title, "", domain.PriorityMedium, nil, uuid.Nil)

	response := MapTodoToResponse(todo)

	if response.Description != "" {
		t.Errorf("Description should be empty, got %v", response.Description)
	}

	// Verify other fields still work
	if response.Title != "Todo with no description" {
		t.Errorf("Title = %v, want %v", response.Title, "Todo with no description")
	}
}
