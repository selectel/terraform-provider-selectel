package selectel

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	vpc "github.com/selectel/vpc-go/pkg/v2"
	"github.com/selectel/vpc-go/pkg/v2/router"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testAccNetworkingRouterV2ResourceName   = "selectel_networking_router_v2.router_tf_acc_test_1"
	testAccNetworkingRouterV2DataSourceName = "data.selectel_networking_router_v2.router_tf_acc_test_1"
)

// networkingRouterV2TestExternalNetwork is the external network the gateway
// tests attach to. It is region-specific, so the test skips rather than
// attaching a gateway to an arbitrary network.
func networkingRouterV2TestExternalNetwork(t *testing.T) string {
	t.Helper()

	networkID := os.Getenv("SELECTEL_ROUTER_EXTERNAL_NETWORK_ID")
	if networkID == "" {
		t.Skip("SELECTEL_ROUTER_EXTERNAL_NETWORK_ID must be set to the id of an external network")
	}

	return networkID
}

// -----------------------------------------------------------------------------
// Schema unit tests.
// -----------------------------------------------------------------------------

// TestNetworkingRouterV2Schema covers REQ-055, REQ-057 and REQ-059: which
// attributes are arguments, which replace the router, and which are read-only.
func TestNetworkingRouterV2Schema(t *testing.T) {
	s := resourceNetworkingRouterV2().Schema

	// REQ-055: project_id and region identify the router and cannot be updated.
	t.Run("project_id and region are required and ForceNew", func(t *testing.T) {
		for _, name := range []string{"project_id", "region"} {
			require.Contains(t, s, name)
			assert.True(t, s[name].Required)
			assert.True(t, s[name].ForceNew)
		}
	})

	// REQ-059: exactly these attributes replace the router.
	t.Run("only the immutable attributes are ForceNew", func(t *testing.T) {
		forceNew := make([]string, 0)
		for name, attr := range s {
			if attr.ForceNew {
				forceNew = append(forceNew, name)
			}
		}

		assert.ElementsMatch(t, []string{
			"project_id", "region", "flavor_id", "availability_zone_hints",
		}, forceNew)
	})

	// REQ-058: the gateway is attached, changed and detached in place.
	t.Run("the gateway attributes are updatable", func(t *testing.T) {
		for _, name := range []string{"external_network_id", "enable_snat"} {
			require.Contains(t, s, name)
			assert.True(t, s[name].Optional)
			assert.False(t, s[name].ForceNew, name+" must be updatable in place")
		}
	})

	// REQ-066: read-only because the policy reserves them for admin.
	for _, name := range []string{"status", "external_fixed_ip", "blocked", "availability_zones", "all_tags"} {
		t.Run(name+" is computed-only", func(t *testing.T) {
			require.Contains(t, s, name)
			assert.True(t, s[name].Computed)
			assert.False(t, s[name].Optional, name+" must not be a configuration argument")
			assert.False(t, s[name].Required)
		})
	}

	// REQ-066: distributed and ha are not exposed at all — the deployed policy
	// reserves even reading them for admin, so there is no actual value to
	// reflect. Same conclusion as subnet segment_id.
	// REQ-057: gateway subnet selection writes the admin-only external_fixed_ips.
	// REQ-065: routes belong to the data source, not to this resource.
	for _, name := range []string{
		"distributed", "ha", "external_subnet_ids", "external_qos_policy_id", "routes",
	} {
		t.Run(name+" is not exposed", func(t *testing.T) {
			assert.NotContains(t, s, name)
		})
	}
}

// TestNetworkingRouterV2GatewayValidation covers REQ-061: enable_snat without
// external_network_id is rejected before the API is called.
func TestNetworkingRouterV2GatewayValidation(t *testing.T) {
	s := resourceNetworkingRouterV2().Schema

	require.Contains(t, s, "enable_snat")
	assert.Equal(t, []string{"external_network_id"}, s["enable_snat"].RequiredWith,
		"enable_snat must declare its dependency on external_network_id so the plan fails locally")
}

