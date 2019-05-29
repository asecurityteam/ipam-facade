package device42

import "time"

type subnetResponse struct {
	Limit      int      `json:"limit"`
	Offset     int      `json:"offset"`
	TotalCount int      `json:"total_count"`
	Subnets    []subnet `json:"subnets"`
}

type subnet struct {
	Allocated             string        `json:"allocated"`
	AllowBroadcastAddress string        `json:"allow_broadcast_address"`
	AllowNetworkAddress   string        `json:"allow_network_address"`
	Assigned              string        `json:"assigned"`
	CanEdit               string        `json:"can_edit"`
	CategoryID            int           `json:"category_id"`
	CategoryName          string        `json:"category_name"`
	CustomFields          []customField `json:"custom_fields"`
	CustomerID            int           `json:"customer_id"`
	Description           string        `json:"description"`
	Gateway               string        `json:"gateway"`
	MaskBits              int           `json:"mask_bits"`
	Name                  string        `json:"name"`
	Network               string        `json:"network"`
	Notes                 string        `json:"notes"`
	ParentSubnetID        int           `json:"parent_subnet_id"`
	ParentVLANID          int           `json:"parent_vlan_id"`
	ParentVLANName        string        `json:"parent_vlan_name"`
	ParentVLANNumber      string        `json:"parent_vlan_number"`
	RangeBegin            string        `json:"range_begin"`
	RangeEnd              string        `json:"range_end"`
	ServiceLevel          string        `json:"service_level"`
	SubnetID              int           `json:"subnet_id"`
	Tags                  []string      `json:"tags"`
	VRFGroupID            int           `json:"vrf_group_id"`
	VRFGroupName          string        `json:"vrf_group_name"`
}

type ipResponse struct {
	Limit      int  `json:"limit"`
	Offset     int  `json:"offset"`
	TotalCount int  `json:"total_count"`
	IPs        []ip `json:"ips"`
}

type ip struct {
	Available    string        `json:"available"`
	CustomFields []customField `json:"custom_fields"`
	Device       string        `json:"device"`
	DeviceID     int           `json:"device_id"`
	ID           int           `json:"id"`
	IP           string        `json:"ip"`
	Label        string        `json:"label"`
	LastUpdated  time.Time     `json:"last_updated"`
	MacAddress   string        `json:"mac_address"`
	MacID        int           `json:"mac_id"`
	Notes        string        `json:"notes"`
	Subnet       string        `json:"subnet"`
	SubnetID     int           `json:"subnet_id"`
	Type         string        `json:"type"`
}

type customersResponse struct {
	Customers []customer `json:"Customers"`
}

type customer struct {
	Contacts     []contact     `json:"Contacts"`
	CustomFields []customField `json:"Custom Fields"`
	ContactInfo  string        `json:"contact_info"`
	DevicesURL   string        `json:"devices_url"`
	Groups       string        `json:"groups"`
	ID           int           `json:"id"`
	Name         string        `json:"name"`
	Notes        string        `json:"notes"`
	SubnetsURL   string        `json:"subnets_url"`
}

type contact struct {
	Address string `json:"address"`
	Email   string `json:"email"`
	Name    string `json:"name"`
	Phone   string `json:"phone"`
	Type    string `json:"type"`
}

type customField struct {
	Key   string      `json:"key"`
	Value interface{} `json:"value"`
	Notes string      `json:"notes"`
}
