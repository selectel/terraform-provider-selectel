package crossregionsubnets

// CrossRegionSubnetOpts represents options for the cross-region subnets Create request.
type CrossRegionSubnetOpts struct {
	// CrossRegionSubnets represents options for all cross-region subnets.
	CrossRegionSubnets []CrossRegionSubnetOpt `json:"cross_region_subnets"`
}

// CrossRegionSubnetOpt represents options for the single cross-region subnet.
type CrossRegionSubnetOpt struct {
	// Quantity represents how many subnets do we need to create.
	Quantity int `json:"quantity"`

	// Regions represents region options for the cross-region subnet.
	Regions []CrossRegionOpt `json:"regions"`

	// CIDR represents a subnet prefix in CIDR notation for the cross-region subnet.
	CIDR string `json:"cidr"`
}

// CrossRegionOpt represents region options for the cross-region subnet.
type CrossRegionOpt struct {
	// Region represents region that cross-region subnet is associated to.
	Region string `json:"region"`
}

// ListOpts represents options for the List request.
type ListOpts struct {
	Detailed bool `param:"detailed"`
}
