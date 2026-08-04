package selectel

import (
	"context"
	"errors"
	"fmt"
	"os"
	"regexp"
	"strconv"
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

// networkingFloatingIPV2OptionalTestValues is the expression naming the subnet an
// address is allocated from, plus the DNS name recorded with it.
//
// The subnet defaults to the first one of the external network, taken from the
// data source: a project cannot read those subnets, but it can see their
// identifiers. SELECTEL_NETWORKING_FLOATINGIP_SUBNET_ID overrides it, because an
// external network may hold subnets no address can be allocated from.
//
// The domain is not here at all — it comes from a private DNS zone the
// configuration creates, since the fork resolves dns_domain against the zones of
// the project.
func networkingFloatingIPV2OptionalTestValues() (subnetID, dnsName string) {
	subnetID = "data.selectel_networking_network_v2.external_tf_acc_test_1.subnets[0]"
	if configured := os.Getenv("SELECTEL_NETWORKING_FLOATINGIP_SUBNET_ID"); configured != "" {
		subnetID = strconv.Quote(configured)
	}

	return subnetID, acctest.RandomWithPrefix("tf-acc-fip")
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

// TestNetworkingFloatingIPV2WriteOnlyDNSFields covers the same rule for dns_name
// and dns_domain. A live floating IP created with both returns neither in the
// create response nor in a following read, so they cannot be Computed and a
// configured value must not be read as a change from an empty stored one —
// otherwise every plan would release and re-allocate the address.
func TestNetworkingFloatingIPV2WriteOnlyDNSFields(t *testing.T) {
	s := resourceNetworkingFloatingIPV2().Schema

	for _, name := range []string{"dns_name", "dns_domain"} {
		t.Run(name+" is write-only and suppressed against an unknown stored value", func(t *testing.T) {
			require.Contains(t, s, name)
			assert.True(t, s[name].Optional)
			assert.False(t, s[name].Computed, "the API returns nothing to populate it from")
			assert.True(t, s[name].ForceNew)
			require.NotNil(t, s[name].DiffSuppressFunc)

			assert.True(t, s[name].DiffSuppressFunc(name, "", "configured", nil),
				"an unknown stored value must not plan a replacement")
			assert.False(t, s[name].DiffSuppressFunc(name, "stored", "configured", nil),
				"a change between two known values is a real change")
			assert.False(t, s[name].DiffSuppressFunc(name, "stored", "", nil),
				"clearing a known value is a real change")
		})
	}
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
	projectID, region := networkingV2TestEnv()
	pool := networkingFloatingIPV2TestPool(t)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheckWithProjectIDAndRegion(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckNetworkingFloatingIPV2Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkingFloatingIPV2Basic(projectID, region, pool),
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
				Config:   testAccNetworkingFloatingIPV2Basic(projectID, region, pool),
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
	projectID, region := networkingV2TestEnv()
	pool := networkingFloatingIPV2TestPool(t)
	zonePrefix := acctest.RandomWithPrefix("tf-acc-zone")
	subnetID, dnsName := networkingFloatingIPV2OptionalTestValues()

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheckWithProjectIDAndRegion(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckNetworkingFloatingIPV2Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkingFloatingIPV2Optional(
					projectID, region, pool, zonePrefix, subnetID, dnsName,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(
						testAccNetworkingFloatingIPV2ResourceName, "address",
					),
					// subnet_id, dns_name and dns_domain are deliberately not asserted
					// here: the API returns none of them, and their diff suppression
					// keeps a configured value out of state entirely. The empty
					// follow-up plan of the next step is what shows they reached the
					// create request without planning a replacement — the behaviour
					// that was broken before the suppression existed.
					resource.TestCheckResourceAttr(
						testAccNetworkingFloatingIPV2ResourceName, "description", "optional attributes",
					),
					resource.TestCheckResourceAttr(
						testAccNetworkingFloatingIPV2ResourceName, "tags.#", "2",
					),
					resource.TestCheckResourceAttr(
						testAccNetworkingFloatingIPV2ResourceName, "all_tags.#", "2",
					),
				),
			},
			{
				Config: testAccNetworkingFloatingIPV2Optional(
					projectID, region, pool, zonePrefix, subnetID, dnsName,
				),
				PlanOnly: true,
			},
		},
	})
}

