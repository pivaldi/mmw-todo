package application

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/pivaldi/mmw/todo/internal/application/ports"
	"github.com/pivaldi/mmw/todo/internal/application/ports/mocks"
	domain "github.com/pivaldi/mmw/todo/internal/domain/todo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// Test helpers

func createTestTodo() *domain.Todo {
	title, _ := domain.NewTaskTitle("Test Todo")
	return domain.NewTodo(title, "Test description", domain.PriorityMedium, nil)
}

func setupService(t *testing.T) (*TodoApplicationService, *mocks.MockTodoRepository, *mocks.MockEventDispatcher, *mocks.MockUnitOfWork) {
	mockRepo := mocks.NewMockTodoRepository(t)
	mockDispatcher := mocks.NewMockEventDispatcher(t)
	mockUoW := mocks.NewMockUnitOfWork(t)
	service := NewTodoApplicationService(mockRepo, mockUoW, mockDispatcher)
	return service, mockRepo, mockDispatcher, mockUoW
}

// CreateTodo Tests

func TestTodoService_CreateTodo_Success(t *testing.T) {
	tests := []struct {
		name       string
		request    CreateTodoRequest
		wantTitle  string
		wantDesc   string
		wantPri    domain.Priority
		wantStatus domain.TaskStatus
	}{
		{
			name: "valid todo with all fields",
			request: CreateTodoRequest{
				Title:       "Buy groceries",
				Description: "Milk, eggs, bread",
				Priority:    domain.PriorityMedium,
			},
			wantTitle:  "Buy groceries",
			wantDesc:   "Milk, eggs, bread",
			wantPri:    domain.PriorityMedium,
			wantStatus: domain.TaskStatusPending,
		},
		{
			name: "valid todo with high priority",
			request: CreateTodoRequest{
				Title:       "Urgent task",
				Description: "Complete ASAP",
				Priority:    domain.PriorityHigh,
			},
			wantTitle:  "Urgent task",
			wantDesc:   "Complete ASAP",
			wantPri:    domain.PriorityHigh,
			wantStatus: domain.TaskStatusPending,
		},
		{
			name: "valid todo with low priority",
			request: CreateTodoRequest{
				Title:    "Optional task",
				Priority: domain.PriorityLow,
			},
			wantTitle:  "Optional task",
			wantDesc:   "",
			wantPri:    domain.PriorityLow,
			wantStatus: domain.TaskStatusPending,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service, mockRepo, mockDispatcher, mockUoW := setupService(t)

			// Setup UoW to execute the transaction closure
			mockUoW.EXPECT().
				WithTransaction(mock.Anything, mock.AnythingOfType("func(context.Context) error")).
				RunAndReturn(func(ctx context.Context, fn func(txCtx context.Context) error) error {
					return fn(ctx)
				})

			// Setup repository to succeed
			mockRepo.EXPECT().
				Save(mock.Anything, mock.AnythingOfType("*domain.Todo")).
				Return(nil)

			// Setup event dispatcher to succeed
			mockDispatcher.EXPECT().
				Dispatch(mock.Anything, mock.MatchedBy(func(events []domain.DomainEvent) bool {
					return len(events) == 1 && events[0].EventType() == "TodoCreated"
				})).
				Return(nil)

			result, err := service.CreateTodo(context.Background(), tt.request)

			require.NoError(t, err)
			require.NotNil(t, result)
			assert.Equal(t, tt.wantTitle, result.Title)
			assert.Equal(t, tt.wantDesc, result.Description)
			assert.Equal(t, tt.wantPri, result.Priority)
			assert.Equal(t, tt.wantStatus, result.Status)
		})
	}
}

