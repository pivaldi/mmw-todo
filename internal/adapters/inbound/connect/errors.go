// modules/todo/internal/adapters/inbound/connect/errors.go
package connect

import (
	"errors"

	"connectrpc.com/connect"
	"github.com/piprim/mmw/pkg/platform"
	"github.com/pivaldi/mmw-contracts/go/application/todo"
	commonv1 "github.com/pivaldi/mmw-contracts/go/network/common/v1"
	"github.com/pivaldi/mmw-todo/internal/adapters/inbound/mapper"
)

// domainConnectCodeMap maps proto error codes (from contracts) to Connect status codes.
//
//nolint:gochecknoglobals // package-level lookup table, not mutable state
var domainConnectCodeMap = map[platform.ErrorCode]connect.Code{
	platform.ErrorCode(todo.ErrorCodeInvalidTitle):            connect.CodeInvalidArgument,
	platform.ErrorCode(todo.ErrorCodeInvalidDueDate):          connect.CodeInvalidArgument,
	platform.ErrorCode(todo.ErrorCodeInvalidID):               connect.CodeInvalidArgument,
	platform.ErrorCode(todo.ErrorCodeNotFound):                connect.CodeNotFound,
	platform.ErrorCode(todo.ErrorCodeAlreadyExists):           connect.CodeAlreadyExists,
	platform.ErrorCode(todo.ErrorCodeCannotCompleteCancelled): connect.CodeFailedPrecondition,
	platform.ErrorCode(todo.ErrorCodeCannotModifyCompleted):   connect.CodeFailedPrecondition,
	platform.ErrorCode(todo.ErrorCodeInvalidStatusTransition): connect.CodeFailedPrecondition,
}

// connectErrorFrom converts any error from the application layer into a *connect.Error.
// DomainErrors are mapped to their Connect code and enriched with a typed proto detail
// so TypeScript clients can call err.findDetails(DomainError) to get { code, message }.
// All other errors become CodeInternal.
//
// Note: this function intentionally duplicates the proto-detail attachment logic
// rather than sharing it via a helper in ogl, since ogl must not depend on
// project-specific contracts (commonv1).
func connectErrorFrom(err error) *connect.Error {
	mappedErr := mapper.DomainErrorFor(err)
	domainErr, ok := errors.AsType[*platform.DomainError](mappedErr)
	if !ok {
		return connect.NewError(connect.CodeInternal, err)
	}

	code, ok := domainConnectCodeMap[domainErr.Code]
	if !ok {
		code = connect.CodeInternal
	}

	cerr := connect.NewError(code, errors.New(domainErr.Message))

	detail, detailErr := connect.NewErrorDetail(&commonv1.DomainError{
		Code:    int32(domainErr.Code),
		Message: domainErr.Message,
	})
	if detailErr == nil {
		cerr.AddDetail(detail)
	}

	return cerr
}
