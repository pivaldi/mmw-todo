// modules/todo/internal/adapters/inbound/mapper/errors_test.go
package mapper_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/piprim/mmw/pkg/platform"
	tododef "github.com/pivaldi/mmw-contracts/definitions/todo"
	"github.com/pivaldi/mmw-todo/internal/adapters/inbound/mapper"
	"github.com/pivaldi/mmw-todo/internal/domain"
)

func TestDomainErrorFor_KnownSentinels(t *testing.T) {
	cases := []struct {
		name     string
		input    error
		wantCode platform.ErrorCode
	}{
		{"ErrInvalidTitle", domain.ErrInvalidTitle, platform.ErrorCode(tododef.ErrorCodeInvalidTitle)},
		{"ErrInvalidDueDate", domain.ErrInvalidDueDate, platform.ErrorCode(tododef.ErrorCodeInvalidDueDate)},
		{"ErrInvalidID", domain.ErrInvalidID, platform.ErrorCode(tododef.ErrorCodeInvalidID)},
		{"ErrTodoNotFound", domain.ErrTodoNotFound, platform.ErrorCode(tododef.ErrorCodeNotFound)},
		{"ErrTodoAlreadyExists", domain.ErrTodoAlreadyExists, platform.ErrorCode(tododef.ErrorCodeAlreadyExists)},
		{"ErrCannotCompleteCancelled", domain.ErrCannotCompleteCancelled, platform.ErrorCode(tododef.ErrorCodeCannotCompleteCancelled)},
		{"ErrCannotModifyCompleted", domain.ErrCannotModifyCompleted, platform.ErrorCode(tododef.ErrorCodeCannotModifyCompleted)},
		{"ErrInvalidStatusTransition", domain.ErrInvalidStatusTransition, platform.ErrorCode(tododef.ErrorCodeInvalidStatusTransition)},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			result := mapper.DomainErrorFor(tc.input)

			domErr, ok := errors.AsType[*platform.DomainError](result)
			if !ok {
				t.Fatalf("expected *platform.DomainError, got %T", result)
			}

			if domErr.Code != tc.wantCode {
				t.Errorf("Code = %v, want %v", domErr.Code, tc.wantCode)
			}

			if domErr.Message == "" {
				t.Error("Message must not be empty")
			}
		})
	}
}

func TestDomainErrorFor_WrappedSentinel_IsUnwrapped(t *testing.T) {
	wrapped := fmt.Errorf("context: %w", domain.ErrInvalidTitle)

	result := mapper.DomainErrorFor(wrapped)

	_, ok := errors.AsType[*platform.DomainError](result)
	if !ok {
		t.Fatal("expected *platform.DomainError for wrapped domain error")
	}
}

func TestDomainErrorFor_NonDomainError_PassesThrough(t *testing.T) {
	infra := errors.New("db connection refused")

	result := mapper.DomainErrorFor(infra)

	if result != infra {
		t.Errorf("expected original error to pass through, got %v", result)
	}
}
