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

func testSecretsManagerVersionState(t *testing.T, fixture secretsManagerVersionResource, version, pin int) *terraform.InstanceState {
	t.Helper()
	d := schema.TestResourceDataRaw(t, fixture.resource.Schema, fixture.config)
	d.SetId(fixture.id)
	require.NoError(t, d.Set("version", version))
	require.NoError(t, d.Set("current_version", pin))

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
			name       string
			oldPin     int
			pin        int
			oldVersion int
			version    int
			rotate     bool
			activate   bool
		}{
			{name: "automatic rotation", oldVersion: 1, version: 2, rotate: true},
			{name: "pinned rotation", oldPin: 1, pin: 1, oldVersion: 1, version: 1, rotate: true},
			{name: "rollback", pin: 1, oldVersion: 2, version: 1, activate: true},
			{name: "pin already active", pin: 2, oldVersion: 2, version: 2},
			{name: "unpin", oldPin: 1, oldVersion: 1, version: 1},
			{name: "rotate and rollback", pin: 1, oldVersion: 2, version: 1, rotate: true, activate: true},
		} {
			t.Run(kind+"/"+scenario.name, func(t *testing.T) {
				fixture := testSecretsManagerVersionResource(kind)
				state := testSecretsManagerVersionState(t, fixture, scenario.oldVersion, scenario.oldPin)
				fixture.config["current_version"] = scenario.pin
				var requests []secretsManagerRequest
				if scenario.rotate {
					fixture.config[fixture.payload] = "new-payload"
					body := `{"value":"bmV3LXBheWxvYWQ="}`
					if kind == "certificate" {
						body = `{"pem":{"certificates":["old-cert"],"private_key":"new-payload"}}`
					}
					requests = append(requests, secretsManagerRequest{
						method: http.MethodPost, path: fixture.path + "/versions?activate=" + strconv.FormatBool(scenario.pin == 0),
						body: body, response: testSecretsManagerUploadResponse(kind, scenario.oldVersion+1),
					})
				}
				if scenario.pin > 0 {
					requests = append(requests, secretsManagerRequest{method: http.MethodGet, path: fixture.path, response: testSecretsManagerVersionResponse(kind, scenario.oldVersion)})
				}
				if scenario.activate {
					requests = append(requests, secretsManagerRequest{method: http.MethodPost, path: fixture.path + "/versions/1/activate", response: testSecretsManagerUploadResponse(kind, 1)})
				}
				requests = append(requests, secretsManagerRequest{method: http.MethodGet, path: fixture.path, response: testSecretsManagerVersionResponse(kind, scenario.version)})
				meta := testSecretsManagerServer(t, requests)
				next := testSecretsManagerApply(t, fixture, state, meta, false)
				assert.Equal(t, strconv.Itoa(scenario.version), next.Attributes["version"])
				assert.Equal(t, strconv.Itoa(scenario.pin), next.Attributes["current_version"])
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
		for _, pin := range []int{0, 1} {
			t.Run(fmt.Sprintf("%s/pin=%d", kind, pin), func(t *testing.T) {
				fixture := testSecretsManagerVersionResource(kind)
				fixture.config["current_version"] = pin
				state := testSecretsManagerVersionState(t, fixture, 1, pin)
				meta := testSecretsManagerServer(t, []secretsManagerRequest{{method: http.MethodGet, path: fixture.path, response: testSecretsManagerVersionResponse(kind, 2)}})
				d := fixture.resource.Data(state)
				require.False(t, fixture.resource.ReadContext(context.Background(), d, meta).HasError())
				assert.Equal(t, 2, d.Get("version"))
				assert.Equal(t, fixture.config[fixture.payload], d.Get(fixture.payload))
				diff, err := fixture.resource.Diff(context.Background(), d.State(), terraform.NewResourceConfigRaw(fixture.config), meta)
				require.NoError(t, err)
				if pin == 0 {
					assert.Equal(t, 0, d.Get("current_version"))
					assert.True(t, diff.Empty())
					return
				}
				assert.Equal(t, 2, d.Get("current_version"))
				require.NotNil(t, diff)
				assert.Equal(t, "1", diff.Attributes["current_version"].New)
			})
		}
	}
}

func TestSecretsManagerVersionValidation(t *testing.T) {
	for _, kind := range []string{"certificate", "secret"} {
		t.Run(kind, func(t *testing.T) {
			fixture := testSecretsManagerVersionResource(kind)
			_, errs := fixture.resource.Schema["current_version"].ValidateFunc(-1, "current_version")
			require.NotEmpty(t, errs)
			for _, pin := range []int{0, 1, 2} {
				_, errs = fixture.resource.Schema["current_version"].ValidateFunc(pin, "current_version")
				require.Empty(t, errs)
			}
			fixture.config["current_version"] = 2
			d := schema.TestResourceDataRaw(t, fixture.resource.Schema, fixture.config)
			// A nil client configuration proves validation precedes authentication and remote creation.
			diags := fixture.resource.CreateContext(context.Background(), d, nil)
			require.True(t, diags.HasError())
			assert.Contains(t, diags[0].Summary, "current_version must be 0 or 1")
			assert.Empty(t, d.Id())
		})
	}
}

func TestSecretsManagerVersionActivationFailureRetry(t *testing.T) {
	for _, kind := range []string{"certificate", "secret"} {
		t.Run(kind, func(t *testing.T) {
			fixture := testSecretsManagerVersionResource(kind)
			state := testSecretsManagerVersionState(t, fixture, 2, 0)
			fixture.config[fixture.payload] = "new-payload"
			fixture.config["current_version"] = 1
			meta := testSecretsManagerServer(t, []secretsManagerRequest{
				{method: http.MethodPost, path: fixture.path + "/versions?activate=false", response: testSecretsManagerUploadResponse(kind, 3)},
				{method: http.MethodGet, path: fixture.path, response: testSecretsManagerVersionResponse(kind, 2)},
				{method: http.MethodPost, path: fixture.path + "/versions/1/activate", status: http.StatusBadRequest, response: `{"error":"activation failed"}`},
				{method: http.MethodGet, path: fixture.path, response: testSecretsManagerVersionResponse(kind, 2)},
				{method: http.MethodPost, path: fixture.path + "/versions/1/activate", response: testSecretsManagerUploadResponse(kind, 1)},
				{method: http.MethodGet, path: fixture.path, response: testSecretsManagerVersionResponse(kind, 1)},
			})
			next := testSecretsManagerApply(t, fixture, state, meta, true)
			assert.Equal(t, "new-payload", next.Attributes[fixture.payload])
			assert.Equal(t, "2", next.Attributes["version"])
			assert.Equal(t, "2", next.Attributes["current_version"])
			next = testSecretsManagerApply(t, fixture, next, meta, false)
			assert.Equal(t, "1", next.Attributes["current_version"])
		})
	}
}

func TestSecretsManagerSecretDescriptionOnly(t *testing.T) {
	fixture := testSecretsManagerVersionResource("secret")
	state := testSecretsManagerVersionState(t, fixture, 2, 0)
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

func TestSecretsManagerVersionActivationConflict(t *testing.T) {
	for _, kind := range []string{"certificate", "secret"} {
		t.Run(kind, func(t *testing.T) {
			fixture := testSecretsManagerVersionResource(kind)
			state := testSecretsManagerVersionState(t, fixture, 2, 0)
			fixture.config["current_version"] = 1
			meta := testSecretsManagerServer(t, []secretsManagerRequest{
				{method: http.MethodGet, path: fixture.path, response: testSecretsManagerVersionResponse(kind, 2)},
				{
					method: http.MethodPost, path: fixture.path + "/versions/1/activate", status: http.StatusConflict,
					response: `{"status_text":"CONFLICT","error_text":"version already active"}`,
				},
				{method: http.MethodGet, path: fixture.path, response: testSecretsManagerVersionResponse(kind, 1)},
			})
			next := testSecretsManagerApply(t, fixture, state, meta, false)
			assert.Equal(t, "1", next.Attributes["version"])
			assert.Equal(t, "1", next.Attributes["current_version"])
			assert.Equal(t, fixture.config[fixture.payload], next.Attributes[fixture.payload])
		})
	}
}

func TestSecretsManagerVersionActivationConflictNotConverged(t *testing.T) {
	for _, kind := range []string{"certificate", "secret"} {
		for _, active := range []int{2, 3} {
			t.Run(fmt.Sprintf("%s/active=%d", kind, active), func(t *testing.T) {
				fixture := testSecretsManagerVersionResource(kind)
				state := testSecretsManagerVersionState(t, fixture, 2, 0)
				fixture.config["current_version"] = 1
				meta := testSecretsManagerServer(t, []secretsManagerRequest{
					{method: http.MethodGet, path: fixture.path, response: testSecretsManagerVersionResponse(kind, 2)},
					{
						method: http.MethodPost, path: fixture.path + "/versions/1/activate", status: http.StatusConflict,
						response: `{"status_text":"CONFLICT","error_text":"activation conflict"}`,
					},
					{method: http.MethodGet, path: fixture.path, response: testSecretsManagerVersionResponse(kind, active)},
				})
				config := terraform.NewResourceConfigRaw(fixture.config)
				diff, err := fixture.resource.Diff(context.Background(), state, config, meta)
				require.NoError(t, err)
				next, diags := fixture.resource.Apply(context.Background(), state, diff, meta)
				require.True(t, diags.HasError())
				assert.Contains(t, diags[0].Summary, "CONFLICT: activation conflict")
				require.NotNil(t, next)
				assert.Equal(t, strconv.Itoa(active), next.Attributes["version"])
				assert.Equal(t, strconv.Itoa(active), next.Attributes["current_version"])
				diff, err = fixture.resource.Diff(context.Background(), next, config, meta)
				require.NoError(t, err)
				require.NotNil(t, diff)
				assert.Equal(t, "1", diff.Attributes["current_version"].New)
			})
		}
	}
}

func TestSecretsManagerVersionUploadFailureRetry(t *testing.T) {
	for _, kind := range []string{"certificate", "secret"} {
		t.Run(kind, func(t *testing.T) {
			fixture := testSecretsManagerVersionResource(kind)
			state := testSecretsManagerVersionState(t, fixture, 1, 0)
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
	state := testSecretsManagerVersionState(t, fixture, 2, 0)
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
		for _, pin := range []int{0, 1} {
			t.Run(fmt.Sprintf("%s/pin=%d", kind, pin), func(t *testing.T) {
				fixture := testSecretsManagerVersionResource(kind)
				fixture.config["current_version"] = pin
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
				assert.Equal(t, strconv.Itoa(pin), next.Attributes["current_version"])
			})
		}
	}
}
