package randomnumbergenerator

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUUIDGeneration(t *testing.T) {
	generator := UUIDGenerator{}
	_, err := generator.NewRandom()
	assert.NoError(t, err)
}
