package inproc

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	defauth "github.com/pivaldi/mmw-contracts/definitions/auth"
	"github.com/pivaldi/mmw-todo/internal/application/ports"
)

// Client adapts the public contract to the Todo application's specific port
type Client struct {
	contract defauth.AuthService
}

func NewClient(contract defauth.AuthService) *Client {
	return &Client{contract: contract}
}

func (c *Client) GetUser(ctx context.Context, id string) (*ports.User, error) {
	// Call the contract (which routes to Inproc or Network seamlessly)
	dto, err := c.contract.GetUser(ctx, id)
	if err != nil {
		return nil, ports.ErrAuthServiceDown
	}
	if dto == nil {
		return nil, fmt.Errorf("auth service returned nil user for id %q", id)
	}

	// Parse the proto string ID into the uuid.UUID expected by the domain port
	userUUID, err := uuid.Parse(dto.Id)
	if err != nil {
		return nil, fmt.Errorf("auth service returned invalid user id %q: %w", dto.Id, err)
	}

	// Map the contract DTO to the Todo domain's expected shape
	return &ports.User{
		UUID:  userUUID,
		Login: dto.Login,
	}, nil
}
