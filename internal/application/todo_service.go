package application

import (
	"context"
	"fmt"
	"time"

	domain "github.com/pivaldi/mmw/todo/internal/domain/todo"
	"github.com/pivaldi/mmw/todo/internal/ports"
)

// TodoService defines the application service interface
// This is the primary port - implemented by TodoApplicationService, called by adapters
type TodoService interface {
	CreateTodo(ctx context.Context, req CreateTodoRequest) (*TodoResponse, error)
	GetTodo(ctx context.Context, id string) (*TodoResponse, error)
	UpdateTodo(ctx context.Context, id string, req UpdateTodoRequest) (*TodoResponse, error)
	CompleteTodo(ctx context.Context, id string) (*TodoResponse, error)
	ReopenTodo(ctx context.Context, id string) (*TodoResponse, error)
	DeleteTodo(ctx context.Context, id string) error
	ListTodos(ctx context.Context, filters ListFilters) (*ListTodosResponse, error)
}

// TodoApplicationService implements the TodoService port
// It orchestrates domain operations and coordinates infrastructure concerns
type TodoApplicationService struct {
	repository ports.TodoRepository
	dispatcher ports.EventDispatcher
}

// NewTodoApplicationService creates a new TodoApplicationService
func NewTodoApplicationService(
	repository ports.TodoRepository,
	dispatcher ports.EventDispatcher,
) *TodoApplicationService {
	return &TodoApplicationService{
		repository: repository,
		dispatcher: dispatcher,
	}
}

// CreateTodo creates a new todo
func (s *TodoApplicationService) CreateTodo(
	ctx context.Context,
	req CreateTodoRequest,
) (*TodoResponse, error) {
	// Create value objects from request
	title, err := domain.NewTaskTitle(req.Title)
	if err != nil {
		return nil, fmt.Errorf("invalid title: %w", err)
	}

	priority, err := domain.NewPriority(req.Priority)
	if err != nil {
		return nil, fmt.Errorf("invalid priority: %w", err)
	}

	var dueDate *domain.DueDate
	if req.DueDate != nil {
		dd, err := domain.NewDueDate(*req.DueDate)
		if err != nil {
			return nil, fmt.Errorf("invalid due date: %w", err)
		}
		dueDate = &dd
	}

	// Create todo using domain factory
	todo := domain.NewTodo(title, req.Description, priority, dueDate)

	// Persist the todo
	if err := s.repository.Save(ctx, todo); err != nil {
		return nil, fmt.Errorf("saving todo: %w", err)
	}

	// Dispatch domain events
	if err := s.dispatcher.Dispatch(ctx, todo.Events()); err != nil {
		return nil, fmt.Errorf("dispatching events: %w", err)
	}

	// Clear events after dispatching
	todo.ClearEvents()

	// Map to response DTO
	return MapTodoToResponse(todo), nil
}

// GetTodo retrieves a todo by ID
func (s *TodoApplicationService) GetTodo(
	ctx context.Context,
	id string,
) (*TodoResponse, error) {
	// Parse and validate ID
	todoID, err := domain.ParseTodoID(id)
	if err != nil {
		return nil, fmt.Errorf("invalid todo ID: %w", err)
	}

	// Retrieve from repository
	todo, err := s.repository.FindByID(ctx, todoID)
	if err != nil {
		return nil, fmt.Errorf("finding todo: %w", err)
	}

	// Map to response DTO
	return MapTodoToResponse(todo), nil
}

// UpdateTodo updates an existing todo
func (s *TodoApplicationService) UpdateTodo(
	ctx context.Context,
	id string,
	req UpdateTodoRequest,
) (*TodoResponse, error) {
	// Parse and validate ID
	todoID, err := domain.ParseTodoID(id)
	if err != nil {
		return nil, fmt.Errorf("invalid todo ID: %w", err)
	}

	// Retrieve existing todo
	todo, err := s.repository.FindByID(ctx, todoID)
	if err != nil {
		return nil, fmt.Errorf("finding todo: %w", err)
	}

	// Update title if provided
	if req.Title != nil {
		title, err := domain.NewTaskTitle(*req.Title)
		if err != nil {
			return nil, fmt.Errorf("invalid title: %w", err)
		}
		if err := todo.UpdateTitle(title); err != nil {
			return nil, fmt.Errorf("updating title: %w", err)
		}
	}

	// Update description if provided
	if req.Description != nil {
		if err := todo.UpdateDescription(*req.Description); err != nil {
			return nil, fmt.Errorf("updating description: %w", err)
		}
	}

	// Update priority if provided
	if req.Priority != nil {
		priority, err := domain.NewPriority(*req.Priority)
		if err != nil {
			return nil, fmt.Errorf("invalid priority: %w", err)
		}
		if err := todo.UpdatePriority(priority); err != nil {
			return nil, fmt.Errorf("updating priority: %w", err)
		}
	}

	// Update due date if provided
	if req.DueDate != nil {
		var dueDate *domain.DueDate
		if *req.DueDate != (time.Time{}) {
			dd, err := domain.NewDueDate(*req.DueDate)
			if err != nil {
				return nil, fmt.Errorf("invalid due date: %w", err)
			}
			dueDate = &dd
		}
		if err := todo.UpdateDueDate(dueDate); err != nil {
			return nil, fmt.Errorf("updating due date: %w", err)
		}
	}

	// Update status if provided
	if req.Status != nil {
		status, err := domain.NewTaskStatus(*req.Status)
		if err != nil {
			return nil, fmt.Errorf("invalid status: %w", err)
		}
		if err := todo.UpdateStatus(status); err != nil {
			return nil, fmt.Errorf("updating status: %w", err)
		}
	}

	// Persist changes
	if err := s.repository.Update(ctx, todo); err != nil {
		return nil, fmt.Errorf("updating todo: %w", err)
	}

	// Dispatch domain events
	if err := s.dispatcher.Dispatch(ctx, todo.Events()); err != nil {
		return nil, fmt.Errorf("dispatching events: %w", err)
	}

	// Clear events after dispatching
	todo.ClearEvents()

	// Map to response DTO
	return MapTodoToResponse(todo), nil
}

