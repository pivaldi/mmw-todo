// Package authctx provides context helpers for propagating the authenticated
// user identity across application layer boundaries.
package authctx

import (
	"context"
	"errors"

	"github.com/google/uuid"
)

type todoContextKey struct{}

// ErrUnauthenticated is returned when the request context contains no userID.
var ErrUnauthenticated = errors.New("unauthenticated")

// WithUserID stores the authenticated userID in the context.
// Called by the auth middleware before passing the request to the handler.
func WithUserID(ctx context.Context, userID uuid.UUID) context.Context {
	return context.WithValue(ctx, todoContextKey{}, userID)
}

// UserIDFromContext extracts the authenticated userID from the context.
// Returns ErrUnauthenticated if absent.
func UserIDFromContext(ctx context.Context) (uuid.UUID, error) {
	v := ctx.Value(todoContextKey{})
	if v == nil {
		return uuid.Nil, ErrUnauthenticated
	}

	id, ok := v.(uuid.UUID)
	if !ok {
		return uuid.Nil, ErrUnauthenticated
	}

	return id, nil
}
