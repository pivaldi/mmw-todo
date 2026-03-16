package ports

import (
	"errors"

	"github.com/google/uuid"
)

var ErrAuthServiceDown = errors.New("authentication service down")

type User struct {
	UUID  uuid.UUID
	Login string
}
