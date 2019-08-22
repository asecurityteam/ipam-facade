package uuidgenerator

import (
	"github.com/google/uuid"
)

// RandomUUIDGenerator implements the RandomNumberGenerator interface
type RandomUUIDGenerator struct{}

// NewUUIDString generates a UUIDv4 as a string
func (u *RandomUUIDGenerator) NewUUIDString() (string, error) {
	id, err := uuid.NewRandom()
	if err != nil {
		return "", err
	}
	return id.String(), nil
}
