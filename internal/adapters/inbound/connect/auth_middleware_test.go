package connect_test

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	authv1 "github.com/pivaldi/mmw-contracts/go/network/auth/v1"
	. "github.com/pivaldi/mmw-todo/internal/adapters/inbound/connect"
	"github.com/pivaldi/mmw-todo/internal/application/authctx"
)

type mockAuthPrivateService struct {
	userID uuid.UUID
	err    error
}

func (m *mockAuthPrivateService) ValidateToken(
	_ context.Context, _ *authv1.ValidateTokenRequest,
) (*authv1.ValidateTokenResponse, error) {
	if m.err != nil {
		return nil, m.err
	}

	return &authv1.ValidateTokenResponse{IsValid: true, UserId: m.userID.String()}, nil
}

func TestAuthMiddleware_ValidToken_CallsNext(t *testing.T) {
	userID := uuid.New()
	svc := &mockAuthPrivateService{userID: userID}
	called := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		id, err := authctx.UserIDFromContext(r.Context())
		require.NoError(t, err)
		assert.Equal(t, userID, id)
		w.WriteHeader(http.StatusOK)
	})

	handler := NewAuthMiddleware(svc, slog.Default(), nil, next)
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	assert.True(t, called)
	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestAuthMiddleware_MissingToken_Returns401(t *testing.T) {
	svc := &mockAuthPrivateService{userID: uuid.New()}
	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := NewAuthMiddleware(svc, slog.Default(), nil, next)
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
}

func TestAuthMiddleware_InvalidToken_Returns401(t *testing.T) {
	svc := &mockAuthPrivateService{err: errors.New("invalid token")}
	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := NewAuthMiddleware(svc, slog.Default(), nil, next)
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.Header.Set("Authorization", "Bearer bad-token")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
}
