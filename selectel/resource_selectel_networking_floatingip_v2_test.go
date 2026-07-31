package selectel

import (
	"context"
	"errors"
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/selectel/vpc-go/pkg/v2/floatingip"
	"github.com/selectel/vpc-go/pkg/v2/network"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testAccNetworkingFloatingIPV2ResourceName   = "selectel_networking_floatingip_v2.floatingip_tf_acc_test_1"
	testAccNetworkingFloatingIPV2DataSourceName = "data.selectel_networking_floatingip_v2.floatingip_tf_acc_test_1"
)

// networkingFloatingIPV2TestPool is the external network the acceptance tests
// allocate addresses from. It is region-specific, so the test skips when it is
// not configured rather than silently allocating from the wrong pool.
func networkingFloatingIPV2TestPool(t *testing.T) string {
	t.Helper()

	pool := os.Getenv("SELECTEL_NETWORKING_FLOATINGIP_POOL")
	if pool == "" {
		t.Skip("SELECTEL_NETWORKING_FLOATINGIP_POOL must be set to the name of an external network")
	}

	return pool
}

func networkingFloatingIPV2OptionalTestValues(t *testing.T) (string, string, string) {
	t.Helper()

	subnetID := os.Getenv("SELECTEL_NETWORKING_FLOATINGIP_SUBNET_ID")
	dnsName := os.Getenv("SELECTEL_NETWORKING_FLOATINGIP_DNS_NAME")
	dnsDomain := os.Getenv("SELECTEL_NETWORKING_FLOATINGIP_DNS_DOMAIN")
	if subnetID == "" || dnsName == "" || dnsDomain == "" {
		t.Skip(
			"SELECTEL_NETWORKING_FLOATINGIP_SUBNET_ID, " +
				"SELECTEL_NETWORKING_FLOATINGIP_DNS_NAME and " +
				"SELECTEL_NETWORKING_FLOATINGIP_DNS_DOMAIN must be set",
		)
	}

	return subnetID, dnsName, dnsDomain
}

// -----------------------------------------------------------------------------
// Schema unit tests.
// -----------------------------------------------------------------------------

// TestNetworkingFloatingIPV2Schema covers REQ-085, REQ-086, REQ-087, REQ-089 and
// REQ-095: which attributes are arguments, which replace the address, and which
// are read-only.
func TestNetworkingFloatingIPV2Schema(t *testing.T) {
	s := resourceNetworkingFloatingIPV2().Schema

	t.Run("project_id, region and pool are required and ForceNew", func(t *testing.T) {
		for _, name := range []string{"project_id", "region", "pool"} {
			require.Contains(t, s, name)
			assert.True(t, s[name].Required)
			assert.True(t, s[name].ForceNew)
		}
	})

	// REQ-089: exactly these attributes replace the floating IP. address is not
	// among them because it cannot be requested at all.
	t.Run("only the immutable attributes are ForceNew", func(t *testing.T) {
		forceNew := make([]string, 0)
		for name, attr := range s {
			if attr.ForceNew {
				forceNew = append(forceNew, name)
			}
		}

		assert.ElementsMatch(t, []string{
			"project_id", "region", "pool", "subnet_id", "dns_name", "dns_domain",
		}, forceNew)
	})

	// REQ-087: the association is updated in place.
	t.Run("port_id and fixed_ip are updatable", func(t *testing.T) {
		for _, name := range []string{"port_id", "fixed_ip"} {
			require.Contains(t, s, name)
			assert.True(t, s[name].Optional)
			assert.True(t, s[name].Computed)
			assert.False(t, s[name].ForceNew, name+" must be updatable in place")
		}
	})

	// REQ-086: the API always assigns the address, and REQ-095: blocked is
	// admin-managed.
	for _, name := range []string{
		"address", "floating_network_id", "router_id", "status", "qos_policy_id", "blocked", "all_tags",
	} {
		t.Run(name+" is computed-only", func(t *testing.T) {
			require.Contains(t, s, name)
			assert.True(t, s[name].Computed)
			assert.False(t, s[name].Optional, name+" must not be a configuration argument")
			assert.False(t, s[name].Required)
		})
	}

	// subnet_ids is not part of the schema: the API takes a single subnet.
	t.Run("subnet_ids is not exposed", func(t *testing.T) {
		assert.NotContains(t, s, "subnet_ids")
	})
}