// CompleteTodo marks a todo as completed
func (s *TodoApplicationService) CompleteTodo(
	ctx context.Context,
	id string,
) (*TodoResponse, error) {
	// Parse and validate ID
	todoID, err := domain.ParseTodoID(id)
	if err != nil {
		return nil, fmt.Errorf("invalid todo ID: %w", err)
	}

	// Retrieve existing todo
	todo, err := s.repository.FindByID(ctx, todoID)
	if err != nil {
		return nil, fmt.Errorf("finding todo: %w", err)
	}

	// Complete the todo
	if err := todo.Complete(); err != nil {
		return nil, fmt.Errorf("completing todo: %w", err)
	}

	// Persist changes
	if err := s.repository.Update(ctx, todo); err != nil {
		return nil, fmt.Errorf("updating todo: %w", err)
	}

	// Dispatch domain events
	if err := s.dispatcher.Dispatch(ctx, todo.Events()); err != nil {
		return nil, fmt.Errorf("dispatching events: %w", err)
	}

	// Clear events after dispatching
	todo.ClearEvents()

	// Map to response DTO
	return MapTodoToResponse(todo), nil
}

// ReopenTodo reopens a completed or cancelled todo
func (s *TodoApplicationService) ReopenTodo(
	ctx context.Context,
	id string,
) (*TodoResponse, error) {
	// Parse and validate ID
	todoID, err := domain.ParseTodoID(id)
	if err != nil {
		return nil, fmt.Errorf("invalid todo ID: %w", err)
	}

	// Retrieve existing todo
	todo, err := s.repository.FindByID(ctx, todoID)
	if err != nil {
		return nil, fmt.Errorf("finding todo: %w", err)
	}

	// Reopen the todo
	if err := todo.Reopen(); err != nil {
		return nil, fmt.Errorf("reopening todo: %w", err)
	}

	// Persist changes
	if err := s.repository.Update(ctx, todo); err != nil {
		return nil, fmt.Errorf("updating todo: %w", err)
	}

	// Dispatch domain events
	if err := s.dispatcher.Dispatch(ctx, todo.Events()); err != nil {
		return nil, fmt.Errorf("dispatching events: %w", err)
	}

	// Clear events after dispatching
	todo.ClearEvents()

	// Map to response DTO
	return MapTodoToResponse(todo), nil
}

// DeleteTodo deletes a todo
func (s *TodoApplicationService) DeleteTodo(
	ctx context.Context,
	id string,
) error {
	// Parse and validate ID
	todoID, err := domain.ParseTodoID(id)
	if err != nil {
		return fmt.Errorf("invalid todo ID: %w", err)
	}

	// Delete from repository
	if err := s.repository.Delete(ctx, todoID); err != nil {
		return fmt.Errorf("deleting todo: %w", err)
	}

	// Create and dispatch deleted event
	deletedEvent := domain.NewTodoDeletedEvent(todoID)
	if err := s.dispatcher.Dispatch(ctx, []domain.DomainEvent{deletedEvent}); err != nil {
		return fmt.Errorf("dispatching events: %w", err)
	}

	return nil
}

// ListTodos retrieves todos with optional filters
func (s *TodoApplicationService) ListTodos(
	ctx context.Context,
	filters ListFilters,
) (*ListTodosResponse, error) {
	// Convert application filters to repository filters
	repoFilters := ports.Filters{
		Limit:  filters.Limit,
		Offset: filters.Offset,
	}

	if filters.Status != nil {
		status, err := domain.NewTaskStatus(*filters.Status)
		if err != nil {
			return nil, fmt.Errorf("invalid status filter: %w", err)
		}
		repoFilters.Status = &status
	}

	if filters.Priority != nil {
		priority, err := domain.NewPriority(*filters.Priority)
		if err != nil {
			return nil, fmt.Errorf("invalid priority filter: %w", err)
		}
		repoFilters.Priority = &priority
	}

	// Retrieve todos from repository
	todos, err := s.repository.FindAll(ctx, repoFilters)
	if err != nil {
		return nil, fmt.Errorf("finding todos: %w", err)
	}

	// Map to response DTOs
	return &ListTodosResponse{
		Todos:      MapTodosToResponse(todos),
		TotalCount: len(todos),
	}, nil
}