// TestNetworkingRouterV2ExpandGateway covers REQ-057 and REQ-058: the gateway is
// built from the two writable sub-attributes, and its absence means detach.
func TestNetworkingRouterV2ExpandGateway(t *testing.T) {
	t.Run("no gateway configured", func(t *testing.T) {
		d := schema.TestResourceDataRaw(t, resourceNetworkingRouterV2().Schema, map[string]interface{}{})

		_, ok := expandNetworkingRouterV2Gateway(d)
		assert.False(t, ok, "an unset external_network_id must report no gateway, which update turns into a detach")
	})

	t.Run("network without an explicit snat setting", func(t *testing.T) {
		d := schema.TestResourceDataRaw(t, resourceNetworkingRouterV2().Schema, map[string]interface{}{
			"external_network_id": "network-id",
		})

		gateway, ok := expandNetworkingRouterV2Gateway(d)

		require.True(t, ok)
		assert.Equal(t, "network-id", gateway.NetworkID)
		assert.Nil(t, gateway.EnableSNAT, "an unset enable_snat must be omitted so the API default applies")
	})

	t.Run("network with snat disabled", func(t *testing.T) {
		d := schema.TestResourceDataRaw(t, resourceNetworkingRouterV2().Schema, map[string]interface{}{
			"external_network_id": "network-id",
			"enable_snat":         false,
		})

		gateway, ok := expandNetworkingRouterV2Gateway(d)

		require.True(t, ok)
		require.NotNil(t, gateway.EnableSNAT)
		assert.False(t, *gateway.EnableSNAT)
	})
}

// TestNetworkingRouterV2FlattenGateway covers REQ-058 SCN-058-03: a detached
// gateway must clear the gateway attributes rather than keep stale values.
func TestNetworkingRouterV2FlattenGateway(t *testing.T) {
	t.Run("gateway present", func(t *testing.T) {
		d := schema.TestResourceDataRaw(t, resourceNetworkingRouterV2().Schema, map[string]interface{}{})

		flattenNetworkingRouterV2(d, &router.Router{
			ID: "router-id",
			ExternalGateway: &router.ExternalGatewayInfo{
				NetworkID:        "network-id",
				EnableSNAT:       true,
				ExternalFixedIPs: []router.ExternalFixedIP{{SubnetID: "subnet-id", IPAddress: "203.0.113.10"}},
			},
		})

		assert.Equal(t, "network-id", d.Get("external_network_id"))
		assert.Equal(t, true, d.Get("enable_snat"))
		assert.Equal(t, "203.0.113.10", d.Get("external_fixed_ip.0.ip_address"))
		assert.Equal(t, "subnet-id", d.Get("external_fixed_ip.0.subnet_id"))
	})

	t.Run("gateway absent", func(t *testing.T) {
		d := schema.TestResourceDataRaw(t, resourceNetworkingRouterV2().Schema, map[string]interface{}{
			"external_network_id": "network-id",
		})

		flattenNetworkingRouterV2(d, &router.Router{ID: "router-id"})

		assert.Empty(t, d.Get("external_network_id"))
		assert.Empty(t, d.Get("external_fixed_ip"))
	})
}

// TestNetworkingRouterV2DataSourceSelectors covers REQ-065.
func TestNetworkingRouterV2DataSourceSelectors(t *testing.T) {
	s := dataSourceNetworkingRouterV2().Schema

	selectors := make([]string, 0)
	for name, attr := range s {
		if attr.Optional || attr.Required {
			selectors = append(selectors, name)
		}
	}

	assert.ElementsMatch(t, []string{
		"project_id", "region", "router_id", "name", "description",
		"admin_state_up", "status", "tags",
	}, selectors)

	// REQ-065: routes is returned by the data source and reports every route,
	// whichever resource added it.
	require.Contains(t, s, "routes")
	assert.True(t, s["routes"].Computed)
	assert.False(t, s["routes"].Optional)

	// REQ-066: not selectable, because the caller cannot read them.
	for _, name := range []string{"distributed", "ha"} {
		assert.NotContains(t, s, name)
	}
}

