package users

// UserOpts represents options for the user Create and Update requests.
type UserOpts struct {
	// Name represents the name of a user.
	Name string `json:"name,omitempty"`

	// Password represents a user's password.
	Password string `json:"password,omitempty"`

	// Enabled shows if user is active or it needs to be disabled.
	Enabled *bool `json:"enabled,omitempty"`
}