// TestNetworkingFloatingIPV2DNSDomainValidation covers the FQDN rule of REQ-086.
func TestNetworkingFloatingIPV2DNSDomainValidation(t *testing.T) {
	validate := resourceNetworkingFloatingIPV2().Schema["dns_domain"].ValidateFunc
	require.NotNil(t, validate)

	t.Run("rejects a domain without a trailing dot", func(t *testing.T) {
		_, errs := validate("example.com", "dns_domain")

		require.NotEmpty(t, errs)
		assert.Contains(t, errs[0].Error(), "dot at the end")
	})

	for _, value := range []string{"", "example.com."} {
		t.Run("accepts "+value, func(t *testing.T) {
			_, errs := validate(value, "dns_domain")
			assert.Empty(t, errs)
		})
	}
}

// TestNetworkingFloatingIPV2SubnetIDDiffSuppression covers SCN-089-04: after an
// import the stored subnet_id is empty, because the API never returns it
// (REQ-086). REQ-089 scopes the replacement rule to a change between two known
// values, so an unknown stored value must not plan a replacement — that would
// release an already allocated address over something the provider could not read.
func TestNetworkingFloatingIPV2SubnetIDDiffSuppression(t *testing.T) {
	cases := []struct {
		name       string
		stored     string
		configured string
		suppressed bool
	}{
		{
			name:       "unknown stored value against a configured subnet is suppressed",
			stored:     "",
			configured: "subnet-id",
			suppressed: true,
		},
		{
			name:       "a real change between two known values still replaces",
			stored:     "subnet-a",
			configured: "subnet-b",
			suppressed: false,
		},
		{
			name:       "identical values",
			stored:     "subnet-a",
			configured: "subnet-a",
			suppressed: false,
		},
		{
			name:       "clearing a known value is a change",
			stored:     "subnet-a",
			configured: "",
			suppressed: false,
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			assert.Equal(t, testCase.suppressed,
				suppressNetworkingFloatingIPV2SubnetIDDiff(
					"subnet_id", testCase.stored, testCase.configured, nil))
		})
	}

	// The attribute must stay out of the computed set: there is nothing to fill it
	// from, so Computed would only mask the fact that the value is unknown.
	t.Run("subnet_id is not computed", func(t *testing.T) {
		s := resourceNetworkingFloatingIPV2().Schema
		require.Contains(t, s, "subnet_id")
		assert.False(t, s["subnet_id"].Computed)
		assert.True(t, s["subnet_id"].ForceNew)
	})
}

// TestNetworkingFloatingIPV2DataSourceSelectors covers REQ-094.
func TestNetworkingFloatingIPV2DataSourceSelectors(t *testing.T) {
	s := dataSourceNetworkingFloatingIPV2().Schema

	selectors := make([]string, 0)
	for name, attr := range s {
		if attr.Optional || attr.Required {
			selectors = append(selectors, name)
		}
	}

	assert.ElementsMatch(t, []string{
		"project_id", "region", "floatingip_id", "address", "description",
		"pool", "port_id", "fixed_ip", "status", "tags",
	}, selectors)

	for _, name := range []string{"blocked", "qos_policy_id", "router_id", "floating_network_id", "all_tags"} {
		require.Contains(t, s, name)
		assert.True(t, s[name].Computed)
		assert.False(t, s[name].Optional, name+" must be returned, not selectable")
	}
}

