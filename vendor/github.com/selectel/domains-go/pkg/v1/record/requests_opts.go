package record

// CreateOpts represents requests options to create a domain record.
type CreateOpts struct {
	// Name represents record name.
	Name string `json:"name"`

	// Type represents record type.
	Type Type `json:"type"`

	// TTL represents record time-to-live.
	TTL int `json:"ttl"`

	// Content represents record content.
	// Absent for SRV.
	Content string `json:"content,omitempty"`

	// Emails represents email of domain's admin.
	// For SOA records only.
	Email string `json:"email,omitempty"`

	// Priority represents records preferences.
	// Lower value means more preferred.
	// For MX/SRV records only.
	Priority *int `json:"priority,omitempty"`

	// Weight represents a relative weight for records with the same priority,
	// higher value means higher chance of getting picked.
	// For SRV records only.
	Weight *int `json:"weight,omitempty"`

	// Port represents the TCP or UDP port on which the service is to be found.
	// For SRV records only.
	Port *int `json:"port,omitempty"`

	// Target represents the canonical hostname of the machine providing the service.
	// For SRV records only.
	Target string `json:"target,omitempty"`
}

// UpdateOpts represents requests options to update a domain record.
type UpdateOpts struct {
	// Name represents record name.
	Name string `json:"name"`

	// Type represents record type.
	Type Type `json:"type"`

	// TTL represents record time-to-live.
	TTL int `json:"ttl"`

	// Content represents record content.
	// Absent for SRV.
	Content string `json:"content,omitempty"`

	// Emails represents email of domain's admin.
	// For SOA records only.
	Email string `json:"email,omitempty"`

	// Priority represents records preferences.
	// Lower value means more preferred.
	// For MX/SRV records only.
	Priority *int `json:"priority,omitempty"`

	// Weight represents a relative weight for records with the same priority,
	// higher value means higher chance of getting picked.
	// For SRV records only.
	Weight *int `json:"weight,omitempty"`

	// Port represents the TCP or UDP port on which the service is to be found.
	// For SRV records only.
	Port *int `json:"port,omitempty"`

	// Target represents the canonical hostname of the machine providing the service.
	// For SRV records only.
	Target string `json:"target,omitempty"`
}
