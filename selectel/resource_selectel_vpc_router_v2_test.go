package selectel

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	vpc "github.com/selectel/vpc-go/pkg/v2"
	"github.com/selectel/vpc-go/pkg/v2/router"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/terraform-providers/terraform-provider-selectel/selectel/internal/httptest"
	"github.com/terraform-providers/terraform-provider-selectel/selectel/internal/schemaconfig"
)

const (
	testAccVPCRouterV2ResourceName                  = "selectel_vpc_router_v2.router_tf_acc_test_1"
	testAccVPCRouterV2ExternalNetworkDataSourceName = "data.selectel_vpc_network_v2.external"
)

func TestVPCRouterV2ExpandGateway(t *testing.T) {
	t.Run("no gateway configured", func(t *testing.T) {
		d := schemaconfig.TestResourceData(t, resourceVPCRouterV2().Schema, map[string]any{})

		_, ok := expandVPCRouterV2Gateway(d)
		assert.False(t, ok, "an unset external_network_id must report no gateway, which update turns into a detach")
	})

	t.Run("network without an explicit snat setting", func(t *testing.T) {
		d := schemaconfig.TestResourceData(t, resourceVPCRouterV2().Schema, map[string]any{
			"external_network_id": "network-id",
		})

		gateway, ok := expandVPCRouterV2Gateway(d)

		require.True(t, ok)
		assert.Equal(t, "network-id", gateway.NetworkID)
		assert.Nil(t, gateway.EnableSNAT, "an unset enable_snat must be omitted so the API default applies")
	})

	t.Run("network with snat disabled", func(t *testing.T) {
		d := schemaconfig.TestResourceData(t, resourceVPCRouterV2().Schema, map[string]any{
			"external_network_id": "network-id",
			"enable_snat":         false,
		})

		gateway, ok := expandVPCRouterV2Gateway(d)

		require.True(t, ok)
		require.NotNil(t, gateway.EnableSNAT)
		assert.False(t, *gateway.EnableSNAT)
	})
}

func TestVPCRouterV2FlattenGateway(t *testing.T) {
	t.Run("gateway present", func(t *testing.T) {
		d := schemaconfig.TestResourceData(t, resourceVPCRouterV2().Schema, map[string]any{})

		flattenVPCRouterV2(d, &router.Router{
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
		d := schemaconfig.TestResourceData(t, resourceVPCRouterV2().Schema, map[string]any{
			"external_network_id": "network-id",
		})

		flattenVPCRouterV2(d, &router.Router{ID: "router-id"})

		assert.Empty(t, d.Get("external_network_id"))
		assert.Empty(t, d.Get("external_fixed_ip"))
	})
}

func TestAccVPCRouterV2Basic(t *testing.T) {
	var found router.Router
	projectID, region := vpcV2TestEnv()

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheckWithProjectIDAndRegion(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckVPCRouterV2Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCRouterV2Basic(projectID, region),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCRouterV2Exists(&found),
					resource.TestCheckResourceAttr(testAccVPCRouterV2ResourceName, "admin_state_up", "true"),
					resource.TestCheckResourceAttrSet(testAccVPCRouterV2ResourceName, "status"),
					resource.TestCheckResourceAttrSet(testAccVPCRouterV2ResourceName, "blocked"),
					resource.TestCheckResourceAttr(testAccVPCRouterV2ResourceName, "external_network_id", ""),
					resource.TestCheckResourceAttr(testAccVPCRouterV2ResourceName, "external_fixed_ip.#", "0"),
				),
			},
			{
				Config:   testAccVPCRouterV2Basic(projectID, region),
				PlanOnly: true,
			},
			{
				ResourceName:      testAccVPCRouterV2ResourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config:             testAccVPCRouterV2Basic(projectID, region),
				Check:              testAccCheckVPCV2DeleteOutOfBand(testAccVPCRouterV2ResourceName, router.Delete),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccVPCRouterV2Update(t *testing.T) {
	var found router.Router
	projectID, region := vpcV2TestEnv()

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheckWithProjectIDAndRegion(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckVPCRouterV2Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCRouterV2Named(projectID, region, "router-1", false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCRouterV2Exists(&found),
					resource.TestCheckResourceAttr(testAccVPCRouterV2ResourceName, "admin_state_up", "false"),
					resource.TestCheckResourceAttr(testAccVPCRouterV2ResourceName, "tags.#", "2"),
				),
			},
			{
				Config: testAccVPCRouterV2Described(projectID, region, "router-1-renamed", "updated by the acceptance test", true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCRouterV2SameID(&found),
					resource.TestCheckResourceAttr(testAccVPCRouterV2ResourceName, "name", "router-1-renamed"),
					resource.TestCheckResourceAttr(testAccVPCRouterV2ResourceName, "description", "updated by the acceptance test"),
					resource.TestCheckResourceAttr(testAccVPCRouterV2ResourceName, "admin_state_up", "true"),
				),
			},
			{
				Config:             testAccVPCRouterV2Named(projectID, region, "router-1-renamed", true),
				Check:              testAccCheckVPCRouterV2RenameOutOfBand(),
				ExpectNonEmptyPlan: true,
			},
			{
				Config: testAccVPCRouterV2Named(projectID, region, "router-1-renamed", true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCRouterV2SameID(&found),
					resource.TestCheckResourceAttr(testAccVPCRouterV2ResourceName, "name", "router-1-renamed"),
				),
			},
		},
	})
}

