package selectel

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/selectel/vpc-go/pkg/v2/subnet"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testAccNetworkingSubnetV2ResourceName = "selectel_networking_subnet_v2.subnet_tf_acc_test_1"

// -----------------------------------------------------------------------------
// Schema unit tests.
// -----------------------------------------------------------------------------

// TestNetworkingSubnetV2SchemaOwnership covers REQ-022: service_types and blocked
// are admin-managed, and segment_id is out of the schema entirely because the SDK
// does not return it. Subnet pools are not shipped, so the attributes that would
// draw a subnet from one are absent too.
func TestNetworkingSubnetV2SchemaOwnership(t *testing.T) {
	s := resourceNetworkingSubnetV2().Schema

	for _, name := range []string{"project_id", "region", "network_id"} {
		t.Run(name+" is required and ForceNew", func(t *testing.T) {
			require.Contains(t, s, name)
			assert.True(t, s[name].Required)
			assert.True(t, s[name].ForceNew)
		})
	}

	for _, name := range []string{"service_types", "blocked"} {
		t.Run(name+" is computed-only", func(t *testing.T) {
			require.Contains(t, s, name)
			assert.True(t, s[name].Computed)
			assert.False(t, s[name].Optional)
			assert.False(t, s[name].Required)
		})
	}

	// Subnet pools are not part of this provider, so a subnet cannot be created
	// from one and the two attributes that would do it are absent.
	t.Run("segment_id, host_routes, subnetpool_id and prefix_length are not part of the resource", func(t *testing.T) {
		for _, name := range []string{"segment_id", "host_routes", "subnetpool_id", "prefix_length"} {
			assert.NotContains(t, s, name)
		}
	})

	// REQ-016: exactly these attributes replace the resource, including the two
	// IPv6 modes that have no API update path.
	t.Run("only the immutable attributes are ForceNew", func(t *testing.T) {
		forceNew := make([]string, 0)
		for name, attr := range s {
			if attr.ForceNew {
				forceNew = append(forceNew, name)
			}
		}

		assert.ElementsMatch(t, []string{
			"project_id", "region", "network_id", "cidr",
			"ip_version", "ipv6_address_mode", "ipv6_ra_mode",
		}, forceNew)
	})

	// REQ-018: the two mutually exclusive pairs are rejected by the schema.
	t.Run("mutually exclusive attributes conflict", func(t *testing.T) {
		assert.Equal(t, []string{"no_gateway"}, s["gateway_ip"].ConflictsWith)
		assert.Equal(t, []string{"gateway_ip"}, s["no_gateway"].ConflictsWith)
	})

	t.Run("defaults match the api", func(t *testing.T) {
		assert.Equal(t, true, s["enable_dhcp"].Default)
		assert.Equal(t, networkingSubnetV2IPVersion4, s["ip_version"].Default)
	})
}

// TestNetworkingSubnetV2CIDRValidation covers SCN-018-05: a value that is not a
// CIDR, or that is not the network address of its prefix, is rejected before the
// API is called.
func TestNetworkingSubnetV2CIDRValidation(t *testing.T) {
	for _, c := range []struct {
		value   string
		wantErr string
	}{
		{value: "192.168.0.0/24"},
		{value: "2001:db8::/64"},
		{value: ""},
		{value: "not-a-cidr", wantErr: "not a valid CIDR"},
		{value: "192.168.0.0", wantErr: "not a valid CIDR"},
		{value: "192.168.0.17/24", wantErr: "must be the network address"},
	} {
		t.Run(c.value, func(t *testing.T) {
			_, errs := validateNetworkingSubnetV2CIDR(c.value, "cidr")

			if c.wantErr == "" {
				assert.Empty(t, errs)

				return
			}

			require.NotEmpty(t, errs)
			assert.Contains(t, errs[0].Error(), c.wantErr)
		})
	}
}

