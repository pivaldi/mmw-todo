package connect

import (
	"context"

	"github.com/google/uuid"
	pfmiddleware "github.com/piprim/mmw/pkg/platform/middleware"
	defauth "github.com/pivaldi/mmw-contracts/go/application/auth"
	authv1 "github.com/pivaldi/mmw-contracts/go/network/auth/v1"
)

// NewTokenValidator wraps an AuthPrivateService as a platform TokenValidator.
// The returned function validates a bearer token by calling svc.ValidateToken
// and parses the user UUID from the response.
func NewTokenValidator(svc defauth.AuthPrivateService) pfmiddleware.TokenValidator {
	return func(ctx context.Context, token string) (uuid.UUID, error) {
		resp, err := svc.ValidateToken(ctx, &authv1.ValidateTokenRequest{Token: token})
		if err != nil {
			//nolint:wrapcheck // err is not wrapped.
			return uuid.Nil, err
		}

		return uuid.Parse(resp.GetUserId())
	}
}