// TestAccNetworkingFloatingIPV2UnknownPool covers SCN-091-02: an unknown pool
// fails before any floating IP is created.
func TestAccNetworkingFloatingIPV2UnknownPool(t *testing.T) {
	projectID, region := networkingV2TestEnv()

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheckWithProjectIDAndRegion(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckNetworkingFloatingIPV2Destroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccNetworkingFloatingIPV2Basic(projectID, region, "tf-acc-no-such-pool"),
				ExpectError: regexp.MustCompile(`does not match any external network`),
			},
		},
	})
}

// TestAccNetworkingFloatingIPV2Update covers SCN-088-01 (description and tags).
func TestAccNetworkingFloatingIPV2Update(t *testing.T) {
	var found floatingip.FloatingIP
	projectID, region := networkingV2TestEnv()
	pool := networkingFloatingIPV2TestPool(t)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheckWithProjectIDAndRegion(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckNetworkingFloatingIPV2Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkingFloatingIPV2Described(projectID, region, pool, "created by the acceptance test"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNetworkingFloatingIPV2Exists(&found),
					resource.TestCheckResourceAttr(testAccNetworkingFloatingIPV2ResourceName, "tags.#", "2"),
				),
			},
			{
				Config: testAccNetworkingFloatingIPV2Described(projectID, region, pool, "updated by the acceptance test"),
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
	projectID, region := networkingV2TestEnv()
	pool := networkingFloatingIPV2TestPool(t)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheckWithProjectIDAndRegion(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckNetworkingFloatingIPV2Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkingFloatingIPV2Associated(projectID, region, pool, "port_a", "192.168.199.70"),
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
				Config: testAccNetworkingFloatingIPV2Associated(projectID, region, pool, "port_b", "192.168.199.71"),
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
				Config:   testAccNetworkingFloatingIPV2Associated(projectID, region, pool, "port_b", "192.168.199.71"),
				PlanOnly: true,
			},
		},
	})
}

// TestAccNetworkingFloatingIPV2ExternalChanges covers SCN-090-01, SCN-090-02 and
// SCN-092-02.
func TestAccNetworkingFloatingIPV2ExternalChanges(t *testing.T) {
	var found floatingip.FloatingIP
	projectID, region := networkingV2TestEnv()
	pool := networkingFloatingIPV2TestPool(t)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheckWithProjectIDAndRegion(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckNetworkingFloatingIPV2Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkingFloatingIPV2Described(projectID, region, pool, "created by the acceptance test"),
				Check:  testAccCheckNetworkingFloatingIPV2Exists(&found),
			},
			{
				Config:             testAccNetworkingFloatingIPV2Described(projectID, region, pool, "created by the acceptance test"),
				Check:              testAccCheckNetworkingFloatingIPV2DescribeOutOfBand(),
				ExpectNonEmptyPlan: true,
			},
			{
				Config:             testAccNetworkingFloatingIPV2Described(projectID, region, pool, "created by the acceptance test"),
				Check:              testAccCheckNetworkingFloatingIPV2DeleteOutOfBand(),
				ExpectNonEmptyPlan: true,
			},
			{
				Config: testAccNetworkingFloatingIPV2Described(projectID, region, pool, "created by the acceptance test"),
				Check:  testAccCheckNetworkingFloatingIPV2Exists(&found),
			},
		},
	})
}

