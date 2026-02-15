package connect

import (
	"context"
	"errors"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/types/known/timestamppb"

	todov1 "github.com/pivaldi/mmw/contracts/go/todo/v1"
	"github.com/pivaldi/mmw/todo/internal/application"
	domain "github.com/pivaldi/mmw/todo/internal/domain/todo"
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
	appReq := application.CreateTodoRequest{
		Title:       req.Msg.Title,
		Description: req.Msg.Description,
		Priority:    mapPriorityFromProto(req.Msg.Priority),
	}

	// Handle optional due date
	if req.Msg.DueDate != nil {
		dueDate := req.Msg.DueDate.AsTime()
		appReq.DueDate = &dueDate
	}

	// Call application service
	todo, err := h.service.CreateTodo(ctx, appReq)
	if err != nil {
		return nil, mapDomainError(err)
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
	todo, err := h.service.GetTodo(ctx, req.Msg.Id)
	if err != nil {
		return nil, mapDomainError(err)
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
	appReq := application.UpdateTodoRequest{}

	if req.Msg.Title != nil {
		appReq.Title = req.Msg.Title
	}

	if req.Msg.Description != nil {
		appReq.Description = req.Msg.Description
	}

	if req.Msg.Priority != nil {
		priority := mapPriorityFromProto(*req.Msg.Priority)
		appReq.Priority = &priority
	}

	if req.Msg.Status != nil {
		status := mapStatusFromProto(*req.Msg.Status)
		appReq.Status = &status
	}

	if req.Msg.DueDate != nil {
		dueDate := req.Msg.DueDate.AsTime()
		appReq.DueDate = &dueDate
	}

	// Call application service
	todo, err := h.service.UpdateTodo(ctx, req.Msg.Id, appReq)
	if err != nil {
		return nil, mapDomainError(err)
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
	todo, err := h.service.CompleteTodo(ctx, req.Msg.Id)
	if err != nil {
		return nil, mapDomainError(err)
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
	todo, err := h.service.ReopenTodo(ctx, req.Msg.Id)
	if err != nil {
		return nil, mapDomainError(err)
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
	err := h.service.DeleteTodo(ctx, req.Msg.Id)
	if err != nil {
		return nil, mapDomainError(err)
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
	filters := application.ListFilters{}

	// Convert int32 pointers to int pointers
	if req.Msg.Limit != nil {
		limit := int(*req.Msg.Limit)
		filters.Limit = &limit
	}

	if req.Msg.Offset != nil {
		offset := int(*req.Msg.Offset)
		filters.Offset = &offset
	}

	if req.Msg.Status != nil {
		status := mapStatusFromProto(*req.Msg.Status)
		filters.Status = &status
	}

	if req.Msg.Priority != nil {
		priority := mapPriorityFromProto(*req.Msg.Priority)
		filters.Priority = &priority
	}

	// Call application service
	result, err := h.service.ListTodos(ctx, filters)
	if err != nil {
		return nil, mapDomainError(err)
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
func mapTodoToProto(todo *application.TodoResponse) *todov1.Todo {
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
func mapStatusToProto(status string) todov1.TaskStatus {
	switch status {
	case "pending":
		return todov1.TaskStatus_TASK_STATUS_PENDING
	case "in_progress":
		return todov1.TaskStatus_TASK_STATUS_IN_PROGRESS
	case "completed":
		return todov1.TaskStatus_TASK_STATUS_COMPLETED
	case "cancelled":
		return todov1.TaskStatus_TASK_STATUS_CANCELLED
	default:
		return todov1.TaskStatus_TASK_STATUS_UNSPECIFIED
	}
}

// mapPriorityToProto converts a priority string to protobuf enum
func mapPriorityToProto(priority string) todov1.Priority {
	switch priority {
	case "low":
		return todov1.Priority_PRIORITY_LOW
	case "medium":
		return todov1.Priority_PRIORITY_MEDIUM
	case "high":
		return todov1.Priority_PRIORITY_HIGH
	case "urgent":
		return todov1.Priority_PRIORITY_URGENT
	default:
		return todov1.Priority_PRIORITY_UNSPECIFIED
	}
}

// mapStatusFromProto converts a protobuf status enum to string
func mapStatusFromProto(status todov1.TaskStatus) string {
	switch status {
	case todov1.TaskStatus_TASK_STATUS_PENDING:
		return "pending"
	case todov1.TaskStatus_TASK_STATUS_IN_PROGRESS:
		return "in_progress"
	case todov1.TaskStatus_TASK_STATUS_COMPLETED:
		return "completed"
	case todov1.TaskStatus_TASK_STATUS_CANCELLED:
		return "cancelled"
	default:
		return "pending"
	}
}

// mapPriorityFromProto converts a protobuf priority enum to string
func mapPriorityFromProto(priority todov1.Priority) string {
	switch priority {
	case todov1.Priority_PRIORITY_LOW:
		return "low"
	case todov1.Priority_PRIORITY_MEDIUM:
		return "medium"
	case todov1.Priority_PRIORITY_HIGH:
		return "high"
	case todov1.Priority_PRIORITY_URGENT:
		return "urgent"
	default:
		return "medium"
	}
}

// mapDomainError converts domain errors to Connect errors with appropriate codes
func mapDomainError(err error) error {
	if err == nil {
		return nil
	}

	// Check for domain-specific errors
	if errors.Is(err, domain.ErrTodoNotFound) {
		return connect.NewError(connect.CodeNotFound, err)
	}

	// Check for validation errors
	var validationErr *domain.ValidationError
	if errors.As(err, &validationErr) {
		return connect.NewError(connect.CodeInvalidArgument, err)
	}

	// Check for business rule errors
	var businessErr *domain.BusinessRuleError
	if errors.As(err, &businessErr) {
		return connect.NewError(connect.CodeFailedPrecondition, err)
	}

	// Default to internal error
	return connect.NewError(connect.CodeInternal, err)
}
