package ipamfetcher

import (
	"context"
	"net/http"
	"net/url"
)

// Device42ClientConfig contains configuration settings for a Device42Client
type Device42ClientConfig struct {
	Endpoint string
	Limit    int
}

// Name is used by the settings library to replace the default naming convention.
func (d *Device42ClientConfig) Name() string {
	return "Device42Client"
}

// NewDevice42ClientComponent generates, a new, unititialized Device42ClientComponent
func NewDevice42ClientComponent() *Device42ClientComponent {
	return &Device42ClientComponent{}
}

// Device42ClientComponent satisfies the settings library Component API,
// and may be used by the settings.NewComponent function.
type Device42ClientComponent struct{}

// Settings populates a set of default valid resource types for the Device42ClientSettings
// if none are provided via config.
func (d *Device42ClientComponent) Settings() *Device42ClientConfig {
	return &Device42ClientConfig{}
}

// New constructs a Device42Client from a config.
func (d *Device42ClientComponent) New(_ context.Context, c *Device42ClientConfig) (*Device42Client, error) {
	u, e := url.Parse(c.Endpoint)
	if e != nil {
		return nil, e
	}
	return &Device42Client{
		Endpoint: u,
		Limit:    c.Limit,
	}, nil
}

// Device42ClientConfig contains values to configure a Device42 client
type Device42Client struct {
	Client   *http.Client
	Endpoint *url.URL
	Limit    int
}
