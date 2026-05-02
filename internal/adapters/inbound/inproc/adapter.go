// modules/todo/internal/adapters/inbound/inproc/adapter.go
package inproc

import (
	"context"

	"github.com/rotisserie/eris"

	deftodo "github.com/pivaldi/mmw-contracts/go/application/todo"
	todov1 "github.com/pivaldi/mmw-contracts/go/network/todo/v1"
	"github.com/pivaldi/mmw-todo/internal/adapters/inbound/mapper"
	"github.com/pivaldi/mmw-todo/internal/application"
	"github.com/pivaldi/mmw-todo/internal/application/dto"
)

// Adapter wraps application.TodoService and implements deftodo.TodoService.
// It maps proto request/response types to internal DTOs so other modules can
// call the todo service in-process without depending on its internal packages.
type Adapter struct {
	svc application.TodoService
}

// compile-time assertion
var _ deftodo.TodoService = (*Adapter)(nil)

func NewAdapter(svc application.TodoService) *Adapter {
	return &Adapter{svc: svc}
}

func (a *Adapter) CreateTodo(ctx context.Context, req *todov1.CreateTodoRequest) (*todov1.CreateTodoResponse, error) {
	appReq := dto.CreateTodoRequest{
		Title:       req.GetTitle(),
		Description: req.GetDescription(),
		Priority:    mapper.PriorityFromProto(req.GetPriority()),
	}

	if req.GetDueDate() != nil {
		t := req.GetDueDate().AsTime()
		appReq.DueDate = &t
	}

	todo, err := a.svc.CreateTodo(ctx, &appReq)
	if err != nil {
		return nil, eris.Wrap(err, "create todo")
	}

	return &todov1.CreateTodoResponse{Todo: mapper.TodoToProto(todo)}, nil
}

func (a *Adapter) GetTodo(ctx context.Context, req *todov1.GetTodoRequest) (*todov1.GetTodoResponse, error) {
	todo, err := a.svc.GetTodo(ctx, req.GetId())
	if err != nil {
		return nil, eris.Wrap(err, "get todo")
	}

	return &todov1.GetTodoResponse{Todo: mapper.TodoToProto(todo)}, nil
}

func (a *Adapter) UpdateTodo(ctx context.Context, req *todov1.UpdateTodoRequest) (*todov1.UpdateTodoResponse, error) {
	appReq := dto.UpdateTodoRequest{}

	if req.Title != nil {
		title := req.GetTitle()
		appReq.Title = &title
	}

	if req.Description != nil {
		desc := req.GetDescription()
		appReq.Description = &desc
	}

	if req.Priority != nil {
		priority := mapper.PriorityFromProto(req.GetPriority())
		appReq.Priority = &priority
	}

	if req.Status != nil {
		status := mapper.StatusFromProto(req.GetStatus())
		appReq.Status = &status
	}

	if req.GetDueDate() != nil {
		t := req.GetDueDate().AsTime()
		appReq.DueDate = &t
	}

	todo, err := a.svc.UpdateTodo(ctx, req.GetId(), &appReq)
	if err != nil {
		return nil, eris.Wrap(err, "update todo")
	}

	return &todov1.UpdateTodoResponse{Todo: mapper.TodoToProto(todo)}, nil
}

func (a *Adapter) CompleteTodo(
	ctx context.Context, req *todov1.CompleteTodoRequest,
) (*todov1.CompleteTodoResponse, error) {
	todo, err := a.svc.CompleteTodo(ctx, req.GetId())
	if err != nil {
		return nil, eris.Wrap(err, "complete todo")
	}

	return &todov1.CompleteTodoResponse{Todo: mapper.TodoToProto(todo)}, nil
}

func (a *Adapter) ReopenTodo(ctx context.Context, req *todov1.ReopenTodoRequest) (*todov1.ReopenTodoResponse, error) {
	todo, err := a.svc.ReopenTodo(ctx, req.GetId())
	if err != nil {
		return nil, eris.Wrap(err, "reopen todo")
	}

	return &todov1.ReopenTodoResponse{Todo: mapper.TodoToProto(todo)}, nil
}

func (a *Adapter) DeleteTodo(ctx context.Context, req *todov1.DeleteTodoRequest) (*todov1.DeleteTodoResponse, error) {
	if err := a.svc.DeleteTodo(ctx, req.GetId()); err != nil {
		return nil, eris.Wrap(err, "delete todo")
	}

	return &todov1.DeleteTodoResponse{}, nil
}

func (a *Adapter) ListTodos(ctx context.Context, req *todov1.ListTodosRequest) (*todov1.ListTodosResponse, error) {
	filters := dto.ListFilters{}

	if req.Status != nil {
		status := mapper.StatusFromProto(req.GetStatus())
		filters.Status = &status
	}

	if req.Priority != nil {
		priority := mapper.PriorityFromProto(req.GetPriority())
		filters.Priority = &priority
	}

	if req.Limit != nil {
		limit := int(req.GetLimit())
		filters.Limit = &limit
	}

	if req.Offset != nil {
		offset := int(req.GetOffset())
		filters.Offset = &offset
	}

	result, err := a.svc.ListTodos(ctx, &filters)
	if err != nil {
		return nil, eris.Wrap(err, "list todos")
	}

	protoTodos := make([]*todov1.Todo, len(result.Todos))
	for i, todo := range result.Todos {
		protoTodos[i] = mapper.TodoToProto(todo)
	}

	return &todov1.ListTodosResponse{
		Todos:      protoTodos,
		TotalCount: int32(result.TotalCount),
	}, nil
}
