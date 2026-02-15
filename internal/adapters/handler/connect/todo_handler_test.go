package connect

import (
	"context"
	"errors"
	"testing"
	"time"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/types/known/timestamppb"

	todov1 "github.com/pivaldi/mmw/contracts/go/todo/v1"
	"github.com/pivaldi/mmw/todo/internal/application"
	domain "github.com/pivaldi/mmw/todo/internal/domain/todo"
)

// MockTodoService is a mock implementation of application.TodoService
type MockTodoService struct {
	CreateTodoFunc   func(ctx context.Context, req application.CreateTodoRequest) (*application.TodoResponse, error)
	GetTodoFunc      func(ctx context.Context, id string) (*application.TodoResponse, error)
	UpdateTodoFunc   func(ctx context.Context, id string, req application.UpdateTodoRequest) (*application.TodoResponse, error)
	CompleteTodoFunc func(ctx context.Context, id string) (*application.TodoResponse, error)
	ReopenTodoFunc   func(ctx context.Context, id string) (*application.TodoResponse, error)
	DeleteTodoFunc   func(ctx context.Context, id string) error
	ListTodosFunc    func(ctx context.Context, filters application.ListFilters) (*application.ListTodosResponse, error)
}

func (m *MockTodoService) CreateTodo(ctx context.Context, req application.CreateTodoRequest) (*application.TodoResponse, error) {
	if m.CreateTodoFunc != nil {
		return m.CreateTodoFunc(ctx, req)
	}
	return nil, errors.New("not implemented")
}

func (m *MockTodoService) GetTodo(ctx context.Context, id string) (*application.TodoResponse, error) {
	if m.GetTodoFunc != nil {
		return m.GetTodoFunc(ctx, id)
	}
	return nil, errors.New("not implemented")
}

func (m *MockTodoService) UpdateTodo(ctx context.Context, id string, req application.UpdateTodoRequest) (*application.TodoResponse, error) {
	if m.UpdateTodoFunc != nil {
		return m.UpdateTodoFunc(ctx, id, req)
	}
	return nil, errors.New("not implemented")
}

func (m *MockTodoService) CompleteTodo(ctx context.Context, id string) (*application.TodoResponse, error) {
	if m.CompleteTodoFunc != nil {
		return m.CompleteTodoFunc(ctx, id)
	}
	return nil, errors.New("not implemented")
}

func (m *MockTodoService) ReopenTodo(ctx context.Context, id string) (*application.TodoResponse, error) {
	if m.ReopenTodoFunc != nil {
		return m.ReopenTodoFunc(ctx, id)
	}
	return nil, errors.New("not implemented")
}

func (m *MockTodoService) DeleteTodo(ctx context.Context, id string) error {
	if m.DeleteTodoFunc != nil {
		return m.DeleteTodoFunc(ctx, id)
	}
	return errors.New("not implemented")
}

func (m *MockTodoService) ListTodos(ctx context.Context, filters application.ListFilters) (*application.ListTodosResponse, error) {
	if m.ListTodosFunc != nil {
		return m.ListTodosFunc(ctx, filters)
	}
	return nil, errors.New("not implemented")
}

func TestTodoHandler_CreateTodo_Success(t *testing.T) {
	mockService := &MockTodoService{
		CreateTodoFunc: func(ctx context.Context, req application.CreateTodoRequest) (*application.TodoResponse, error) {
			// Verify request mapping
			if req.Title != "Test Todo" {
				t.Errorf("Title = %v, want %v", req.Title, "Test Todo")
			}
			if req.Priority != "medium" {
				t.Errorf("Priority = %v, want %v", req.Priority, "medium")
			}

			return &application.TodoResponse{
				ID:          "123",
				Title:       "Test Todo",
				Description: "Test description",
				Status:      "pending",
				Priority:    "medium",
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			}, nil
		},
	}

	handler := NewTodoHandler(mockService)

	req := connect.NewRequest(&todov1.CreateTodoRequest{
		Title:       "Test Todo",
		Description: "Test description",
		Priority:    todov1.Priority_PRIORITY_MEDIUM,
	})

	resp, err := handler.CreateTodo(context.Background(), req)

	if err != nil {
		t.Fatalf("CreateTodo() unexpected error: %v", err)
	}

	if resp.Msg.Todo.Title != "Test Todo" {
		t.Errorf("Response title = %v, want %v", resp.Msg.Todo.Title, "Test Todo")
	}

	if resp.Msg.Todo.Status != todov1.TaskStatus_TASK_STATUS_PENDING {
		t.Errorf("Response status = %v, want %v", resp.Msg.Todo.Status, todov1.TaskStatus_TASK_STATUS_PENDING)
	}
}

