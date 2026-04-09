package connect

import (
	"log/slog"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/rotisserie/eris"

	defauth "github.com/pivaldi/mmw-contracts/go/application/auth"
	authv1 "github.com/pivaldi/mmw-contracts/go/network/auth/v1"
	"github.com/pivaldi/mmw-todo/internal/application/authctx"
)

// NewAuthMiddleware returns an HTTP handler that validates the Bearer token
// by calling authSvc.ValidateToken, then injects the userID into
// the request context before delegating to next.
// Routes starting with any of the excludedPaths will bypass authentication.
func NewAuthMiddleware(
	authSvc defauth.AuthPrivateService,
	logger *slog.Logger,
	excludedPaths []string,
	next http.Handler,
) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		for _, path := range excludedPaths {
			if strings.HasPrefix(r.URL.Path, path) || strings.Contains(r.URL.Path, "/debug/") {
				next.ServeHTTP(w, r)
				return
			}
		}

		token := extractBearer(r)
		if token == "" {
			writeUnauthorized(w)

			return
		}

		resp, err := authSvc.ValidateToken(r.Context(), &authv1.ValidateTokenRequest{Token: token})
		if err != nil {
			logger.Error("token validation failed", "err", eris.ToString(err, true), "path", r.URL.Path)
			writeUnauthorized(w)

			return
		}

		userID, err := uuid.Parse(resp.GetUserId())
		if err != nil {
			logger.Error("invalid user_id in token response", "err", err, "path", r.URL.Path)
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
