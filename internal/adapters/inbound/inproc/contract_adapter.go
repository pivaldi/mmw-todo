package inproc

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"

	tododef "github.com/pivaldi/mmw-contracts/go/application/todo"
	todov1 "github.com/pivaldi/mmw-contracts/go/network/todo/v1"
	"github.com/pivaldi/mmw-todo/internal/adapters/inbound/mapper"
	"github.com/pivaldi/mmw-todo/internal/application"
	"github.com/pivaldi/mmw-todo/internal/application/dto"
	"github.com/pivaldi/mmw-todo/internal/domain"
)

// ContractAdapter wraps TodoService and implements tododef.TodoService,
// translating between proto-typed requests/responses and domain-idiomatic signatures.
type ContractAdapter struct {
	svc application.TodoService
}

var _ tododef.TodoService = (*ContractAdapter)(nil)

// NewContractAdapter creates a ContractAdapter around svc.
func NewContractAdapter(svc application.TodoService) *ContractAdapter {
	return &ContractAdapter{svc: svc}
}

func (a *ContractAdapter) CreateTodo(
	ctx context.Context, req *todov1.CreateTodoRequest,
) (*todov1.CreateTodoResponse, error) {
	r, err := a.svc.CreateTodo(ctx, &dto.CreateTodoRequest{
		Title:       req.GetTitle(),
		Description: req.GetDescription(),
		Priority:    protoPriorityToDomain(req.GetPriority()),
		DueDate:     protoToTime(req.GetDueDate()),
	})
	if err != nil {
		return nil, fmt.Errorf("create todo: %w", mapper.DomainErrorFor(err))
	}

	return &todov1.CreateTodoResponse{Todo: todoResponseToProto(r)}, nil
}

func (a *ContractAdapter) GetTodo(
	ctx context.Context, req *todov1.GetTodoRequest,
) (*todov1.GetTodoResponse, error) {
	r, err := a.svc.GetTodo(ctx, req.GetId())
	if err != nil {
		return nil, fmt.Errorf("get todo: %w", mapper.DomainErrorFor(err))
	}

	return &todov1.GetTodoResponse{Todo: todoResponseToProto(r)}, nil
}

func (a *ContractAdapter) UpdateTodo(
	ctx context.Context, req *todov1.UpdateTodoRequest,
) (*todov1.UpdateTodoResponse, error) {
	update := &dto.UpdateTodoRequest{}
	if req.Title != nil {
		update.Title = req.Title
	}
	if req.Description != nil {
		update.Description = req.Description
	}
	if req.Priority != nil {
		p := protoPriorityToDomain(req.GetPriority())
		update.Priority = &p
	}
	if req.DueDate != nil { //nolint:protogetter // optional field: nil check requires direct access
		update.DueDate = protoToTime(req.GetDueDate())
	}
	if req.Status != nil {
		s := protoStatusToDomain(req.GetStatus())
		update.Status = &s
	}
	r, err := a.svc.UpdateTodo(ctx, req.GetId(), update)
	if err != nil {
		return nil, fmt.Errorf("update todo: %w", mapper.DomainErrorFor(err))
	}

	return &todov1.UpdateTodoResponse{Todo: todoResponseToProto(r)}, nil
}

func (a *ContractAdapter) CompleteTodo(
	ctx context.Context, req *todov1.CompleteTodoRequest,
) (*todov1.CompleteTodoResponse, error) {
	r, err := a.svc.CompleteTodo(ctx, req.GetId())
	if err != nil {
		return nil, fmt.Errorf("complete todo: %w", mapper.DomainErrorFor(err))
	}

	return &todov1.CompleteTodoResponse{Todo: todoResponseToProto(r)}, nil
}

func (a *ContractAdapter) ReopenTodo(
	ctx context.Context, req *todov1.ReopenTodoRequest,
) (*todov1.ReopenTodoResponse, error) {
	r, err := a.svc.ReopenTodo(ctx, req.GetId())
	if err != nil {
		return nil, fmt.Errorf("reopen todo: %w", mapper.DomainErrorFor(err))
	}

	return &todov1.ReopenTodoResponse{Todo: todoResponseToProto(r)}, nil
}

