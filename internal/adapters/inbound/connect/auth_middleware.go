package connect

import (
	"net/http"
	"strings"

	defauth "github.com/pivaldi/mmw-contracts/definitions/auth"
	"github.com/pivaldi/mmw-todo/internal/application/authctx"
)

// NewAuthMiddleware returns an HTTP handler that validates the Bearer token
// by calling authSvc.ValidateToken, then injects the userID into
// the request context before delegating to next.
func NewAuthMiddleware(authSvc defauth.AuthService, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := extractBearer(r)
		if token == "" {
			writeUnauthorized(w)

			return
		}

		userID, err := authSvc.ValidateToken(r.Context(), token)
		if err != nil {
			writeUnauthorized(w)

			return
		}

		ctx := authctx.WithUserID(r.Context(), userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func extractBearer(r *http.Request) string {
	auth := r.Header.Get("Authorization")
	if !strings.HasPrefix(auth, "Bearer ") {
		return ""
	}

	return strings.TrimPrefix(auth, "Bearer ")
}

func writeUnauthorized(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	_, _ = w.Write([]byte(`{"code":"unauthenticated","message":"missing or invalid token"}`))
}
