package ipamfetcher

import (
	"context"

	"github.com/asecurityteam/ipam-facade/pkg/domain"
)

// Client implements the IPAMDataFetcher interface
type Client struct {
	CustomerFetcher domain.CustomerFetcher
	SubnetFetcher   domain.SubnetFetcher
	DeviceFetcher   domain.DeviceFetcher
}

// FetchIPAMData implements the IPAMDataFetcher interface to retrieve data from Device42
func (c *Client) FetchIPAMData(ctx context.Context) (domain.IPAMData, error) {
	customers, err := c.CustomerFetcher.FetchCustomers(ctx)
	if err != nil {
		return domain.IPAMData{}, err
	}
	subnets, err := c.SubnetFetcher.FetchSubnets(ctx)
	if err != nil {
		return domain.IPAMData{}, err
	}
	devices, err := c.DeviceFetcher.FetchDevices(ctx)
	if err != nil {
		return domain.IPAMData{}, err
	}

	return domain.IPAMData{
		Customers: customers,
		Subnets:   subnets,
		Devices:   devices,
	}, nil
}
