package application

import (
	"context"
	"fmt"

	"github.com/pivaldi/mmw-todo/internal/application/command"
	"github.com/pivaldi/mmw-todo/internal/application/dto"
	"github.com/pivaldi/mmw-todo/internal/application/ports"
	"github.com/pivaldi/mmw-todo/internal/application/query"
)

// TodoService defines the application service interface
// This is the primary port - implemented by TodoApplicationService, called by adapters
type TodoService interface {
	// Create Todo
	CreateTodo(ctx context.Context, req *dto.CreateTodoRequest) (*dto.TodoResponse, error)
	// Getter Todo
	GetTodo(ctx context.Context, id string) (*dto.TodoResponse, error)
	// Update todo
	UpdateTodo(ctx context.Context, id string, req *dto.UpdateTodoRequest) (*dto.TodoResponse, error)
	// Make a todo completed
	CompleteTodo(ctx context.Context, id string) (*dto.TodoResponse, error)
	// Reopen a todo
	ReopenTodo(ctx context.Context, id string) (*dto.TodoResponse, error)
	// Delete a todo
	DeleteTodo(ctx context.Context, id string) error
	// Get a list of todo
	ListTodos(ctx context.Context, filters *dto.ListFilters) (*dto.ListTodosResponse, error)
	// Check the application health
	Health(ctx context.Context) (any, error)
}

// TodoApplicationService implements the TodoService port
// It delegates to commands and queries
type TodoApplicationService struct {
	repository      ports.TodoRepository
	createTodoCmd   *command.CreateTodoCommand
	updateTodoCmd   *command.UpdateTodoCommand
	completeTodoCmd *command.TodoStatusChangeCommand
	reopenTodoCmd   *command.TodoStatusChangeCommand
	deleteTodoCmd   *command.DeleteTodoCommand
	getTodoQuery    *query.GetTodoQuery
	listTodosQuery  *query.ListTodosQuery
}

// NewTodoApplicationService creates a new TodoApplicationService
func NewTodoApplicationService(
	repository ports.TodoRepository,
	unitOfWork ports.UnitOfWork,
	eventDispatcher ports.EventDispatcher,
) TodoService {
	return &TodoApplicationService{
		repository:      repository,
		createTodoCmd:   command.NewCreateTodoCommand(repository, unitOfWork, eventDispatcher),
		updateTodoCmd:   command.NewUpdateTodoCommand(repository, eventDispatcher),
		completeTodoCmd: command.NewCompleteTodoCommand(repository, eventDispatcher),
		reopenTodoCmd:   command.NewReopenTodoCommand(repository, eventDispatcher),
		deleteTodoCmd:   command.NewDeleteTodoCommand(repository, eventDispatcher),
		getTodoQuery:    query.NewGetTodoQuery(repository),
		listTodosQuery:  query.NewListTodosQuery(repository),
	}
}

// CreateTodo creates a new todo
func (s *TodoApplicationService) CreateTodo(
	ctx context.Context,
	req *dto.CreateTodoRequest,
) (*dto.TodoResponse, error) {
	result, err := s.createTodoCmd.Execute(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("create todo: %w", err)
	}

	return result, nil
}

// GetTodo retrieves a todo by ID
func (s *TodoApplicationService) GetTodo(
	ctx context.Context,
	id string,
) (*dto.TodoResponse, error) {
	result, err := s.getTodoQuery.Execute(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get todo: %w", err)
	}

	return result, nil
}

// UpdateTodo updates an existing todo
func (s *TodoApplicationService) UpdateTodo(
	ctx context.Context,
	id string,
	req *dto.UpdateTodoRequest,
) (*dto.TodoResponse, error) {
	result, err := s.updateTodoCmd.Execute(ctx, id, req)
	if err != nil {
		return nil, fmt.Errorf("update todo: %w", err)
	}

	return result, nil
}

// CompleteTodo marks a todo as completed
func (s *TodoApplicationService) CompleteTodo(
	ctx context.Context,
	id string,
) (*dto.TodoResponse, error) {
	result, err := s.completeTodoCmd.Execute(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("complete todo: %w", err)
	}

	return result, nil
}

// ReopenTodo reopens a completed or cancelled todo
func (s *TodoApplicationService) ReopenTodo(
	ctx context.Context,
	id string,
) (*dto.TodoResponse, error) {
	result, err := s.reopenTodoCmd.Execute(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("reopen todo: %w", err)
	}

	return result, nil
}

// DeleteTodo deletes a todo
func (s *TodoApplicationService) DeleteTodo(
	ctx context.Context,
	id string,
) error {
	if err := s.deleteTodoCmd.Execute(ctx, id); err != nil {
		return fmt.Errorf("delete todo: %w", err)
	}

	return nil
}

// ListTodos retrieves todos with optional filters
func (s *TodoApplicationService) ListTodos(
	ctx context.Context,
	filters *dto.ListFilters,
) (*dto.ListTodosResponse, error) {
	result, err := s.listTodosQuery.Execute(ctx, filters)
	if err != nil {
		return nil, fmt.Errorf("list todos: %w", err)
	}

	return result, nil
}

// Health return a simple database health check
func (s *TodoApplicationService) Health(ctx context.Context) (any, error) {
	count, err := s.repository.Health(ctx)
	if err != nil {
		return nil, fmt.Errorf("health check: %w", err)
	}

	return count, nil
}