// TestNetworkingSubnetV2ConfigValidation covers SCN-018-02 and SCN-018-04: the
// checks the schema cannot express are made before any API call.
func TestNetworkingSubnetV2ConfigValidation(t *testing.T) {
	newData := func(raw map[string]interface{}) *schema.ResourceData {
		return schema.TestResourceDataRaw(t, resourceNetworkingSubnetV2().Schema, raw)
	}

	t.Run("duplicate dns_nameservers are rejected", func(t *testing.T) {
		err := validateNetworkingSubnetV2Config(newData(map[string]interface{}{
			"dns_nameservers": []interface{}{"8.8.8.8", "8.8.8.8"},
		}))

		require.Error(t, err)
		assert.Contains(t, err.Error(), "duplicate")
	})

	t.Run("ipv6 modes require ip_version 6", func(t *testing.T) {
		err := validateNetworkingSubnetV2Config(newData(map[string]interface{}{
			"ip_version":        networkingSubnetV2IPVersion4,
			"ipv6_address_mode": "slaac",
		}))

		require.Error(t, err)
		assert.Contains(t, err.Error(), "ip_version")
	})

	t.Run("a valid configuration passes", func(t *testing.T) {
		err := validateNetworkingSubnetV2Config(newData(map[string]interface{}{
			"ip_version":      networkingSubnetV2IPVersion4,
			"dns_nameservers": []interface{}{"8.8.8.8", "1.1.1.1"},
		}))

		assert.NoError(t, err)
	})
}

func TestNetworkingSubnetV2AllocationPools(t *testing.T) {
	raw := []interface{}{
		map[string]interface{}{"start": "192.168.0.10", "end": "192.168.0.20"},
	}

	pools := expandNetworkingSubnetV2AllocationPools(raw)

	require.Len(t, pools, 1)
	assert.Equal(t, subnet.AllocationPool{Start: "192.168.0.10", End: "192.168.0.20"}, pools[0])
	assert.Equal(t, raw, flattenNetworkingSubnetV2AllocationPools(pools))
}

func TestNetworkingSubnetV2HostRoutes(t *testing.T) {
	routes := []subnet.HostRoute{{Destination: "10.0.0.0/8", NextHop: "192.168.0.1"}}

	assert.Equal(t, []interface{}{
		map[string]interface{}{"destination_cidr": "10.0.0.0/8", "next_hop": "192.168.0.1"},
	}, flattenNetworkingSubnetV2HostRoutes(routes))
}

// -----------------------------------------------------------------------------
// Acceptance tests against a staging environment (TF_ACC=1).
// -----------------------------------------------------------------------------

// TestAccNetworkingSubnetV2Basic covers SCN-013-01 (create with an explicit CIDR),
// SCN-020-01 (import) and SCN-019-01 (delete).
func TestAccNetworkingSubnetV2Basic(t *testing.T) {
	var found subnet.Subnet
	projectID, region := networkingV2TestEnv()

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheckWithProjectIDAndRegion(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckNetworkingSubnetV2Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkingSubnetV2Basic(projectID, region, "192.168.199.0/24"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNetworkingSubnetV2Exists(testAccNetworkingSubnetV2ResourceName, &found),
					resource.TestCheckResourceAttr(testAccNetworkingSubnetV2ResourceName, "cidr", "192.168.199.0/24"),
					resource.TestCheckResourceAttr(testAccNetworkingSubnetV2ResourceName, "ip_version", "4"),
					resource.TestCheckResourceAttr(testAccNetworkingSubnetV2ResourceName, "enable_dhcp", "true"),
					// The API computes the gateway and the allocation pool.
					resource.TestCheckResourceAttrSet(testAccNetworkingSubnetV2ResourceName, "gateway_ip"),
					resource.TestCheckResourceAttrSet(testAccNetworkingSubnetV2ResourceName, "allocation_pool.#"),
					resource.TestCheckResourceAttrSet(testAccNetworkingSubnetV2ResourceName, "blocked"),
				),
			},
			{
				// SCN-013-01: no drift means an empty plan.
				Config:   testAccNetworkingSubnetV2Basic(projectID, region, "192.168.199.0/24"),
				PlanOnly: true,
			},
			{
				// SCN-020-01: import by the plain subnet identifier.
				ResourceName:      testAccNetworkingSubnetV2ResourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

// TestAccNetworkingSubnetV2OptionalAttributes covers SCN-014-01 … SCN-014-07.
func TestAccNetworkingSubnetV2OptionalAttributes(t *testing.T) {
	var found subnet.Subnet
	projectID, region := networkingV2TestEnv()
	subnetName := acctest.RandomWithPrefix("tf-acc-subnet")

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheckWithProjectIDAndRegion(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckNetworkingSubnetV2Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkingSubnetV2OptionalAttributes(projectID, region, subnetName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNetworkingSubnetV2Exists(testAccNetworkingSubnetV2ResourceName, &found),
					resource.TestCheckResourceAttr(testAccNetworkingSubnetV2ResourceName, "name", subnetName),
					resource.TestCheckResourceAttr(testAccNetworkingSubnetV2ResourceName, "enable_dhcp", "false"),
					resource.TestCheckResourceAttr(testAccNetworkingSubnetV2ResourceName, "dns_nameservers.#", "2"),
					resource.TestCheckResourceAttr(testAccNetworkingSubnetV2ResourceName, "allocation_pool.#", "1"),
					resource.TestCheckResourceAttr(testAccNetworkingSubnetV2ResourceName, "gateway_ip", "192.168.199.1"),
				),
			},
			{
				Config:   testAccNetworkingSubnetV2OptionalAttributes(projectID, region, subnetName),
				PlanOnly: true,
			},
		},
	})
}

