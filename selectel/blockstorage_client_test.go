package selectel

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/selectel/blockstorage-go/pkg/v1/volumetype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFrameworkBlockStorageClient(t *testing.T) {
	for _, scenario := range []string{"scoped client", "authentication error", "missing service", "missing region", "empty endpoint"} {
		t.Run(scenario, func(t *testing.T) {
			testFrameworkBlockStorageClient(t, scenario)
		})
	}
}

func testFrameworkBlockStorageClient(t *testing.T, scenario string) {
	t.Helper()
	var endpoint string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path == "/volumes/types" {
			assert.Equal(t, "scoped-token", r.Header.Get("X-Auth-Token"))
			assert.Contains(t, r.Header.Get("User-Agent"), "provider-test")
			assert.NoError(t, json.NewEncoder(w).Encode(map[string]any{"volume_types": []any{}}))

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
		service := BlockStorage
		region := "ru-1"
		storageEndpoint := endpoint + "/volumes"
		if scenario == "missing service" {
			service = "other"
		}
		if scenario == "missing region" {
			region = "other"
		}
		if scenario == "empty endpoint" {
			storageEndpoint = ""
		}
		w.Header().Set("X-Subject-Token", "scoped-token")
		if r.Method == http.MethodPost {
			w.WriteHeader(http.StatusCreated)
		}
		assert.NoError(t, json.NewEncoder(w).Encode(map[string]any{"token": map[string]any{
			"expires_at": "2099-01-01T00:00:00.000000Z", "issued_at": "2026-01-01T00:00:00.000000Z", "methods": []string{"password"}, "project": map[string]string{"id": "project"},
			"catalog": []any{
				map[string]any{"type": "identity", "name": "keystone", "endpoints": []any{map[string]string{"interface": "public", "region": "ru-1", "region_id": "ru-1", "url": endpoint + "/v3"}}},
				map[string]any{"type": service, "name": "cinder", "endpoints": []any{map[string]string{"interface": "public", "region": "display-label", "region_id": region, "url": storageEndpoint}}},
			},
		}}))
	}))
	t.Cleanup(server.Close)
	endpoint = server.URL
	config := &Config{AuthURL: endpoint + "/v3", AuthRegion: "ru-1", Username: "user", Password: "test-password", DomainName: "domain", UserAgent: "provider-test", ProjectID: "project", Region: "ru-1"}
	project, region := config.BlockStorageDefaults()
	assert.Equal(t, "project", project)
	assert.Equal(t, "ru-1", region)
	client, err := config.BlockStorageClient(context.Background(), project, region)
	if scenario != "scoped client" {
		require.Error(t, err)
		assert.Nil(t, client)
		return
	}
	require.NoError(t, err)
	_, err = volumetype.List(context.Background(), client, volumetype.ListOpts{})
	require.NoError(t, err)
}
