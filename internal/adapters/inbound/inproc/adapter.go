// modules/todo/internal/adapters/inbound/inproc/adapter.go
package inproc

import (
	"context"

	"github.com/rotisserie/eris"
	"google.golang.org/protobuf/types/known/timestamppb"

	deftodo "github.com/pivaldi/mmw-contracts/definitions/todo"
	todov1 "github.com/pivaldi/mmw-contracts/gen/go/todo/v1"
	"github.com/pivaldi/mmw-todo/internal/application"
	"github.com/pivaldi/mmw-todo/internal/application/dto"
	"github.com/pivaldi/mmw-todo/internal/domain"
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
		Priority:    mapPriorityFromProto(req.GetPriority()),
	}

	if req.GetDueDate() != nil {
		t := req.GetDueDate().AsTime()
		appReq.DueDate = &t
	}

	todo, err := a.svc.CreateTodo(ctx, &appReq)
	if err != nil {
		return nil, eris.Wrap(err, "create todo")
	}

	return &todov1.CreateTodoResponse{Todo: mapTodoToProto(todo)}, nil
}

func (a *Adapter) GetTodo(ctx context.Context, req *todov1.GetTodoRequest) (*todov1.GetTodoResponse, error) {
	todo, err := a.svc.GetTodo(ctx, req.GetId())
	if err != nil {
		return nil, eris.Wrap(err, "get todo")
	}

	return &todov1.GetTodoResponse{Todo: mapTodoToProto(todo)}, nil
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
		priority := mapPriorityFromProto(req.GetPriority())
		appReq.Priority = &priority
	}

	if req.Status != nil {
		status := mapStatusFromProto(req.GetStatus())
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

	return &todov1.UpdateTodoResponse{Todo: mapTodoToProto(todo)}, nil
}

func (a *Adapter) CompleteTodo(ctx context.Context, req *todov1.CompleteTodoRequest) (*todov1.CompleteTodoResponse, error) {
	todo, err := a.svc.CompleteTodo(ctx, req.GetId())
	if err != nil {
		return nil, eris.Wrap(err, "complete todo")
	}

	return &todov1.CompleteTodoResponse{Todo: mapTodoToProto(todo)}, nil
}

func (a *Adapter) ReopenTodo(ctx context.Context, req *todov1.ReopenTodoRequest) (*todov1.ReopenTodoResponse, error) {
	todo, err := a.svc.ReopenTodo(ctx, req.GetId())
	if err != nil {
		return nil, eris.Wrap(err, "reopen todo")
	}

	return &todov1.ReopenTodoResponse{Todo: mapTodoToProto(todo)}, nil
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
		status := mapStatusFromProto(req.GetStatus())
		filters.Status = &status
	}

	if req.Priority != nil {
		priority := mapPriorityFromProto(req.GetPriority())
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
		protoTodos[i] = mapTodoToProto(todo)
	}

	return &todov1.ListTodosResponse{
		Todos:      protoTodos,
		TotalCount: int32(result.TotalCount),
	}, nil
}

// — mapping helpers —

func mapTodoToProto(todo *dto.TodoResponse) *todov1.Todo {
	protoTodo := &todov1.Todo{
		Id:          todo.ID,
		Title:       todo.Title,
		Description: todo.Description,
		Status:      mapStatusToProto(todo.Status),
		Priority:    mapPriorityToProto(todo.Priority),
		CreatedAt:   timestamppb.New(todo.CreatedAt),
		UpdatedAt:   timestamppb.New(todo.UpdatedAt),
	}

	if todo.DueDate != nil {
		protoTodo.DueDate = timestamppb.New(*todo.DueDate)
	}

	return protoTodo
}

func mapStatusToProto(status domain.TaskStatus) todov1.TaskStatus {
	switch status {
	case domain.TaskStatusPending:
		return todov1.TaskStatus_TASK_STATUS_PENDING
	case domain.TaskStatusInProgress:
		return todov1.TaskStatus_TASK_STATUS_IN_PROGRESS
	case domain.TaskStatusCompleted:
		return todov1.TaskStatus_TASK_STATUS_COMPLETED
	case domain.TaskStatusCancelled:
		return todov1.TaskStatus_TASK_STATUS_CANCELLED
	default:
		return todov1.TaskStatus_TASK_STATUS_UNSPECIFIED
	}
}

func mapPriorityToProto(priority domain.Priority) todov1.Priority {
	switch priority {
	case domain.PriorityLow:
		return todov1.Priority_PRIORITY_LOW
	case domain.PriorityMedium:
		return todov1.Priority_PRIORITY_MEDIUM
	case domain.PriorityHigh:
		return todov1.Priority_PRIORITY_HIGH
	case domain.PriorityUrgent:
		return todov1.Priority_PRIORITY_URGENT
	default:
		return todov1.Priority_PRIORITY_UNSPECIFIED
	}
}

func mapStatusFromProto(status todov1.TaskStatus) domain.TaskStatus {
	switch status {
	case todov1.TaskStatus_TASK_STATUS_IN_PROGRESS:
		return domain.TaskStatusInProgress
	case todov1.TaskStatus_TASK_STATUS_COMPLETED:
		return domain.TaskStatusCompleted
	case todov1.TaskStatus_TASK_STATUS_CANCELLED:
		return domain.TaskStatusCancelled
	default:
		return domain.TaskStatusPending
	}
}

func mapPriorityFromProto(priority todov1.Priority) domain.Priority {
	switch priority {
	case todov1.Priority_PRIORITY_LOW:
		return domain.PriorityLow
	case todov1.Priority_PRIORITY_HIGH:
		return domain.PriorityHigh
	case todov1.Priority_PRIORITY_URGENT:
		return domain.PriorityUrgent
	default:
		return domain.PriorityMedium
	}
}
