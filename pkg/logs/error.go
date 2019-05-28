package logs

// InvalidInput is logged when a request is made to fetch an asset by an invalid IP address.
type InvalidInput struct {
	Message string `logevent:"message,default=invalid-input"`
	Reason  string `logevent:"reason"`
}

// AssetNotFound is logged when there no asset was found in storage with a given IP address.
type AssetNotFound struct {
	Message string `logevent:"message,default=asset-not-found"`
	Reason  string `logevent:"reason"`
}

// AssetFetcherFailure is logged when an unexpected error occurs attempting to fetch an asset from storage.
type AssetFetcherFailure struct {
	Message string `logevent:"message,default=asset-fetch-failure"`
	Reason  string `logevent:"reason"`
}

// IPAMDataFetcherFailure is logged when fetching IPAM data from the CMDB data source fails.
type IPAMDataFetcherFailure struct {
	Message string `logevent:"message,default=ipam-data-fetcher-failure"`
	Reason  string `logevent:"reason"`
}

// IPAMDataStorerFailure is logged when storing IPAM data fails.
type IPAMDataStorerFailure struct {
	Message string `logevent:"message,default=ipam-data-storer-failure"`
	Reason  string `logevent:"reason"`
}
