package connect

import (
	"context"
	"errors"
	"testing"
	"time"

	"connectrpc.com/connect"
	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/timestamppb"

	todov1 "github.com/pivaldi/mmw-contracts/go/network/todo/v1"
	pfauthctx "github.com/piprim/mmw/pkg/platform/authctx"
	"github.com/pivaldi/mmw-todo/internal/domain"
	"github.com/pivaldi/mmw-todo/internal/testhelpers"
)

// newTestHandler wires a real TodoApplicationService with in-memory fakes
// and returns the handler and repo for test setup.
func newTestHandler(t *testing.T) (*TodoHandler, *testhelpers.InMemoryTodoRepo) {
	t.Helper()
	svc, repo := testhelpers.NewTestService(t)
	return NewTodoHandler(svc), repo
}

// testCtx returns a context with a fixed userID injected, required by the
// application layer for all operations.
func testCtx() context.Context {
	return pfauthctx.WithUserID(context.Background(), uuid.MustParse("00000000-0000-0000-0000-000000000001"))
}

func TestTodoHandler_CreateTodo_Success(t *testing.T) {
	handler, _ := newTestHandler(t)

	req := connect.NewRequest(&todov1.CreateTodoRequest{
		Title:       "Test Todo",
		Description: "Test description",
		Priority:    todov1.Priority_PRIORITY_MEDIUM,
	})

	resp, err := handler.CreateTodo(testCtx(), req)
	if err != nil {
		t.Fatalf("CreateTodo() unexpected error: %v", err)
	}
	if resp.Msg.Todo.Title != "Test Todo" {
		t.Errorf("Title = %v, want %v", resp.Msg.Todo.Title, "Test Todo")
	}
	if resp.Msg.Todo.Status != todov1.TaskStatus_TASK_STATUS_PENDING {
		t.Errorf("Status = %v, want PENDING", resp.Msg.Todo.Status)
	}
}

func TestTodoHandler_CreateTodo_WithDueDate_Success(t *testing.T) {
	handler, _ := newTestHandler(t)
	dueDate := time.Now().Add(24 * time.Hour)

	req := connect.NewRequest(&todov1.CreateTodoRequest{
		Title:       "Test Todo",
		Description: "Test description",
		Priority:    todov1.Priority_PRIORITY_HIGH,
		DueDate:     timestamppb.New(dueDate),
	})

	resp, err := handler.CreateTodo(testCtx(), req)
	if err != nil {
		t.Fatalf("CreateTodo() unexpected error: %v", err)
	}
	if resp.Msg.Todo.DueDate == nil {
		t.Error("Expected due date in response")
	}
}

func TestTodoHandler_GetTodo_Success(t *testing.T) {
	handler, _ := newTestHandler(t)
	ctx := testCtx()

	createResp, err := handler.CreateTodo(ctx, connect.NewRequest(&todov1.CreateTodoRequest{
		Title:    "Fetch Me",
		Priority: todov1.Priority_PRIORITY_MEDIUM,
	}))
	if err != nil {
		t.Fatalf("CreateTodo() unexpected error: %v", err)
	}
	id := createResp.Msg.Todo.Id

	resp, err := handler.GetTodo(ctx, connect.NewRequest(&todov1.GetTodoRequest{Id: id}))
	if err != nil {
		t.Fatalf("GetTodo() unexpected error: %v", err)
	}
	if resp.Msg.Todo.Id != id {
		t.Errorf("ID = %v, want %v", resp.Msg.Todo.Id, id)
	}
}

func TestTodoHandler_GetTodo_NotFound_ReturnsNotFoundError(t *testing.T) {
	handler, _ := newTestHandler(t)

	_, err := handler.GetTodo(testCtx(), connect.NewRequest(&todov1.GetTodoRequest{
		Id: uuid.New().String(), // valid UUID format, but not in the repo
	}))

	if err == nil {
		t.Fatal("GetTodo() expected error, got nil")
	}
	var connectErr *connect.Error
	if !errors.As(err, &connectErr) {
		t.Fatalf("Expected *connect.Error, got %T", err)
	}
	if connectErr.Code() != connect.CodeNotFound {
		t.Errorf("Code() = %v, want CodeNotFound", connectErr.Code())
	}
}

func TestTodoHandler_UpdateTodo_Success(t *testing.T) {
	handler, _ := newTestHandler(t)
	ctx := testCtx()

	createResp, err := handler.CreateTodo(ctx, connect.NewRequest(&todov1.CreateTodoRequest{
		Title:    "Original Title",
		Priority: todov1.Priority_PRIORITY_MEDIUM,
	}))
	if err != nil {
		t.Fatalf("CreateTodo() unexpected error: %v", err)
	}
	id := createResp.Msg.Todo.Id
	newTitle := "Updated Title"

	resp, err := handler.UpdateTodo(ctx, connect.NewRequest(&todov1.UpdateTodoRequest{
		Id:    id,
		Title: &newTitle,
	}))
	if err != nil {
		t.Fatalf("UpdateTodo() unexpected error: %v", err)
	}
	if resp.Msg.Todo.Title != newTitle {
		t.Errorf("Title = %v, want %v", resp.Msg.Todo.Title, newTitle)
	}
}

