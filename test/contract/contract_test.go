// Package contract tests that inproc.Adapter correctly implements the
// deftodo.TodoService contract — including DTO mapping, enum translation,
// and error propagation. No infrastructure required.
package contract

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	todov1 "github.com/pivaldi/mmw-contracts/go/network/todo/v1"
	tododef "github.com/pivaldi/mmw-contracts/go/application/todo"
	"github.com/pivaldi/mmw-todo/internal/adapters/inbound/inproc"
	"github.com/pivaldi/mmw-todo/internal/adapters/inbound/mapper"
	pfauthctx "github.com/piprim/mmw/pkg/platform/authctx"
	"github.com/pivaldi/mmw-todo/internal/domain"
	"github.com/pivaldi/mmw-todo/internal/testhelpers"
	"github.com/piprim/mmw/pkg/platform"
)

func newAdapter(t *testing.T) (*inproc.Adapter, *testhelpers.InMemoryTodoRepo) {
	t.Helper()
	svc, repo := testhelpers.NewTestService(t)
	return inproc.NewAdapter(svc), repo
}

func authedCtx() context.Context {
	return pfauthctx.WithUserID(context.Background(), uuid.MustParse("00000000-0000-0000-0000-000000000001"))
}

// isDomainErrorCode checks whether err carries the given domain error code.
// The application layer converts domain sentinels into *platform.DomainError,
// so tests must match by code rather than by sentinel identity.
// want must be a value castable to platform.ErrorCode (e.g. tododef.ErrorCodeNotFound).
func isDomainErrorCode(err error, want platform.ErrorCode) bool {
	mapped := mapper.DomainErrorFor(err)
	de, ok := errors.AsType[*platform.DomainError](mapped)
	if !ok {
		return false
	}
	return de.Code == want
}

func TestContract_CreateTodo(t *testing.T) {
	adapter, _ := newAdapter(t)
	ctx := authedCtx()

	resp, err := adapter.CreateTodo(ctx, &todov1.CreateTodoRequest{
		Title:       "Buy groceries",
		Description: "Milk, eggs, bread",
		Priority:    todov1.Priority_PRIORITY_HIGH,
	})
	if err != nil {
		t.Fatalf("CreateTodo() unexpected error: %v", err)
	}
	if resp.Todo == nil {
		t.Fatal("CreateTodo() returned nil Todo")
	}
	if resp.Todo.Title != "Buy groceries" {
		t.Errorf("Title = %q, want %q", resp.Todo.Title, "Buy groceries")
	}
	if resp.Todo.Description != "Milk, eggs, bread" {
		t.Errorf("Description = %q, want %q", resp.Todo.Description, "Milk, eggs, bread")
	}
	if resp.Todo.Priority != todov1.Priority_PRIORITY_HIGH {
		t.Errorf("Priority = %v, want PRIORITY_HIGH", resp.Todo.Priority)
	}
	if resp.Todo.Status != todov1.TaskStatus_TASK_STATUS_PENDING {
		t.Errorf("Status = %v, want TASK_STATUS_PENDING", resp.Todo.Status)
	}
	if resp.Todo.Id == "" {
		t.Error("ID must not be empty")
	}
	if resp.Todo.CreatedAt == nil {
		t.Error("CreatedAt must not be nil")
	}
}

func TestContract_GetTodo_NotFound(t *testing.T) {
	adapter, _ := newAdapter(t)

	_, err := adapter.GetTodo(authedCtx(), &todov1.GetTodoRequest{
		Id: uuid.New().String(),
	})

	if err == nil {
		t.Fatal("GetTodo() expected error for missing ID, got nil")
	}
	// The application layer converts domain.ErrTodoNotFound into *platform.DomainError
	// with code ErrorCodeNotFound. Match by code, not by sentinel identity.
	if !isDomainErrorCode(err, platform.ErrorCode(tododef.ErrorCodeNotFound)) {
		t.Errorf("error should carry ErrorCodeNotFound, got: %v", err)
	}
}

