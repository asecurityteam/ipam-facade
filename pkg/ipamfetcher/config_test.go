package ipamfetcher

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestName(t *testing.T) {
	d := &Device42ClientSettings{}
	assert.Equal(t, "Device42Client", d.Name())
}

func TestDefaultConfig(t *testing.T) {
	c := &Device42ClientComponent{}
	settings := c.Settings()

	assert.Equal(t, "", settings.Endpoint)
	assert.Equal(t, 0, settings.Limit)
}