// TestNetworkingRouterV2AllowlistErrorClasses covers SCN-062-03 and SCN-062-05:
// the two allowlist failures differ by diagnostic class, and the 404 must not be
// treated as a missing managed resource.
//
// Staging runs under the administrator role only, so a non-privileged caller
// cannot be produced there; per the recorded resolution these two scenarios are
// proven against the shared classifier from TASK-003.
func TestNetworkingRouterV2AllowlistErrorClasses(t *testing.T) {
	t.Run("a value outside a non-empty allowlist is an authorization failure", func(t *testing.T) {
		err := newTestAPIError(http.StatusForbidden, vpc.ErrorClassForbidden)

		assert.True(t, networkingIsForbidden(err))
		assert.False(t, networkingIsNotFound(err))
	})

	t.Run("an unset allowlist is a not-found failure that keeps the resource", func(t *testing.T) {
		err := newTestAPIError(http.StatusNotFound, vpc.ErrorClassNotFound)

		assert.True(t, networkingIsNotFound(err))
		assert.False(t, networkingIsForbidden(err))

		// The two classes are distinguishable, which is what lets Create report
		// the 404 as an error instead of dropping the resource from state.
		d := schema.TestResourceDataRaw(t, resourceNetworkingRouterV2().Schema, map[string]interface{}{})
		d.SetId("router-id")
		diags := networkingCheckDeleted(d, err, networkingRouterV2Object)
		assert.Nil(t, diags, "on Read the same class clears the id")
		assert.Empty(t, d.Id())
	})
}

// -----------------------------------------------------------------------------
// Acceptance tests against a staging environment (TF_ACC=1).
// -----------------------------------------------------------------------------

// TestAccNetworkingRouterV2Basic covers SCN-055-02, SCN-056-01, SCN-056-02,
// SCN-062-04, SCN-064-01 and SCN-063-02.
func TestAccNetworkingRouterV2Basic(t *testing.T) {
	var found router.Router
	projectID, region := networkingV2TestEnv()

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheckWithProjectIDAndRegion(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckNetworkingRouterV2Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkingRouterV2Basic(projectID, region),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNetworkingRouterV2Exists(&found),
					// SCN-056-01: assigned by the API, and no gateway.
					resource.TestCheckResourceAttrSet(testAccNetworkingRouterV2ResourceName, "admin_state_up"),
					resource.TestCheckResourceAttrSet(testAccNetworkingRouterV2ResourceName, "status"),
					resource.TestCheckResourceAttrSet(testAccNetworkingRouterV2ResourceName, "blocked"),
					resource.TestCheckResourceAttr(testAccNetworkingRouterV2ResourceName, "external_network_id", ""),
					resource.TestCheckResourceAttr(testAccNetworkingRouterV2ResourceName, "external_fixed_ip.#", "0"),
				),
			},
			{
				// SCN-056-02: the follow-up plan is empty.
				Config:   testAccNetworkingRouterV2Basic(projectID, region),
				PlanOnly: true,
			},
			{
				// SCN-064-01: import fills project_id and region from the provider.
				ResourceName:      testAccNetworkingRouterV2ResourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

// TestAccNetworkingRouterV2Update covers SCN-057-03, SCN-058-01 and SCN-060-01.
func TestAccNetworkingRouterV2Update(t *testing.T) {
	var found router.Router
	projectID, region := networkingV2TestEnv()

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheckWithProjectIDAndRegion(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckNetworkingRouterV2Destroy,
		Steps: []resource.TestStep{
			{
				// SCN-057-03: an explicit admin_state_up=false differs from the default.
				Config: testAccNetworkingRouterV2Named(projectID, region, "router-1", false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNetworkingRouterV2Exists(&found),
					resource.TestCheckResourceAttr(testAccNetworkingRouterV2ResourceName, "admin_state_up", "false"),
					resource.TestCheckResourceAttr(testAccNetworkingRouterV2ResourceName, "tags.#", "2"),
				),
			},
			{
				// SCN-058-01: name, description and admin_state_up change in place.
				Config: testAccNetworkingRouterV2Named(projectID, region, "router-1-renamed", true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNetworkingRouterV2SameID(&found),
					resource.TestCheckResourceAttr(testAccNetworkingRouterV2ResourceName, "name", "router-1-renamed"),
					resource.TestCheckResourceAttr(testAccNetworkingRouterV2ResourceName, "admin_state_up", "true"),
				),
			},
			{
				// SCN-060-01: an out-of-band change surfaces as drift.
				Config:             testAccNetworkingRouterV2Named(projectID, region, "router-1-renamed", true),
				Check:              testAccCheckNetworkingRouterV2RenameOutOfBand(),
				ExpectNonEmptyPlan: true,
			},
			{
				Config: testAccNetworkingRouterV2Named(projectID, region, "router-1-renamed", true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNetworkingRouterV2SameID(&found),
					resource.TestCheckResourceAttr(testAccNetworkingRouterV2ResourceName, "name", "router-1-renamed"),
				),
			},
		},
	})
}

