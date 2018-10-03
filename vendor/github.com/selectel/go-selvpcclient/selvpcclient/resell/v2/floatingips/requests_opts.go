package floatingips

// FloatingIPOpts represents options for the floating ips Create request.
type FloatingIPOpts struct {
	// FloatingIPs represents options for all floating ips.
	FloatingIPs []FloatingIPOpt `json:"floatingips"`
}

// FloatingIPOpt represents options for the single floating ip.
type FloatingIPOpt struct {
	// Region represents an Identity service region of where the floating ips should reside.
	Region string `json:"region"`

	// Quantity represents how many floating ips do we need to create in a single request.
	Quantity int `json:"quantity"`
}

// ListOpts represents options for the floating ips List request.
type ListOpts struct {
	Detailed bool `param:"detailed"`
}
