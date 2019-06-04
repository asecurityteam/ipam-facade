package ipamfetcher

import (
	"context"
	"net/url"
)

// Device42ClientSettings contains configuration settings for a Device42Client
type Device42ClientSettings struct {
	Endpoint string
	Limit    int
}

// Name is used by the settings library to replace the default naming convention.
func (d *Device42ClientSettings) Name() string {
	return "Device42Client"
}

// Device42ClientComponent satisfies the settings library Component API,
// and may be used by the settings.NewComponent function.
type Device42ClientComponent struct{}

// Settings populates a set of default valid resource types for the Device42ClientSettings
// if none are provided via config.
func (d *Device42ClientComponent) Settings() *Device42ClientSettings {
	return &Device42ClientSettings{
		Endpoint: "",
		Limit:    0,
	}
}

// New constructs a Device42Client from a config.
func (d *Device42ClientComponent) New(_ context.Context, c *Device42ClientSettings) (*Device42ClientConfig, error) {
	u, e := url.Parse(c.Endpoint)
	if e != nil {
		return nil, e
	}
	return &Device42ClientConfig{
		Endpoint: u,
		Limit:    c.Limit,
	}, nil
}

// Device42ClientConfig contains values to configure a Device42 client
type Device42ClientConfig struct {
	Endpoint *url.URL
	Limit    int
}
