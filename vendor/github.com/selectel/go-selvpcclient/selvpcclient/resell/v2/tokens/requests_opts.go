package tokens

// TokenOpts represents options to create a new Identity token.
type TokenOpts struct {
	// ProjectID can be used to create a project-scoped Identity token.
	ProjectID string `json:"project_id,omitempty"`

	// DomainName can be used to create a domain-scoped Identity token.
	DomainName string `json:"domain_name,omitempty"`
}
