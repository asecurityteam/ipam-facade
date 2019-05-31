package device42

import (
	"context"
	"encoding/json"
	"strconv"

	"github.com/asecurityteam/ipam-facade/pkg/domain"
)

type subnetResponse struct {
	Limit      int      `json:"limit"`
	Offset     int      `json:"offset"`
	TotalCount int      `json:"total_count"`
	Subnets    []subnet `json:"subnets"`
}

type subnet struct {
	Allocated             string       `json:"allocated"`
	AllowBroadcastAddress string       `json:"allow_broadcast_address"`
	AllowNetworkAddress   string       `json:"allow_network_address"`
	Assigned              string       `json:"assigned"`
	CanEdit               string       `json:"can_edit"`
	CategoryID            int          `json:"category_id"`
	CategoryName          string       `json:"category_name"`
	CustomFields          customFields `json:"custom_fields"`
	CustomerID            int          `json:"customer_id"`
	Description           string       `json:"description"`
	Gateway               string       `json:"gateway"`
	MaskBits              int          `json:"mask_bits"`
	Name                  string       `json:"name"`
	Network               string       `json:"network"`
	Notes                 string       `json:"notes"`
	ParentSubnetID        int          `json:"parent_subnet_id"`
	ParentVLANID          int          `json:"parent_vlan_id"`
	ParentVLANName        string       `json:"parent_vlan_name"`
	ParentVLANNumber      string       `json:"parent_vlan_number"`
	RangeBegin            string       `json:"range_begin"`
	RangeEnd              string       `json:"range_end"`
	ServiceLevel          string       `json:"service_level"`
	SubnetID              int          `json:"subnet_id"`
	Tags                  []string     `json:"tags"`
	VRFGroupID            int          `json:"vrf_group_id"`
	VRFGroupName          string       `json:"vrf_group_name"`
}

// Device42SubnetFetcher implements the SubnetFetcher interface to retrieve subnet information
// from Device42
type Device42SubnetFetcher struct {
	Paginator Paginator
}

// FetchSubnets retrieves subnet information from Device42
func (d *Device42SubnetFetcher) FetchSubnets(ctx context.Context) ([]domain.Subnet, error) {
	getSubnetsResponse, err := d.Paginator.BatchPagedRequests(ctx)
	if err != nil {
		return nil, err
	}

	subnets := make([]domain.Subnet, 0)
	for _, response := range getSubnetsResponse {
		var getSubnetPayload subnetResponse
		if err := json.Unmarshal(response, &getSubnetPayload); err != nil {
			return nil, err
		}

		for _, subnet := range getSubnetPayload.Subnets {
			subnets = append(subnets, domain.Subnet{
				ID:         strconv.Itoa(subnet.SubnetID),
				Network:    subnet.Network,
				MaskBits:   int8(subnet.MaskBits),
				Location:   subnet.CustomFields.GetValue("Location"),
				CustomerID: strconv.Itoa(subnet.CustomerID),
			})
		}
	}

	return subnets, nil
}
