package ipamfetcher

import (
	"context"
	"encoding/json"
	"net/url"
	"path"
	"strconv"

	"github.com/asecurityteam/ipam-facade/pkg/domain"
)

type subnetResponse struct {
	Limit      int      `json:"limit"`
	Offset     int      `json:"offset"`
	TotalCount int      `json:"total_count"`
	Subnets    []subnet `json:"subnets"`
}

type subnet struct {
	CustomFields customFields `json:"custom_fields"`
	CustomerID   int          `json:"customer_id"`
	MaskBits     int          `json:"mask_bits"`
	Network      string       `json:"network"`
	SubnetID     int          `json:"subnet_id"`
}

// NewDevice42SubnetFetcher generates a new Device42SubnetFetcher
func NewDevice42SubnetFetcher(dc Device42Client) *Device42DeviceFetcher {
	resourceEndpoint, _ := url.Parse(dc.Endpoint.String())
	resourceEndpoint.Path = path.Join(resourceEndpoint.Path, "api", "1.0", "subnets")
	return &Device42DeviceFetcher{
		PageFetcher: &Device42PageFetcher{
			Client:   dc.Client,
			Endpoint: resourceEndpoint,
		},
		Limit: dc.Limit,
	}
}

// Device42SubnetFetcher implements the SubnetFetcher interface to retrieve subnet information
// from Device42
type Device42SubnetFetcher struct {
	PageFetcher PageFetcher
	Limit       int
}

// FetchSubnets retrieves subnet information from Device42
func (d *Device42SubnetFetcher) FetchSubnets(ctx context.Context) ([]domain.Subnet, error) {
	iterator := Device42PageIterator{
		Context:     ctx,
		Limit:       d.Limit,
		PageFetcher: d.PageFetcher,
	}

	subnets := make([]domain.Subnet, 0)
	for iterator.Next() {
		var subnetsResponse subnetResponse
		currentPage := iterator.Current()
		if err := json.Unmarshal(currentPage.Body, &subnetsResponse); err != nil {
			return nil, err
		}
		for _, subnet := range subnetsResponse.Subnets {
			subnets = append(subnets, domain.Subnet{
				ID:         strconv.Itoa(subnet.SubnetID),
				Network:    subnet.Network,
				MaskBits:   int8(subnet.MaskBits),
				Location:   subnet.CustomFields.GetValue("Location"),
				CustomerID: strconv.Itoa(subnet.CustomerID),
			})
		}
	}
	return subnets, iterator.Close()
}
