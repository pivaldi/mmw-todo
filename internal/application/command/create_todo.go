package command

import (
	"context"

	"github.com/rotisserie/eris"

	"github.com/pivaldi/mmw-todo/internal/application/authctx"
	"github.com/pivaldi/mmw-todo/internal/application/dto"
	"github.com/pivaldi/mmw-todo/internal/application/ports"
	"github.com/pivaldi/mmw-todo/internal/domain"
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
	userID, err := authctx.UserIDFromContext(ctx)
	if err != nil {
		return nil, eris.Wrap(err, "create todo")
	}

	// 1. Create value objects from request (Pure Domain Logic - no UoW needed yet)
	title, err := domain.NewTaskTitle(req.Title)
	if err != nil {
		return nil, eris.Wrap(err, "invalid title")
	}

	if !req.Priority.IsValid() {
		return nil, eris.Wrap(domain.ErrInvalidPriority, "invalid priority")
	}

	var dueDate *domain.DueDate
	if req.DueDate != nil {
		dd, err := domain.NewDueDate(*req.DueDate)
		if err != nil {
			return nil, eris.Wrap(err, "invalid due date")
		}
		dueDate = &dd
	}

	todo := domain.NewTodo(title, req.Description, req.Priority, dueDate, userID)

	// Execute Infrastructure operations within the Unit of Work so with transaction.
	err = c.unitOfWork.WithTransaction(ctx, func(txCtx context.Context) error {
		// Use txCtx here so the repository uses the transaction if any!
		if err := c.repository.Save(txCtx, todo); err != nil {
			return eris.Wrap(err, "saving todo")
		}

		// Dispatch events using txCtx (e.g., saving to an Outbox table in the same DB)
		if err := c.eventDispatcher.Dispatch(txCtx, todo.Events()); err != nil {
			return eris.Wrap(err, "dispatching events")
		}

		return nil
	})
	// Handle UoW failure => Rollback already happened automatically
	if err != nil {
		return nil, eris.Wrap(err, "uow execution failed")
	}

	// 5. Cleanup and Return
	todo.ClearEvents()

	return dto.MapTodoToResponse(todo), nil
}
