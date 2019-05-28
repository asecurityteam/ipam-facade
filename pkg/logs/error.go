package logs

// AssetNotFound is logged when there no asset was found in storage with a given IP address.
type AssetNotFound struct {
	Message string `logevent:"message,default=asset-not-found"`
	Reason  string `logevent:"reason"`
}

// AssetFetchFailure is logged when an unexpected error occurs attempting to fetch an asset from storage.
type AssetFetchFailure struct {
	Message string `logevent:"message,default=asset-fetch-failure"`
	Reason  string `logevent:"reason"`
}
