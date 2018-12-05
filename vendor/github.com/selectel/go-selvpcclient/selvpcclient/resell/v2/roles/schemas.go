package roles

// Role represents a single Resell role.
type Role struct {
	// ProjectID represents an associated Identity service project.
	ProjectID string `json:"project_id"`

	// UserID represents an associated Identity service user.
	UserID string `json:"user_id"`
}
