package v1

// JobMetadata contains the aysnc task ID assigned to the
// sync request used to check for completion
type JobMetadata struct {
	JobID string `json:"jobID"`
}
