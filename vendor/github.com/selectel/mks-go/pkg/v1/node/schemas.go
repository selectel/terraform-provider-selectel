package node

import "time"

// View represents an unmarshalled node body from an API response.
type View struct {
	// ID is the identifier of the node.
	ID string `json:"id"`

	// CreatedAt is the timestamp in UTC timezone of when the node has been created.
	CreatedAt *time.Time `json:"created_at"`

	// UpdatedAt is the timestamp in UTC timezone of when the node has been updated.
	UpdatedAt *time.Time `json:"updated_at"`

	// Hostname represents a hostname of the node.
	Hostname string `json:"hostname"`

	// IP represents IP address of the node.
	IP string `json:"ip"`

	// NodegroupID contains nodegroup identifier.
	NodegroupID string `json:"nodegroup_id"`

	// OSServerID contains OpenStack server identifier.
	OSServerID string `json:"os_server_id"`
}
