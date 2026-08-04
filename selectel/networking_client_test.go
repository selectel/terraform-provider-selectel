package selectel

import (
	"errors"
	"reflect"
	"testing"

	"github.com/selectel/vpc-go/pkg/v2/port"
	"github.com/selectel/vpc-go/pkg/v2/router"
	"github.com/selectel/vpc-go/pkg/v2/securitygroup"
	"github.com/selectel/vpc-go/pkg/v2/subnet"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testNetworkingToken    = "test-project-scoped-value" //nolint:gosec // test fixture, not a credential
	testNetworkingRegion   = "ru-3"
	testNetworkingEndpoint = "https://api.selectel.ru/network"
)

// testTokenCatalog is a stub of the go-selvpcclient token/catalog pair.
type testTokenCatalog struct {
	token             string
	endpoint          string
	endpointErr       error
	gotServiceType    string
	gotEndpointRegion string
}

func (c *testTokenCatalog) GetXAuthToken() string {
	return c.token
}

func (c *testTokenCatalog) GetEndpointURL(serviceType, regionName string) (string, error) {
	c.gotServiceType = serviceType
	c.gotEndpointRegion = regionName

	if c.endpointErr != nil {
		return "", c.endpointErr
	}

	return c.endpoint, nil
}

func TestNetworkingV2Client(t *testing.T) {
	t.Run("builds an sdk client from the token and the resolved endpoint", func(t *testing.T) {
		catalog := &testTokenCatalog{token: testNetworkingToken, endpoint: testNetworkingEndpoint}

		client, err := newNetworkingV2Client(catalog, testNetworkingRegion, "terraform-provider-selectel/test")

		require.NoError(t, err)
		require.NotNil(t, client)

		// The networking service type and the requested region reach the catalog.
		assert.Equal(t, networkingV2ServiceType, catalog.gotServiceType)
		assert.Equal(t, testNetworkingRegion, catalog.gotEndpointRegion)
	})

	t.Run("returns a diagnosable error when the region has no endpoint", func(t *testing.T) {
		catalog := &testTokenCatalog{
			token:       testNetworkingToken,
			endpointErr: errors.New("endpoint not found"),
		}

		client, err := newNetworkingV2Client(catalog, testNetworkingRegion, "")

		require.Error(t, err)
		assert.Nil(t, client)
		assert.Contains(t, err.Error(), networkingV2ServiceType)
		assert.Contains(t, err.Error(), testNetworkingRegion)
		assert.Contains(t, err.Error(), "endpoint not found")
	})

	t.Run("returns an error when the resolved endpoint is empty", func(t *testing.T) {
		catalog := &testTokenCatalog{token: testNetworkingToken}

		client, err := newNetworkingV2Client(catalog, testNetworkingRegion, "")

		require.Error(t, err)
		assert.Nil(t, client)
		assert.Contains(t, err.Error(), "is empty")
	})

	t.Run("returns an error when no region is given", func(t *testing.T) {
		catalog := &testTokenCatalog{token: testNetworkingToken, endpoint: testNetworkingEndpoint}

		client, err := newNetworkingV2Client(catalog, "", "")

		require.Error(t, err)
		assert.Nil(t, client)
		assert.Contains(t, err.Error(), "region")
	})

	t.Run("returns an error when the token is empty", func(t *testing.T) {
		catalog := &testTokenCatalog{endpoint: testNetworkingEndpoint}

		client, err := newNetworkingV2Client(catalog, testNetworkingRegion, "")

		require.Error(t, err)
		assert.Nil(t, client)
		assert.Contains(t, err.Error(), "token")
	})
}