func TestAccVPCRouterV2TagsOutOfBand(t *testing.T) {
	var found router.Router
	projectID, region := vpcV2TestEnv()
	routerName := acctest.RandomWithPrefix("tf-acc-router")

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheckWithProjectIDAndRegion(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckVPCRouterV2Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCRouterV2Named(projectID, region, routerName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCRouterV2Exists(&found),
					testAccCheckVPCTagsV2AddOutOfBand(testAccVPCRouterV2TagOperations, "foreign"),
				),
			},
			{
				Config: testAccVPCRouterV2TagsUpdated(projectID, region, routerName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCRouterV2SameID(&found),
					resource.TestCheckResourceAttr(testAccVPCRouterV2ResourceName, "tags.#", "1"),
					resource.TestCheckTypeSetElemAttr(testAccVPCRouterV2ResourceName, "all_tags.*", "foreign"),
					testAccCheckVPCTagsV2Match(testAccVPCRouterV2TagOperations, []string{"tf-acc", "foreign"}),
				),
			},
			{
				Config:   testAccVPCRouterV2TagsUpdated(projectID, region, routerName),
				PlanOnly: true,
			},
		},
	})
}

func TestAccVPCRouterV2Gateway(t *testing.T) {
	var found router.Router
	projectID, region := vpcV2TestEnv()

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheckWithProjectIDAndRegion(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckVPCRouterV2Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCRouterV2Gateway(projectID, region, ""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCRouterV2Exists(&found),
					resource.TestCheckResourceAttrPair(
						testAccVPCRouterV2ResourceName, "external_network_id",
						testAccVPCRouterV2ExternalNetworkDataSourceName, "network_id"),
					resource.TestCheckResourceAttrSet(testAccVPCRouterV2ResourceName, "enable_snat"),
					resource.TestCheckResourceAttrSet(testAccVPCRouterV2ResourceName, "external_fixed_ip.0.ip_address"),
				),
			},
			{
				Config: testAccVPCRouterV2Gateway(projectID, region, "enable_snat = false"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCRouterV2SameID(&found),
					resource.TestCheckResourceAttr(testAccVPCRouterV2ResourceName, "enable_snat", "false"),
				),
			},
			{
				Config: testAccVPCRouterV2GatewayDetached(projectID, region),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCRouterV2SameID(&found),
					resource.TestCheckResourceAttr(testAccVPCRouterV2ResourceName, "external_network_id", ""),
					resource.TestCheckResourceAttr(testAccVPCRouterV2ResourceName, "external_fixed_ip.#", "0"),
				),
			},
			{
				Config: testAccVPCRouterV2Gateway(projectID, region, ""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCRouterV2SameID(&found),
					resource.TestCheckResourceAttrPair(
						testAccVPCRouterV2ResourceName, "external_network_id",
						testAccVPCRouterV2ExternalNetworkDataSourceName, "network_id"),
				),
			},
		},
	})
}

func testAccVPCRouterV2TagOperations(s *terraform.State) (vpcTagsV2AccOperations, error) {
	handle, err := testAccVPCNetworkV2Handle(s, testAccVPCRouterV2ResourceName)
	if err != nil {
		return nil, err
	}

	return router.TagOperations(handle.client, handle.id), nil
}

func testAccCheckVPCRouterV2Exists(found *router.Router) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		handle, err := testAccVPCNetworkV2Handle(s, testAccVPCRouterV2ResourceName)
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