func TestTodoHandler_CreateTodo_WithDueDate_Success(t *testing.T) {
	dueDate := time.Now().Add(24 * time.Hour)

	mockService := &MockTodoService{
		CreateTodoFunc: func(ctx context.Context, req application.CreateTodoRequest) (*application.TodoResponse, error) {
			if req.DueDate == nil {
				t.Error("Expected due date to be set")
			}

			return &application.TodoResponse{
				ID:          "123",
				Title:       "Test Todo",
				Description: "Test description",
				Status:      "pending",
				Priority:    "high",
				DueDate:     &dueDate,
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			}, nil
		},
	}

	handler := NewTodoHandler(mockService)

	req := connect.NewRequest(&todov1.CreateTodoRequest{
		Title:       "Test Todo",
		Description: "Test description",
		Priority:    todov1.Priority_PRIORITY_HIGH,
		DueDate:     timestamppb.New(dueDate),
	})

	resp, err := handler.CreateTodo(context.Background(), req)

	if err != nil {
		t.Fatalf("CreateTodo() unexpected error: %v", err)
	}

	if resp.Msg.Todo.DueDate == nil {
		t.Error("Expected due date in response")
	}
}

func TestTodoHandler_GetTodo_Success(t *testing.T) {
	mockService := &MockTodoService{
		GetTodoFunc: func(ctx context.Context, id string) (*application.TodoResponse, error) {
			if id != "123" {
				t.Errorf("ID = %v, want %v", id, "123")
			}

			return &application.TodoResponse{
				ID:          "123",
				Title:       "Test Todo",
				Description: "Test description",
				Status:      "pending",
				Priority:    "medium",
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			}, nil
		},
	}

	handler := NewTodoHandler(mockService)

	req := connect.NewRequest(&todov1.GetTodoRequest{
		Id: "123",
	})

	resp, err := handler.GetTodo(context.Background(), req)

	if err != nil {
		t.Fatalf("GetTodo() unexpected error: %v", err)
	}

	if resp.Msg.Todo.Id != "123" {
		t.Errorf("Response ID = %v, want %v", resp.Msg.Todo.Id, "123")
	}
}

func TestTodoHandler_GetTodo_NotFound_ReturnsNotFoundError(t *testing.T) {
	mockService := &MockTodoService{
		GetTodoFunc: func(ctx context.Context, id string) (*application.TodoResponse, error) {
			return nil, domain.ErrTodoNotFound
		},
	}

	handler := NewTodoHandler(mockService)

	req := connect.NewRequest(&todov1.GetTodoRequest{
		Id: "nonexistent",
	})

	_, err := handler.GetTodo(context.Background(), req)

	if err == nil {
		t.Fatal("GetTodo() expected error, got nil")
	}

	var connectErr *connect.Error
	if !errors.As(err, &connectErr) {
		t.Fatalf("Expected connect.Error, got %T", err)
	}

	if connectErr.Code() != connect.CodeNotFound {
		t.Errorf("Error code = %v, want %v", connectErr.Code(), connect.CodeNotFound)
	}
}

func TestTodoHandler_UpdateTodo_Success(t *testing.T) {
	newTitle := "Updated Title"

	mockService := &MockTodoService{
		UpdateTodoFunc: func(ctx context.Context, id string, req application.UpdateTodoRequest) (*application.TodoResponse, error) {
			if req.Title == nil || *req.Title != newTitle {
				t.Error("Title not updated correctly")
			}

			return &application.TodoResponse{
				ID:          id,
				Title:       newTitle,
				Description: "Test description",
				Status:      "pending",
				Priority:    "medium",
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			}, nil
		},
	}

	handler := NewTodoHandler(mockService)

	req := connect.NewRequest(&todov1.UpdateTodoRequest{
		Id:    "123",
		Title: &newTitle,
	})

	resp, err := handler.UpdateTodo(context.Background(), req)

	if err != nil {
		t.Fatalf("UpdateTodo() unexpected error: %v", err)
	}

	if resp.Msg.Todo.Title != newTitle {
		t.Errorf("Response title = %v, want %v", resp.Msg.Todo.Title, newTitle)
	}
}

func TestTodoHandler_CompleteTodo_Success(t *testing.T) {
	mockService := &MockTodoService{
		CompleteTodoFunc: func(ctx context.Context, id string) (*application.TodoResponse, error) {
			return &application.TodoResponse{
				ID:          id,
				Title:       "Test Todo",
				Description: "Test description",
				Status:      "completed",
				Priority:    "medium",
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			}, nil
		},
	}

	handler := NewTodoHandler(mockService)

	req := connect.NewRequest(&todov1.CompleteTodoRequest{
		Id: "123",
	})

	resp, err := handler.CompleteTodo(context.Background(), req)

	if err != nil {
		t.Fatalf("CompleteTodo() unexpected error: %v", err)
	}

	if resp.Msg.Todo.Status != todov1.TaskStatus_TASK_STATUS_COMPLETED {
		t.Errorf("Response status = %v, want %v", resp.Msg.Todo.Status, todov1.TaskStatus_TASK_STATUS_COMPLETED)
	}
}

