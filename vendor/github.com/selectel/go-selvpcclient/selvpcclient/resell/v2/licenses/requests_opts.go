package licenses

// LicenseOpts represents options for the licenses Create request.
type LicenseOpts struct {
	// Licenses represents options for all licenses.
	Licenses []LicenseOpt `json:"licenses"`
}

// LicenseOpt represents options for the single license.
type LicenseOpt struct {
	// Region represents a region of where the licenses should reside.
	Region string `json:"region"`

	// Quantity represents how many licenses do we need to create in a single request.
	Quantity int `json:"quantity"`

	// Type represents needed type of the license.
	Type string `json:"type"`
}
