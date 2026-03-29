// modules/todo/internal/application/errors.go
package application

import (
	"errors"

	"github.com/ovya/ogl/platform"
	deftodo "github.com/pivaldi/mmw-contracts/definitions/todo"
	"github.com/pivaldi/mmw-todo/internal/domain"
)

// DomainErrorFor translates a domain sentinel error into a *platform.DomainError
// so the inbound adapter can map it to a typed Connect error detail.
// Non-domain errors (infra, unexpected) are returned unchanged.
//
// Exported as DomainErrorFor for testability; service.go uses the same symbol directly.
func DomainErrorFor(err error) error {
	switch {
	case errors.Is(err, domain.ErrInvalidTitle):
		return &platform.DomainError{
			Code:    platform.ErrorCode(deftodo.ErrorCodeInvalidTitle),
			Message: err.Error(),
		}

	case errors.Is(err, domain.ErrInvalidDueDate):
		return &platform.DomainError{
			Code:    platform.ErrorCode(deftodo.ErrorCodeInvalidDueDate),
			Message: err.Error(),
		}

	case errors.Is(err, domain.ErrInvalidID):
		return &platform.DomainError{
			Code:    platform.ErrorCode(deftodo.ErrorCodeInvalidID),
			Message: err.Error(),
		}

	case errors.Is(err, domain.ErrTodoNotFound):
		return &platform.DomainError{
			Code:    platform.ErrorCode(deftodo.ErrorCodeNotFound),
			Message: err.Error(),
		}

	case errors.Is(err, domain.ErrTodoAlreadyExists):
		return &platform.DomainError{
			Code:    platform.ErrorCode(deftodo.ErrorCodeAlreadyExists),
			Message: err.Error(),
		}

	case errors.Is(err, domain.ErrCannotCompleteCancelled):
		return &platform.DomainError{
			Code:    platform.ErrorCode(deftodo.ErrorCodeCannotCompleteCancelled),
			Message: err.Error(),
		}

	case errors.Is(err, domain.ErrCannotModifyCompleted):
		return &platform.DomainError{
			Code:    platform.ErrorCode(deftodo.ErrorCodeCannotModifyCompleted),
			Message: err.Error(),
		}

	case errors.Is(err, domain.ErrInvalidStatusTransition):
		return &platform.DomainError{
			Code:    platform.ErrorCode(deftodo.ErrorCodeInvalidStatusTransition),
			Message: err.Error(),
		}
	}

	return err
}
