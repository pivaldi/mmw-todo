package application

import (
	"context"
	"errors"
	"testing"
	"time"

	domain "github.com/pivaldi/mmw/todo/internal/domain/todo"
	"github.com/pivaldi/mmw/todo/internal/ports"
)

// Mock implementations

type MockTodoRepository struct {
	SaveFunc     func(ctx context.Context, todo *domain.Todo) error
	FindByIDFunc func(ctx context.Context, id domain.TodoID) (*domain.Todo, error)
	FindAllFunc  func(ctx context.Context, filters ports.Filters) ([]*domain.Todo, error)
	UpdateFunc   func(ctx context.Context, todo *domain.Todo) error
	DeleteFunc   func(ctx context.Context, id domain.TodoID) error
}

func (m *MockTodoRepository) Save(ctx context.Context, todo *domain.Todo) error {
	if m.SaveFunc != nil {
		return m.SaveFunc(ctx, todo)
	}
	return nil
}

func (m *MockTodoRepository) FindByID(ctx context.Context, id domain.TodoID) (*domain.Todo, error) {
	if m.FindByIDFunc != nil {
		return m.FindByIDFunc(ctx, id)
	}
	return nil, domain.ErrTodoNotFound
}

func (m *MockTodoRepository) FindAll(ctx context.Context, filters ports.Filters) ([]*domain.Todo, error) {
	if m.FindAllFunc != nil {
		return m.FindAllFunc(ctx, filters)
	}
	return []*domain.Todo{}, nil
}

func (m *MockTodoRepository) Update(ctx context.Context, todo *domain.Todo) error {
	if m.UpdateFunc != nil {
		return m.UpdateFunc(ctx, todo)
	}
	return nil
}

func (m *MockTodoRepository) Delete(ctx context.Context, id domain.TodoID) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(ctx, id)
	}
	return nil
}

type MockEventDispatcher struct {
	DispatchFunc     func(ctx context.Context, events []domain.DomainEvent) error
	DispatchedEvents []domain.DomainEvent
}

func (m *MockEventDispatcher) Dispatch(ctx context.Context, events []domain.DomainEvent) error {
	m.DispatchedEvents = append(m.DispatchedEvents, events...)
	if m.DispatchFunc != nil {
		return m.DispatchFunc(ctx, events)
	}
	return nil
}

// Test helpers

func createTestTodo() *domain.Todo {
	title, _ := domain.NewTaskTitle("Test Todo")
	return domain.NewTodo(title, "Test description", domain.PriorityMedium, nil)
}

// Tests

func TestTodoService_CreateTodo_ValidRequest_Success(t *testing.T) {
	mockRepo := &MockTodoRepository{}
	mockDispatcher := &MockEventDispatcher{}
	service := NewTodoApplicationService(mockRepo, mockDispatcher)

	req := CreateTodoRequest{
		Title:       "Buy groceries",
		Description: "Milk, eggs, bread",
		Priority:    "medium",
	}

	result, err := service.CreateTodo(context.Background(), req)

	if err != nil {
		t.Fatalf("CreateTodo() unexpected error: %v", err)
	}

	if result == nil {
		t.Fatal("CreateTodo() returned nil result")
	}

	if result.Title != "Buy groceries" {
		t.Errorf("Title = %v, want %v", result.Title, "Buy groceries")
	}

	if result.Description != "Milk, eggs, bread" {
		t.Errorf("Description = %v, want %v", result.Description, "Milk, eggs, bread")
	}

	if result.Priority != "medium" {
		t.Errorf("Priority = %v, want %v", result.Priority, "medium")
	}

	if result.Status != "pending" {
		t.Errorf("Status = %v, want %v", result.Status, "pending")
	}

	// Verify event was dispatched
	if len(mockDispatcher.DispatchedEvents) != 1 {
		t.Errorf("Expected 1 event dispatched, got %d", len(mockDispatcher.DispatchedEvents))
	}

	if mockDispatcher.DispatchedEvents[0].EventType() != "TodoCreated" {
		t.Errorf("Event type = %v, want %v", mockDispatcher.DispatchedEvents[0].EventType(), "TodoCreated")
	}
}

