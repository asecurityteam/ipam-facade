package logs

// InvalidSubnet is logged when a Subnet is returned from Device42 which is invalid or incomplete
type InvalidSubnet struct {
	Message string `logevent:"message,default=invalid-subnet"`
	ID      int    `logevent:"id"`
	Reason  string `logevent:"reason"`
}
