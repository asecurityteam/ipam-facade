package ipamfetcher

import (
	"context"
	"encoding/json"
	"strconv"
	"time"

	"github.com/asecurityteam/ipam-facade/pkg/domain"
)

type ipResponse struct {
	Limit      int  `json:"limit"`
	Offset     int  `json:"offset"`
	TotalCount int  `json:"total_count"`
	IPs        []ip `json:"ips"`
}

type ip struct {
	Available    string       `json:"available"`
	CustomFields customFields `json:"custom_fields"`
	Device       string       `json:"device"`
	DeviceID     int          `json:"device_id"`
	ID           int          `json:"id"`
	IP           string       `json:"ip"`
	Label        string       `json:"label"`
	LastUpdated  time.Time    `json:"last_updated"`
	MacAddress   string       `json:"mac_address"`
	MacID        int          `json:"mac_id"`
	Notes        string       `json:"notes"`
	Subnet       string       `json:"subnet"`
	SubnetID     int          `json:"subnet_id"`
	Type         string       `json:"type"`
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
