package servers

import (
	"net/http"
)

// newFakeClient creates a new ServiceClient with the given endpoint and transport.
func newFakeClient(endpoint string, transport http.RoundTripper) *ServiceClient {
	return &ServiceClient{
		HTTPClient: &http.Client{Transport: transport},
		Endpoint:   endpoint,
	}
}

const (
	invalidJSONBody = `{
			"result": [
				invalid
			]
		}`

	httpErrorBody    = "Not Found"
	httpErrorMessage = "got the 404 status code from the server: Not Found"
)
