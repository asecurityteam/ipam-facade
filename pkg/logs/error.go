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
	JobID   string `logevent:"jobId"`
}

// AssetStorerFailure is logged when storing IPAM data fails.
type AssetStorerFailure struct {
	Message string `logevent:"message,default=ipam-data-storer-failure"`
	Reason  string `logevent:"reason"`
	JobID   string `logevent:"jobId"`
}

// SyncError is emitted if the IPAM data sync fails
type SyncError struct {
	Message string `logevent:"message,default=sync-error"`
	Reason  string `logevent:"reason"`
}

// ProducerError is emitted when producer fails to enqueue
type ProducerError struct {
	Message string `logevent:"message,default=producer-error"`
	Reason  string `logevent:"reason"`
}