func (a *ContractAdapter) DeleteTodo(
	ctx context.Context, req *todov1.DeleteTodoRequest,
) (*todov1.DeleteTodoResponse, error) {
	if err := a.svc.DeleteTodo(ctx, req.GetId()); err != nil {
		return nil, fmt.Errorf("delete todo: %w", mapper.DomainErrorFor(err))
	}

	return &todov1.DeleteTodoResponse{}, nil
}

func (a *ContractAdapter) ListTodos(
	ctx context.Context, req *todov1.ListTodosRequest,
) (*todov1.ListTodosResponse, error) {
	filters := &dto.ListFilters{}
	if req.Status != nil {
		s := protoStatusToDomain(req.GetStatus())
		filters.Status = &s
	}
	if req.Priority != nil {
		p := protoPriorityToDomain(req.GetPriority())
		filters.Priority = &p
	}
	if req.Limit != nil {
		l := int(req.GetLimit())
		filters.Limit = &l
	}
	if req.Offset != nil {
		o := int(req.GetOffset())
		filters.Offset = &o
	}
	result, err := a.svc.ListTodos(ctx, filters)
	if err != nil {
		return nil, fmt.Errorf("list todos: %w", mapper.DomainErrorFor(err))
	}
	todos := make([]*todov1.Todo, len(result.Todos))
	for i, t := range result.Todos {
		todos[i] = todoResponseToProto(t)
	}

	return &todov1.ListTodosResponse{
		Todos:      todos,
		TotalCount: int32(result.TotalCount),
	}, nil
}

// ── mapping helpers ───────────────────────────────────────────────────────────

func todoResponseToProto(r *dto.TodoResponse) *todov1.Todo {
	return &todov1.Todo{
		Id:          r.ID,
		Title:       r.Title,
		Description: r.Description,
		Status:      domainStatusToProto(r.Status),
		Priority:    domainPriorityToProto(r.Priority),
		CreatedAt:   timestamppb.New(r.CreatedAt),
		UpdatedAt:   timestamppb.New(r.UpdatedAt),
		DueDate:     timeToProto(r.DueDate),
	}
}

func domainStatusToProto(s domain.TaskStatus) todov1.TaskStatus {
	switch s {
	case domain.TaskStatusPending:
		return todov1.TaskStatus_TASK_STATUS_PENDING
	case domain.TaskStatusCompleted:
		return todov1.TaskStatus_TASK_STATUS_COMPLETED
	case domain.TaskStatusCancelled:
		return todov1.TaskStatus_TASK_STATUS_CANCELLED
	default:
		return todov1.TaskStatus_TASK_STATUS_UNSPECIFIED
	}
}

func protoStatusToDomain(s todov1.TaskStatus) domain.TaskStatus {
	switch s {
	case todov1.TaskStatus_TASK_STATUS_COMPLETED:
		return domain.TaskStatusCompleted
	case todov1.TaskStatus_TASK_STATUS_CANCELLED:
		return domain.TaskStatusCancelled
	default:
		return domain.TaskStatusPending
	}
}

func domainPriorityToProto(p domain.Priority) todov1.Priority {
	switch p {
	case domain.PriorityLow:
		return todov1.Priority_PRIORITY_LOW
	case domain.PriorityMedium:
		return todov1.Priority_PRIORITY_MEDIUM
	case domain.PriorityHigh:
		return todov1.Priority_PRIORITY_HIGH
	default:
		return todov1.Priority_PRIORITY_UNSPECIFIED
	}
}

func protoPriorityToDomain(p todov1.Priority) domain.Priority {
	switch p {
	case todov1.Priority_PRIORITY_LOW:
		return domain.PriorityLow
	case todov1.Priority_PRIORITY_HIGH:
		return domain.PriorityHigh
	default:
		return domain.PriorityMedium
	}
}

func timeToProto(t *time.Time) *timestamppb.Timestamp {
	if t == nil {
		return nil
	}

	return timestamppb.New(*t)
}

func protoToTime(ts *timestamppb.Timestamp) *time.Time {
	if ts == nil {
		return nil
	}

	t := ts.AsTime()

	return &t
}
