package httptest

import (
	"io"
	"net/http"
	"strings"
)

// RoundTripFunc lets us use a function as an http.RoundTripper.
type RoundTripFunc func(req *http.Request) (*http.Response, error)

func (f RoundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

// NewFakeResponse creates a fake *http.Response with the provided status and body.
func NewFakeResponse(status int, body string) *http.Response {
	return &http.Response{
		StatusCode: status,
		Body:       io.NopCloser(strings.NewReader(body)),
	}
}

// NewFakeTransport returns a fake transport with the given response and error.
func NewFakeTransport(resp *http.Response, err error) RoundTripFunc {
	return RoundTripFunc(func(_ *http.Request) (*http.Response, error) {
		return resp, err
	})
}
