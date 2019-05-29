package domain

import "context"

// IPAMDataStorer stores IPAM data fetched from a CMDB data source into local storage.
type IPAMDataStorer interface {
	StoreIPAMData(context.Context, IPAMData) error
}