func TestTodoHandler_ReopenTodo_Success(t *testing.T) {
	mockService := &MockTodoService{
		ReopenTodoFunc: func(ctx context.Context, id string) (*application.TodoResponse, error) {
			return &application.TodoResponse{
				ID:          id,
				Title:       "Test Todo",
				Description: "Test description",
				Status:      "pending",
				Priority:    "medium",
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			}, nil
		},
	}

	handler := NewTodoHandler(mockService)

	req := connect.NewRequest(&todov1.ReopenTodoRequest{
		Id: "123",
	})

	resp, err := handler.ReopenTodo(context.Background(), req)

	if err != nil {
		t.Fatalf("ReopenTodo() unexpected error: %v", err)
	}

	if resp.Msg.Todo.Status != todov1.TaskStatus_TASK_STATUS_PENDING {
		t.Errorf("Response status = %v, want %v", resp.Msg.Todo.Status, todov1.TaskStatus_TASK_STATUS_PENDING)
	}
}

func TestTodoHandler_DeleteTodo_Success(t *testing.T) {
	mockService := &MockTodoService{
		DeleteTodoFunc: func(ctx context.Context, id string) error {
			if id != "123" {
				t.Errorf("ID = %v, want %v", id, "123")
			}
			return nil
		},
	}

	handler := NewTodoHandler(mockService)

	req := connect.NewRequest(&todov1.DeleteTodoRequest{
		Id: "123",
	})

	_, err := handler.DeleteTodo(context.Background(), req)

	if err != nil {
		t.Fatalf("DeleteTodo() unexpected error: %v", err)
	}
}

func TestTodoHandler_ListTodos_Success(t *testing.T) {
	mockService := &MockTodoService{
		ListTodosFunc: func(ctx context.Context, filters application.ListFilters) (*application.ListTodosResponse, error) {
			// Verify filters
			if filters.Status != nil && *filters.Status != "pending" {
				t.Errorf("Status filter = %v, want %v", *filters.Status, "pending")
			}

			todo1 := &application.TodoResponse{
				ID:          "1",
				Title:       "Todo 1",
				Description: "Description 1",
				Status:      "pending",
				Priority:    "medium",
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			}

			todo2 := &application.TodoResponse{
				ID:          "2",
				Title:       "Todo 2",
				Description: "Description 2",
				Status:      "pending",
				Priority:    "high",
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			}

			return &application.ListTodosResponse{
				Todos:      []*application.TodoResponse{todo1, todo2},
				TotalCount: 2,
			}, nil
		},
	}

	handler := NewTodoHandler(mockService)

	status := todov1.TaskStatus_TASK_STATUS_PENDING
	limit := int32(10)
	offset := int32(0)
	req := connect.NewRequest(&todov1.ListTodosRequest{
		Status: &status,
		Limit:  &limit,
		Offset: &offset,
	})

	resp, err := handler.ListTodos(context.Background(), req)

	if err != nil {
		t.Fatalf("ListTodos() unexpected error: %v", err)
	}

	if len(resp.Msg.Todos) != 2 {
		t.Errorf("Response todos count = %v, want %v", len(resp.Msg.Todos), 2)
	}

	if resp.Msg.TotalCount != 2 {
		t.Errorf("Response total count = %v, want %v", resp.Msg.TotalCount, 2)
	}
}

func TestTodoHandler_ValidationError_ReturnsInvalidArgument(t *testing.T) {
	mockService := &MockTodoService{
		CreateTodoFunc: func(ctx context.Context, req application.CreateTodoRequest) (*application.TodoResponse, error) {
			return nil, &domain.ValidationError{
				Field:   "title",
				Message: "title is required",
			}
		},
	}

	handler := NewTodoHandler(mockService)

	req := connect.NewRequest(&todov1.CreateTodoRequest{
		Title:    "",
		Priority: todov1.Priority_PRIORITY_MEDIUM,
	})

	_, err := handler.CreateTodo(context.Background(), req)

	if err == nil {
		t.Fatal("CreateTodo() expected error, got nil")
	}

	var connectErr *connect.Error
	if !errors.As(err, &connectErr) {
		t.Fatalf("Expected connect.Error, got %T", err)
	}

	if connectErr.Code() != connect.CodeInvalidArgument {
		t.Errorf("Error code = %v, want %v", connectErr.Code(), connect.CodeInvalidArgument)
	}
}

func TestTodoHandler_BusinessRuleError_ReturnsFailedPrecondition(t *testing.T) {
	mockService := &MockTodoService{
		CompleteTodoFunc: func(ctx context.Context, id string) (*application.TodoResponse, error) {
			return nil, &domain.BusinessRuleError{
				Rule:    "complete_cancelled",
				Message: "cannot complete a cancelled task",
			}
		},
	}

	handler := NewTodoHandler(mockService)

	req := connect.NewRequest(&todov1.CompleteTodoRequest{
		Id: "123",
	})

	_, err := handler.CompleteTodo(context.Background(), req)

	if err == nil {
		t.Fatal("CompleteTodo() expected error, got nil")
	}

	var connectErr *connect.Error
	if !errors.As(err, &connectErr) {
		t.Fatalf("Expected connect.Error, got %T", err)
	}

	if connectErr.Code() != connect.CodeFailedPrecondition {
		t.Errorf("Error code = %v, want %v", connectErr.Code(), connect.CodeFailedPrecondition)
	}
}
