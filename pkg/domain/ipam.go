package domain

import "context"

// Device represents a physical device with a network interface on the stored IP address.
type Device struct {
	ID       string
	IP       string
	SubnetID string
}

// Subnet represents a block of IP addresses allocated to a ResourceOwner.
type Subnet struct {
	ID         string
	Network    string
	MaskBits   int8
	Location   string
	CustomerID string
}

// Customer represents a person and team most directly responsible for a Subnet.
type Customer struct {
	ID            string
	ResourceOwner string
	BusinessUnit  string
}

// IPAMData represents the full collection of IPAM data stored by the IPAM Facade.
type IPAMData struct {
	Devices   []Device
	Subnets   []Subnet
	Customers []Customer
}

// SubnetFetcher is an interface to fetch Subnet information
type SubnetFetcher interface {
	FetchSubnets(ctx context.Context) ([]Subnet, error)
}

// DeviceFetcher is an interface to fetch Device information
type DeviceFetcher interface {
	FetchDevices(ctx context.Context) ([]Device, error)
}

// CustomerFetcher provides an interface for fetching customer data
type CustomerFetcher interface {
	FetchCustomers(ctx context.Context) ([]Customer, error)
}