// TestAccNetworkingSubnetV2NoGateway covers SCN-014-03: no_gateway creates the
// subnet without a gateway.
func TestAccNetworkingSubnetV2NoGateway(t *testing.T) {
	var found subnet.Subnet
	projectID, region := networkingV2TestEnv()

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheckWithProjectIDAndRegion(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckNetworkingSubnetV2Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkingSubnetV2NoGateway(projectID, region),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNetworkingSubnetV2Exists(testAccNetworkingSubnetV2ResourceName, &found),
					resource.TestCheckResourceAttr(testAccNetworkingSubnetV2ResourceName, "no_gateway", "true"),
					resource.TestCheckResourceAttr(testAccNetworkingSubnetV2ResourceName, "gateway_ip", ""),
				),
			},
			{
				Config:   testAccNetworkingSubnetV2NoGateway(projectID, region),
				PlanOnly: true,
			},
		},
	})
}

// TestAccNetworkingSubnetV2Update covers SCN-015-01 … SCN-015-04.
func TestAccNetworkingSubnetV2Update(t *testing.T) {
	var found subnet.Subnet
	projectID, region := networkingV2TestEnv()
	subnetName := acctest.RandomWithPrefix("tf-acc-subnet")
	updatedName := acctest.RandomWithPrefix("tf-acc-subnet-upd")

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheckWithProjectIDAndRegion(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckNetworkingSubnetV2Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkingSubnetV2OptionalAttributes(projectID, region, subnetName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNetworkingSubnetV2Exists(testAccNetworkingSubnetV2ResourceName, &found),
				),
			},
			{
				Config: testAccNetworkingSubnetV2Updated(projectID, region, updatedName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNetworkingSubnetV2SameID(testAccNetworkingSubnetV2ResourceName, &found),
					resource.TestCheckResourceAttr(testAccNetworkingSubnetV2ResourceName, "name", updatedName),
					resource.TestCheckResourceAttr(testAccNetworkingSubnetV2ResourceName, "enable_dhcp", "true"),
					resource.TestCheckResourceAttr(testAccNetworkingSubnetV2ResourceName, "dns_nameservers.#", "1"),
					resource.TestCheckResourceAttr(testAccNetworkingSubnetV2ResourceName, "gateway_ip", "192.168.199.2"),
				),
			},
			{
				Config:   testAccNetworkingSubnetV2Updated(projectID, region, updatedName),
				PlanOnly: true,
			},
		},
	})
}

// TestAccNetworkingSubnetV2ForceNew covers SCN-016-01: changing the CIDR
// replaces the subnet instead of updating it.
func TestAccNetworkingSubnetV2ForceNew(t *testing.T) {
	var found subnet.Subnet
	projectID, region := networkingV2TestEnv()

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheckWithProjectIDAndRegion(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckNetworkingSubnetV2Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkingSubnetV2Basic(projectID, region, "192.168.199.0/24"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNetworkingSubnetV2Exists(testAccNetworkingSubnetV2ResourceName, &found),
				),
			},
			{
				Config: testAccNetworkingSubnetV2Basic(projectID, region, "192.168.200.0/24"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNetworkingSubnetV2DifferentID(testAccNetworkingSubnetV2ResourceName, &found),
					resource.TestCheckResourceAttr(testAccNetworkingSubnetV2ResourceName, "cidr", "192.168.200.0/24"),
				),
			},
		},
	})
}