func TestTodoService_CreateTodo_InvalidTitle_ReturnsError(t *testing.T) {
	mockRepo := &MockTodoRepository{}
	mockDispatcher := &MockEventDispatcher{}
	service := NewTodoApplicationService(mockRepo, mockDispatcher)

	req := CreateTodoRequest{
		Title:    "", // Invalid: empty title
		Priority: "medium",
	}

	_, err := service.CreateTodo(context.Background(), req)

	if err == nil {
		t.Error("CreateTodo() expected error for empty title, got nil")
	}
}

func TestTodoService_CreateTodo_InvalidPriority_ReturnsError(t *testing.T) {
	mockRepo := &MockTodoRepository{}
	mockDispatcher := &MockEventDispatcher{}
	service := NewTodoApplicationService(mockRepo, mockDispatcher)

	req := CreateTodoRequest{
		Title:    "Test",
		Priority: "invalid", // Invalid priority
	}

	_, err := service.CreateTodo(context.Background(), req)

	if err == nil {
		t.Error("CreateTodo() expected error for invalid priority, got nil")
	}
}

func TestTodoService_CreateTodo_RepositoryError_ReturnsError(t *testing.T) {
	mockRepo := &MockTodoRepository{
		SaveFunc: func(ctx context.Context, todo *domain.Todo) error {
			return errors.New("database error")
		},
	}
	mockDispatcher := &MockEventDispatcher{}
	service := NewTodoApplicationService(mockRepo, mockDispatcher)

	req := CreateTodoRequest{
		Title:    "Test",
		Priority: "medium",
	}

	_, err := service.CreateTodo(context.Background(), req)

	if err == nil {
		t.Error("CreateTodo() expected error when repository fails, got nil")
	}
}

func TestTodoService_GetTodo_ExistingTodo_Success(t *testing.T) {
	testTodo := createTestTodo()
	mockRepo := &MockTodoRepository{
		FindByIDFunc: func(ctx context.Context, id domain.TodoID) (*domain.Todo, error) {
			return testTodo, nil
		},
	}
	mockDispatcher := &MockEventDispatcher{}
	service := NewTodoApplicationService(mockRepo, mockDispatcher)

	result, err := service.GetTodo(context.Background(), testTodo.ID().String())

	if err != nil {
		t.Fatalf("GetTodo() unexpected error: %v", err)
	}

	if result.ID != testTodo.ID().String() {
		t.Errorf("ID = %v, want %v", result.ID, testTodo.ID().String())
	}
}

func TestTodoService_GetTodo_InvalidID_ReturnsError(t *testing.T) {
	mockRepo := &MockTodoRepository{}
	mockDispatcher := &MockEventDispatcher{}
	service := NewTodoApplicationService(mockRepo, mockDispatcher)

	_, err := service.GetTodo(context.Background(), "invalid-id")

	if err == nil {
		t.Error("GetTodo() expected error for invalid ID, got nil")
	}
}

func TestTodoService_GetTodo_NotFound_ReturnsError(t *testing.T) {
	mockRepo := &MockTodoRepository{
		FindByIDFunc: func(ctx context.Context, id domain.TodoID) (*domain.Todo, error) {
			return nil, domain.ErrTodoNotFound
		},
	}
	mockDispatcher := &MockEventDispatcher{}
	service := NewTodoApplicationService(mockRepo, mockDispatcher)

	validID := domain.NewTodoID()
	_, err := service.GetTodo(context.Background(), validID.String())

	if err == nil {
		t.Error("GetTodo() expected error for non-existent todo, got nil")
	}
}

