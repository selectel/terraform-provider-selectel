package selectel

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"sync/atomic"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type secretsManagerRequest struct {
	method   string
	path     string
	body     string
	response string
	status   int
}

func testSecretsManagerServer(t *testing.T, requests []secretsManagerRequest) *Config {
	t.Helper()
	var index atomic.Int32
	var endpoint string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path == "/v3/auth/tokens" {
			w.Header().Set("X-Subject-Token", "test-auth")
			if r.Method == http.MethodPost {
				w.WriteHeader(http.StatusCreated)
			}
			assert.NoError(t, json.NewEncoder(w).Encode(map[string]any{"token": map[string]any{
				"expires_at": "2099-01-01T00:00:00.000000Z",
				"issued_at":  "2026-01-01T00:00:00.000000Z",
				"methods":    []string{"password"},
				"project":    map[string]string{"id": "project"},
				"catalog": []any{
					map[string]any{"type": "identity", "endpoints": []any{map[string]string{"interface": "public", "region": "ru-1", "url": endpoint + "/v3"}}},
					map[string]any{"type": SecretsManager, "endpoints": []any{map[string]string{"interface": "public", "region": "ru-1", "region_id": "ru-1", "url": endpoint + "/sm"}}},
					map[string]any{"type": CertificateManager, "endpoints": []any{map[string]string{"interface": "public", "region": "ru-1", "region_id": "ru-1", "url": endpoint + "/cm"}}},
				},
			}}))

			return
		}
		i := int(index.Add(1)) - 1
		if i >= len(requests) {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.RequestURI())
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		expected := requests[i]
		assert.Equal(t, expected.method, r.Method)
		assert.Equal(t, expected.path, r.URL.RequestURI())
		if expected.body != "" {
			var body json.RawMessage
			if assert.NoError(t, json.NewDecoder(r.Body).Decode(&body)) {
				assert.JSONEq(t, expected.body, string(body))
			}
		}
		status := expected.status
		if status == 0 {
			status = http.StatusOK
		}
		w.WriteHeader(status)
		_, err := w.Write([]byte(expected.response))
		assert.NoError(t, err)
	}))
	endpoint = server.URL
	t.Cleanup(func() {
		server.Close()
		assert.Equal(t, len(requests), int(index.Load()), "all API requests must be consumed")
	})

	return &Config{AuthURL: endpoint + "/v3", AuthRegion: "ru-1", Username: "user", Password: "test-password", DomainName: "domain", ProjectID: "project"}
}

type secretsManagerVersionResource struct {
	resource *schema.Resource
	config   map[string]any
	id       string
	path     string
	payload  string
}

func testSecretsManagerVersionResource(kind string) secretsManagerVersionResource {
	if kind == "certificate" {
		return secretsManagerVersionResource{
			resource: resourceSecretsManagerCertificateV1(),
			config: map[string]any{
				"name": "test", "project_id": "project",
				"certificates": []any{"old-cert"}, "private_key": "old-key",
			},
			id: "certificate-id", path: "/cm/v1/cert/certificate-id", payload: "private_key",
		}
	}

	return secretsManagerVersionResource{
		resource: resourceSecretsManagerSecretV1(),
		config:   map[string]any{"key": "test", "project_id": "project", "value": "old-value", "description": "old-description"},
		id:       "project/test", path: "/sm/v1/test", payload: "value",
	}
}

func testSecretsManagerVersionResponse(kind string, version int) string {
	if kind == "certificate" {
		return fmt.Sprintf(`{"id":"certificate-id","name":"test","version":%d}`, version)
	}

	return fmt.Sprintf(`{"name":"test","description":"old-description","version":{"version_id":%d,"created_at":"2026-01-01","value":"b2xkLXZhbHVl"}}`, version)
}

func testSecretsManagerUploadResponse(kind string, version int) string {
	if kind == "certificate" {
		return testSecretsManagerVersionResponse(kind, version)
	}

	return fmt.Sprintf(`{"version_id":%d,"created_at":"2026-01-01"}`, version)
}

func testSecretsManagerVersionState(t *testing.T, fixture secretsManagerVersionResource, version int) *terraform.InstanceState {
	t.Helper()
	d := schema.TestResourceDataRaw(t, fixture.resource.Schema, fixture.config)
	d.SetId(fixture.id)
	require.NoError(t, d.Set("version", version))

	return d.State()
}

