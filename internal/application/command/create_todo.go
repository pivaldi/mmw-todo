package command

import (
	"context"
	"fmt"

	"github.com/pivaldi/mmw/todo/internal/application/dto"
	"github.com/pivaldi/mmw/todo/internal/application/ports"
	domain "github.com/pivaldi/mmw/todo/internal/domain/todo"
)

// CreateTodoCommand handles the creation of new todos
type CreateTodoCommand struct {
	repository      ports.TodoRepository
	unitOfWork      ports.UnitOfWork
	eventDispatcher ports.EventDispatcher
}

// NewCreateTodoCommand creates a new CreateTodoCommand
func NewCreateTodoCommand(
	repository ports.TodoRepository,
	unitOfWork ports.UnitOfWork,
	eventDispatcher ports.EventDispatcher,
) *CreateTodoCommand {
	return &CreateTodoCommand{
		repository:      repository,
		unitOfWork:      unitOfWork,
		eventDispatcher: eventDispatcher,
	}
}

// Execute creates a new todo
// Use a transaction for demonstration purpose.
func (c *CreateTodoCommand) Execute(
	ctx context.Context,
	req *dto.CreateTodoRequest,
) (*dto.TodoResponse, error) {
	// 1. Create value objects from request (Pure Domain Logic - no UoW needed yet)
	title, err := domain.NewTaskTitle(req.Title)
	if err != nil {
		return nil, fmt.Errorf("invalid title: %w", err)
	}

	if !req.Priority.IsValid() {
		return nil, fmt.Errorf("invalid priority: %w", domain.ErrInvalidPriority)
	}

	var dueDate *domain.DueDate
	if req.DueDate != nil {
		dd, err := domain.NewDueDate(*req.DueDate)
		if err != nil {
			return nil, fmt.Errorf("invalid due date: %w", err)
		}
		dueDate = &dd
	}

	todo := domain.NewTodo(title, req.Description, req.Priority, dueDate)

	// Execute Infrastructure operations within the Unit of Work so with transaction.
	err = c.unitOfWork.WithTransaction(ctx, func(txCtx context.Context) error {
		// Use txCtx here so the repository uses the transaction if any!
		if err := c.repository.Save(txCtx, todo); err != nil {
			return fmt.Errorf("saving todo: %w", err)
		}

		// Dispatch events using txCtx (e.g., saving to an Outbox table in the same DB)
		if err := c.eventDispatcher.Dispatch(txCtx, todo.Events()); err != nil {
			return fmt.Errorf("dispatching events: %w", err)
		}

		return nil
	})

	// Handle UoW failure => Rollback already happened automatically
	if err != nil {
		return nil, fmt.Errorf("uow execution failed: %w", err)
	}

	// 5. Cleanup and Return
	todo.ClearEvents()

	return dto.MapTodoToResponse(todo), nil
}
