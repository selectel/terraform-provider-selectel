package subnets

import "github.com/selectel/go-selvpcclient/selvpcclient"

// SubnetOpts represents options for the subnets Create request.
type SubnetOpts struct {
	// Subnets represents options for all subnets.
	Subnets []SubnetOpt `json:"subnets"`
}

// SubnetOpt represents options for the single subnet.
type SubnetOpt struct {
	// Region represents a region of where the subnet should reside.
	Region string `json:"region"`

	// Quantity represents how many subnets do we need to create.
	Quantity int `json:"quantity"`

	// Type represents ip version type.
	Type selvpcclient.IPVersion `json:"type"`

	// PrefixLength represents length of the subnet prefix.
	PrefixLength int `json:"prefix_length"`
}

// ListOpts represents options for the licenses List request.
type ListOpts struct {
	Detailed bool `param:"detailed"`
}