func TestTodoService_UpdateTodo_UpdateTitle_Success(t *testing.T) {
	testTodo := createTestTodo()
	testTodo.ClearEvents() // Clear creation events

	mockRepo := &MockTodoRepository{
		FindByIDFunc: func(ctx context.Context, id domain.TodoID) (*domain.Todo, error) {
			// Return a fresh copy with cleared events
			title, _ := domain.NewTaskTitle("Test Todo")
			fresh := domain.ReconstituteTodo(
				testTodo.ID(),
				title,
				testTodo.Description(),
				testTodo.Status(),
				testTodo.Priority(),
				testTodo.DueDate(),
				testTodo.CreatedAt(),
				testTodo.UpdatedAt(),
				testTodo.CompletedAt(),
			)
			return fresh, nil
		},
	}
	mockDispatcher := &MockEventDispatcher{}
	service := NewTodoApplicationService(mockRepo, mockDispatcher)

	newTitle := "Updated Title"
	req := UpdateTodoRequest{
		Title: &newTitle,
	}

	result, err := service.UpdateTodo(context.Background(), testTodo.ID().String(), req)

	if err != nil {
		t.Fatalf("UpdateTodo() unexpected error: %v", err)
	}

	if result.Title != newTitle {
		t.Errorf("Title = %v, want %v", result.Title, newTitle)
	}

	// Verify event was dispatched
	if len(mockDispatcher.DispatchedEvents) != 1 {
		t.Errorf("Expected 1 event dispatched, got %d", len(mockDispatcher.DispatchedEvents))
	}
}

func TestTodoService_CompleteTodo_PendingTodo_Success(t *testing.T) {
	testTodo := createTestTodo()
	mockRepo := &MockTodoRepository{
		FindByIDFunc: func(ctx context.Context, id domain.TodoID) (*domain.Todo, error) {
			return testTodo, nil
		},
	}
	mockDispatcher := &MockEventDispatcher{}
	service := NewTodoApplicationService(mockRepo, mockDispatcher)

	result, err := service.CompleteTodo(context.Background(), testTodo.ID().String())

	if err != nil {
		t.Fatalf("CompleteTodo() unexpected error: %v", err)
	}

	if result.Status != "completed" {
		t.Errorf("Status = %v, want %v", result.Status, "completed")
	}

	// Verify TodoCompleted event was dispatched
	foundCompletedEvent := false
	for _, event := range mockDispatcher.DispatchedEvents {
		if event.EventType() == "TodoCompleted" {
			foundCompletedEvent = true
			break
		}
	}

	if !foundCompletedEvent {
		t.Error("Expected TodoCompleted event to be dispatched")
	}
}

func TestTodoService_ReopenTodo_CompletedTodo_Success(t *testing.T) {
	testTodo := createTestTodo()
	testTodo.Complete() // Mark as completed first
	testTodo.ClearEvents()

	mockRepo := &MockTodoRepository{
		FindByIDFunc: func(ctx context.Context, id domain.TodoID) (*domain.Todo, error) {
			return testTodo, nil
		},
	}
	mockDispatcher := &MockEventDispatcher{}
	service := NewTodoApplicationService(mockRepo, mockDispatcher)

	result, err := service.ReopenTodo(context.Background(), testTodo.ID().String())

	if err != nil {
		t.Fatalf("ReopenTodo() unexpected error: %v", err)
	}

	if result.Status != "pending" {
		t.Errorf("Status = %v, want %v", result.Status, "pending")
	}

	// Verify TodoReopened event was dispatched
	foundReopenedEvent := false
	for _, event := range mockDispatcher.DispatchedEvents {
		if event.EventType() == "TodoReopened" {
			foundReopenedEvent = true
			break
		}
	}

	if !foundReopenedEvent {
		t.Error("Expected TodoReopened event to be dispatched")
	}
}