func TestNetworkingFloatingIPV2PoolID(t *testing.T) {
	t.Run("not found", func(t *testing.T) {
		_, err := networkingFloatingIPV2PoolID(nil, "public")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "does not match any external network")
	})

	t.Run("unique", func(t *testing.T) {
		id, err := networkingFloatingIPV2PoolID([]network.Network{{ID: "network-id"}}, "public")
		require.NoError(t, err)
		assert.Equal(t, "network-id", id)
	})

	t.Run("ambiguous", func(t *testing.T) {
		_, err := networkingFloatingIPV2PoolID(
			[]network.Network{{ID: "one"}, {ID: "two"}}, "public",
		)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "matches 2 external networks")
	})
}

func TestNetworkingFloatingIPV2DataSourceResult(t *testing.T) {
	t.Run("not found", func(t *testing.T) {
		_, err := networkingFloatingIPV2DataSourceResult(nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "no networking floating ip found")
	})

	t.Run("unique", func(t *testing.T) {
		result, err := networkingFloatingIPV2DataSourceResult(
			[]floatingip.FloatingIP{{ID: "floating-ip-id"}},
		)
		require.NoError(t, err)
		assert.Equal(t, "floating-ip-id", result.ID)
	})

	t.Run("ambiguous", func(t *testing.T) {
		_, err := networkingFloatingIPV2DataSourceResult(
			[]floatingip.FloatingIP{{ID: "one"}, {ID: "two"}},
		)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "2 networking floating ips found")
	})
}

// TestNetworkingFloatingIPV2Flatten covers the state side of SCN-090-01 and
// SCN-095-02: mutable API drift, including tags, and the read-only blocked
// status must be reflected by Read.
func TestNetworkingFloatingIPV2Flatten(t *testing.T) {
	d := schema.TestResourceDataRaw(t, resourceNetworkingFloatingIPV2().Schema, map[string]interface{}{
		"project_id": "project-id",
		"region":     "ru-3",
		"pool":       "external-network",
		"tags":       []interface{}{"configured"},
	})

	portID := "port-id"
	fixedIP := "192.0.2.10"
	flattenNetworkingFloatingIPV2Resource(d, &floatingip.FloatingIP{
		ID:                "floating-ip-id",
		FloatingNetworkID: "external-network-id",
		FloatingIPAddress: "203.0.113.10",
		PortID:            &portID,
		FixedIPAddress:    &fixedIP,
		Description:       "drifted outside terraform",
		Status:            "ACTIVE",
		Tags:              []string{"actual", "external"},
		Blocked:           true,
	})

	assert.Equal(t, "203.0.113.10", d.Get("address"))
	assert.Equal(t, portID, d.Get("port_id"))
	assert.Equal(t, fixedIP, d.Get("fixed_ip"))
	assert.Equal(t, "drifted outside terraform", d.Get("description"))
	assert.True(t, d.Get("blocked").(bool))
	assert.ElementsMatch(t, []interface{}{"actual", "external"}, d.Get("tags").(*schema.Set).List())
	assert.ElementsMatch(t, []interface{}{"actual", "external"}, d.Get("all_tags").(*schema.Set).List())
}

func TestNetworkingFloatingIPV2DataSourceFlattenPreservesTagSelector(t *testing.T) {
	d := schema.TestResourceDataRaw(t, dataSourceNetworkingFloatingIPV2().Schema, map[string]interface{}{
		"project_id": "project-id",
		"region":     "ru-3",
		"tags":       []interface{}{"selector"},
	})

	flattenNetworkingFloatingIPV2(d, &floatingip.FloatingIP{
		Tags: []string{"selector", "external"},
	})

	assert.ElementsMatch(t, []interface{}{"selector"}, d.Get("tags").(*schema.Set).List())
	assert.ElementsMatch(t, []interface{}{"selector", "external"}, d.Get("all_tags").(*schema.Set).List())
}

