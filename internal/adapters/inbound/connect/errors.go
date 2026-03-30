// modules/todo/internal/adapters/inbound/connect/errors.go
package connect

import (
	"errors"

	"connectrpc.com/connect"
	"github.com/piprim/mmw/pkg/platform"
	deftodo "github.com/pivaldi/mmw-contracts/definitions/todo"
	commonv1 "github.com/pivaldi/mmw-contracts/gen/go/common/v1"
)

// domainConnectCodeMap maps todo domain error codes to Connect status codes.
//
//nolint:gochecknoglobals // package-level lookup table, not mutable state
var domainConnectCodeMap = map[platform.ErrorCode]connect.Code{
	platform.ErrorCode(deftodo.ErrorCodeInvalidTitle):            connect.CodeInvalidArgument,
	platform.ErrorCode(deftodo.ErrorCodeInvalidDueDate):          connect.CodeInvalidArgument,
	platform.ErrorCode(deftodo.ErrorCodeInvalidID):               connect.CodeInvalidArgument,
	platform.ErrorCode(deftodo.ErrorCodeNotFound):                connect.CodeNotFound,
	platform.ErrorCode(deftodo.ErrorCodeAlreadyExists):           connect.CodeAlreadyExists,
	platform.ErrorCode(deftodo.ErrorCodeCannotCompleteCancelled): connect.CodeFailedPrecondition,
	platform.ErrorCode(deftodo.ErrorCodeCannotModifyCompleted):   connect.CodeFailedPrecondition,
	platform.ErrorCode(deftodo.ErrorCodeInvalidStatusTransition): connect.CodeFailedPrecondition,
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
	domainErr, ok := errors.AsType[*platform.DomainError](err)
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