// TestAccNetworkingRouterV2Gateway covers SCN-057-01, SCN-057-02, SCN-058-02 and
// SCN-058-03.
func TestAccNetworkingRouterV2Gateway(t *testing.T) {
	var found router.Router
	projectID, region := networkingV2TestEnv()
	externalNetworkID := networkingRouterV2TestExternalNetwork(t)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheckWithProjectIDAndRegion(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckNetworkingRouterV2Destroy,
		Steps: []resource.TestStep{
			{
				// SCN-057-01: a gateway without an explicit enable_snat.
				Config: testAccNetworkingRouterV2Gateway(projectID, region, externalNetworkID, ""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNetworkingRouterV2Exists(&found),
					resource.TestCheckResourceAttr(testAccNetworkingRouterV2ResourceName, "external_network_id", externalNetworkID),
					resource.TestCheckResourceAttrSet(testAccNetworkingRouterV2ResourceName, "enable_snat"),
					// The API assigns the gateway addresses.
					resource.TestCheckResourceAttrSet(testAccNetworkingRouterV2ResourceName, "external_fixed_ip.0.ip_address"),
				),
			},
			{
				// SCN-057-02: an explicit enable_snat=false, in place.
				Config: testAccNetworkingRouterV2Gateway(projectID, region, externalNetworkID, "enable_snat = false"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNetworkingRouterV2SameID(&found),
					resource.TestCheckResourceAttr(testAccNetworkingRouterV2ResourceName, "enable_snat", "false"),
				),
			},
			{
				// SCN-058-03: removing external_network_id detaches the gateway.
				Config: testAccNetworkingRouterV2Basic(projectID, region),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNetworkingRouterV2SameID(&found),
					resource.TestCheckResourceAttr(testAccNetworkingRouterV2ResourceName, "external_network_id", ""),
					resource.TestCheckResourceAttr(testAccNetworkingRouterV2ResourceName, "external_fixed_ip.#", "0"),
				),
			},
			{
				// SCN-058-02: attaching it again to a router that has no gateway.
				Config: testAccNetworkingRouterV2Gateway(projectID, region, externalNetworkID, ""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNetworkingRouterV2SameID(&found),
					resource.TestCheckResourceAttr(testAccNetworkingRouterV2ResourceName, "external_network_id", externalNetworkID),
				),
			},
		},
	})
}

// TestAccNetworkingRouterV2Replacement covers SCN-059-01 and SCN-059-02.
func TestAccNetworkingRouterV2Replacement(t *testing.T) {
	var found router.Router
	projectID, region := networkingV2TestEnv()

	zone := os.Getenv("SELECTEL_ROUTER_ZONE")
	otherZone := os.Getenv("SELECTEL_ROUTER_ZONE_ALTERNATIVE")
	if zone == "" || otherZone == "" {
		t.Skip("SELECTEL_ROUTER_ZONE and SELECTEL_ROUTER_ZONE_ALTERNATIVE must name two zones from allowed_router_zones")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheckWithProjectIDAndRegion(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckNetworkingRouterV2Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkingRouterV2Zone(projectID, region, zone),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNetworkingRouterV2Exists(&found),
					resource.TestCheckResourceAttr(testAccNetworkingRouterV2ResourceName, "availability_zone_hints.0", zone),
				),
			},
			{
				// SCN-059-01: a different zone hint replaces the router.
				Config: testAccNetworkingRouterV2Zone(projectID, region, otherZone),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNetworkingRouterV2Replaced(&found),
					resource.TestCheckResourceAttr(testAccNetworkingRouterV2ResourceName, "availability_zone_hints.0", otherZone),
				),
			},
		},
	})
}