func TestTodoService_DeleteTodo_ExistingTodo_Success(t *testing.T) {
	testTodo := createTestTodo()
	mockRepo := &MockTodoRepository{}
	mockDispatcher := &MockEventDispatcher{}
	service := NewTodoApplicationService(mockRepo, mockDispatcher)

	err := service.DeleteTodo(context.Background(), testTodo.ID().String())

	if err != nil {
		t.Fatalf("DeleteTodo() unexpected error: %v", err)
	}

	// Verify TodoDeleted event was dispatched
	foundDeletedEvent := false
	for _, event := range mockDispatcher.DispatchedEvents {
		if event.EventType() == "TodoDeleted" {
			foundDeletedEvent = true
			break
		}
	}

	if !foundDeletedEvent {
		t.Error("Expected TodoDeleted event to be dispatched")
	}
}

func TestTodoService_ListTodos_NoFilters_ReturnsAll(t *testing.T) {
	testTodo1 := createTestTodo()
	testTodo2 := createTestTodo()

	mockRepo := &MockTodoRepository{
		FindAllFunc: func(ctx context.Context, filters ports.Filters) ([]*domain.Todo, error) {
			return []*domain.Todo{testTodo1, testTodo2}, nil
		},
	}
	mockDispatcher := &MockEventDispatcher{}
	service := NewTodoApplicationService(mockRepo, mockDispatcher)

	result, err := service.ListTodos(context.Background(), ListFilters{})

	if err != nil {
		t.Fatalf("ListTodos() unexpected error: %v", err)
	}

	if len(result.Todos) != 2 {
		t.Errorf("Expected 2 todos, got %d", len(result.Todos))
	}

	if result.TotalCount != 2 {
		t.Errorf("TotalCount = %d, want 2", result.TotalCount)
	}
}

func TestTodoService_ListTodos_WithStatusFilter_FiltersCorrectly(t *testing.T) {
	mockRepo := &MockTodoRepository{
		FindAllFunc: func(ctx context.Context, filters ports.Filters) ([]*domain.Todo, error) {
			// Verify filter was passed correctly
			if filters.Status == nil {
				t.Error("Expected status filter to be set")
			} else if *filters.Status != domain.StatusPending {
				t.Errorf("Status filter = %v, want %v", *filters.Status, domain.StatusPending)
			}
			return []*domain.Todo{}, nil
		},
	}
	mockDispatcher := &MockEventDispatcher{}
	service := NewTodoApplicationService(mockRepo, mockDispatcher)

	statusFilter := "pending"
	_, err := service.ListTodos(context.Background(), ListFilters{
		Status: &statusFilter,
	})

	if err != nil {
		t.Fatalf("ListTodos() unexpected error: %v", err)
	}
}

func TestTodoService_CreateTodo_WithDueDate_Success(t *testing.T) {
	mockRepo := &MockTodoRepository{}
	mockDispatcher := &MockEventDispatcher{}
	service := NewTodoApplicationService(mockRepo, mockDispatcher)

	futureDate := time.Now().Add(24 * time.Hour)
	req := CreateTodoRequest{
		Title:    "Test with due date",
		Priority: "high",
		DueDate:  &futureDate,
	}

	result, err := service.CreateTodo(context.Background(), req)

	if err != nil {
		t.Fatalf("CreateTodo() unexpected error: %v", err)
	}

	if result.DueDate == nil {
		t.Error("Expected due date to be set")
	}
}

func TestTodoService_CreateTodo_WithPastDueDate_ReturnsError(t *testing.T) {
	mockRepo := &MockTodoRepository{}
	mockDispatcher := &MockEventDispatcher{}
	service := NewTodoApplicationService(mockRepo, mockDispatcher)

	pastDate := time.Now().Add(-24 * time.Hour)
	req := CreateTodoRequest{
		Title:    "Test with past due date",
		Priority: "high",
		DueDate:  &pastDate,
	}

	_, err := service.CreateTodo(context.Background(), req)

	if err == nil {
		t.Error("CreateTodo() expected error for past due date, got nil")
	}
}
