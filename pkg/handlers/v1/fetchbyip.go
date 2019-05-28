package v1

import "context"

// IPAddressQuery contains an IP address on which to search for physical assets.
type IPAddressQuery struct {
	IPAddress string `json:"ipAddress"`
}

// PhysicalAssetDetails provides the response structure for PhysicalAsset records returned from storage.
type PhysicalAssetDetails struct {
	IP            string `json:"ip"`
	ResourceOwner string `json:"resourceOwner"`
	BusinessUnit  string `json:"businessUnit"`
	Tags          tags   `json:"tags"`
}

// tags is the key-value pair structure that provides less important information than the
// root keys of the PhysicalAssetDetails response.
type tags struct {
	SubnetID string `json:"subnetID"`
	Network  string `json:"network"`
	DeviceID string `json:"deviceID"`
	Location string `json:"location"`
}

// FetchByIPAddressHandler uses its PhysicalAssetFetcher implementation to serve fetch requests for
// physical assets by IP address.
type FetchByIPAddressHandler struct {
	PhysicalAssetFetcher domain.PhysicalAssetFetcher
	LogFn                domain.LogFn
}

// Handle processes an incoming IPAddressQuery and returns a PhysicalAssetDetails response or an error.
func (h *FetchByIPAddressHandler) Handle(ctx context.Context, query IPAddressQuery) (PhysicalAssetDetails, error) {
	logger := h.LogFn(ctx)

	asset, err := h.PhysicalAsetFetcher(ctx, query.IPAddress)
	switch err.(type) {
	case nil:
		response := h.physicalAssetToResponse(asset)
		return response, nil
	case domain.AssetNotFound:
		logger.Error(logs.AssetNotFound{Reason: err.Error()})
		return PhysicalAsset{}, err
	default:
		logger.Error(logs.AssetFetchError{Reason: err.Error()})
		return PhysicalAsset{}, err
	}
}

// physicalAssetToResponse converts a PhysicalAsset structure into a PhysicalAssetDetails structure for the
// handler's HTTP response body.
func (h *FetchByIpAddressHandler) physicalAssetToResponse(asset domain.PhysicalAsset) PhysicalAssetDetails {
	return PhysicalAssetDetails{
		IP: asset.IP
		ResourceOwner: asset.ResourceOwner
		BusinessUnit: asset.BusinessUnit
		Tags: tags{
			SubnetID: asset.SubnetID
			Network: asset.Network
			DeviceID: asset.DeviceID
			Location: asset.Location
		}
	}
}
