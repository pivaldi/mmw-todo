package dto

import (
	"time"

	"github.com/pivaldi/mmw-todo/internal/domain"
)

// CreateTodoRequest represents the data needed to create a new todo
type CreateTodoRequest struct {
	Title       string
	Description string
	Priority    domain.Priority
	DueDate     *time.Time
}

// UpdateTodoRequest represents the data for updating a todo
// All fields are optional (pointers indicate which fields to update)
type UpdateTodoRequest struct {
	Title       *string
	Description *string
	Priority    *domain.Priority
	DueDate     *time.Time
	Status      *domain.TaskStatus
}

// TodoResponse represents a todo for API responses
type TodoResponse struct {
	ID          string
	Title       string
	Description string
	Status      domain.TaskStatus
	Priority    domain.Priority
	DueDate     *time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// ListFilters represents filtering options for listing todos
type ListFilters struct {
	Status   *domain.TaskStatus
	Priority *domain.Priority
	Limit    *int
	Offset   *int
}

// ListTodosResponse represents the response for listing todos
type ListTodosResponse struct {
	Todos      []*TodoResponse
	TotalCount int
}

// MapTodoToResponse converts a domain Todo to a TodoResponse DTO
func MapTodoToResponse(todo *domain.Todo) *TodoResponse {
	response := &TodoResponse{
		ID:          todo.ID().String(),
		Title:       todo.Title().String(),
		Description: todo.Description(),
		Status:      todo.Status(),
		Priority:    todo.Priority(),
		CreatedAt:   todo.CreatedAt(),
		UpdatedAt:   todo.UpdatedAt(),
	}

	if todo.DueDate() != nil {
		dueDate := todo.DueDate().Time()
		response.DueDate = &dueDate
	}

	return response
}

// MapTodosToResponse converts multiple domain Todos to TodoResponse DTOs
func MapTodosToResponse(todos []*domain.Todo) []*TodoResponse {
	responses := make([]*TodoResponse, len(todos))
	for i, todo := range todos {
		responses[i] = MapTodoToResponse(todo)
	}

	return responses
}
