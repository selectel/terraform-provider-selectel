package subnets

import "github.com/selectel/go-selvpcclient/selvpcclient/resell/v2/servers"

// Subnet represents a single Resell subnet.
type Subnet struct {
	// ID is a unique id of a subnet.
	ID int `json:"id"`

	// Status shows if subnet is used.
	Status string `json:"status"`

	// Servers contains info about servers to which subnet is associated to.
	Servers []servers.Server `json:"servers"`

	// Region represents a region of where the subnet resides.
	Region string `json:"region"`

	// CIDR is a subnet prefix in CIDR notation.
	CIDR string `json:"cidr"`

	// NetworkID represents id of the associated network in the Networking service.
	NetworkID string `json:"network_id"`

	// SubnetID represents id of the associated subnet in the Networking service.
	SubnetID string `json:"subnet_id"`

	// ProjectID represents an associated Identity service project.
	ProjectID string `json:"project_id"`
}
