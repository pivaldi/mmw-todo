package application

import (
	"context"

	"github.com/google/uuid"

	"github.com/pivaldi/mmw/todo/internal/application/authctx"
)

// ErrUnauthenticated is returned when the request context contains no userID.
var ErrUnauthenticated = authctx.ErrUnauthenticated

// WithUserID stores the authenticated userID in the context.
// Called by the auth middleware before passing the request to the handler.
func WithUserID(ctx context.Context, userID uuid.UUID) context.Context {
	return authctx.WithUserID(ctx, userID)
}

// UserIDFromContext extracts the authenticated userID from the context.
// Returns ErrUnauthenticated if absent.
func UserIDFromContext(ctx context.Context) (uuid.UUID, error) {
	return authctx.UserIDFromContext(ctx)
}
