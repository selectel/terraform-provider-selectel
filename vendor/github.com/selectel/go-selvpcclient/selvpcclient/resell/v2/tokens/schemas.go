package tokens

// Token represent a single Identity service token.
type Token struct {
	// ID contains and Identity token id and can be used to requests calls against
	// the different OpenStack services.
	ID string `json:"id"`
}