// TestAccNetworkingSubnetV2ExternalChanges covers SCN-017-01, SCN-017-02 and
// SCN-019-02.
func TestAccNetworkingSubnetV2ExternalChanges(t *testing.T) {
	var found subnet.Subnet
	projectID, region := networkingV2TestEnv()
	subnetName := acctest.RandomWithPrefix("tf-acc-subnet")

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheckWithProjectIDAndRegion(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckNetworkingSubnetV2Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkingSubnetV2OptionalAttributes(projectID, region, subnetName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNetworkingSubnetV2Exists(testAccNetworkingSubnetV2ResourceName, &found),
				),
			},
			{
				// SCN-017-01: an out-of-band rename shows up as drift.
				Config:             testAccNetworkingSubnetV2OptionalAttributes(projectID, region, subnetName),
				Check:              testAccCheckNetworkingSubnetV2RenameOutOfBand(&found),
				ExpectNonEmptyPlan: true,
			},
			{
				// SCN-017-02 and SCN-019-02: an out-of-band delete plans a
				// re-create instead of failing.
				Config:             testAccNetworkingSubnetV2OptionalAttributes(projectID, region, subnetName),
				Check:              testAccCheckNetworkingSubnetV2DeleteOutOfBand(),
				ExpectNonEmptyPlan: true,
			},
			{
				Config: testAccNetworkingSubnetV2OptionalAttributes(projectID, region, subnetName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNetworkingSubnetV2Exists(testAccNetworkingSubnetV2ResourceName, &found),
				),
			},
		},
	})
}

// TestAccNetworkingSubnetV2DataSource covers SCN-021-01 … SCN-021-03.
func TestAccNetworkingSubnetV2DataSource(t *testing.T) {
	projectID, region := networkingV2TestEnv()
	subnetName := acctest.RandomWithPrefix("tf-acc-subnet")

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheckWithProjectIDAndRegion(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckNetworkingSubnetV2Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkingSubnetV2DataSource(projectID, region, subnetName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.selectel_networking_subnet_v2.subnet_tf_acc_test_1", "name", subnetName),
					resource.TestCheckResourceAttrPair(
						"data.selectel_networking_subnet_v2.subnet_tf_acc_test_1", "subnet_id",
						testAccNetworkingSubnetV2ResourceName, "id",
					),
					// host_routes is exported by the data source only.
					resource.TestCheckResourceAttrSet(
						"data.selectel_networking_subnet_v2.subnet_tf_acc_test_1", "host_routes.#"),
					resource.TestCheckResourceAttrPair(
						"data.selectel_networking_subnet_v2.subnet_tf_acc_test_1", "blocked",
						testAccNetworkingSubnetV2ResourceName, "blocked",
					),
				),
			},
			{
				Config:      testAccNetworkingSubnetV2DataSourceNoMatch(projectID, region),
				ExpectError: regexp.MustCompile(`no networking subnet found`),
			},
		},
	})
}

// -----------------------------------------------------------------------------
// Acceptance test helpers.
// -----------------------------------------------------------------------------

func testAccNetworkingSubnetV2Handle(s *terraform.State) (*networkingNetworkV2TestHandle, error) {
	rs, ok := s.RootModule().Resources[testAccNetworkingSubnetV2ResourceName]
	if !ok {
		return nil, fmt.Errorf("not found: %s", testAccNetworkingSubnetV2ResourceName)
	}

	if rs.Primary.ID == "" {
		return nil, errors.New("no ID is set")
	}

	client, err := networkingV2Client(
		testAccProvider.Meta(),
		rs.Primary.Attributes["project_id"],
		rs.Primary.Attributes["region"],
	)
	if err != nil {
		return nil, err
	}

	return &networkingNetworkV2TestHandle{client: client, id: rs.Primary.ID}, nil
}

func testAccCheckNetworkingSubnetV2Exists(_ string, found *subnet.Subnet) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		handle, err := testAccNetworkingSubnetV2Handle(s)
		if err != nil {
			return err
		}

		result, err := subnet.Get(context.Background(), handle.client, handle.id)
		if err != nil {
			return err
		}

		if result.ID != handle.id {
			return errors.New("networking subnet not found")
		}

		*found = *result

		return nil
	}
}

func testAccCheckNetworkingSubnetV2SameID(resourceName string, found *subnet.Subnet) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("not found: %s", resourceName)
		}

		if rs.Primary.ID != found.ID {
			return fmt.Errorf("networking subnet was replaced: %s -> %s", found.ID, rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckNetworkingSubnetV2DifferentID(resourceName string, found *subnet.Subnet) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("not found: %s", resourceName)
		}

		if rs.Primary.ID == found.ID {
			return errors.New("networking subnet was updated in place instead of being replaced")
		}

		found.ID = rs.Primary.ID

		return nil
	}
}

func testAccCheckNetworkingSubnetV2RenameOutOfBand(found *subnet.Subnet) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		handle, err := testAccNetworkingSubnetV2Handle(s)
		if err != nil {
			return err
		}

		driftedName := found.Name + "-drifted"
		_, err = subnet.Update(context.Background(), handle.client, handle.id, subnet.UpdateRequest{
			Name: &driftedName,
		})

		return err
	}
}

