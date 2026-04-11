package connect_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	authv1 "github.com/pivaldi/mmw-contracts/go/network/auth/v1"
	. "github.com/pivaldi/mmw-todo/internal/adapters/inbound/connect"
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

func TestNewTokenValidator_ValidService_ReturnsUUID(t *testing.T) {
	userID := uuid.New()
	svc := &mockAuthPrivateService{userID: userID}

	validate := NewTokenValidator(svc)
	got, err := validate(context.Background(), "any-token")

	require.NoError(t, err)
	assert.Equal(t, userID, got)
}

func TestNewTokenValidator_ServiceError_PropagatesError(t *testing.T) {
	svc := &mockAuthPrivateService{err: errors.New("invalid token")}

	validate := NewTokenValidator(svc)
	_, err := validate(context.Background(), "bad-token")

	assert.Error(t, err)
}
