package command

import (
	"context"
	"fmt"
	"time"

	"github.com/pivaldi/mmw/todo/internal/application/authctx"
	"github.com/pivaldi/mmw/todo/internal/application/dto"
	"github.com/pivaldi/mmw/todo/internal/application/ports"
	domain "github.com/pivaldi/mmw/todo/internal/domain/todo"
)

// UpdateTodoCommand handles updating existing todos
type UpdateTodoCommand struct {
	repository      ports.TodoRepository
	eventDispatcher ports.EventDispatcher
}

// NewUpdateTodoCommand creates a new UpdateTodoCommand
func NewUpdateTodoCommand(
	repository ports.TodoRepository,
	eventDispatcher ports.EventDispatcher,
) *UpdateTodoCommand {
	return &UpdateTodoCommand{
		repository:      repository,
		eventDispatcher: eventDispatcher,
	}
}

// Execute updates an existing todo
func (c *UpdateTodoCommand) Execute(
	ctx context.Context,
	id string,
	req *dto.UpdateTodoRequest,
) (*dto.TodoResponse, error) {
	userID, err := authctx.UserIDFromContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("update todo: %w", err)
	}

	// Parse and validate ID
	todoID, err := domain.ParseTodoID(id)
	if err != nil {
		return nil, fmt.Errorf("invalid todo ID: %w", err)
	}

	// Retrieve existing todo
	todo, err := c.repository.FindByID(ctx, todoID, userID)
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
		// Validate priority enum
		if !req.Priority.IsValid() {
			return nil, fmt.Errorf("invalid priority: %w", domain.ErrInvalidPriority)
		}
		if err := todo.UpdatePriority(*req.Priority); err != nil {
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
		if err := todo.UpdateStatus(*req.Status); err != nil {
			return nil, fmt.Errorf("updating status: %w", err)
		}
	}

	// Persist changes
	if err := c.repository.Update(ctx, todo, userID); err != nil {
		return nil, fmt.Errorf("updating todo: %w", err)
	}

	// Dispatch domain events
	if err := c.eventDispatcher.Dispatch(ctx, todo.Events()); err != nil {
		return nil, fmt.Errorf("dispatching events: %w", err)
	}

	// Clear events after dispatching
	todo.ClearEvents()

	// Map to response DTO
	return dto.MapTodoToResponse(todo), nil
}
