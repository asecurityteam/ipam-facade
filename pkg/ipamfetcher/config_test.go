package ipamfetcher

import (
	"context"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestName(t *testing.T) {
	d := &Device42ClientConfig{}
	assert.Equal(t, "Device42Client", d.Name())
}

func TestDefaultConfig(t *testing.T) {
	component := NewDevice42ClientComponent()
	config := component.Settings()
	client, err := component.New(context.Background(), config)
	zeroURL, _ := url.Parse("")
	assert.Equal(t, client.Endpoint, zeroURL)
	assert.Equal(t, client.Limit, 0)
	assert.NoError(t, err)
}

func TestBadEndpoint(t *testing.T) {
	component := NewDevice42ClientComponent()
	config := &Device42ClientConfig{
		Endpoint: "https://lo\\<calhost:443",
		HTTP:     component.HTTP.Settings(),
	}
	_, err := component.New(context.Background(), config)
	assert.Error(t, err)
}
