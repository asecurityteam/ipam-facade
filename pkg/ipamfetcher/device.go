package ipamfetcher

import (
	"context"
	"encoding/json"
	"net/url"
	"path"
	"strconv"

	"github.com/asecurityteam/ipam-facade/pkg/domain"
)

type ipResponse struct {
	Limit      int  `json:"limit"`
	Offset     int  `json:"offset"`
	TotalCount int  `json:"total_count"`
	IPs        []ip `json:"ips"`
}

type ip struct {
	DeviceID int    `json:"device_id"`
	IP       string `json:"ip"`
	SubnetID int    `json:"subnet_id"`
}

// NewDevice42DeviceFetcher generates a new Device42DeviceFetcher
func NewDevice42DeviceFetcher(dc *Device42Client) *Device42DeviceFetcher {
	resourceEndpoint, _ := url.Parse(dc.Endpoint.String())
	resourceEndpoint.Path = path.Join(resourceEndpoint.Path, "api", "1.0", "ips")
	return &Device42DeviceFetcher{
		PageFetcher: &Device42PageFetcher{
			Client:   dc.Client,
			Endpoint: resourceEndpoint,
		},
		Limit: dc.Limit,
	}
}

// Device42DeviceFetcher implements the DeviceFetcher interface to retrieve device information
// from Device42
type Device42DeviceFetcher struct {
	PageFetcher PageFetcher
	Limit       int
}

// FetchDevices retrieve device information from Device42
func (d *Device42DeviceFetcher) FetchDevices(ctx context.Context) ([]domain.Device, error) {
	iterator := &Device42PageIterator{
		Context:     ctx,
		Limit:       d.Limit,
		PageFetcher: d.PageFetcher,
	}

	assets := make([]domain.Device, 0)
	for iterator.Next() {
		var devicesResponse ipResponse
		currentPage := iterator.Current()
		if err := json.Unmarshal(currentPage.Body, &devicesResponse); err != nil {
			return nil, err
		}
		for _, asset := range devicesResponse.IPs {
			assets = append(assets, domain.Device{
				IP:       asset.IP,
				ID:       strconv.Itoa(asset.DeviceID),
				SubnetID: strconv.Itoa(asset.SubnetID),
			})
		}
	}
	return assets, iterator.Close()
}