func testAccCheckNetworkingSubnetV2DeleteOutOfBand() resource.TestCheckFunc {
	return func(s *terraform.State) error {
		handle, err := testAccNetworkingSubnetV2Handle(s)
		if err != nil {
			return err
		}

		return subnet.Delete(context.Background(), handle.client, handle.id)
	}
}

func testAccCheckNetworkingSubnetV2Destroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "selectel_networking_subnet_v2" {
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

		if _, err := subnet.Get(context.Background(), client, rs.Primary.ID); err == nil {
			return errors.New("networking subnet still exists")
		} else if !networkingIsNotFound(err) {
			return err
		}
	}

	return nil
}

// -----------------------------------------------------------------------------
// Acceptance test configurations.
// -----------------------------------------------------------------------------

func testAccNetworkingSubnetV2Network(projectID, region string) string {
	return testAccNetworkingV2Base(projectID) + fmt.Sprintf(`
resource "selectel_networking_network_v2" "network_tf_acc_test_1" {
  project_id = local.project_id
  region     = "%s"
  name       = "tf-acc-subnet-parent"
}
`, region)
}

func testAccNetworkingSubnetV2Basic(projectID, region, cidr string) string {
	return testAccNetworkingSubnetV2Network(projectID, region) + fmt.Sprintf(`
resource "selectel_networking_subnet_v2" "subnet_tf_acc_test_1" {
  project_id = local.project_id
  region     = "%s"
  network_id = selectel_networking_network_v2.network_tf_acc_test_1.id
  cidr       = "%s"
}
`, region, cidr)
}

func testAccNetworkingSubnetV2NoGateway(projectID, region string) string {
	return testAccNetworkingSubnetV2Network(projectID, region) + fmt.Sprintf(`
resource "selectel_networking_subnet_v2" "subnet_tf_acc_test_1" {
  project_id = local.project_id
  region     = "%s"
  network_id = selectel_networking_network_v2.network_tf_acc_test_1.id
  cidr       = "192.168.199.0/24"
  no_gateway = true
}
`, region)
}

func testAccNetworkingSubnetV2OptionalAttributes(projectID, region, subnetName string) string {
	return testAccNetworkingSubnetV2Network(projectID, region) + fmt.Sprintf(`
resource "selectel_networking_subnet_v2" "subnet_tf_acc_test_1" {
  project_id           = local.project_id
  region               = "%s"
  network_id           = selectel_networking_network_v2.network_tf_acc_test_1.id
  cidr                 = "192.168.199.0/24"
  name                 = "%s"
  description          = "created by the acceptance test"
  gateway_ip           = "192.168.199.1"
  enable_dhcp          = false
  dns_nameservers      = ["8.8.8.8", "1.1.1.1"]
  tags                 = ["tf-acc"]

  allocation_pool {
    start = "192.168.199.10"
    end   = "192.168.199.20"
  }
}
`, region, subnetName)
}

func testAccNetworkingSubnetV2Updated(projectID, region, subnetName string) string {
	return testAccNetworkingSubnetV2Network(projectID, region) + fmt.Sprintf(`
resource "selectel_networking_subnet_v2" "subnet_tf_acc_test_1" {
  project_id           = local.project_id
  region               = "%s"
  network_id           = selectel_networking_network_v2.network_tf_acc_test_1.id
  cidr                 = "192.168.199.0/24"
  name                 = "%s"
  description          = "updated by the acceptance test"
  gateway_ip           = "192.168.199.2"
  enable_dhcp          = true
  dns_nameservers      = ["8.8.8.8"]
  tags                 = ["tf-acc", "updated"]

  allocation_pool {
    start = "192.168.199.30"
    end   = "192.168.199.40"
  }
}
`, region, subnetName)
}

func testAccNetworkingSubnetV2DataSource(projectID, region, subnetName string) string {
	return testAccNetworkingSubnetV2OptionalAttributes(projectID, region, subnetName) + fmt.Sprintf(`
data "selectel_networking_subnet_v2" "subnet_tf_acc_test_1" {
  project_id = local.project_id
  region     = "%s"
  name       = selectel_networking_subnet_v2.subnet_tf_acc_test_1.name
}
`, region)
}

func testAccNetworkingSubnetV2DataSourceNoMatch(projectID, region string) string {
	return testAccNetworkingSubnetV2Network(projectID, region) + fmt.Sprintf(`
data "selectel_networking_subnet_v2" "subnet_tf_acc_test_1" {
  project_id = local.project_id
  region     = "%s"
  name       = "tf-acc-no-such-subnet"
}
`, region)
}
