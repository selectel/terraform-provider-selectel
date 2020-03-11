package licenses

import "github.com/selectel/go-selvpcclient/selvpcclient/resell/v2/servers"

// License represents a single Resell License.
type License struct {
	// ID is a unique id of the license.
	ID int `json:"id"`

	// ProjectID represents an associated Identity service project.
	ProjectID string `json:"project_id"`

	// Region represents a region of where the license resides.
	Region string `json:"region"`

	// Servers contains info about servers to which license is associated to.
	Servers []servers.Server `json:"servers"`

	// Status represents a current status of the license.
	Status string `json:"status"`

	// Type represent a license type.
	Type string `json:"type"`

	// NetworkID represents id of the associated network in the Networking service.
	NetworkID string `json:"network_id"`

	// SubnetID represents id of the associated subnet in the Networking service.
	SubnetID string `json:"subnet_id"`

	// PortID represents id of the associated ports in the Networking service.
	PortID string `json:"port_id"`
}