func TestTodoService_CreateTodo_WithDueDate(t *testing.T) {
	tests := []struct {
		name      string
		dueDate   time.Time
		wantError bool
	}{
		{
			name:      "future due date succeeds",
			dueDate:   time.Now().Add(24 * time.Hour),
			wantError: false,
		},
		{
			name:      "past due date fails",
			dueDate:   time.Now().Add(-24 * time.Hour),
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service, mockRepo, mockDispatcher, mockUoW := setupService(t)

			if !tt.wantError {
				// Only setup expectations if we expect success
				mockUoW.EXPECT().
					WithTransaction(mock.Anything, mock.AnythingOfType("func(context.Context) error")).
					RunAndReturn(func(ctx context.Context, fn func(txCtx context.Context) error) error {
						return fn(ctx)
					})

				mockRepo.EXPECT().
					Save(mock.Anything, mock.AnythingOfType("*domain.Todo")).
					Return(nil)

				mockDispatcher.EXPECT().
					Dispatch(mock.Anything, mock.Anything).
					Return(nil)
			}

			req := CreateTodoRequest{
				Title:    "Test with due date",
				Priority: domain.PriorityHigh,
				DueDate:  &tt.dueDate,
			}

			result, err := service.CreateTodo(context.Background(), req)

			if tt.wantError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				require.NotNil(t, result)
				assert.NotNil(t, result.DueDate)
			}
		})
	}
}

func TestTodoService_CreateTodo_ValidationErrors(t *testing.T) {
	tests := []struct {
		name    string
		request CreateTodoRequest
	}{
		{
			name: "empty title fails",
			request: CreateTodoRequest{
				Title:    "",
				Priority: domain.PriorityMedium,
			},
		},
		{
			name: "invalid priority fails",
			request: CreateTodoRequest{
				Title:    "Test",
				Priority: "invalid",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service, _, _, _ := setupService(t)

			result, err := service.CreateTodo(context.Background(), tt.request)

			assert.Error(t, err)
			assert.Nil(t, result)
		})
	}
}

func TestTodoService_CreateTodo_RepositoryError(t *testing.T) {
	service, mockRepo, _, mockUoW := setupService(t)

	// Setup UoW to execute the transaction closure
	mockUoW.EXPECT().
		WithTransaction(mock.Anything, mock.AnythingOfType("func(context.Context) error")).
		RunAndReturn(func(ctx context.Context, fn func(txCtx context.Context) error) error {
			return fn(ctx)
		})

	// Setup repository to fail
	mockRepo.EXPECT().
		Save(mock.Anything, mock.AnythingOfType("*domain.Todo")).
		Return(errors.New("database error"))

	// NOTE: We deliberately DO NOT set an expectation for mockDispatcher.
	// If the service tries to dispatch events after a save failure, the test will fail!

	req := CreateTodoRequest{
		Title:    "Test",
		Priority: domain.PriorityMedium,
	}

	result, err := service.CreateTodo(context.Background(), req)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "saving todo")
}

// GetTodo Tests

func TestTodoService_GetTodo_Success(t *testing.T) {
	testTodo := createTestTodo()
	service, mockRepo, _, _ := setupService(t)

	mockRepo.EXPECT().
		FindByID(mock.Anything, testTodo.ID()).
		Return(testTodo, nil)

	result, err := service.GetTodo(context.Background(), testTodo.ID().String())

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, testTodo.ID().String(), result.ID)
	assert.Equal(t, "Test Todo", result.Title)
}

func TestTodoService_GetTodo_Errors(t *testing.T) {
	tests := []struct {
		name      string
		id        string
		setupMock func(*mocks.MockTodoRepository)
	}{
		{
			name: "invalid ID format",
			id:   "invalid-id",
			setupMock: func(m *mocks.MockTodoRepository) {
				// No setup needed, validation happens before repository call
			},
		},
		{
			name: "todo not found",
			id:   domain.NewTodoID().String(),
			setupMock: func(m *mocks.MockTodoRepository) {
				m.EXPECT().
					FindByID(mock.Anything, mock.AnythingOfType("domain.TodoID")).
					Return(nil, domain.ErrTodoNotFound)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service, mockRepo, _, _ := setupService(t)
			if tt.setupMock != nil {
				tt.setupMock(mockRepo)
			}

			result, err := service.GetTodo(context.Background(), tt.id)

			assert.Error(t, err)
			assert.Nil(t, result)
		})
	}
}

// UpdateTodo Tests

