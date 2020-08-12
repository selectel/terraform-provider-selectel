package tokens

// TokenOpts represents options to create a new Identity token.
type TokenOpts struct {
	// ProjectID can be used to create a project-scoped Identity token.
	ProjectID string `json:"project_id,omitempty"`

	// AccountName can be used to create a domain-scoped Identity token.
	AccountName string `json:"account_name,omitempty"`
}
