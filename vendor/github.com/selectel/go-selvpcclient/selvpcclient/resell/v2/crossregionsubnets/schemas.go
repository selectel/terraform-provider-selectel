package crossregionsubnets

import (
	"github.com/selectel/go-selvpcclient/selvpcclient/resell/v2/servers"
	"github.com/selectel/go-selvpcclient/selvpcclient/resell/v2/subnets"
)

// CrossRegionSubnet represents a single Resell cross-region subnet.
type CrossRegionSubnet struct {
	// ID is a unique id of a cross-region subnet.
	ID int `json:"id"`

	// CIDR is a cross-region subnet prefix in CIDR notation.
	CIDR string `json:"cidr"`

	// VLANID represents id of the associated VLAN in the Networking service.
	VLANID int `json:"vlan_id"`

	// Status shows if cross-region subnet is used.
	Status string `json:"status"`

	// Servers contains info about servers to which cross-region subnet is associated to.
	Servers []servers.Server `json:"servers"`

	// Subnets contains standard subnets in every region that cross-region subnet is attached to.
	Subnets []subnets.Subnet `json:"subnets"`
}