func TestTodoService_UpdateTodo_Success(t *testing.T) {
	testTodo := createTestTodo()
	service, mockRepo, mockDispatcher, _ := setupService(t)

	mockRepo.EXPECT().
		FindByID(mock.Anything, testTodo.ID()).
		Return(testTodo, nil)

	mockRepo.EXPECT().
		Update(mock.Anything, mock.AnythingOfType("*domain.Todo")).
		Return(nil)

	mockDispatcher.EXPECT().
		Dispatch(mock.Anything, mock.MatchedBy(func(events []domain.DomainEvent) bool {
			return len(events) >= 1
		})).
		Return(nil)

	newTitle := "Updated Title"
	req := UpdateTodoRequest{
		Title: &newTitle,
	}

	result, err := service.UpdateTodo(context.Background(), testTodo.ID().String(), req)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, newTitle, result.Title)
}

// CompleteTodo Tests

func TestTodoService_CompleteTodo_Success(t *testing.T) {
	testTodo := createTestTodo()
	service, mockRepo, mockDispatcher, _ := setupService(t)

	mockRepo.EXPECT().
		FindByID(mock.Anything, testTodo.ID()).
		Return(testTodo, nil)

	mockRepo.EXPECT().
		Update(mock.Anything, mock.AnythingOfType("*domain.Todo")).
		Return(nil)

	mockDispatcher.EXPECT().
		Dispatch(mock.Anything, mock.MatchedBy(func(events []domain.DomainEvent) bool {
			// Verify TodoCompleted event is present
			for _, event := range events {
				if event.EventType() == "TodoCompleted" {
					return true
				}
			}
			return false
		})).
		Return(nil)

	result, err := service.CompleteTodo(context.Background(), testTodo.ID().String())

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, domain.TaskStatusCompleted, result.Status)
}

// ReopenTodo Tests

func TestTodoService_ReopenTodo_Success(t *testing.T) {
	testTodo := createTestTodo()
	testTodo.Complete() // Mark as completed first
	testTodo.ClearEvents()

	service, mockRepo, mockDispatcher, _ := setupService(t)

	mockRepo.EXPECT().
		FindByID(mock.Anything, testTodo.ID()).
		Return(testTodo, nil)

	mockRepo.EXPECT().
		Update(mock.Anything, mock.AnythingOfType("*domain.Todo")).
		Return(nil)

	mockDispatcher.EXPECT().
		Dispatch(mock.Anything, mock.MatchedBy(func(events []domain.DomainEvent) bool {
			// Verify TodoReopened event is present
			for _, event := range events {
				if event.EventType() == "TodoReopened" {
					return true
				}
			}
			return false
		})).
		Return(nil)

	result, err := service.ReopenTodo(context.Background(), testTodo.ID().String())

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, domain.TaskStatusPending, result.Status)
}

// DeleteTodo Tests

func TestTodoService_DeleteTodo_Success(t *testing.T) {
	testTodo := createTestTodo()
	service, mockRepo, mockDispatcher, _ := setupService(t)

	mockRepo.EXPECT().
		Delete(mock.Anything, testTodo.ID()).
		Return(nil)

	mockDispatcher.EXPECT().
		Dispatch(mock.Anything, mock.MatchedBy(func(events []domain.DomainEvent) bool {
			// Verify TodoDeleted event is present
			for _, event := range events {
				if event.EventType() == "TodoDeleted" {
					return true
				}
			}
			return false
		})).
		Return(nil)

	err := service.DeleteTodo(context.Background(), testTodo.ID().String())

	require.NoError(t, err)
}

// ListTodos Tests

func TestTodoService_ListTodos_Success(t *testing.T) {
	testTodo1 := createTestTodo()
	testTodo2 := createTestTodo()

	service, mockRepo, _, _ := setupService(t)

	mockRepo.EXPECT().
		FindAll(mock.Anything, mock.AnythingOfType("ports.Filters")).
		Return([]*domain.Todo{testTodo1, testTodo2}, nil)

	result, err := service.ListTodos(context.Background(), ListFilters{})

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Len(t, result.Todos, 2)
	assert.Equal(t, 2, result.TotalCount)
}

func TestTodoService_ListTodos_WithStatusFilter(t *testing.T) {
	service, mockRepo, _, _ := setupService(t)

	mockRepo.EXPECT().
		FindAll(mock.Anything, mock.MatchedBy(func(filters ports.Filters) bool {
			// Verify status filter is passed correctly
			return filters.Status != nil && *filters.Status == domain.TaskStatusPending
		})).
		Return([]*domain.Todo{}, nil)

	statusFilter := domain.TaskStatusPending
	result, err := service.ListTodos(context.Background(), ListFilters{
		Status: &statusFilter,
	})

	require.NoError(t, err)
	require.NotNil(t, result)
}

