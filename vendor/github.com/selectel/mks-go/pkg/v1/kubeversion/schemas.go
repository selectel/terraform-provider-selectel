package kubeversion

// View represents an unmarshalled Kubernetes version body from an API response.
type View struct {
	// Version represents the supported Kubernetes version in format: "X.Y.Z".
	Version string `json:"version"`

	// IsDefault flag indicates if kubernetes version is default.
	IsDefault bool `json:"is_default"`
}