// TestAccNetworkingRouterV2ZoneOutsideAllowlist covers SCN-062-01: a privileged
// caller is not restricted by allowed_router_zones. It needs the name of a zone
// that really exists in the region but is outside the allowlist — an arbitrary
// name would be rejected by ordinary availability-zone validation instead, so
// the test would pass for the wrong reason.
func TestAccNetworkingRouterV2ZoneOutsideAllowlist(t *testing.T) {
	var found router.Router
	projectID, region := networkingV2TestEnv()

	zone := os.Getenv("SELECTEL_ROUTER_ZONE_OUTSIDE_ALLOWLIST")
	if zone == "" {
		t.Skip("SELECTEL_ROUTER_ZONE_OUTSIDE_ALLOWLIST must name a real zone outside allowed_router_zones")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheckWithProjectIDAndRegion(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckNetworkingRouterV2Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkingRouterV2Zone(projectID, region, zone),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNetworkingRouterV2Exists(&found),
					resource.TestCheckResourceAttr(testAccNetworkingRouterV2ResourceName, "availability_zone_hints.0", zone),
				),
			},
		},
	})
}

// TestAccNetworkingRouterV2DeleteConflict covers SCN-063-01: a router with an
// attached interface is refused with a conflict instead of being detached.
func TestAccNetworkingRouterV2DeleteConflict(t *testing.T) {
	projectID, region := networkingV2TestEnv()
	var attached networkingRouterV2AttachedInterface

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheckWithProjectIDAndRegion(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckNetworkingRouterV2Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkingRouterV2WithInterface(projectID, region),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(testAccNetworkingRouterV2ResourceName, "id"),
					// The interface is attached outside Terraform on purpose. Declaring
					// it as a resource would make Terraform detach it before the router,
					// so the conflict this scenario is about could never happen.
					testAccCheckNetworkingRouterV2AttachInterfaceOutOfBand(&attached),
				),
			},
			{
				// Destroying only the router, while its interface stays, must be
				// refused with a conflict.
				Config:      testAccNetworkingRouterV2InterfaceOnly(projectID, region),
				ExpectError: regexp.MustCompile(`(?i)conflict|in ?use|still has ports`),
			},
			{
				// The interface has to go before the framework can destroy anything:
				// it is not in the state, so nothing else would ever remove it.
				PreConfig: attached.detach,
				Config:    testAccNetworkingRouterV2InterfaceOnly(projectID, region),
			},
		},
	})
}

// TestAccNetworkingRouterV2ExternalDeletion covers SCN-060-02 and SCN-063-03.
func TestAccNetworkingRouterV2ExternalDeletion(t *testing.T) {
	var found router.Router
	projectID, region := networkingV2TestEnv()

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheckWithProjectIDAndRegion(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckNetworkingRouterV2Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkingRouterV2Basic(projectID, region),
				Check:  testAccCheckNetworkingRouterV2Exists(&found),
			},
			{
				Config:             testAccNetworkingRouterV2Basic(projectID, region),
				Check:              testAccCheckNetworkingRouterV2DeleteOutOfBand(),
				ExpectNonEmptyPlan: true,
			},
			{
				Config: testAccNetworkingRouterV2Basic(projectID, region),
				Check:  testAccCheckNetworkingRouterV2Exists(&found),
			},
		},
	})
}

// TestAccNetworkingRouterV2DataSource covers SCN-065-01, SCN-065-02, SCN-065-03
// and SCN-066-03.
func TestAccNetworkingRouterV2DataSource(t *testing.T) {
	projectID, region := networkingV2TestEnv()
	// The project is shared, so the name the data source searches on must be
	// unique to this run.
	routerName := acctest.RandomWithPrefix("tf-acc-router")

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheckWithProjectIDAndRegion(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckNetworkingRouterV2Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkingRouterV2DataSource(projectID, region, routerName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(
						testAccNetworkingRouterV2DataSourceName, "router_id",
						testAccNetworkingRouterV2ResourceName, "id",
					),
					// SCN-066-03: the same blocked value as the resource reports.
					resource.TestCheckResourceAttrPair(
						testAccNetworkingRouterV2DataSourceName, "blocked",
						testAccNetworkingRouterV2ResourceName, "blocked",
					),
					resource.TestCheckResourceAttrSet(testAccNetworkingRouterV2DataSourceName, "routes.#"),
				),
			},
			{
				Config:      testAccNetworkingRouterV2DataSourceNoMatch(projectID, region),
				ExpectError: regexp.MustCompile(`no networking router found`),
			},
		},
	})
}