// Transaction and Rollback Tests

func TestTodoService_TransactionRollback_OnError(t *testing.T) {
	service, mockRepo, _, mockUoW := setupService(t)

	// Setup UoW to execute the transaction closure
	mockUoW.EXPECT().
		WithTransaction(mock.Anything, mock.AnythingOfType("func(context.Context) error")).
		RunAndReturn(func(ctx context.Context, fn func(txCtx context.Context) error) error {
			return fn(ctx)
		})

	// Setup repository to fail
	mockRepo.EXPECT().
		Save(mock.Anything, mock.AnythingOfType("*domain.Todo")).
		Return(errors.New("database duplicate key error"))

	// NOTE: We deliberately DO NOT set an expectation for mockDispatcher.
	// If the service tries to dispatch events after a save failure, the test will fail!

	req := CreateTodoRequest{
		Title:    "Fail Test",
		Priority: domain.PriorityLow,
	}

	_, err := service.CreateTodo(context.Background(), req)

	assert.Error(t, err)
}

// Examples using mockery-generated mocks with proper transaction handling

func TestCreateTodo_Success(t *testing.T) {
	// 1. Instantiate the auto-generated v3 mocks
	// Note: Because we set structname: "Mock{{.InterfaceName}}", the constructor is NewMock...
	mockRepo := mocks.NewMockTodoRepository(t)
	mockUoW := mocks.NewMockUnitOfWork(t)
	mockDispatcher := mocks.NewMockEventDispatcher(t)

	ctx := context.Background()

	// 2. The UoW Closure Magic
	// We tell the mock: "When WithTransaction is called, intercept it and run the closure!"
	mockUoW.EXPECT().
		WithTransaction(mock.Anything, mock.AnythingOfType("func(context.Context) error")).
		RunAndReturn(func(ctx context.Context, fn func(txCtx context.Context) error) error {
			return fn(ctx) // Instantly execute the inner block
		})

	// 3. Setup Expectations inside the closure
	mockRepo.EXPECT().
		Save(mock.Anything, mock.AnythingOfType("*domain.Todo")).
		Return(nil)

	mockDispatcher.EXPECT().
		Dispatch(mock.Anything, mock.Anything).
		Return(nil)

	// 4. Instantiate your Service
	service := NewTodoApplicationService(mockRepo, mockUoW, mockDispatcher)

	// 5. Execute the actual Request
	req := CreateTodoRequest{
		Title:    "Learn Go Clean Architecture",
		Priority: domain.PriorityHigh,
	}

	resp, err := service.CreateTodo(ctx, req)

	// 6. Assertions
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "Learn Go Clean Architecture", resp.Title)

	// Note: We don't need `mockRepo.AssertExpectations(t)` because
	// passing `t` to the constructor (NewMock...) handles that automatically!
}

func TestCreateTodo_RollbackOnSaveFailure(t *testing.T) {
	mockRepo := mocks.NewMockTodoRepository(t)
	mockUoW := mocks.NewMockUnitOfWork(t)
	mockDispatcher := mocks.NewMockEventDispatcher(t)

	ctx := context.Background()

	// Execute closure exactly like success case
	mockUoW.EXPECT().
		WithTransaction(mock.Anything, mock.AnythingOfType("func(context.Context) error")).
		RunAndReturn(func(ctx context.Context, fn func(txCtx context.Context) error) error {
			return fn(ctx)
		})

	// Simulate a database failure!
	mockRepo.EXPECT().
		Save(mock.Anything, mock.AnythingOfType("*domain.Todo")).
		Return(assert.AnError)

	// NOTE: We deliberately DO NOT set an expectation for mockDispatcher.
	// If the service tries to dispatch events after a save failure, the test will correctly fail!

	service := NewTodoApplicationService(mockRepo, mockUoW, mockDispatcher)

	req := CreateTodoRequest{
		Title:    "Trigger Failure",
		Priority: domain.PriorityLow,
	}

	resp, err := service.CreateTodo(ctx, req)

	// Assert the failure bubbled up
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.ErrorIs(t, err, assert.AnError)
}
