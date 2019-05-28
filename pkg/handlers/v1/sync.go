package v1

import (
	"context"

	"github.com/asecurityteam/ipam-facade/pkg/domain"
	"github.com/asecurityteam/ipam-facade/pkg/logs"
)

// SyncIPAMDataHandler uses its IPAMDataFetcher implementation to serve sync requests
// for refreshing the local IPAM data from the CMDB data source.
type SyncIPAMDataHandler struct {
	IPAMDataFetcher domain.IPAMDataFetcher
	IPAMDataStorer  domain.IPAMDataStorer
	LogFn           domain.LogFn
}

// Handle fetches IPAM data from a CMDB and stores the data locally.
func (h *SyncIPAMDataHandler) Handle(ctx context.Context) error {
	logger := h.LogFn(ctx)

	ipamData, err := h.IPAMDataFetcher.FetchIPAMData(ctx)
	if err != nil {
		logger.Error(logs.IPAMDataFetcherFailure{Reason: err.Error()})
		return err
	}

	if err := h.IPAMDataStorer.StoreIPAMData(ctx, ipamData); err != nil {
		logger.Error(logs.IPAMDataStorerFailure{Reason: err.Error()})
		return err
	}

	return nil
}
