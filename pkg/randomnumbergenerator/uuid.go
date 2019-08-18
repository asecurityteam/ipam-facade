package randomnumbergenerator

import (
	"github.com/google/uuid"
)

// UUIDGenerator implements the RandomNumberGenerator interface
type UUIDGenerator struct{}

// NewRandom returns a sufficiently random UUID
func (u *UUIDGenerator) NewRandom() (uuid.UUID, error) {
	return uuid.NewRandom()
}