func testAccCheckVPCRouterV2SameID(found *router.Router) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[testAccVPCRouterV2ResourceName]
		if !ok {
			return fmt.Errorf("not found: %s", testAccVPCRouterV2ResourceName)
		}

		if rs.Primary.ID != found.ID {
			return fmt.Errorf("VPC router was replaced: %s -> %s", found.ID, rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckVPCRouterV2RenameOutOfBand() resource.TestCheckFunc {
	return func(s *terraform.State) error {
		handle, err := testAccVPCNetworkV2Handle(s, testAccVPCRouterV2ResourceName)
		if err != nil {
			return err
		}

		drifted := testAccVPCV2DriftedName
		_, err = router.Update(context.Background(), handle.client, handle.id, router.UpdateRequest{
			Name: &drifted,
		})

		return err
	}
}

func testAccCheckVPCRouterV2Destroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "selectel_vpc_router_v2" {
			continue
		}

		client, err := vpcV2Client(
			testAccProvider.Meta(),
			rs.Primary.Attributes["project_id"],
			rs.Primary.Attributes["region"],
		)
		if err != nil {
			return err
		}

		if _, err := router.Get(context.Background(), client, rs.Primary.ID); err == nil {
			return errors.New("VPC router still exists")
		} else if !vpc.IsErrorClass(err, vpc.ErrorClassNotFound) {
			return err
		}
	}

	return nil
}

func testAccVPCRouterV2Basic(projectID, region string) string {
	return testAccVPCV2Base(projectID) + fmt.Sprintf(`
resource "selectel_vpc_router_v2" "router_tf_acc_test_1" {
  project_id = local.project_id
  region     = "%s"
}
`, region)
}

func testAccVPCRouterV2Named(projectID, region, name string, adminStateUp bool) string {
	return testAccVPCRouterV2Described(projectID, region, name, "Managed by the acceptance test", adminStateUp)
}

func testAccVPCRouterV2Described(projectID, region, name, description string, adminStateUp bool) string {
	return testAccVPCV2Base(projectID) + fmt.Sprintf(`
resource "selectel_vpc_router_v2" "router_tf_acc_test_1" {
  project_id     = local.project_id
  region         = "%s"
  name           = "%s"
  description    = "%s"
  admin_state_up = %t
  tags           = ["tf-acc", "VPC"]
}
`, region, name, description, adminStateUp)
}

func testAccVPCRouterV2TagsUpdated(projectID, region, name string) string {
	return testAccVPCV2Base(projectID) + fmt.Sprintf(`
resource "selectel_vpc_router_v2" "router_tf_acc_test_1" {
  project_id     = local.project_id
  region         = "%s"
  name           = "%s"
  description    = "Managed by the acceptance test"
  admin_state_up = true
  tags           = ["tf-acc"]
}
`, region, name)
}

func testAccVPCRouterV2GatewaySupport(projectID, region string) string {
	return testAccVPCV2Base(projectID) + fmt.Sprintf(`
data "selectel_vpc_network_v2" "external" {
  project_id = local.project_id
  region     = "%s"
  external   = true
}
`, region)
}

func testAccVPCRouterV2Gateway(projectID, region, extra string) string {
	return testAccVPCRouterV2GatewaySupport(projectID, region) + fmt.Sprintf(`
resource "selectel_vpc_router_v2" "router_tf_acc_test_1" {
  project_id          = local.project_id
  region              = "%s"
  external_network_id = data.selectel_vpc_network_v2.external.network_id
  %s
}
`, region, extra)
}

func testAccVPCRouterV2GatewayDetached(projectID, region string) string {
	return testAccVPCRouterV2GatewaySupport(projectID, region) + fmt.Sprintf(`
resource "selectel_vpc_router_v2" "router_tf_acc_test_1" {
  project_id = local.project_id
  region     = "%s"
}
`, region)
}

func TestVPCRouterV2CheckCreated(t *testing.T) {
	t.Run("active", func(t *testing.T) {
		require.NoError(t, checkVPCRouterV2Created(&router.Router{ID: "r-1", Status: vpcV2StatusActive}))
	})

	t.Run("error", func(t *testing.T) {
		err := checkVPCRouterV2Created(&router.Router{ID: "r-1", Status: vpcV2StatusError})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "VPC router r-1 is in ERROR status")
	})
}

