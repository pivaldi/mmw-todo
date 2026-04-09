// modules/todo/internal/adapters/inbound/connect/errors_test.go
package connect

import (
	"errors"
	"testing"

	"connectrpc.com/connect"
	"github.com/piprim/mmw/pkg/platform"
	tododef "github.com/pivaldi/mmw-contracts/go/application/todo"
)

func TestConnectErrorFrom_DomainError_MapsToCorrectCode(t *testing.T) {
	cases := []struct {
		name     string
		code     platform.ErrorCode
		wantCode connect.Code
	}{
		{"InvalidTitle", platform.ErrorCode(tododef.ErrorCodeInvalidTitle), connect.CodeInvalidArgument},
		{"InvalidDueDate", platform.ErrorCode(tododef.ErrorCodeInvalidDueDate), connect.CodeInvalidArgument},
		{"InvalidID", platform.ErrorCode(tododef.ErrorCodeInvalidID), connect.CodeInvalidArgument},
		{"NotFound", platform.ErrorCode(tododef.ErrorCodeNotFound), connect.CodeNotFound},
		{"AlreadyExists", platform.ErrorCode(tododef.ErrorCodeAlreadyExists), connect.CodeAlreadyExists},
		{"CannotCompleteCancelled", platform.ErrorCode(tododef.ErrorCodeCannotCompleteCancelled), connect.CodeFailedPrecondition},
		{"CannotModifyCompleted", platform.ErrorCode(tododef.ErrorCodeCannotModifyCompleted), connect.CodeFailedPrecondition},
		{"InvalidStatusTransition", platform.ErrorCode(tododef.ErrorCodeInvalidStatusTransition), connect.CodeFailedPrecondition},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			domainErr := &platform.DomainError{Code: tc.code, Message: "test error"}

			result := connectErrorFrom(domainErr)

			var connectErr *connect.Error
			if !errors.As(result, &connectErr) {
				t.Fatalf("expected *connect.Error, got %T", result)
			}

			if connectErr.Code() != tc.wantCode {
				t.Errorf("Code() = %v, want %v", connectErr.Code(), tc.wantCode)
			}
		})
	}
}

func TestConnectErrorFrom_DomainError_HasDetail(t *testing.T) {
	domainErr := &platform.DomainError{
		Code:    platform.ErrorCode(tododef.ErrorCodeNotFound),
		Message: "todo not found",
	}

	result := connectErrorFrom(domainErr)

	var connectErr *connect.Error
	if !errors.As(result, &connectErr) {
		t.Fatalf("expected *connect.Error, got %T", result)
	}

	if len(connectErr.Details()) == 0 {
		t.Error("expected at least one error detail")
	}
}

func TestConnectErrorFrom_UnknownDomainCode_IsInternal(t *testing.T) {
	domainErr := &platform.DomainError{Code: 9999, Message: "unknown"}

	result := connectErrorFrom(domainErr)

	var connectErr *connect.Error
	if !errors.As(result, &connectErr) {
		t.Fatalf("expected *connect.Error, got %T", result)
	}

	if connectErr.Code() != connect.CodeInternal {
		t.Errorf("Code() = %v, want %v", connectErr.Code(), connect.CodeInternal)
	}
}

func TestConnectErrorFrom_NonDomainError_IsInternal(t *testing.T) {
	infraErr := errors.New("db connection refused")

	result := connectErrorFrom(infraErr)

	var connectErr *connect.Error
	if !errors.As(result, &connectErr) {
		t.Fatalf("expected *connect.Error, got %T", result)
	}

	if connectErr.Code() != connect.CodeInternal {
		t.Errorf("Code() = %v, want %v", connectErr.Code(), connect.CodeInternal)
	}
}
