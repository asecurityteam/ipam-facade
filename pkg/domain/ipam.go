package domain

// Device represents a physical device with a network interface on the stored IP address.
type Device struct {
	ID       string
	IP       string
	SubnetID string
}

// Subnet represents a block of IP addresses allocated to a ResourceOwner.
type Subnet struct {
	ID         string
	Network    string
	MaskBits   int8
	Location   string
	CustomerID string
}

// Customer represents a person and team most directly responsible for a Subnet.
type Customer struct {
	ID            string
	ResourceOwner string
	BusinessUnit  string
}

// IPAMData represents the full collection of IPAM data stored by the IPAM Facade.
type IPAMData struct {
	Devices   []Device
	Subnets   []Subnet
	Customers []Customer
}
