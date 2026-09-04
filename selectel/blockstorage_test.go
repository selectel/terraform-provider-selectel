package selectel

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/selectel/blockstorage-go/pkg/v1/volumetype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testBlockStorageProjectID = "project-id"
	testBlockStorageRegion    = "ru-2"
	testBlockStorageToken     = "project-scoped-token"
	testBlockStorageUserAgent = "terraform-provider-selectel/test"
)

func TestUnitBlockStorageClientUsesProjectRegionTokenAndUserAgent(t *testing.T) {
	var requestToken string
	var requestUserAgent string
	var requestMethod string
	var requestPath string

	server := newBlockStorageTestServer(t, func(response http.ResponseWriter, request *http.Request) {
		requestToken = request.Header.Get("X-Auth-Token")
		requestUserAgent = request.Header.Get("User-Agent")
		requestMethod = request.Method
		requestPath = request.URL.Path
		response.Header().Set("Content-Type", "application/json")
		response.WriteHeader(http.StatusOK)
		_, _ = response.Write([]byte(`{"volume_types":[]}`))
	})
	defer server.Close()

	config := testBlockStorageConfig(server.URL)
	resourceData := testBlockStorageResourceData(t, testBlockStorageProjectID, testBlockStorageRegion)

	client, diagnostics := getBlockStorageClient(resourceData, config)
	require.False(t, diagnostics.HasError(), diagnostics)
	require.NotNil(t, client)

	assert.Equal(t, testBlockStorageProjectID, *server.projectID)

	_, err := volumetype.List(context.Background(), client, volumetype.ListOpts{})
	require.NoError(t, err)
	assert.Equal(t, http.MethodGet, requestMethod)
	assert.Equal(
		t,
		"/volumev3/"+testBlockStorageRegion+"/"+testBlockStorageProjectID+"/types",
		requestPath,
	)
	assert.Equal(t, testBlockStorageToken, requestToken)
	assert.True(t, strings.HasPrefix(requestUserAgent, testBlockStorageUserAgent+" blockstorage-go/"))

	anotherClient, diagnostics := getBlockStorageClient(resourceData, config)
	require.False(t, diagnostics.HasError(), diagnostics)
	assert.NotSame(t, client, anotherClient)
}

func TestUnitBlockStorageClientReportsProjectScopeError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, _ *http.Request) {
		http.Error(response, "authentication failed", http.StatusUnauthorized)
	}))
	defer server.Close()

	client, diagnostics := getBlockStorageClient(
		testBlockStorageResourceData(t, testBlockStorageProjectID, testBlockStorageRegion),
		testBlockStorageConfig(server.URL),
	)

	assert.Nil(t, client)
	require.True(t, diagnostics.HasError())
	assert.Contains(t, diagnostics[0].Summary, "failed to get project-scoped Selectel VPC client")
	assert.Contains(t, diagnostics[0].Summary, testBlockStorageProjectID)
}

func TestUnitBlockStorageClientReportsRegionEndpointError(t *testing.T) {
	server := newBlockStorageTestServer(t, nil)
	defer server.Close()

	client, diagnostics := getBlockStorageClient(
		testBlockStorageResourceData(t, testBlockStorageProjectID, "ru-missing"),
		testBlockStorageConfig(server.URL),
	)

	assert.Nil(t, client)
	require.True(t, diagnostics.HasError())
	assert.Contains(t, diagnostics[0].Summary, "failed to validate Block Storage region")
	assert.Contains(t, diagnostics[0].Summary, "ru-missing")
	assert.Contains(t, diagnostics[0].Summary, "ru-1")
	assert.Contains(t, diagnostics[0].Summary, testBlockStorageRegion)
	assert.Contains(t, diagnostics[0].Summary, BlockStorage)
}

func TestUnitBlockStorageClientReportsEmptyEndpoint(t *testing.T) {
	server := newBlockStorageTestServer(t, nil, "")
	defer server.Close()

	client, diagnostics := getBlockStorageClient(
		testBlockStorageResourceData(t, testBlockStorageProjectID, testBlockStorageRegion),
		testBlockStorageConfig(server.URL),
	)

	assert.Nil(t, client)
	require.True(t, diagnostics.HasError())
	assert.Contains(t, diagnostics[0].Summary, "Block Storage endpoint is empty")
	assert.Contains(t, diagnostics[0].Summary, testBlockStorageRegion)
	assert.Contains(t, diagnostics[0].Summary, testBlockStorageProjectID)
	assert.Contains(t, diagnostics[0].Summary, BlockStorage)
}

