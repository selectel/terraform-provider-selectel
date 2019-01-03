package keypairs

// KeypairOpts represents options for the keypair Create request.
type KeypairOpts struct {
	// Name contains a human-readable name for the keypair.
	Name string `json:"name"`

	// PublicKey contains a public part of the keypair.
	PublicKey string `json:"public_key"`

	// Regions contains a list of OpenStack Identity service regions where users
	// can use this keypair.
	Regions []string `json:"regions,omitempty"`

	// UserID contains an ID of an OpenStack Identity service user that owns
	// this keypair.
	UserID string `json:"user_id"`
}