// TestAccNetworkingFloatingIPV2DataSource covers SCN-094-01 … SCN-094-03.
func TestAccNetworkingFloatingIPV2DataSource(t *testing.T) {
	projectID, region := networkingV2TestEnv()
	pool := networkingFloatingIPV2TestPool(t)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheckWithProjectIDAndRegion(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckNetworkingFloatingIPV2Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkingFloatingIPV2DataSource(projectID, region, pool),
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
				Config:      testAccNetworkingFloatingIPV2DataSourceNoMatch(projectID, region),
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
			return err
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

func testAccNetworkingFloatingIPV2Basic(projectID, region, pool string) string {
	return testAccNetworkingV2Base(projectID) + fmt.Sprintf(`
resource "selectel_networking_floatingip_v2" "floatingip_tf_acc_test_1" {
  project_id = local.project_id
  region     = "%s"
  pool       = "%s"
}
`, region, pool)
}

func testAccNetworkingFloatingIPV2Described(projectID, region, pool, description string) string {
	return testAccNetworkingV2Base(projectID) + fmt.Sprintf(`
resource "selectel_networking_floatingip_v2" "floatingip_tf_acc_test_1" {
  project_id  = local.project_id
  region      = "%s"
  pool        = "%s"
  description = "%s"
  tags        = ["tf-acc", "networking"]
}
`, region, pool, description)
}

func testAccNetworkingFloatingIPV2Optional(
	projectID, region, pool, zonePrefix, subnetID, dnsName string,
) string {
	return testAccNetworkingV2Base(projectID) +
		testAccNetworkingV2PrivateDNSZones(region, zonePrefix) + fmt.Sprintf(`
data "selectel_networking_network_v2" "external_tf_acc_test_1" {
  project_id = local.project_id
  region     = "%[1]s"
  name       = "%[2]s"
  external   = true
}

resource "selectel_networking_floatingip_v2" "floatingip_tf_acc_test_1" {
  project_id  = local.project_id
  region      = "%[1]s"
  pool        = "%[2]s"
  subnet_id   = %[3]s
  description = "optional attributes"
  tags        = ["tf-acc", "networking"]
  dns_name    = "%[4]s"
  dns_domain  = selectel_private_dns_zone_v1.zone_tf_acc_test_1.domain
}
`, region, pool, subnetID, dnsName)
}

// testAccNetworkingFloatingIPV2Gateway routes the test subnet to the external
// network the addresses come from.
//
// Neutron refuses to associate a floating IP with a port whose subnet has no path
// to that external network — "External network ... is not reachable from subnet
// ..." (HTTP 404) — so every configuration that associates an address needs a
// router with a gateway and an interface in the subnet. The external network is
// resolved from its name through the data source, because the pool is a name while
// the gateway takes an identifier.
func testAccNetworkingFloatingIPV2Gateway(region, pool string) string {
	return fmt.Sprintf(`
data "selectel_networking_network_v2" "external_tf_acc_test_1" {
  project_id = local.project_id
  region     = "%[1]s"
  name       = "%[2]s"
  external   = true
}

resource "selectel_networking_router_v2" "router_tf_acc_test_1" {
  project_id          = local.project_id
  region              = "%[1]s"
  external_network_id = data.selectel_networking_network_v2.external_tf_acc_test_1.network_id
}

resource "selectel_networking_router_interface_v2" "interface_tf_acc_test_1" {
  project_id = local.project_id
  region     = "%[1]s"
  router_id  = selectel_networking_router_v2.router_tf_acc_test_1.id
  subnet_id  = selectel_networking_subnet_v2.subnet_tf_acc_test_1.id
}
`, region, pool)
}

func testAccNetworkingFloatingIPV2Associated(projectID, region, pool, selectedPort, fixedIP string) string {
	return testAccNetworkingPortV2Subnet(projectID, region) +
		testAccNetworkingFloatingIPV2Gateway(region, pool) + fmt.Sprintf(`
resource "selectel_networking_port_v2" "port_a" {
  project_id = local.project_id
  region     = "%s"
  network_id = selectel_networking_network_v2.network_tf_acc_test_1.id

  fixed_ip {
    subnet_id  = selectel_networking_subnet_v2.subnet_tf_acc_test_1.id
    ip_address = "192.168.199.70"
  }
}

resource "selectel_networking_port_v2" "port_b" {
  project_id = local.project_id
  region     = "%s"
  network_id = selectel_networking_network_v2.network_tf_acc_test_1.id

  fixed_ip {
    subnet_id  = selectel_networking_subnet_v2.subnet_tf_acc_test_1.id
    ip_address = "192.168.199.71"
  }
}

resource "selectel_networking_floatingip_v2" "floatingip_tf_acc_test_1" {
  project_id  = local.project_id
  region      = "%s"
  pool        = "%s"
  port_id     = selectel_networking_port_v2.%s.id
  fixed_ip    = "%s"

  # The route to the external network is what makes the association possible, and
  # nothing else in this resource refers to it.
  depends_on = [selectel_networking_router_interface_v2.interface_tf_acc_test_1]
}
`, region, region, region, pool, selectedPort, fixedIP)
}

func testAccNetworkingFloatingIPV2DataSource(projectID, region, pool string) string {
	return testAccNetworkingFloatingIPV2Basic(projectID, region, pool) + fmt.Sprintf(`
data "selectel_networking_floatingip_v2" "floatingip_tf_acc_test_1" {
  project_id = local.project_id
  region     = "%s"
  address    = selectel_networking_floatingip_v2.floatingip_tf_acc_test_1.address
}
`, region)
}

func testAccNetworkingFloatingIPV2DataSourceNoMatch(projectID, region string) string {
	return testAccNetworkingV2Base(projectID) + fmt.Sprintf(`
data "selectel_networking_floatingip_v2" "floatingip_tf_acc_test_1" {
  project_id = local.project_id
  region     = "%s"
  address    = "203.0.113.254"
}
`, region)
}
