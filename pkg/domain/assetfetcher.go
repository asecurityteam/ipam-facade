package domain

import (
	"context"
	"fmt"
)

// PhysicalAsset represents a non-cloud device with a network interface.
type PhysicalAsset struct {
	IP            string
	ResourceOwner string
	BusinessUnit  string
	Network       string
	Location      string
	DeviceID      int64
	SubnetID      int64
	CustomerID    int64
}

// AssetSubnet represents a network subnet to which assets are allocated
type AssetSubnet struct {
	Network       string
	ResourceOwner string
	BusinessUnit  string
	Location      string
}

// AssetIP represents IP address info for an asset
type AssetIP struct {
	IP            string
	Network       string
	ResourceOwner string
	BusinessUnit  string
	Location      string
}

// PhysicalAssetFetcher retrieves a PhysicalAsset by its IP Address.
type PhysicalAssetFetcher interface {
	FetchPhysicalAsset(ctx context.Context, ipAddress string) (PhysicalAsset, error)
}

// SubnetsFetcher fetches a pages response for network subnets
type SubnetsFetcher interface {
	FetchSubnets(ctx context.Context, limit, offset int) ([]AssetSubnet, error)
}

// IPsFetcher fetches a pages response for ip addresses
type IPsFetcher interface {
	FetchIPs(ctx context.Context, limit, offset int) ([]AssetIP, error)
}

// InvalidInput occurs when a physical asset is requested by an invalid IP address.
type InvalidInput struct {
	IP string
}

func (e InvalidInput) Error() string {
	return fmt.Sprintf("%v is not a valid IP address", e.IP)
}

// AssetNotFound is used to indicate that no physical asset with the given IP address exists in storage.
type AssetNotFound struct {
	Inner error
	IP    string
}

func (e AssetNotFound) Error() string {
	return fmt.Sprintf("no asset with IP address %s found in storage: %v", e.IP, e.Inner)
}

// AssetFetchError is used to indicate an unexpected error occurred while querying storage
// for an asset with the given IP address.
type AssetFetchError struct {
	Inner error
	IP    string
}

func (e AssetFetchError) Error() string {
	return fmt.Sprintf("unexpected error occurred querying storage for asset with IP address %s: %v", e.IP, e.Inner)
}
