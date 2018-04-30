package resell

import "github.com/selectel/go-selvpcclient/selvpcclient"

const (
	// ServiceType contains the name of the Selectel VPC service for which this
	// package is intended.
	ServiceType = "resell"

	// Endpoint contains the base url for all versions of the Resell client.
	Endpoint = selvpcclient.DefaultEndpoint + "/" + ServiceType

	// UserAgent contains the user agent for all versions of the Resell client.
	UserAgent = selvpcclient.DefaultUserAgent
)