func TestTodoHandler_CompleteTodo_Success(t *testing.T) {
	handler, _ := newTestHandler(t)
	ctx := testCtx()

	createResp, err := handler.CreateTodo(ctx, connect.NewRequest(&todov1.CreateTodoRequest{
		Title:    "Complete Me",
		Priority: todov1.Priority_PRIORITY_MEDIUM,
	}))
	if err != nil {
		t.Fatalf("CreateTodo() unexpected error: %v", err)
	}

	resp, err := handler.CompleteTodo(ctx, connect.NewRequest(&todov1.CompleteTodoRequest{
		Id: createResp.Msg.Todo.Id,
	}))
	if err != nil {
		t.Fatalf("CompleteTodo() unexpected error: %v", err)
	}
	if resp.Msg.Todo.Status != todov1.TaskStatus_TASK_STATUS_COMPLETED {
		t.Errorf("Status = %v, want COMPLETED", resp.Msg.Todo.Status)
	}
}

func TestTodoHandler_ReopenTodo_Success(t *testing.T) {
	handler, _ := newTestHandler(t)
	ctx := testCtx()

	createResp, _ := handler.CreateTodo(ctx, connect.NewRequest(&todov1.CreateTodoRequest{
		Title:    "Reopen Me",
		Priority: todov1.Priority_PRIORITY_MEDIUM,
	}))
	_, _ = handler.CompleteTodo(ctx, connect.NewRequest(&todov1.CompleteTodoRequest{
		Id: createResp.Msg.Todo.Id,
	}))

	resp, err := handler.ReopenTodo(ctx, connect.NewRequest(&todov1.ReopenTodoRequest{
		Id: createResp.Msg.Todo.Id,
	}))
	if err != nil {
		t.Fatalf("ReopenTodo() unexpected error: %v", err)
	}
	if resp.Msg.Todo.Status != todov1.TaskStatus_TASK_STATUS_PENDING {
		t.Errorf("Status = %v, want PENDING", resp.Msg.Todo.Status)
	}
}

func TestTodoHandler_DeleteTodo_Success(t *testing.T) {
	handler, _ := newTestHandler(t)
	ctx := testCtx()

	createResp, _ := handler.CreateTodo(ctx, connect.NewRequest(&todov1.CreateTodoRequest{
		Title:    "Delete Me",
		Priority: todov1.Priority_PRIORITY_MEDIUM,
	}))
	id := createResp.Msg.Todo.Id

	_, err := handler.DeleteTodo(ctx, connect.NewRequest(&todov1.DeleteTodoRequest{Id: id}))
	if err != nil {
		t.Fatalf("DeleteTodo() unexpected error: %v", err)
	}

	_, err = handler.GetTodo(ctx, connect.NewRequest(&todov1.GetTodoRequest{Id: id}))
	if err == nil {
		t.Fatal("GetTodo() after delete expected error, got nil")
	}
}

func TestTodoHandler_ListTodos_Success(t *testing.T) {
	handler, _ := newTestHandler(t)
	ctx := testCtx()

	_, _ = handler.CreateTodo(ctx, connect.NewRequest(&todov1.CreateTodoRequest{
		Title:    "Todo 1",
		Priority: todov1.Priority_PRIORITY_MEDIUM,
	}))
	_, _ = handler.CreateTodo(ctx, connect.NewRequest(&todov1.CreateTodoRequest{
		Title:    "Todo 2",
		Priority: todov1.Priority_PRIORITY_HIGH,
	}))

	status := todov1.TaskStatus_TASK_STATUS_PENDING
	limit := int32(10)
	offset := int32(0)
	resp, err := handler.ListTodos(ctx, connect.NewRequest(&todov1.ListTodosRequest{
		Status: &status,
		Limit:  &limit,
		Offset: &offset,
	}))
	if err != nil {
		t.Fatalf("ListTodos() unexpected error: %v", err)
	}
	if len(resp.Msg.Todos) != 2 {
		t.Errorf("len(Todos) = %v, want 2", len(resp.Msg.Todos))
	}
}

func TestTodoHandler_DomainError_InvalidTitle_ReturnsInvalidArgument(t *testing.T) {
	handler, _ := newTestHandler(t)

	_, err := handler.CreateTodo(testCtx(), connect.NewRequest(&todov1.CreateTodoRequest{
		Title:    "",
		Priority: todov1.Priority_PRIORITY_MEDIUM,
	}))

	if err == nil {
		t.Fatal("CreateTodo() expected error, got nil")
	}
	var connectErr *connect.Error
	if !errors.As(err, &connectErr) {
		t.Fatalf("Expected *connect.Error, got %T", err)
	}
	if connectErr.Code() != connect.CodeInvalidArgument {
		t.Errorf("Code() = %v, want CodeInvalidArgument", connectErr.Code())
	}
}

func TestTodoHandler_DomainError_CannotCompleteCancelled_ReturnsFailedPrecondition(t *testing.T) {
	handler, repo := newTestHandler(t)
	ctx := testCtx()
	userID := uuid.MustParse("00000000-0000-0000-0000-000000000001")

	// Seed a cancelled todo directly via the repo so we can try to complete it.
	title, _ := domain.NewTaskTitle("Cancelled Todo")
	todo := domain.NewTodo(title, "", domain.PriorityMedium, nil, userID)
	_ = todo.Cancel()
	_ = repo.Save(ctx, todo)

	_, err := handler.CompleteTodo(ctx, connect.NewRequest(&todov1.CompleteTodoRequest{
		Id: todo.ID().String(),
	}))
	if err == nil {
		t.Fatal("CompleteTodo() on cancelled todo expected error, got nil")
	}
	var connectErr *connect.Error
	if !errors.As(err, &connectErr) {
		t.Fatalf("Expected *connect.Error, got %T", err)
	}
	if connectErr.Code() != connect.CodeFailedPrecondition {
		t.Errorf("Code() = %v, want CodeFailedPrecondition", connectErr.Code())
	}
}
