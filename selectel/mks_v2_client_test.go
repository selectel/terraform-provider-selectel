package selectel

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/selectel/mks-go/v2/pkg/kubeversion"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFrameworkMKSV2Client(t *testing.T) {
	for _, scenario := range []string{"scoped client", "second pool", "authentication error", "missing service", "unknown pool"} {
		t.Run(scenario, func(t *testing.T) {
			testFrameworkMKSV2Client(t, scenario)
		})
	}
}

// testFrameworkMKSV2Client fakes Keystone with a catalog of two pools
// (ru-1 and ru-2) and takes a kube version request through the returned
// client. The pool the request reaches names itself in the answer, so a
// client scoped for the second pool is told apart from one scoped for the
// first.
func testFrameworkMKSV2Client(t *testing.T, scenario string) {
	t.Helper()
	var endpoint string
	reachedPool := ""
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if strings.HasSuffix(r.URL.Path, "/kubeversions") {
			assert.Equal(t, "scoped-token", r.Header.Get("X-Auth-Token"))
			assert.Contains(t, r.Header.Get("User-Agent"), "provider-test")
			reachedPool = strings.Trim(strings.TrimSuffix(r.URL.Path, "/v2/kubeversions"), "/")
			assert.NoError(t, json.NewEncoder(w).Encode(map[string]any{"kube_versions": []any{}}))

			return
		}
		assert.Equal(t, "/v3/auth/tokens", r.URL.Path)
		if scenario == "authentication error" {
			w.WriteHeader(http.StatusUnauthorized)

			return
		}
		if r.Method == http.MethodPost {
			var body struct {
				Auth struct {
					Scope struct {
						Project struct {
							ID string `json:"id"`
						} `json:"project"`
					} `json:"scope"`
				} `json:"auth"`
			}
			assert.NoError(t, json.NewDecoder(r.Body).Decode(&body))
			assert.Equal(t, "project", body.Auth.Scope.Project.ID)
		}
		service := MKSV2
		if scenario == "missing service" {
			service = "other"
		}
		poolHost := map[string]string{"ru-1": "mks-ru-1", "ru-2": "mks-ru-2"}
		if scenario == "unknown pool" {
			delete(poolHost, "ru-2")
		}
		endpoints := make([]any, 0, len(poolHost))
		for pool := range poolHost {
			// Every pool of the fake shares the server URL: the client only
			// reads this endpoint, the request itself answers on the same
			// server whatever the URL host claims.
			endpoints = append(endpoints, map[string]string{
				"interface": "public", "region": pool, "region_id": pool,
				"url": endpoint + "/" + pool + "/v2",
			})
		}
		w.Header().Set("X-Subject-Token", "scoped-token")
		if r.Method == http.MethodPost {
			w.WriteHeader(http.StatusCreated)
		}
		assert.NoError(t, json.NewEncoder(w).Encode(map[string]any{"token": map[string]any{
			"expires_at": "2099-01-01T00:00:00.000000Z", "issued_at": "2026-01-01T00:00:00.000000Z", "methods": []string{"password"}, "project": map[string]string{"id": "project"},
			"catalog": []any{
				map[string]any{"type": "identity", "name": "keystone", "endpoints": []any{map[string]string{"interface": "public", "region": "ru-1", "region_id": "ru-1", "url": endpoint + "/v3"}}},
				map[string]any{"type": service, "name": "managed-kubernetes", "endpoints": endpoints},
			},
		}}))
	}))
	t.Cleanup(server.Close)
	endpoint = server.URL
	config := &Config{AuthURL: endpoint + "/v3", AuthRegion: "ru-1", Username: "user", Password: "test-password", DomainName: "domain", UserAgent: "provider-test", ProjectID: "project", Region: "ru-1"}
	project, pool := config.MKSV2Defaults()
	assert.Equal(t, "project", project)
	assert.Equal(t, "ru-1", pool)

	pool = "ru-1"
	if scenario == "second pool" {
		pool = "ru-2"
	}
	if scenario == "unknown pool" {
		pool = "ru-9"
	}
	client, err := config.MKSV2Client(context.Background(), project, pool)
	if scenario == "unknown pool" {
		require.Error(t, err)
		assert.Nil(t, client)
		assert.Contains(t, err.Error(), "can't validate pool")
		assert.Contains(t, err.Error(), "region value must contain one of the values")
		assert.Contains(t, err.Error(), "ru-1")

		return
	}
	if scenario != "scoped client" && scenario != "second pool" {
		require.Error(t, err)
		assert.Nil(t, client)

		return
	}
	require.NoError(t, err)
	_, err = kubeversion.List(context.Background(), client)
	require.NoError(t, err)
	assert.Equal(t, pool, reachedPool)
}

func TestMKSV2BaseURL(t *testing.T) {
	for _, tc := range []struct {
		endpoint, want string
	}{
		{"https://mks-ru-1.example.com/v2", "https://mks-ru-1.example.com"},
		{"https://mks-ru-2.example.com/v2/", "https://mks-ru-2.example.com"},
		{"https://mks-ru-1.example.com", "https://mks-ru-1.example.com"},
		{"https://mks-ru-1.example.com:8443/v2", "https://mks-ru-1.example.com:8443"},
	} {
		assert.Equal(t, tc.want, mksV2BaseURL(tc.endpoint), tc.endpoint)
	}
}
