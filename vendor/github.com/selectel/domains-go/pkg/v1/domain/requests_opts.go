package domain

// CreateOpts represents requests options to create a domain.
type CreateOpts struct {
	// Name represents domain name.
	Name string `json:"name,omitempty"`

	// BindZone represents zone files in BIND format.
	BindZone string `json:"bind_zone,omitempty"`
}
