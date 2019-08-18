package domain

import (
	"github.com/google/uuid"
)

// RandomNumberGenerator generates a sufficiently random UUID
type RandomNumberGenerator interface {
	NewRandom() (uuid.UUID, error)
}
