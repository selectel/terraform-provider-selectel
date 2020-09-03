package nodegroup

// CreateOpts represents options for the nodegroup Create request.
type CreateOpts struct {
	// Count represents nodes count for this nodegroup.
	Count int `json:"count,omitempty"`

	// FlavorID contains reference to a pre-created flavor.
	// It can be omitted in most cases.
	FlavorID string `json:"flavor_id,omitempty"`

	// CPUs represents CPU count for each node.
	// It can be omitted only in cases when flavor_id is set.
	CPUs int `json:"cpus,omitempty"`

	// RAMMB represents RAM count in MB for each node.
	// It can be omitted only in cases when flavor_id is set.
	RAMMB int `json:"ram_mb,omitempty"`

	// VolumeGB represents volume size in GB for each node.
	// It can be omitted only in cases when flavor_id is set and volume is local.
	VolumeGB int `json:"volume_gb,omitempty"`

	// VolumeType represents blockstorage volume type for each node.
	// It can be omitted only in cases when flavor_id is set and volume is local.
	VolumeType string `json:"volume_type,omitempty"`

	// LocalVolume represents if nodes will use local volume.
	LocalVolume bool `json:"local_volume,omitempty"`

	// KeypairName contains name of the SSH key that will be added to all nodes.
	KeypairName string `json:"keypair_name,omitempty"`

	// AffinityPolicy is an optional parameter to tune nodes affinity.
	AffinityPolicy string `json:"affinity_policy,omitempty"`

	// AvailabilityZone should contain the valid zone in the selected region of the created cluster.
	AvailabilityZone string `json:"availability_zone,omitempty"`

	// Labels represents an object containing a set of Kubernetes labels that will be applied
	// for each node in the group. The keys must be user-defined.
	Labels map[string]string `json:"labels"`

	// Taints represents a list of nodegroup taints.
	Taints []Taint `json:"taints"`

	// EnableAutoscale reflects if the nodegroup is allowed to be scaled automatically.
	// Disabled by default.
	EnableAutoscale *bool `json:"enable_autoscale,omitempty"`

	// AutoscaleMinNodes represents minimum possible number of worker nodes in the nodegroup.
	AutoscaleMinNodes *int `json:"autoscale_min_nodes,omitempty"`

	// AutoscaleMaxNodes represents maximum possible number of worker nodes in the nodegroup.
	AutoscaleMaxNodes *int `json:"autoscale_max_nodes,omitempty"`
}

// ResizeOpts represents options for the nodegroup Resize request.
type ResizeOpts struct {
	// Desired represents desired amount of nodes for this nodegroup.
	Desired int `json:"desired"`
}

// UpdateOpts represents options for the nodegroup Update request.
type UpdateOpts struct {
	// Labels represents an object containing a set of Kubernetes labels that will be applied
	// for each node in the group. The keys must be user-defined.
	Labels map[string]string `json:"labels"`

	// EnableAutoscale reflects if the nodegroup is allowed to be scaled automatically.
	EnableAutoscale *bool `json:"enable_autoscale,omitempty"`

	// AutoscaleMinNodes represents minimum possible number of worker nodes in the nodegroup.
	AutoscaleMinNodes *int `json:"autoscale_min_nodes,omitempty"`

	// AutoscaleMaxNodes represents maximum possible number of worker nodes in the nodegroup.
	AutoscaleMaxNodes *int `json:"autoscale_max_nodes,omitempty"`
}
