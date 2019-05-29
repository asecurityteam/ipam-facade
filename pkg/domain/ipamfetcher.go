package domain

import "context"

// IPAMDataFetcher fetches IPAM data from a CMDB like Device42 for local storage.
type IPAMDataFetcher interface {
	FetchIPAMData(context.Context) (IPAMData, error)
}
