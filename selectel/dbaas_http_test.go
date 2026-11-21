package selectel

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewDBaaSHTTPClient(t *testing.T) {
	var gotCacheControl, gotToken string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotCacheControl = r.Header.Get("Cache-Control")
		gotToken = r.Header.Get("X-Auth-Token")
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	client := newDBaaSHTTPClient()
	assert.Equal(t, dbaasHTTPTimeout, client.Timeout)

	req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, srv.URL+"/v1/datastores/id", nil)
	require.NoError(t, err)
	req.Header.Set("X-Auth-Token", "token")

	resp, err := client.Do(req)
	require.NoError(t, err)
	resp.Body.Close()

	assert.Equal(t, "no-cache", gotCacheControl)
	assert.Equal(t, "token", gotToken)
	assert.Empty(t, req.Header.Get("Cache-Control"), "original request must not be modified")
}
