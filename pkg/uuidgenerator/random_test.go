package uuidgenerator

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUUIDGeneration(t *testing.T) {
	generator := RandomUUIDGenerator{}
	_, err := generator.NewUUIDString()
	assert.NoError(t, err)
}