func testBlockStorageConfig(authURL string) *Config {
	return &Config{
		AuthURL:    authURL + "/v3",
		AuthRegion: testBlockStorageRegion,
		Username:   "service-user",
		Password:   "password",
		DomainName: "domain",
		UserAgent:  testBlockStorageUserAgent,
	}
}

func testBlockStorageResourceData(t *testing.T, projectID, region string) *schema.ResourceData {
	t.Helper()

	return schema.TestResourceDataRaw(t, map[string]*schema.Schema{
		"project_id": {
			Type:     schema.TypeString,
			Required: true,
		},
		"region": {
			Type:     schema.TypeString,
			Required: true,
		},
	}, map[string]any{
		"project_id": projectID,
		"region":     region,
	})
}

type blockStorageTestServer struct {
	*httptest.Server
	projectID *string
}

func newBlockStorageTestServer(
	t *testing.T,
	blockStorageHandler http.HandlerFunc,
	blockStorageEndpointURL ...string,
) *blockStorageTestServer {
	t.Helper()

	var projectID string
	var server *httptest.Server

	server = httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		switch {
		case request.URL.Path == "/v3/auth/tokens" && request.Method == http.MethodPost:
			var body struct {
				Auth struct {
					Scope struct {
						Project struct {
							ID string `json:"id"`
						} `json:"project"`
					} `json:"scope"`
				} `json:"auth"`
			}
			require.NoError(t, json.NewDecoder(request.Body).Decode(&body))
			projectID = body.Auth.Scope.Project.ID
			writeBlockStorageTokenResponse(
				t, response, server.URL, projectID, http.StatusCreated, blockStorageEndpointURL...,
			)
		case request.URL.Path == "/v3/auth/tokens" && request.Method == http.MethodGet:
			assert.Equal(t, testBlockStorageToken, request.Header.Get("X-Subject-Token"))
			writeBlockStorageTokenResponse(
				t, response, server.URL, projectID, http.StatusOK, blockStorageEndpointURL...,
			)
		case strings.HasPrefix(request.URL.Path, "/volumev3/"):
			if blockStorageHandler != nil {
				blockStorageHandler(response, request)

				return
			}
			response.WriteHeader(http.StatusNoContent)
		default:
			http.NotFound(response, request)
		}
	}))

	return &blockStorageTestServer{Server: server, projectID: &projectID}
}

func writeBlockStorageTokenResponse(
	t *testing.T,
	response http.ResponseWriter,
	serverURL string,
	projectID string,
	status int,
	blockStorageEndpointURL ...string,
) {
	t.Helper()
	testRegionEndpointURL := serverURL + "/volumev3/" + testBlockStorageRegion + "/" + projectID
	if len(blockStorageEndpointURL) != 0 {
		testRegionEndpointURL = blockStorageEndpointURL[0]
	}

	response.Header().Set("Content-Type", "application/json")
	response.Header().Set("X-Subject-Token", testBlockStorageToken)
	response.WriteHeader(status)

	err := json.NewEncoder(response).Encode(map[string]any{
		"token": map[string]any{
			"expires_at": "2099-01-01T00:00:00.000000Z",
			"issued_at":  "2026-01-01T00:00:00.000000Z",
			"methods":    []string{"password"},
			"project": map[string]string{
				"id": projectID,
			},
			"catalog": []map[string]any{
				{
					"type": "identity",
					"name": "keystone",
					"endpoints": []map[string]string{
						{
							"interface": "public",
							"region":    testBlockStorageRegion,
							"region_id": testBlockStorageRegion,
							"url":       serverURL + "/v3",
						},
					},
				},
				{
					"type": BlockStorage,
					"name": "cinder",
					"endpoints": []map[string]string{
						{
							"interface": "public",
							"region":    "ru-1",
							"region_id": "ru-1",
							"url":       serverURL + "/volumev3/ru-1/" + projectID,
						},
						{
							"interface": "public",
							"region":    testBlockStorageRegion,
							"region_id": testBlockStorageRegion,
							"url":       testRegionEndpointURL,
						},
					},
				},
			},
		},
	})
	require.NoError(t, err)
}
