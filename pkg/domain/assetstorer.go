package domain

import "context"

// PhysicalAssetStorer stores IPAM data fetched from a CMDB data source into local storage.
type PhysicalAssetStorer interface {
	StorePhysicalAssets(context.Context, IPAMData) error
}