// -----------------------------------------------------------------------------
// Acceptance tests against a staging environment (TF_ACC=1).
// -----------------------------------------------------------------------------

// TestAccNetworkingFloatingIPV2Basic covers SCN-085-01, SCN-085-02 (empty
// follow-up plan), SCN-093-01 (import) and SCN-092-01 (delete).
func TestAccNetworkingFloatingIPV2Basic(t *testing.T) {
	var found floatingip.FloatingIP
	projectName := acctest.RandomWithPrefix("tf-acc")
	pool := networkingFloatingIPV2TestPool(t)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckNetworkingFloatingIPV2Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkingFloatingIPV2Basic(projectName, pool),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNetworkingFloatingIPV2Exists(&found),
					// The API assigns the address and resolves the pool.
					resource.TestCheckResourceAttrSet(testAccNetworkingFloatingIPV2ResourceName, "address"),
					resource.TestCheckResourceAttrSet(testAccNetworkingFloatingIPV2ResourceName, "floating_network_id"),
					resource.TestCheckResourceAttrSet(testAccNetworkingFloatingIPV2ResourceName, "status"),
					resource.TestCheckResourceAttrSet(testAccNetworkingFloatingIPV2ResourceName, "blocked"),
				),
			},
			{
				Config:   testAccNetworkingFloatingIPV2Basic(projectName, pool),
				PlanOnly: true,
			},
			{
				ResourceName:      testAccNetworkingFloatingIPV2ResourceName,
				ImportState:       true,
				ImportStateVerify: true,
				// pool is a name the provider resolves; the API returns only the
				// resolved identifier.
				ImportStateVerifyIgnore: []string{"pool"},
			},
		},
	})
}

// TestAccNetworkingFloatingIPV2Optional covers SCN-086-01, SCN-086-02,
// SCN-086-04 and SCN-086-05. SCN-086-06 is covered by the schema test because
// it must fail before any API request.
func TestAccNetworkingFloatingIPV2Optional(t *testing.T) {
	projectName := acctest.RandomWithPrefix("tf-acc")
	pool := networkingFloatingIPV2TestPool(t)
	subnetID, dnsName, dnsDomain := networkingFloatingIPV2OptionalTestValues(t)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckNetworkingFloatingIPV2Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkingFloatingIPV2Optional(
					projectName, pool, subnetID, dnsName, dnsDomain,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(
						testAccNetworkingFloatingIPV2ResourceName, "address",
					),
					resource.TestCheckResourceAttr(
						testAccNetworkingFloatingIPV2ResourceName, "subnet_id", subnetID,
					),
					resource.TestCheckResourceAttr(
						testAccNetworkingFloatingIPV2ResourceName, "description", "optional attributes",
					),
					resource.TestCheckResourceAttr(
						testAccNetworkingFloatingIPV2ResourceName, "tags.#", "2",
					),
					resource.TestCheckResourceAttr(
						testAccNetworkingFloatingIPV2ResourceName, "all_tags.#", "2",
					),
					resource.TestCheckResourceAttr(
						testAccNetworkingFloatingIPV2ResourceName, "dns_name", dnsName,
					),
					resource.TestCheckResourceAttr(
						testAccNetworkingFloatingIPV2ResourceName, "dns_domain", dnsDomain,
					),
				),
			},
			{
				Config: testAccNetworkingFloatingIPV2Optional(
					projectName, pool, subnetID, dnsName, dnsDomain,
				),
				PlanOnly: true,
			},
		},
	})
}

// TestAccNetworkingFloatingIPV2UnknownPool covers SCN-091-02: an unknown pool
// fails before any floating IP is created.
func TestAccNetworkingFloatingIPV2UnknownPool(t *testing.T) {
	projectName := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckNetworkingFloatingIPV2Destroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccNetworkingFloatingIPV2Basic(projectName, "tf-acc-no-such-pool"),
				ExpectError: regexp.MustCompile(`does not match any external network`),
			},
		},
	})
}

