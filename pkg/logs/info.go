package logs

// InvalidSubnet is logged when a Subnet is returned from Device42 which is invalid or incomplete
type InvalidSubnet struct {
	Message string `logevent:"message,default=invalid-subnet"`
	ID      string `logevent:"id"`
	Reason  string `logevent:"reason"`
}

// DataSyncJobComplete is logged when an asynchronous job to synchronize the local
// IPAM data cache completes.
type DataSyncJobComplete struct {
	Message string `logevent:"message,default=ipam-sync-complete"`
	JobID   string `logevent:"jobid"`
	Reason  string `logevent:"reason"`
}