func TestContract_StatusMapping(t *testing.T) {
	adapter, _ := newAdapter(t)
	ctx := authedCtx()

	createResp, err := adapter.CreateTodo(ctx, &todov1.CreateTodoRequest{
		Title:    "Status test",
		Priority: todov1.Priority_PRIORITY_MEDIUM,
	})
	if err != nil {
		t.Fatalf("CreateTodo(): %v", err)
	}
	id := createResp.Todo.Id

	cases := []struct {
		name   string
		action func() (todov1.TaskStatus, error)
		want   todov1.TaskStatus
	}{
		{
			name: "pending after create",
			action: func() (todov1.TaskStatus, error) {
				r, err := adapter.GetTodo(ctx, &todov1.GetTodoRequest{Id: id})
				if err != nil {
					return 0, err
				}
				return r.Todo.Status, nil
			},
			want: todov1.TaskStatus_TASK_STATUS_PENDING,
		},
		{
			name: "completed after CompleteTodo",
			action: func() (todov1.TaskStatus, error) {
				r, err := adapter.CompleteTodo(ctx, &todov1.CompleteTodoRequest{Id: id})
				if err != nil {
					return 0, err
				}
				return r.Todo.Status, nil
			},
			want: todov1.TaskStatus_TASK_STATUS_COMPLETED,
		},
		{
			name: "pending after ReopenTodo",
			action: func() (todov1.TaskStatus, error) {
				r, err := adapter.ReopenTodo(ctx, &todov1.ReopenTodoRequest{Id: id})
				if err != nil {
					return 0, err
				}
				return r.Todo.Status, nil
			},
			want: todov1.TaskStatus_TASK_STATUS_PENDING,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := tc.action()
			if err != nil {
				t.Fatalf("action error: %v", err)
			}
			if got != tc.want {
				t.Errorf("Status = %v, want %v", got, tc.want)
			}
		})
	}

	// The adapter exposes no CancelTodo or MarkInProgressTodo method, so
	// TASK_STATUS_CANCELLED and TASK_STATUS_IN_PROGRESS cannot be reached via
	// the public API. We seed these states directly in the in-memory repo and
	// verify that mapStatusToProto inside GetTodo translates them correctly.
	userID := uuid.MustParse("00000000-0000-0000-0000-000000000001")

	t.Run("cancelled via seeded repo state", func(t *testing.T) {
		svc, seedRepo := testhelpers.NewTestService(t)
		seedAdapter := inproc.NewAdapter(svc)

		title, _ := domain.NewTaskTitle("Cancel test")
		todo := domain.NewTodo(title, "", domain.PriorityMedium, nil, userID)
		if err := todo.Cancel(); err != nil {
			t.Fatalf("Cancel(): %v", err)
		}
		if err := seedRepo.Save(authedCtx(), todo); err != nil {
			t.Fatalf("Save(): %v", err)
		}

		resp, err := seedAdapter.GetTodo(authedCtx(), &todov1.GetTodoRequest{Id: string(todo.ID())})
		if err != nil {
			t.Fatalf("GetTodo(): %v", err)
		}
		if resp.Todo.Status != todov1.TaskStatus_TASK_STATUS_CANCELLED {
			t.Errorf("Status = %v, want TASK_STATUS_CANCELLED", resp.Todo.Status)
		}
	})

	t.Run("in_progress via seeded repo state", func(t *testing.T) {
		svc, seedRepo := testhelpers.NewTestService(t)
		seedAdapter := inproc.NewAdapter(svc)

		title, _ := domain.NewTaskTitle("InProgress test")
		todo := domain.NewTodo(title, "", domain.PriorityMedium, nil, userID)
		if err := todo.MarkInProgress(); err != nil {
			t.Fatalf("MarkInProgress(): %v", err)
		}
		if err := seedRepo.Save(authedCtx(), todo); err != nil {
			t.Fatalf("Save(): %v", err)
		}

		resp, err := seedAdapter.GetTodo(authedCtx(), &todov1.GetTodoRequest{Id: string(todo.ID())})
		if err != nil {
			t.Fatalf("GetTodo(): %v", err)
		}
		if resp.Todo.Status != todov1.TaskStatus_TASK_STATUS_IN_PROGRESS {
			t.Errorf("Status = %v, want TASK_STATUS_IN_PROGRESS", resp.Todo.Status)
		}
	})
}

func TestContract_PriorityMapping(t *testing.T) {
	cases := []struct {
		in   todov1.Priority
		want todov1.Priority
	}{
		{todov1.Priority_PRIORITY_LOW, todov1.Priority_PRIORITY_LOW},
		{todov1.Priority_PRIORITY_MEDIUM, todov1.Priority_PRIORITY_MEDIUM},
		{todov1.Priority_PRIORITY_HIGH, todov1.Priority_PRIORITY_HIGH},
		{todov1.Priority_PRIORITY_URGENT, todov1.Priority_PRIORITY_URGENT},
	}

	for _, tc := range cases {
		t.Run(tc.in.String(), func(t *testing.T) {
			adapter, _ := newAdapter(t)
			resp, err := adapter.CreateTodo(authedCtx(), &todov1.CreateTodoRequest{
				Title:    "Priority test",
				Priority: tc.in,
			})
			if err != nil {
				t.Fatalf("CreateTodo(): %v", err)
			}
			if resp.Todo.Priority != tc.want {
				t.Errorf("Priority = %v, want %v", resp.Todo.Priority, tc.want)
			}
		})
	}
}

func TestContract_ErrorPropagation(t *testing.T) {
	adapter, _ := newAdapter(t)
	ctx := authedCtx()

	// Create a todo, delete it, then verify that subsequent operations correctly
	// propagate the typed error through the full chain:
	//   domain.ErrTodoNotFound
	//     → DomainErrorFor → *platform.DomainError{Code: ErrorCodeNotFound}
	//     → eris.Wrap in the adapter
	// This proves errors.As traverses eris-wrapped platform.DomainError values.
	createResp, err := adapter.CreateTodo(ctx, &todov1.CreateTodoRequest{
		Title:    "Error propagation test",
		Priority: todov1.Priority_PRIORITY_MEDIUM,
	})
	if err != nil {
		t.Fatalf("CreateTodo(): %v", err)
	}
	id := createResp.Todo.Id

	_, err = adapter.DeleteTodo(ctx, &todov1.DeleteTodoRequest{Id: id})
	if err != nil {
		t.Fatalf("DeleteTodo(): %v", err)
	}

	// GetTodo after deletion must propagate a typed DomainError with ErrorCodeNotFound.
	_, err = adapter.GetTodo(ctx, &todov1.GetTodoRequest{Id: id})
	if err == nil {
		t.Fatal("GetTodo() after deletion expected error, got nil")
	}
	if !isDomainErrorCode(err, platform.ErrorCode(tododef.ErrorCodeNotFound)) {
		t.Errorf("error should carry ErrorCodeNotFound, got: %v", err)
	}

	// DeleteTodo on a non-existent ID must also propagate ErrorCodeNotFound.
	_, err = adapter.DeleteTodo(ctx, &todov1.DeleteTodoRequest{Id: uuid.New().String()})
	if err == nil {
		t.Fatal("DeleteTodo() on missing ID expected error, got nil")
	}
	if !isDomainErrorCode(err, platform.ErrorCode(tododef.ErrorCodeNotFound)) {
		t.Errorf("delete of missing ID should carry ErrorCodeNotFound, got: %v", err)
	}
}