// TestAccNetworkingFloatingIPV2Update covers SCN-088-01 (description and tags).
func TestAccNetworkingFloatingIPV2Update(t *testing.T) {
	var found floatingip.FloatingIP
	projectName := acctest.RandomWithPrefix("tf-acc")
	pool := networkingFloatingIPV2TestPool(t)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckNetworkingFloatingIPV2Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkingFloatingIPV2Described(projectName, pool, "created by the acceptance test"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNetworkingFloatingIPV2Exists(&found),
					resource.TestCheckResourceAttr(testAccNetworkingFloatingIPV2ResourceName, "tags.#", "2"),
				),
			},
			{
				Config: testAccNetworkingFloatingIPV2Described(projectName, pool, "updated by the acceptance test"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNetworkingFloatingIPV2SameID(&found),
					resource.TestCheckResourceAttr(testAccNetworkingFloatingIPV2ResourceName, "description", "updated by the acceptance test"),
				),
			},
		},
	})
}

// TestAccNetworkingFloatingIPV2Association covers SCN-087-01 and SCN-087-02:
// Create associates the address with the selected fixed IP, and Update moves it
// to another port without replacing the floating IP.
func TestAccNetworkingFloatingIPV2Association(t *testing.T) {
	var found floatingip.FloatingIP
	projectName := acctest.RandomWithPrefix("tf-acc")
	pool := networkingFloatingIPV2TestPool(t)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckNetworkingFloatingIPV2Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkingFloatingIPV2Associated(projectName, pool, "port_a", "192.168.199.70"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNetworkingFloatingIPV2Exists(&found),
					resource.TestCheckResourceAttrPair(
						testAccNetworkingFloatingIPV2ResourceName, "port_id",
						"selectel_networking_port_v2.port_a", "id",
					),
					resource.TestCheckResourceAttr(
						testAccNetworkingFloatingIPV2ResourceName, "fixed_ip", "192.168.199.70",
					),
				),
			},
			{
				Config: testAccNetworkingFloatingIPV2Associated(projectName, pool, "port_b", "192.168.199.71"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNetworkingFloatingIPV2SameID(&found),
					resource.TestCheckResourceAttrPair(
						testAccNetworkingFloatingIPV2ResourceName, "port_id",
						"selectel_networking_port_v2.port_b", "id",
					),
					resource.TestCheckResourceAttr(
						testAccNetworkingFloatingIPV2ResourceName, "fixed_ip", "192.168.199.71",
					),
				),
			},
			{
				Config:   testAccNetworkingFloatingIPV2Associated(projectName, pool, "port_b", "192.168.199.71"),
				PlanOnly: true,
			},
		},
	})
}

// TestAccNetworkingFloatingIPV2ExternalChanges covers SCN-090-01, SCN-090-02 and
// SCN-092-02.
func TestAccNetworkingFloatingIPV2ExternalChanges(t *testing.T) {
	var found floatingip.FloatingIP
	projectName := acctest.RandomWithPrefix("tf-acc")
	pool := networkingFloatingIPV2TestPool(t)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckNetworkingFloatingIPV2Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkingFloatingIPV2Described(projectName, pool, "created by the acceptance test"),
				Check:  testAccCheckNetworkingFloatingIPV2Exists(&found),
			},
			{
				Config:             testAccNetworkingFloatingIPV2Described(projectName, pool, "created by the acceptance test"),
				Check:              testAccCheckNetworkingFloatingIPV2DescribeOutOfBand(),
				ExpectNonEmptyPlan: true,
			},
			{
				Config:             testAccNetworkingFloatingIPV2Described(projectName, pool, "created by the acceptance test"),
				Check:              testAccCheckNetworkingFloatingIPV2DeleteOutOfBand(),
				ExpectNonEmptyPlan: true,
			},
			{
				Config: testAccNetworkingFloatingIPV2Described(projectName, pool, "created by the acceptance test"),
				Check:  testAccCheckNetworkingFloatingIPV2Exists(&found),
			},
		},
	})
}

