package roles

// RoleOpts represents options for several Resell roles.
type RoleOpts struct {
	Roles []RoleOpt `json:"roles"`
}

// RoleOpt represents options for a single Resell role.
type RoleOpt struct {
	// ProjectID represents Identity service project.
	ProjectID string `json:"project_id"`

	// UserID represents Identity service user.
	UserID string `json:"user_id"`
}
