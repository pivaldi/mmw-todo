package application

import (
	"context"

	"github.com/pivaldi/mmw/todo/internal/application/command"
	"github.com/pivaldi/mmw/todo/internal/application/dto"
	"github.com/pivaldi/mmw/todo/internal/application/ports"
	"github.com/pivaldi/mmw/todo/internal/application/query"
	"github.com/rotisserie/eris"
)

// TodoService defines the application service interface
// This is the primary port - implemented by TodoApplicationService, called by adapters
type TodoService interface {
	CreateTodo(ctx context.Context, req *dto.CreateTodoRequest) (*dto.TodoResponse, error)
	GetTodo(ctx context.Context, id string) (*dto.TodoResponse, error)
	UpdateTodo(ctx context.Context, id string, req *dto.UpdateTodoRequest) (*dto.TodoResponse, error)
	CompleteTodo(ctx context.Context, id string) (*dto.TodoResponse, error)
	ReopenTodo(ctx context.Context, id string) (*dto.TodoResponse, error)
	DeleteTodo(ctx context.Context, id string) error
	ListTodos(ctx context.Context, filters *dto.ListFilters) (*dto.ListTodosResponse, error)
}

// TodoApplicationService implements the TodoService port
// It delegates to commands and queries
type TodoApplicationService struct {
	createTodoCmd   *command.CreateTodoCommand
	updateTodoCmd   *command.UpdateTodoCommand
	completeTodoCmd *command.CompleteTodoCommand
	reopenTodoCmd   *command.ReopenTodoCommand
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
		return nil, eris.Wrap(err, "failed to create todo")
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
		return nil, eris.Wrap(err, "failed to get todo")
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
		return nil, eris.Wrap(err, "failed to update todo")
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
		return nil, eris.Wrap(err, "failed to complete todo")
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
		return nil, eris.Wrap(err, "failed to reopen todo")
	}

	return result, nil
}

// DeleteTodo deletes a todo
func (s *TodoApplicationService) DeleteTodo(
	ctx context.Context,
	id string,
) error {
	err := s.deleteTodoCmd.Execute(ctx, id)
	if err != nil {
		return eris.Wrap(err, "failed to delete todo")
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
		return nil, eris.Wrap(err, "failed to list todos")
	}

	return result, nil
}
