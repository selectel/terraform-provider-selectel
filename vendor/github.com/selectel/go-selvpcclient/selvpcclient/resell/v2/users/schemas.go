package users

// User represents a single user of the Identity service project.
type User struct {
	// ID is a unique id of a user.
	ID string `json:"id"`

	// Name represents the human-readable name of a user.
	Name string `json:"name"`

	// Enabled shows if user is active or it was disabled.
	Enabled bool `json:"enabled"`
}