// TestAccNetworkingFloatingIPV2DataSource covers SCN-094-01 … SCN-094-03.
func TestAccNetworkingFloatingIPV2DataSource(t *testing.T) {
	projectName := acctest.RandomWithPrefix("tf-acc")
	pool := networkingFloatingIPV2TestPool(t)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckNetworkingFloatingIPV2Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkingFloatingIPV2DataSource(projectName, pool),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(
						testAccNetworkingFloatingIPV2DataSourceName, "floatingip_id",
						testAccNetworkingFloatingIPV2ResourceName, "id",
					),
					resource.TestCheckResourceAttrPair(
						testAccNetworkingFloatingIPV2DataSourceName, "blocked",
						testAccNetworkingFloatingIPV2ResourceName, "blocked",
					),
				),
			},
			{
				Config:      testAccNetworkingFloatingIPV2DataSourceNoMatch(projectName),
				ExpectError: regexp.MustCompile(`no networking floating ip found`),
			},
		},
	})
}

// -----------------------------------------------------------------------------
// Helpers.
// -----------------------------------------------------------------------------

func testAccCheckNetworkingFloatingIPV2Exists(found *floatingip.FloatingIP) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		handle, err := testAccNetworkingNetworkV2Handle(s, testAccNetworkingFloatingIPV2ResourceName)
		if err != nil {
			return err
		}

		result, err := floatingip.Get(context.Background(), handle.client, handle.id)
		if err != nil {
			return err
		}

		*found = *result

		return nil
	}
}

func testAccCheckNetworkingFloatingIPV2SameID(found *floatingip.FloatingIP) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[testAccNetworkingFloatingIPV2ResourceName]
		if !ok {
			return fmt.Errorf("not found: %s", testAccNetworkingFloatingIPV2ResourceName)
		}

		if rs.Primary.ID != found.ID {
			return fmt.Errorf("networking floating ip was replaced: %s -> %s", found.ID, rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckNetworkingFloatingIPV2DescribeOutOfBand() resource.TestCheckFunc {
	return func(s *terraform.State) error {
		handle, err := testAccNetworkingNetworkV2Handle(s, testAccNetworkingFloatingIPV2ResourceName)
		if err != nil {
			return err
		}

		drifted := "drifted outside terraform"
		_, err = floatingip.Update(context.Background(), handle.client, handle.id, floatingip.UpdateRequest{
			Description: &drifted,
		})

		return err
	}
}

func testAccCheckNetworkingFloatingIPV2DeleteOutOfBand() resource.TestCheckFunc {
	return func(s *terraform.State) error {
		handle, err := testAccNetworkingNetworkV2Handle(s, testAccNetworkingFloatingIPV2ResourceName)
		if err != nil {
			return err
		}

		return floatingip.Delete(context.Background(), handle.client, handle.id)
	}
}

func testAccCheckNetworkingFloatingIPV2Destroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "selectel_networking_floatingip_v2" {
			continue
		}

		client, err := networkingV2Client(
			testAccProvider.Meta(),
			rs.Primary.Attributes["project_id"],
			rs.Primary.Attributes["region"],
		)
		if err != nil {
			return nil //nolint:nilerr // the project is gone together with the address
		}

		if _, err := floatingip.Get(context.Background(), client, rs.Primary.ID); err == nil {
			return errors.New("networking floating ip still exists")
		} else if !networkingIsNotFound(err) {
			return err
		}
	}

	return nil
}

// -----------------------------------------------------------------------------
// Configurations.
// -----------------------------------------------------------------------------

func testAccNetworkingFloatingIPV2Basic(projectName, pool string) string {
	return testAccNetworkingNetworkV2Project(projectName) + fmt.Sprintf(`
resource "selectel_networking_floatingip_v2" "floatingip_tf_acc_test_1" {
  project_id = selectel_vpc_project_v2.project_tf_acc_test_1.id
  region     = "%s"
  pool       = "%s"
}
`, networkingV2TestRegion(), pool)
}