func testSecretsManagerApply(t *testing.T, fixture secretsManagerVersionResource, state *terraform.InstanceState, meta *Config, wantError bool) *terraform.InstanceState {
	t.Helper()
	diff, err := fixture.resource.Diff(context.Background(), state, terraform.NewResourceConfigRaw(fixture.config), meta)
	require.NoError(t, err)
	require.NotNil(t, diff)
	next, diags := fixture.resource.Apply(context.Background(), state, diff, meta)
	require.Equal(t, wantError, diags.HasError(), "%+v", diags)
	require.NotNil(t, next)

	return next
}

func TestSecretsManagerVersioning(t *testing.T) {
	for _, kind := range []string{"certificate", "secret"} {
		for _, scenario := range []struct {
			name     string
			oldFlag  any
			flag     any
			version  int
			rotate   bool
			activate bool
		}{
			{name: "default activation", version: 2, rotate: true, activate: true},
			{name: "upload without activation", flag: false, version: 1, rotate: true},
			{name: "explicit activation on upload", oldFlag: false, flag: true, version: 2, rotate: true, activate: true},
			{name: "disable activation without upload", flag: false, version: 1},
			{name: "enable activation without upload", oldFlag: false, flag: true, version: 1, activate: true},
		} {
			t.Run(kind+"/"+scenario.name, func(t *testing.T) {
				fixture := testSecretsManagerVersionResource(kind)
				if scenario.oldFlag != nil {
					fixture.config["is_current"] = scenario.oldFlag
				}
				state := testSecretsManagerVersionState(t, fixture, 1)
				if scenario.flag != nil {
					fixture.config["is_current"] = scenario.flag
				}
				var requests []secretsManagerRequest
				if scenario.rotate {
					fixture.config[fixture.payload] = "new-payload"
					body := `{"value":"bmV3LXBheWxvYWQ="}`
					if kind == "certificate" {
						body = `{"pem":{"certificates":["old-cert"],"private_key":"new-payload"}}`
					}
					requests = append(requests, secretsManagerRequest{
						method: http.MethodPost, path: fixture.path + "/versions?activate=" + strconv.FormatBool(scenario.activate),
						body: body, response: testSecretsManagerUploadResponse(kind, 2),
					})
				}
				requests = append(requests, secretsManagerRequest{method: http.MethodGet, path: fixture.path, response: testSecretsManagerVersionResponse(kind, scenario.version)})
				meta := testSecretsManagerServer(t, requests)
				next := testSecretsManagerApply(t, fixture, state, meta, false)
				assert.Equal(t, strconv.Itoa(scenario.version), next.Attributes["version"])
				assert.Equal(t, strconv.FormatBool(scenario.activate), next.Attributes["is_current"])
				assert.Equal(t, fixture.config[fixture.payload], next.Attributes[fixture.payload])
				diff, err := fixture.resource.Diff(context.Background(), next, terraform.NewResourceConfigRaw(fixture.config), meta)
				require.NoError(t, err)
				assert.True(t, diff.Empty(), "repeated configuration must be idempotent: %+v", diff)
			})
		}
	}
}

func TestSecretsManagerVersionDrift(t *testing.T) {
	for _, kind := range []string{"certificate", "secret"} {
		for _, activate := range []bool{true, false} {
			t.Run(fmt.Sprintf("%s/is_current=%t", kind, activate), func(t *testing.T) {
				fixture := testSecretsManagerVersionResource(kind)
				fixture.config["is_current"] = activate
				state := testSecretsManagerVersionState(t, fixture, 1)
				meta := testSecretsManagerServer(t, []secretsManagerRequest{{method: http.MethodGet, path: fixture.path, response: testSecretsManagerVersionResponse(kind, 2)}})
				d := fixture.resource.Data(state)
				require.False(t, fixture.resource.ReadContext(context.Background(), d, meta).HasError())
				assert.Equal(t, 2, d.Get("version"))
				assert.Equal(t, fixture.config[fixture.payload], d.Get(fixture.payload))
				diff, err := fixture.resource.Diff(context.Background(), d.State(), terraform.NewResourceConfigRaw(fixture.config), meta)
				require.NoError(t, err)
				assert.Equal(t, activate, d.Get("is_current"))
				assert.True(t, diff.Empty())
			})
		}
	}
}

