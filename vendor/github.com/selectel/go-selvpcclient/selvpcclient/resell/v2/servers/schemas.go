package servers

import (
	"time"
)

// Server represents a simple server entity that is used by some other packages of
// the go-selvpcclient.
type Server struct {
	// ID is a unique id of the server.
	ID string `json:"id"`

	// Name is a human-readable name of the server.
	Name string `json:"name"`

	// Status represents a current status of the server.
	Status string `json:"status"`

	// Updated contains the ISO-8601 timestamp of when the state of the server
	// last changed.
	Updated time.Time `json:"updated"`
}
