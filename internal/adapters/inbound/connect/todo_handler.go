package connect

import (
	"context"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/types/known/timestamppb"

	todov1 "github.com/pivaldi/mmw-contracts/go/network/todo/v1"
	"github.com/pivaldi/mmw-todo/internal/application"
	"github.com/pivaldi/mmw-todo/internal/application/dto"
	"github.com/pivaldi/mmw-todo/internal/domain"
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
	// Convert protobuf request to application DTO
	appReq := dto.CreateTodoRequest{
		Title:       req.Msg.GetTitle(),
		Description: req.Msg.GetDescription(),
		Priority:    mapPriorityFromProto(req.Msg.GetPriority()),
	}

	// Handle optional due date
	if req.Msg.GetDueDate() != nil {
		dueDate := req.Msg.GetDueDate().AsTime()
		appReq.DueDate = &dueDate
	}

	// Call application service
	todo, err := h.service.CreateTodo(ctx, &appReq)
	if err != nil {
		return nil, connectErrorFrom(err)
	}

	// Convert response to protobuf
	response := &todov1.CreateTodoResponse{
		Todo: mapTodoToProto(todo),
	}

	return connect.NewResponse(response), nil
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

	response := &todov1.GetTodoResponse{
		Todo: mapTodoToProto(todo),
	}

	return connect.NewResponse(response), nil
}

// UpdateTodo updates an existing todo
func (h *TodoHandler) UpdateTodo(
	ctx context.Context,
	req *connect.Request[todov1.UpdateTodoRequest],
) (*connect.Response[todov1.UpdateTodoResponse], error) {
	// Convert protobuf request to application DTO
	appReq := dto.UpdateTodoRequest{}

	if req.Msg.Title != nil {
		appReq.Title = new(req.Msg.GetTitle())
	}

	if req.Msg.Description != nil {
		appReq.Description = new(req.Msg.GetDescription())
	}

	if req.Msg.Priority != nil {
		priority := mapPriorityFromProto(req.Msg.GetPriority())
		appReq.Priority = &priority
	}

	if req.Msg.Status != nil {
		status := mapStatusFromProto(req.Msg.GetStatus())
		appReq.Status = &status
	}

	if req.Msg.GetDueDate() != nil {
		dueDate := req.Msg.GetDueDate().AsTime()
		appReq.DueDate = &dueDate
	}

	// Call application service
	todo, err := h.service.UpdateTodo(ctx, req.Msg.GetId(), &appReq)
	if err != nil {
		return nil, connectErrorFrom(err)
	}

	response := &todov1.UpdateTodoResponse{
		Todo: mapTodoToProto(todo),
	}

	return connect.NewResponse(response), nil
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

	response := &todov1.CompleteTodoResponse{
		Todo: mapTodoToProto(todo),
	}

	return connect.NewResponse(response), nil
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

	response := &todov1.ReopenTodoResponse{
		Todo: mapTodoToProto(todo),
	}

	return connect.NewResponse(response), nil
}

// DeleteTodo deletes a todo
func (h *TodoHandler) DeleteTodo(
	ctx context.Context,
	req *connect.Request[todov1.DeleteTodoRequest],
) (*connect.Response[todov1.DeleteTodoResponse], error) {
	err := h.service.DeleteTodo(ctx, req.Msg.GetId())
	if err != nil {
		return nil, connectErrorFrom(err)
	}

	response := &todov1.DeleteTodoResponse{}

	return connect.NewResponse(response), nil
}

// ListTodos lists todos with optional filters
func (h *TodoHandler) ListTodos(
	ctx context.Context,
	req *connect.Request[todov1.ListTodosRequest],
) (*connect.Response[todov1.ListTodosResponse], error) {
	// Convert protobuf filters to application filters
	filters := dto.ListFilters{}

	// Convert int32 pointers to int pointers
	if req.Msg.Limit != nil {
		limit := int(req.Msg.GetLimit())
		filters.Limit = &limit
	}

	if req.Msg.Offset != nil {
		offset := int(req.Msg.GetOffset())
		filters.Offset = &offset
	}

	if req.Msg.Status != nil {
		status := mapStatusFromProto(req.Msg.GetStatus())
		filters.Status = &status
	}

	if req.Msg.Priority != nil {
		priority := mapPriorityFromProto(req.Msg.GetPriority())
		filters.Priority = &priority
	}

	// Call application service
	result, err := h.service.ListTodos(ctx, &filters)
	if err != nil {
		return nil, connectErrorFrom(err)
	}

	// Convert todos to protobuf
	protoTodos := make([]*todov1.Todo, len(result.Todos))
	for i, todo := range result.Todos {
		protoTodos[i] = mapTodoToProto(todo)
	}

	response := &todov1.ListTodosResponse{
		Todos:      protoTodos,
		TotalCount: int32(result.TotalCount),
	}

	return connect.NewResponse(response), nil
}

// mapTodoToProto converts an application TodoResponse to protobuf Todo
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

// mapStatusToProto converts a status string to protobuf enum
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

// mapPriorityToProto converts a priority enum to protobuf enum
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

// mapStatusFromProto converts a protobuf status enum to string
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

// mapPriorityFromProto converts a protobuf priority enum to domain enum
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
