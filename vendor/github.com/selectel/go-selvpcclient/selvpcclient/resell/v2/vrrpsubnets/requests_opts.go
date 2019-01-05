package vrrpsubnets

import "github.com/selectel/go-selvpcclient/selvpcclient"

// VRRPSubnetOpts represents options for the VRRP subnets Create request.
type VRRPSubnetOpts struct {
	// VRRPSubnets represents options for all VRRP subnets.
	VRRPSubnets []VRRPSubnetOpt `json:"vrrp_subnets"`
}

// VRRPSubnetOpt represents options for the single VRRP subnet.
type VRRPSubnetOpt struct {
	// Quantity represents how many subnets do we need to create.
	Quantity int `json:"quantity"`

	// RegionOpt represents region options for the VRRP subnet.
	Regions VRRPRegionOpt `json:"regions"`

	// Type represents ip version type.
	Type selvpcclient.IPVersion `json:"type"`

	// PrefixLength represents length of the subnet prefix.
	PrefixLength int `json:"prefix_length"`
}

// VRRPRegionOpt represents region options for the single VRRP subnet.
type VRRPRegionOpt struct {
	// Master represent region that contains a master VRRP router.
	Master string `json:"master"`

	// Slave represent region that contains a slave VRRP router.
	Slave string `json:"slave"`
}

// ListOpts represents options for the VRRP subnets List request.
type ListOpts struct {
	Detailed bool `param:"detailed"`
}