func TestMigrateVPCRouterV2HA(t *testing.T) {
	newClient := func(t *testing.T, statuses ...int) (*vpc.Client, *[]string) {
		t.Helper()
		bodies := &[]string{}
		client, err := vpc.NewClient(vpc.Config{
			Endpoint: "https://vpc.example.test",
			Token:    "token",
			HTTPClient: &http.Client{
				Transport: httptest.RoundTripFunc(func(req *http.Request) (*http.Response, error) {
					assert.Equal(t, http.MethodPut, req.Method)
					assert.Equal(t, "/v2.0/routers/r-1", req.URL.Path)
					raw, _ := io.ReadAll(req.Body)
					*bodies = append(*bodies, string(raw))
					status := http.StatusOK
					if len(*bodies) <= len(statuses) {
						status = statuses[len(*bodies)-1]
					}
					if status != http.StatusOK {
						return httptest.NewFakeResponse(status,
							`{"NeutronError": {"type": "BadRequest", "message": "boom"}}`), nil
					}

					return httptest.NewFakeResponse(http.StatusOK, `{"router": {"id": "r-1"}}`), nil
				}),
			},
		})
		require.NoError(t, err)

		return client, bodies
	}

	t.Run("an enabled router is disabled around the change", func(t *testing.T) {
		client, bodies := newClient(t)
		require.NoError(t, migrateVPCRouterV2HA(context.Background(), client, "r-1", true, true))
		require.Len(t, *bodies, 3)
		assert.JSONEq(t, `{"router": {"admin_state_up": false}}`, (*bodies)[0])
		assert.JSONEq(t, `{"router": {"ha": true}}`, (*bodies)[1])
		assert.JSONEq(t, `{"router": {"admin_state_up": true}}`, (*bodies)[2])
	})

	t.Run("a disabled router changes in one call", func(t *testing.T) {
		client, bodies := newClient(t)
		require.NoError(t, migrateVPCRouterV2HA(context.Background(), client, "r-1", false, false))
		require.Len(t, *bodies, 1)
		assert.JSONEq(t, `{"router": {"ha": false}}`, (*bodies)[0])
	})

	t.Run("a failed change still re-enables the router", func(t *testing.T) {
		client, bodies := newClient(t, http.StatusOK, http.StatusBadRequest, http.StatusOK)
		err := migrateVPCRouterV2HA(context.Background(), client, "r-1", true, true)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "error migrating VPC router r-1 to ha=true")
		assert.NotContains(t, err.Error(), "re-enabling")
		require.Len(t, *bodies, 3)
		assert.JSONEq(t, `{"router": {"admin_state_up": true}}`, (*bodies)[2])
	})

	t.Run("a failed re-enable is reported together with the change", func(t *testing.T) {
		client, bodies := newClient(t, http.StatusOK, http.StatusBadRequest, http.StatusInternalServerError)
		err := migrateVPCRouterV2HA(context.Background(), client, "r-1", true, true)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "error migrating VPC router r-1 to ha=true")
		assert.Contains(t, err.Error(), "error re-enabling VPC router r-1")
		require.Len(t, *bodies, 3)
	})

	t.Run("a failed disable stops before the change", func(t *testing.T) {
		client, bodies := newClient(t, http.StatusInternalServerError)
		err := migrateVPCRouterV2HA(context.Background(), client, "r-1", true, true)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "error disabling VPC router r-1")
		require.Len(t, *bodies, 1)
	})
}

func TestAccVPCRouterV2HA(t *testing.T) {
	projectID, region := vpcV2TestEnv()

	var found router.Router

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheckWithProjectIDAndRegion(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckVPCRouterV2Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCRouterV2HA(projectID, region, true, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCRouterV2Exists(&found),
					resource.TestCheckResourceAttr(testAccVPCRouterV2ResourceName, "ha", "true"),
					resource.TestCheckResourceAttr(testAccVPCRouterV2ResourceName, "admin_state_up", "true"),
				),
			},
			{
				Config: testAccVPCRouterV2HA(projectID, region, false, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCRouterV2SameID(&found),
					resource.TestCheckResourceAttr(testAccVPCRouterV2ResourceName, "ha", "false"),
					resource.TestCheckResourceAttr(testAccVPCRouterV2ResourceName, "admin_state_up", "true"),
				),
			},
			{
				Config: testAccVPCRouterV2HA(projectID, region, true, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCRouterV2SameID(&found),
					resource.TestCheckResourceAttr(testAccVPCRouterV2ResourceName, "ha", "true"),
					resource.TestCheckResourceAttr(testAccVPCRouterV2ResourceName, "admin_state_up", "false"),
				),
			},
		},
	})
}

func testAccVPCRouterV2HA(projectID, region string, ha, adminStateUp bool) string {
	return testAccVPCV2Base(projectID) + fmt.Sprintf(`
resource "selectel_vpc_router_v2" "router_tf_acc_test_1" {
  project_id     = local.project_id
  region         = "%s"
  name           = "router-ha"
  ha             = %t
  admin_state_up = %t
}
`, region, ha, adminStateUp)
}
