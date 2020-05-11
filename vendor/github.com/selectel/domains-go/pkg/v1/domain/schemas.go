package domain

// View represents an unmarshalled domain body from API response.
type View struct {
	// ID is the identifier of the domain.
	ID int `json:"id"`

	// CreateDate represents Unix timestamp when domain has been created.
	CreateDate int `json:"create_date"`

	// ChangeDate represents Unix timestamp when domain has been modified.
	ChangeDate int `json:"change_date"`

	// UserID is the Selectel user's identifier.
	UserID int `json:"user_id"`

	// Name represents domain name.
	Name string `json:"name"`

	// Tags is the list of tags applied for the domain.
	Tags []string `json:"tags"`
}
