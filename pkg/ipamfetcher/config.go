package ipamfetcher

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"path"

	httpclient "github.com/asecurityteam/component-httpclient"
)

// Device42ClientConfig contains configuration settings for a Device42Client
type Device42ClientConfig struct {
	Endpoint string
	Limit    int
	HTTP     *httpclient.Config
}

// Name is used by the settings library to replace the default naming convention.
func (d *Device42ClientConfig) Name() string {
	return "Device42Client"
}

// NewDevice42ClientComponent generates a new, unititialized Device42ClientComponent
func NewDevice42ClientComponent() *Device42ClientComponent {
	return &Device42ClientComponent{
		HTTP: httpclient.NewComponent(),
	}
}

// Device42ClientComponent satisfies the settings library Component API,
// and may be used by the settings.NewComponent function.
type Device42ClientComponent struct {
	HTTP *httpclient.Component
}

// Settings populates a set of default valid resource types for the Device42ClientSettings
// if none are provided via config.
func (d *Device42ClientComponent) Settings() *Device42ClientConfig {
	return &Device42ClientConfig{
		HTTP: d.HTTP.Settings(),
	}
}

// New constructs a Device42Client from a config.
func (d *Device42ClientComponent) New(ctx context.Context, c *Device42ClientConfig) (*Device42Client, error) {
	rt, e := d.HTTP.New(ctx, c.HTTP)
	if e != nil {
		return nil, e
	}
	u, e := url.Parse(c.Endpoint)
	if e != nil {
		return nil, e
	}
	return &Device42Client{
		Endpoint: u,
		Limit:    c.Limit,
		Client: &http.Client{
			Transport: rt,
		},
	}, nil
}

// Device42Client contains values to configure a Device42 client
type Device42Client struct {
	Client   *http.Client
	Endpoint *url.URL
	Limit    int
}

// CheckDependencies makes a call to Endpoint, no path is involved. This is the only
// Because Device42Client is the only shared dependency shared amongst Device42DeviceFetcher,
// Device42SubnetFetcher, and Device42CustomerFetcher, we don't need to test each of
// those components for dependencies
func (d *Device42Client) CheckDependencies(ctx context.Context) error {
	u, _ := url.Parse(d.Endpoint.String())
	u.Path = path.Join(u.Path, "api", "1.0", "vrfgroup")
	req, _ := http.NewRequest(http.MethodGet, u.String(), http.NoBody)
	res, err := d.Client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		return fmt.Errorf("Device42Client unexpectedly returned non-200 response code: %d attempting to GET: %s", res.StatusCode, u.String())
	}
	return nil
}