func TestSecretsManagerSecretDescriptionOnly(t *testing.T) {
	fixture := testSecretsManagerVersionResource("secret")
	state := testSecretsManagerVersionState(t, fixture, 2)
	fixture.config["description"] = "new-description"
	meta := testSecretsManagerServer(t, []secretsManagerRequest{
		{method: http.MethodPut, path: fixture.path, body: `{"description":"new-description","value":""}`},
		{method: http.MethodGet, path: fixture.path, response: `{"name":"test","description":"new-description","version":{"version_id":2}}`},
	})
	next := testSecretsManagerApply(t, fixture, state, meta, false)
	assert.Equal(t, "new-description", next.Attributes["description"])
	assert.Equal(t, "old-value", next.Attributes["value"])
	assert.Equal(t, "2", next.Attributes["version"])
}

func TestSecretsManagerVersionUploadFailureRetry(t *testing.T) {
	for _, kind := range []string{"certificate", "secret"} {
		t.Run(kind, func(t *testing.T) {
			fixture := testSecretsManagerVersionResource(kind)
			state := testSecretsManagerVersionState(t, fixture, 1)
			fixture.config[fixture.payload] = "new-payload"
			meta := testSecretsManagerServer(t, []secretsManagerRequest{
				{
					method: http.MethodPost, path: fixture.path + "/versions?activate=true", status: http.StatusBadRequest,
					response: `{"status_text":"INCORRECT_REQUEST","error_text":"invalid content"}`,
				},
				{method: http.MethodPost, path: fixture.path + "/versions?activate=true", response: testSecretsManagerUploadResponse(kind, 2)},
				{method: http.MethodGet, path: fixture.path, response: testSecretsManagerVersionResponse(kind, 2)},
			})
			next := testSecretsManagerApply(t, fixture, state, meta, true)
			assert.Equal(t, state.Attributes[fixture.payload], next.Attributes[fixture.payload])
			assert.Equal(t, "1", next.Attributes["version"])
			next = testSecretsManagerApply(t, fixture, next, meta, false)
			assert.Equal(t, "new-payload", next.Attributes[fixture.payload])
			assert.Equal(t, "2", next.Attributes["version"])
		})
	}
}

func TestSecretsManagerCertificateNameOnly(t *testing.T) {
	fixture := testSecretsManagerVersionResource("certificate")
	state := testSecretsManagerVersionState(t, fixture, 2)
	fixture.config["name"] = "renamed"
	meta := testSecretsManagerServer(t, []secretsManagerRequest{
		{method: http.MethodPut, path: fixture.path, body: `{"name":"renamed"}`},
		{method: http.MethodGet, path: fixture.path, response: `{"id":"certificate-id","name":"renamed","version":2}`},
	})
	next := testSecretsManagerApply(t, fixture, state, meta, false)
	assert.Equal(t, "renamed", next.Attributes["name"])
	assert.Equal(t, "2", next.Attributes["version"])
}

func TestSecretsManagerVersionCreate(t *testing.T) {
	for _, kind := range []string{"certificate", "secret"} {
		for _, activate := range []bool{true, false} {
			t.Run(fmt.Sprintf("%s/is_current=%t", kind, activate), func(t *testing.T) {
				fixture := testSecretsManagerVersionResource(kind)
				fixture.config["is_current"] = activate
				path := fixture.path
				body := `{"description":"old-description","value":"b2xkLXZhbHVl"}`
				response := ""
				if kind == "certificate" {
					path = "/cm/v1/certs"
					body = `{"name":"test","pem":{"certificates":["old-cert"],"private_key":"old-key"}}`
					response = testSecretsManagerVersionResponse(kind, 1)
				}
				meta := testSecretsManagerServer(t, []secretsManagerRequest{
					{method: http.MethodPost, path: path, body: body, response: response},
					{method: http.MethodGet, path: fixture.path, response: testSecretsManagerVersionResponse(kind, 1)},
				})
				next := testSecretsManagerApply(t, fixture, nil, meta, false)
				assert.Equal(t, fixture.id, next.ID)
				assert.Equal(t, "1", next.Attributes["version"])
				assert.Equal(t, strconv.FormatBool(activate), next.Attributes["is_current"])
			})
		}
	}
}