// -----------------------------------------------------------------------------
// Helpers.
// -----------------------------------------------------------------------------

// networkingRouterV2AttachedInterface remembers an interface attached outside
// Terraform so a later step can detach it: PreConfig runs before the step's
// refresh and has no access to the state.
type networkingRouterV2AttachedInterface struct {
	client   *vpc.Client
	routerID string
	subnetID string
}

func (attached *networkingRouterV2AttachedInterface) detach() {
	if attached.client == nil {
		return
	}

	_, _ = router.RemoveInterface(context.Background(), attached.client,
		attached.routerID, router.BySubnet(attached.subnetID))
}

// testAccCheckNetworkingRouterV2AttachInterfaceOutOfBand attaches the test subnet
// to the test router without Terraform, which is what makes the router
// undeletable while the interface is there.
func testAccCheckNetworkingRouterV2AttachInterfaceOutOfBand(
	attached *networkingRouterV2AttachedInterface,
) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		routerState, ok := s.RootModule().Resources[testAccNetworkingRouterV2ResourceName]
		if !ok {
			return fmt.Errorf("not found: %s", testAccNetworkingRouterV2ResourceName)
		}
		subnetState, ok := s.RootModule().Resources[testAccNetworkingSubnetV2ResourceName]
		if !ok {
			return fmt.Errorf("not found: %s", testAccNetworkingSubnetV2ResourceName)
		}

		client, err := networkingV2Client(
			testAccProvider.Meta(),
			routerState.Primary.Attributes["project_id"],
			routerState.Primary.Attributes["region"],
		)
		if err != nil {
			return err
		}

		if _, err := router.AddInterface(context.Background(), client,
			routerState.Primary.ID, router.BySubnet(subnetState.Primary.ID)); err != nil {
			return err
		}

		attached.client = client
		attached.routerID = routerState.Primary.ID
		attached.subnetID = subnetState.Primary.ID

		return nil
	}
}

func testAccCheckNetworkingRouterV2Exists(found *router.Router) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		handle, err := testAccNetworkingNetworkV2Handle(s, testAccNetworkingRouterV2ResourceName)
		if err != nil {
			return err
		}

		result, err := router.Get(context.Background(), handle.client, handle.id)
		if err != nil {
			return err
		}

		*found = *result

		return nil
	}
}

func testAccCheckNetworkingRouterV2SameID(found *router.Router) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[testAccNetworkingRouterV2ResourceName]
		if !ok {
			return fmt.Errorf("not found: %s", testAccNetworkingRouterV2ResourceName)
		}

		if rs.Primary.ID != found.ID {
			return fmt.Errorf("networking router was replaced: %s -> %s", found.ID, rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckNetworkingRouterV2Replaced(found *router.Router) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[testAccNetworkingRouterV2ResourceName]
		if !ok {
			return fmt.Errorf("not found: %s", testAccNetworkingRouterV2ResourceName)
		}

		if rs.Primary.ID == found.ID {
			return fmt.Errorf("networking router %s was updated in place instead of replaced", found.ID)
		}

		found.ID = rs.Primary.ID

		return nil
	}
}

func testAccCheckNetworkingRouterV2RenameOutOfBand() resource.TestCheckFunc {
	return func(s *terraform.State) error {
		handle, err := testAccNetworkingNetworkV2Handle(s, testAccNetworkingRouterV2ResourceName)
		if err != nil {
			return err
		}

		drifted := testAccNetworkingV2DriftedName
		_, err = router.Update(context.Background(), handle.client, handle.id, router.UpdateRequest{
			Name: &drifted,
		})

		return err
	}
}

func testAccCheckNetworkingRouterV2DeleteOutOfBand() resource.TestCheckFunc {
	return func(s *terraform.State) error {
		handle, err := testAccNetworkingNetworkV2Handle(s, testAccNetworkingRouterV2ResourceName)
		if err != nil {
			return err
		}

		return router.Delete(context.Background(), handle.client, handle.id)
	}
}

