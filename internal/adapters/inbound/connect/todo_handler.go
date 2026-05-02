package connect

import (
	"context"

	"connectrpc.com/connect"

	todov1 "github.com/pivaldi/mmw-contracts/go/network/todo/v1"
	"github.com/pivaldi/mmw-todo/internal/adapters/inbound/mapper"
	"github.com/pivaldi/mmw-todo/internal/application"
	"github.com/pivaldi/mmw-todo/internal/application/dto"
)

// TodoHandler implements the Connect TodoServiceHandler interface
// It bridges HTTP/gRPC requests to the application service
type TodoHandler struct {
	service application.TodoService
}

// NewTodoHandler creates a new TodoHandler
func NewTodoHandler(service application.TodoService) *TodoHandler {
	return &TodoHandler{
		service: service,
	}
}

// CreateTodo creates a new todo item
func (h *TodoHandler) CreateTodo(
	ctx context.Context,
	req *connect.Request[todov1.CreateTodoRequest],
) (*connect.Response[todov1.CreateTodoResponse], error) {
	appReq := dto.CreateTodoRequest{
		Title:       req.Msg.GetTitle(),
		Description: req.Msg.GetDescription(),
		Priority:    mapper.PriorityFromProto(req.Msg.GetPriority()),
	}

	if req.Msg.GetDueDate() != nil {
		dueDate := req.Msg.GetDueDate().AsTime()
		appReq.DueDate = &dueDate
	}

	todo, err := h.service.CreateTodo(ctx, &appReq)
	if err != nil {
		return nil, connectErrorFrom(err)
	}

	return connect.NewResponse(&todov1.CreateTodoResponse{Todo: mapper.TodoToProto(todo)}), nil
}

// GetTodo retrieves a todo by ID
func (h *TodoHandler) GetTodo(
	ctx context.Context,
	req *connect.Request[todov1.GetTodoRequest],
) (*connect.Response[todov1.GetTodoResponse], error) {
	todo, err := h.service.GetTodo(ctx, req.Msg.GetId())
	if err != nil {
		return nil, connectErrorFrom(err)
	}

	return connect.NewResponse(&todov1.GetTodoResponse{Todo: mapper.TodoToProto(todo)}), nil
}

// UpdateTodo updates an existing todo
func (h *TodoHandler) UpdateTodo(
	ctx context.Context,
	req *connect.Request[todov1.UpdateTodoRequest],
) (*connect.Response[todov1.UpdateTodoResponse], error) {
	appReq := dto.UpdateTodoRequest{}

	if req.Msg.Title != nil {
		appReq.Title = new(req.Msg.GetTitle())
	}

	if req.Msg.Description != nil {
		appReq.Description = new(req.Msg.GetDescription())
	}

	if req.Msg.Priority != nil {
		priority := mapper.PriorityFromProto(req.Msg.GetPriority())
		appReq.Priority = &priority
	}

	if req.Msg.Status != nil {
		status := mapper.StatusFromProto(req.Msg.GetStatus())
		appReq.Status = &status
	}

	if req.Msg.GetDueDate() != nil {
		dueDate := req.Msg.GetDueDate().AsTime()
		appReq.DueDate = &dueDate
	}

	todo, err := h.service.UpdateTodo(ctx, req.Msg.GetId(), &appReq)
	if err != nil {
		return nil, connectErrorFrom(err)
	}

	return connect.NewResponse(&todov1.UpdateTodoResponse{Todo: mapper.TodoToProto(todo)}), nil
}

// CompleteTodo marks a todo as completed
func (h *TodoHandler) CompleteTodo(
	ctx context.Context,
	req *connect.Request[todov1.CompleteTodoRequest],
) (*connect.Response[todov1.CompleteTodoResponse], error) {
	todo, err := h.service.CompleteTodo(ctx, req.Msg.GetId())
	if err != nil {
		return nil, connectErrorFrom(err)
	}

	return connect.NewResponse(&todov1.CompleteTodoResponse{Todo: mapper.TodoToProto(todo)}), nil
}

// ReopenTodo reopens a completed or cancelled todo
func (h *TodoHandler) ReopenTodo(
	ctx context.Context,
	req *connect.Request[todov1.ReopenTodoRequest],
) (*connect.Response[todov1.ReopenTodoResponse], error) {
	todo, err := h.service.ReopenTodo(ctx, req.Msg.GetId())
	if err != nil {
		return nil, connectErrorFrom(err)
	}

	return connect.NewResponse(&todov1.ReopenTodoResponse{Todo: mapper.TodoToProto(todo)}), nil
}

// DeleteTodo deletes a todo
func (h *TodoHandler) DeleteTodo(
	ctx context.Context,
	req *connect.Request[todov1.DeleteTodoRequest],
) (*connect.Response[todov1.DeleteTodoResponse], error) {
	if err := h.service.DeleteTodo(ctx, req.Msg.GetId()); err != nil {
		return nil, connectErrorFrom(err)
	}

	return connect.NewResponse(&todov1.DeleteTodoResponse{}), nil
}

// ListTodos lists todos with optional filters
func (h *TodoHandler) ListTodos(
	ctx context.Context,
	req *connect.Request[todov1.ListTodosRequest],
) (*connect.Response[todov1.ListTodosResponse], error) {
	filters := dto.ListFilters{}

	if req.Msg.Limit != nil {
		limit := int(req.Msg.GetLimit())
		filters.Limit = &limit
	}

	if req.Msg.Offset != nil {
		offset := int(req.Msg.GetOffset())
		filters.Offset = &offset
	}

	if req.Msg.Status != nil {
		status := mapper.StatusFromProto(req.Msg.GetStatus())
		filters.Status = &status
	}

	if req.Msg.Priority != nil {
		priority := mapper.PriorityFromProto(req.Msg.GetPriority())
		filters.Priority = &priority
	}

	result, err := h.service.ListTodos(ctx, &filters)
	if err != nil {
		return nil, connectErrorFrom(err)
	}

	protoTodos := make([]*todov1.Todo, len(result.Todos))
	for i, todo := range result.Todos {
		protoTodos[i] = mapper.TodoToProto(todo)
	}

	return connect.NewResponse(&todov1.ListTodosResponse{
		Todos:      protoTodos,
		TotalCount: int32(result.TotalCount),
	}), nil
}
