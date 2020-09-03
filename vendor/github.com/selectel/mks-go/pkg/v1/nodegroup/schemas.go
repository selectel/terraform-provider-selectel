package nodegroup

import (
	"time"

	"github.com/selectel/mks-go/pkg/v1/node"
)

// View represents an unmarshalled nodegroup body from an API response.
type View struct {
	// ID is the identifier of the nodegroup.
	ID string `json:"id"`

	// CreatedAt is the timestamp in UTC timezone of when the nodegroup has been created.
	CreatedAt *time.Time `json:"created_at"`

	// UpdatedAt is the timestamp in UTC timezone of when the nodegroup has been updated.
	UpdatedAt *time.Time `json:"updated_at"`

	// ClusterID contains cluster identifier.
	ClusterID string `json:"cluster_id"`

	// FlavorID contains OpenStack flavor identifier for all nodes in the nodegroup.
	FlavorID string `json:"flavor_id"`

	// VolumeGB represents initial volume size in GB for each node.
	VolumeGB int `json:"volume_gb"`

	// VolumeType represents initial blockstorage volume type for each node.
	VolumeType string `json:"volume_type"`

	// LocalVolume represents if nodes use local volume.
	LocalVolume bool `json:"local_volume"`

	// AvailabilityZone represents OpenStack availability zone for all nodes in the nodegroup.
	AvailabilityZone string `json:"availability_zone"`

	// Nodes contains list of all nodes in the nodegroup.
	Nodes []*node.View `json:"nodes"`

	// Labels represents an object containing a set of Kubernetes labels that will be applied
	// for each node in the group. The keys must be user-defined.
	Labels map[string]string `json:"labels"`

	// Taints represents a list of nodegroup taints.
	Taints []Taint `json:"taints"`
}

// TaintEffect represents an effect of the node's taint.
type TaintEffect string

const (
	NoScheduleEffect       TaintEffect = "NoSchedule"
	NoExecuteEffect        TaintEffect = "NoExecute"
	PreferNoScheduleEffect TaintEffect = "PreferNoSchedule"
)

// Taint represents k8s node's taint.
type Taint struct {
	// Key is the key of the taint.
	Key string `json:"key"`

	// Value is the value of the taint.
	Value string `json:"value"`

	// Effect is the effect of the taint.
	Effect TaintEffect `json:"effect"`
}