// TestNetworkingSDKContract is the verification of TASK-172: it compiles only if
// the four attributes agreed with the SDK owner are present in the SDK's write
// models, and it asserts the one that is deliberately create-only.
//
// The provider's specs require dns_publish_fixed_ip on subnet (REQ-014/REQ-015),
// extra_dhcp_opts and binding:vnic_type on port (REQ-040/REQ-041) and a
// create-only flavor_id on router (REQ-057/REQ-059).
func TestNetworkingSDKContract(t *testing.T) {
	t.Run("subnet accepts dns_publish_fixed_ip on create and update", func(t *testing.T) {
		enabled := true
		createReq := subnet.CreateRequest{DNSPublishFixedIP: &enabled}
		updateReq := subnet.UpdateRequest{DNSPublishFixedIP: &enabled}

		require.NotNil(t, createReq.DNSPublishFixedIP)
		require.NotNil(t, updateReq.DNSPublishFixedIP)
		assert.True(t, *createReq.DNSPublishFixedIP)
	})

	t.Run("port accepts extra_dhcp_opts and binding vnic_type on create and update", func(t *testing.T) {
		value := "10.0.0.1"
		options := []port.ExtraDHCPOption{{Name: "tftp-server", Value: value}}
		vnicType := "normal"

		// An update takes the option type whose value is nullable, because a removal
		// is expressed as a name with a null value.
		updateOptions := []port.UpdateExtraDHCPOption{
			{Name: "tftp-server", Value: &value},
			{Name: "bootfile-name"},
		}

		createReq := port.CreateRequest{ExtraDHCPOptions: &options, BindingVNICType: &vnicType}
		updateReq := port.UpdateRequest{ExtraDHCPOptions: &updateOptions, BindingVNICType: &vnicType}

		require.NotNil(t, createReq.ExtraDHCPOptions)
		require.NotNil(t, createReq.BindingVNICType)
		require.NotNil(t, updateReq.ExtraDHCPOptions)
		require.NotNil(t, updateReq.BindingVNICType)
		assert.Equal(t, "tftp-server", (*createReq.ExtraDHCPOptions)[0].Name)
		assert.Nil(t, (*updateReq.ExtraDHCPOptions)[1].Value)
	})

	t.Run("router accepts flavor_id on create only", func(t *testing.T) {
		flavorID := "6f8a7b1c-2d3e-4f50-8a9b-0c1d2e3f4a5b"
		createReq := router.CreateRequest{FlavorID: &flavorID}

		require.NotNil(t, createReq.FlavorID)

		// flavor_id is immutable, so the update model must not carry it; the
		// provider models it as ForceNew (REQ-059) for exactly this reason.
		updateType := reflect.TypeOf(router.UpdateRequest{})
		_, hasFlavorID := updateType.FieldByName("FlavorID")
		assert.False(t, hasFlavorID, "flavor_id must stay create-only in the SDK")
	})

	// TASK-173: the security group stateful attribute, agreed with the SDK owner
	// separately from the four above, after its availability to a project member
	// was confirmed — its api-def sets no enforce_policy, so only the base
	// create_security_group/update_security_group rules apply (REQ-103).
	t.Run("security group accepts stateful on create and update", func(t *testing.T) {
		stateful := false
		createReq := securitygroup.CreateRequest{Stateful: &stateful}
		updateReq := securitygroup.UpdateRequest{Stateful: &stateful}

		require.NotNil(t, createReq.Stateful)
		require.NotNil(t, updateReq.Stateful)
		assert.False(t, *createReq.Stateful)

		// A pointer keeps an omitted value out of the request, so the API default
		// of true survives rather than being overwritten with false.
		for _, typ := range []reflect.Type{
			reflect.TypeOf(securitygroup.CreateRequest{}),
			reflect.TypeOf(securitygroup.UpdateRequest{}),
		} {
			field, exists := typ.FieldByName("Stateful")
			require.True(t, exists)
			assert.Equal(t, reflect.Ptr, field.Type.Kind(), typ.Name()+" must carry stateful as a pointer")
		}
	})

	// REQ-106: shared reports whether an RBAC policy grants another project
	// access. It is read-only on the API — its api-def sets neither allow_post nor
	// allow_put — so it must be readable but absent from both write models.
	t.Run("security group exposes shared for reading only", func(t *testing.T) {
		var group securitygroup.SecurityGroup

		assert.False(t, group.Shared)
		assert.False(t, group.Stateful)

		for _, typ := range []reflect.Type{
			reflect.TypeOf(securitygroup.CreateRequest{}),
			reflect.TypeOf(securitygroup.UpdateRequest{}),
		} {
			_, hasShared := typ.FieldByName("Shared")
			assert.False(t, hasShared, typ.Name()+" must not carry shared")
		}
	})

	t.Run("subnet still excludes segment_id", func(t *testing.T) {
		// segment_id is admin-only and deliberately out of scope, so neither the
		// model nor the requests may carry it.
		for _, typ := range []reflect.Type{
			reflect.TypeOf(subnet.Subnet{}),
			reflect.TypeOf(subnet.CreateRequest{}),
			reflect.TypeOf(subnet.UpdateRequest{}),
		} {
			_, hasSegmentID := typ.FieldByName("SegmentID")
			assert.False(t, hasSegmentID, typ.Name()+" must not carry segment_id")
		}
	})
}
