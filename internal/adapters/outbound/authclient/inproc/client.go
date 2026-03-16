package inproc

import (
	"context"

	defauth "github.com/pivaldi/mmw/contracts/definitions/auth"
	"github.com/pivaldi/mmw/todo/internal/application/ports"
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

	// Map the contract DTO to the Todo domain's expected shape
	return &ports.User{
		UUID:  dto.UUID,
		Login: dto.Login,
	}, nil
}