func testAccNetworkingFloatingIPV2Described(projectName, pool, description string) string {
	return testAccNetworkingNetworkV2Project(projectName) + fmt.Sprintf(`
resource "selectel_networking_floatingip_v2" "floatingip_tf_acc_test_1" {
  project_id  = selectel_vpc_project_v2.project_tf_acc_test_1.id
  region      = "%s"
  pool        = "%s"
  description = "%s"
  tags        = ["tf-acc", "networking"]
}
`, networkingV2TestRegion(), pool, description)
}

func testAccNetworkingFloatingIPV2Optional(
	projectName, pool, subnetID, dnsName, dnsDomain string,
) string {
	return testAccNetworkingNetworkV2Project(projectName) + fmt.Sprintf(`
resource "selectel_networking_floatingip_v2" "floatingip_tf_acc_test_1" {
  project_id  = selectel_vpc_project_v2.project_tf_acc_test_1.id
  region      = "%s"
  pool        = "%s"
  subnet_id   = "%s"
  description = "optional attributes"
  tags        = ["tf-acc", "networking"]
  dns_name    = "%s"
  dns_domain  = "%s"
}
`, networkingV2TestRegion(), pool, subnetID, dnsName, dnsDomain)
}

func testAccNetworkingFloatingIPV2Associated(projectName, pool, selectedPort, fixedIP string) string {
	return testAccNetworkingPortV2Subnet(projectName) + fmt.Sprintf(`
resource "selectel_networking_port_v2" "port_a" {
  project_id = selectel_vpc_project_v2.project_tf_acc_test_1.id
  region     = "%s"
  network_id = selectel_networking_network_v2.network_tf_acc_test_1.id

  fixed_ip {
    subnet_id  = selectel_networking_subnet_v2.subnet_tf_acc_test_1.id
    ip_address = "192.168.199.70"
  }
}

resource "selectel_networking_port_v2" "port_b" {
  project_id = selectel_vpc_project_v2.project_tf_acc_test_1.id
  region     = "%s"
  network_id = selectel_networking_network_v2.network_tf_acc_test_1.id

  fixed_ip {
    subnet_id  = selectel_networking_subnet_v2.subnet_tf_acc_test_1.id
    ip_address = "192.168.199.71"
  }
}

resource "selectel_networking_floatingip_v2" "floatingip_tf_acc_test_1" {
  project_id  = selectel_vpc_project_v2.project_tf_acc_test_1.id
  region      = "%s"
  pool        = "%s"
  port_id     = selectel_networking_port_v2.%s.id
  fixed_ip    = "%s"
}
`, networkingV2TestRegion(), networkingV2TestRegion(), networkingV2TestRegion(), pool, selectedPort, fixedIP)
}

func testAccNetworkingFloatingIPV2DataSource(projectName, pool string) string {
	return testAccNetworkingFloatingIPV2Basic(projectName, pool) + fmt.Sprintf(`
data "selectel_networking_floatingip_v2" "floatingip_tf_acc_test_1" {
  project_id = selectel_vpc_project_v2.project_tf_acc_test_1.id
  region     = "%s"
  address    = selectel_networking_floatingip_v2.floatingip_tf_acc_test_1.address
}
`, networkingV2TestRegion())
}

func testAccNetworkingFloatingIPV2DataSourceNoMatch(projectName string) string {
	return testAccNetworkingNetworkV2Project(projectName) + fmt.Sprintf(`
data "selectel_networking_floatingip_v2" "floatingip_tf_acc_test_1" {
  project_id = selectel_vpc_project_v2.project_tf_acc_test_1.id
  region     = "%s"
  address    = "203.0.113.254"
}
`, networkingV2TestRegion())
}