func testAccCheckNetworkingRouterV2Destroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "selectel_networking_router_v2" {
			continue
		}

		client, err := networkingV2Client(
			testAccProvider.Meta(),
			rs.Primary.Attributes["project_id"],
			rs.Primary.Attributes["region"],
		)
		if err != nil {
			return err
		}

		if _, err := router.Get(context.Background(), client, rs.Primary.ID); err == nil {
			return errors.New("networking router still exists")
		} else if !networkingIsNotFound(err) {
			return err
		}
	}

	return nil
}

// -----------------------------------------------------------------------------
// Configurations.
// -----------------------------------------------------------------------------

func testAccNetworkingRouterV2Basic(projectID, region string) string {
	return testAccNetworkingV2Base(projectID) + fmt.Sprintf(`
resource "selectel_networking_router_v2" "router_tf_acc_test_1" {
  project_id = local.project_id
  region     = "%s"
}
`, region)
}

func testAccNetworkingRouterV2Named(projectID, region, name string, adminStateUp bool) string {
	return testAccNetworkingV2Base(projectID) + fmt.Sprintf(`
resource "selectel_networking_router_v2" "router_tf_acc_test_1" {
  project_id     = local.project_id
  region         = "%s"
  name           = "%s"
  description    = "Managed by the acceptance test"
  admin_state_up = %t
  tags           = ["tf-acc", "networking"]
}
`, region, name, adminStateUp)
}

func testAccNetworkingRouterV2Gateway(projectID, region, externalNetworkID, extra string) string {
	return testAccNetworkingV2Base(projectID) + fmt.Sprintf(`
resource "selectel_networking_router_v2" "router_tf_acc_test_1" {
  project_id          = local.project_id
  region              = "%s"
  external_network_id = "%s"
  %s
}
`, region, externalNetworkID, extra)
}

func testAccNetworkingRouterV2Zone(projectID, region, zone string) string {
	return testAccNetworkingV2Base(projectID) + fmt.Sprintf(`
resource "selectel_networking_router_v2" "router_tf_acc_test_1" {
  project_id              = local.project_id
  region                  = "%s"
  availability_zone_hints = ["%s"]
}
`, region, zone)
}

func testAccNetworkingRouterV2WithInterface(projectID, region string) string {
	return testAccNetworkingRouterV2InterfaceOnly(projectID, region) + fmt.Sprintf(`
resource "selectel_networking_router_v2" "router_tf_acc_test_1" {
  project_id = local.project_id
  region     = "%s"
}
`, region)
}

// testAccNetworkingRouterV2InterfaceOnly keeps the network and subnet that the
// router interface is attached to, without the router resource itself, so the
// destroy step leaves the interface in place.
func testAccNetworkingRouterV2InterfaceOnly(projectID, region string) string {
	return testAccNetworkingV2Base(projectID) + fmt.Sprintf(`
resource "selectel_networking_network_v2" "network_tf_acc_test_1" {
  project_id = local.project_id
  region     = "%s"
  name       = "network-tf-acc-test-1"
}

resource "selectel_networking_subnet_v2" "subnet_tf_acc_test_1" {
  project_id = local.project_id
  region     = "%s"
  network_id = selectel_networking_network_v2.network_tf_acc_test_1.id
  cidr       = "192.168.199.0/24"
}
`, region, region)
}

func testAccNetworkingRouterV2DataSource(projectID, region, name string) string {
	return testAccNetworkingRouterV2Named(projectID, region, name, true) + fmt.Sprintf(`
data "selectel_networking_router_v2" "router_tf_acc_test_1" {
  project_id = local.project_id
  region     = "%s"
  name       = selectel_networking_router_v2.router_tf_acc_test_1.name
}
`, region)
}

func testAccNetworkingRouterV2DataSourceNoMatch(projectID, region string) string {
	return testAccNetworkingV2Base(projectID) + fmt.Sprintf(`
data "selectel_networking_router_v2" "router_tf_acc_test_1" {
  project_id = local.project_id
  region     = "%s"
  name       = "tf-acc-no-such-router"
}
`, region)
}
