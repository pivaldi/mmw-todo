// modules/todo/internal/adapters/inbound/mapper/errors.go
package mapper

import (
	"errors"

	"github.com/piprim/mmw/pkg/platform"
	tododef "github.com/pivaldi/mmw-contracts/definitions/todo"
	"github.com/pivaldi/mmw-todo/internal/domain"
)

// DomainErrorFor translates a domain sentinel error into a *platform.DomainError
// using the error codes from contracts (definitions/todo). This is the application
// layer's responsibility: binding domain errors to the shared wire protocol so that
// callers — including other modules communicating in-process — receive a typed error
// that carries no domain-specific knowledge.
// Non-domain errors (infra, unexpected) are returned unchanged.
func DomainErrorFor(err error) error {
	switch {
	case errors.Is(err, domain.ErrInvalidTitle):
		return &platform.DomainError{
			Code:    platform.ErrorCode(tododef.ErrorCodeInvalidTitle),
			Message: err.Error(),
		}

	case errors.Is(err, domain.ErrInvalidDueDate):
		return &platform.DomainError{
			Code:    platform.ErrorCode(tododef.ErrorCodeInvalidDueDate),
			Message: err.Error(),
		}

	case errors.Is(err, domain.ErrInvalidID):
		return &platform.DomainError{
			Code:    platform.ErrorCode(tododef.ErrorCodeInvalidID),
			Message: err.Error(),
		}

	case errors.Is(err, domain.ErrTodoNotFound):
		return &platform.DomainError{
			Code:    platform.ErrorCode(tododef.ErrorCodeNotFound),
			Message: err.Error(),
		}

	case errors.Is(err, domain.ErrTodoAlreadyExists):
		return &platform.DomainError{
			Code:    platform.ErrorCode(tododef.ErrorCodeAlreadyExists),
			Message: err.Error(),
		}

	case errors.Is(err, domain.ErrCannotCompleteCancelled):
		return &platform.DomainError{
			Code:    platform.ErrorCode(tododef.ErrorCodeCannotCompleteCancelled),
			Message: err.Error(),
		}

	case errors.Is(err, domain.ErrCannotModifyCompleted):
		return &platform.DomainError{
			Code:    platform.ErrorCode(tododef.ErrorCodeCannotModifyCompleted),
			Message: err.Error(),
		}

	case errors.Is(err, domain.ErrInvalidStatusTransition):
		return &platform.DomainError{
			Code:    platform.ErrorCode(tododef.ErrorCodeInvalidStatusTransition),
			Message: err.Error(),
		}
	}

	return err
}
