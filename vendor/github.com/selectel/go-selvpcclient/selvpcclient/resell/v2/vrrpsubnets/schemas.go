package vrrpsubnets

import (
	"github.com/selectel/go-selvpcclient/selvpcclient/resell/v2/servers"
	"github.com/selectel/go-selvpcclient/selvpcclient/resell/v2/subnets"
)

// VRRPSubnet represents a single Resell VRRP subnet.
type VRRPSubnet struct {
	// ID is a unique id of a VRRP subnet.
	ID int `json:"id"`

	// Status shows if VRRP subnet is used.
	Status string `json:"status"`

	// Servers contains info about servers to which VRRP subnet is associated to.
	Servers []servers.Server `json:"servers"`

	// MasterRegion is a reference to a region that contains a master VRRP router.
	MasterRegion string `json:"master_region"`

	// MasterRegion is a reference to a region that contains a slave VRRP router.
	SlaveRegion string `json:"slave_region"`

	// CIDR is a VRRP subnet prefix in CIDR notation.
	CIDR string `json:"cidr"`

	// Subnets contains standard subnets in every region that VRRP subnet is
	// attached to.
	Subnets []subnets.Subnet `json:"subnets"`

	// ProjectID represents an associated Identity service project.
	ProjectID string `json:"project_id"`
}
