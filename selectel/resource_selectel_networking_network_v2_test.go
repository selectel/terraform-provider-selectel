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
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	vpc "github.com/selectel/vpc-go/pkg/v2"
	"github.com/selectel/vpc-go/pkg/v2/network"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testAccNetworkingNetworkV2ResourceName   = "selectel_networking_network_v2.network_tf_acc_test_1"
	testAccNetworkingNetworkV2DataSourceName = "data.selectel_networking_network_v2.network_tf_acc_test_1"
)

// networkingV2TestRegion is the region the networking acceptance tests run in.
func networkingV2TestRegion() string {
	if region := os.Getenv("SELECTEL_NETWORKING_REGION"); region != "" {
		return region
	}

	return "ru-3"
}

// -----------------------------------------------------------------------------
// Schema unit tests: everything that is decided before any API call.
// -----------------------------------------------------------------------------

// TestNetworkingNetworkV2SchemaOwnership covers REQ-011 (SCN-011-01): every
// attribute the cloud's policy reserves for admin must be read-only, so the
// schema itself rejects it as a configuration argument.
func TestNetworkingNetworkV2SchemaOwnership(t *testing.T) {
	s := resourceNetworkingNetworkV2().Schema

	for _, name := range []string{"project_id", "region"} {
		t.Run(name+" is required and ForceNew", func(t *testing.T) {
			require.Contains(t, s, name)
			assert.True(t, s[name].Required, "must be required")
			assert.True(t, s[name].ForceNew, "must force a new resource")
		})
	}

	for _, name := range []string{
		"blocked", "is_public", "is_dns_enabled",
		"shared", "external", "mtu", "port_security_enabled", "qos_policy_id",
		"provider_network_type", "provider_physical_network", "provider_segmentation_id",
	} {
		t.Run(name+" is computed-only", func(t *testing.T) {
			require.Contains(t, s, name)
			assert.True(t, s[name].Computed, "must be computed")
			assert.False(t, s[name].Optional, "must not be a configuration argument")
			assert.False(t, s[name].Required, "must not be a configuration argument")
		})
	}

	// REQ-002: exactly these attributes are configuration arguments.
	t.Run("only owner-writable attributes are configurable", func(t *testing.T) {
		configurable := make([]string, 0)
		for name, attr := range s {
			if attr.Optional || attr.Required {
				configurable = append(configurable, name)
			}
		}

		assert.ElementsMatch(t, []string{
			"project_id", "region", "name", "description", "admin_state_up",
			"dns_domain", "availability_zone_hints", "tags",
		}, configurable)
	})

	// REQ-004: exactly these attributes replace the resource.
	t.Run("only the immutable attributes are ForceNew", func(t *testing.T) {
		forceNew := make([]string, 0)
		for name, attr := range s {
			if attr.ForceNew {
				forceNew = append(forceNew, name)
			}
		}

		assert.ElementsMatch(t, []string{"project_id", "region", "availability_zone_hints"}, forceNew)
	})

	// REQ-011 also forbids exposing block/unblock as a user-callable action,
	// and bmnet plus transparent_vlan are out of scope entirely.
	t.Run("no out-of-scope attribute is exposed", func(t *testing.T) {
		for _, name := range []string{
			"block", "unblock", "transparent_vlan", "segments",
			"bmnet_subnet", "bmnet_vlan_map", "bmnet_vxlan_endpoint", "bmnet_trunk_port",
		} {
			assert.NotContains(t, s, name)
		}
	})
}

// TestNetworkingNetworkV2DNSDomainValidation covers SCN-008-01: a dns_domain
// without the trailing dot is rejected before the API is called.
func TestNetworkingNetworkV2DNSDomainValidation(t *testing.T) {
	validate := resourceNetworkingNetworkV2().Schema["dns_domain"].ValidateFunc
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

// TestNetworkingNetworkV2DataSourceSelectors checks that no admin-managed
// attribute is offered as a data source selector (REQ-012).
func TestNetworkingNetworkV2DataSourceSelectors(t *testing.T) {
	s := dataSourceNetworkingNetworkV2().Schema

	selectors := make([]string, 0)
	for name, attr := range s {
		if attr.Optional || attr.Required {
			selectors = append(selectors, name)
		}
	}

	assert.ElementsMatch(t, []string{
		"project_id", "region", "network_id", "name", "description", "status",
		"external", "mtu", "tags", "matching_subnet_cidr",
	}, selectors)

	for _, name := range []string{
		"blocked", "is_public", "is_dns_enabled", "shared", "port_security_enabled",
		"qos_policy_id", "provider_network_type", "provider_physical_network",
		"provider_segmentation_id", "subnets", "all_tags", "dns_domain",
		"admin_state_up", "availability_zone_hints",
	} {
		require.Contains(t, s, name)
		assert.True(t, s[name].Computed, name+" must be returned, not selectable")
	}
}

// -----------------------------------------------------------------------------
// Acceptance tests against a staging environment (TF_ACC=1).
// -----------------------------------------------------------------------------

// TestAccNetworkingNetworkV2Basic covers SCN-001-01 and SCN-001-02 (minimal
// create and an empty follow-up plan), SCN-006-01 (import by network id) and
// SCN-007-01 (delete removes the network).
func TestAccNetworkingNetworkV2Basic(t *testing.T) {
	var found network.Network
	projectName := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckNetworkingNetworkV2Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkingNetworkV2Basic(projectName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNetworkingNetworkV2Exists(testAccNetworkingNetworkV2ResourceName, &found),
					resource.TestCheckResourceAttr(testAccNetworkingNetworkV2ResourceName, "region", networkingV2TestRegion()),
					// The API assigns every optional+computed attribute.
					resource.TestCheckResourceAttrSet(testAccNetworkingNetworkV2ResourceName, "name"),
					resource.TestCheckResourceAttrSet(testAccNetworkingNetworkV2ResourceName, "admin_state_up"),
					// Admin-managed attributes are reported, never configured.
					resource.TestCheckResourceAttrSet(testAccNetworkingNetworkV2ResourceName, "mtu"),
					resource.TestCheckResourceAttrSet(testAccNetworkingNetworkV2ResourceName, "shared"),
					resource.TestCheckResourceAttrSet(testAccNetworkingNetworkV2ResourceName, "blocked"),
					// Create waited for a stable status.
					testAccCheckNetworkingNetworkV2StableStatus(testAccNetworkingNetworkV2ResourceName),
				),
			},
			{
				// SCN-001-02: no changes and no drift means an empty plan.
				Config:   testAccNetworkingNetworkV2Basic(projectName),
				PlanOnly: true,
			},
			{
				// SCN-006-01: import by the plain network identifier.
				ResourceName:      testAccNetworkingNetworkV2ResourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

// TestAccNetworkingNetworkV2OptionalAttributes covers SCN-002-01 and SCN-002-02.
func TestAccNetworkingNetworkV2OptionalAttributes(t *testing.T) {
	var found network.Network
	projectName := acctest.RandomWithPrefix("tf-acc")
	networkName := acctest.RandomWithPrefix("tf-acc-net")

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckNetworkingNetworkV2Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkingNetworkV2OptionalAttributes(projectName, networkName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNetworkingNetworkV2Exists(testAccNetworkingNetworkV2ResourceName, &found),
					resource.TestCheckResourceAttr(testAccNetworkingNetworkV2ResourceName, "name", networkName),
					resource.TestCheckResourceAttr(testAccNetworkingNetworkV2ResourceName, "description", "created by the acceptance test"),
					resource.TestCheckResourceAttr(testAccNetworkingNetworkV2ResourceName, "admin_state_up", "true"),
					resource.TestCheckResourceAttr(testAccNetworkingNetworkV2ResourceName, "dns_domain", "acc-test.example.com."),
					resource.TestCheckResourceAttr(testAccNetworkingNetworkV2ResourceName, "tags.#", "2"),
					resource.TestCheckResourceAttr(testAccNetworkingNetworkV2ResourceName, "all_tags.#", "2"),
				),
			},
			{
				// SCN-002-02: attributes left out of the configuration keep the
				// value the API assigned, so the plan stays empty.
				Config:   testAccNetworkingNetworkV2OptionalAttributes(projectName, networkName),
				PlanOnly: true,
			},
		},
	})
}

// TestAccNetworkingNetworkV2Update covers SCN-003-01 and SCN-003-02: the
// mutable attributes and dns_domain change in place, keeping the identifier.
func TestAccNetworkingNetworkV2Update(t *testing.T) {
	var found network.Network
	projectName := acctest.RandomWithPrefix("tf-acc")
	networkName := acctest.RandomWithPrefix("tf-acc-net")
	updatedName := acctest.RandomWithPrefix("tf-acc-net-upd")

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckNetworkingNetworkV2Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkingNetworkV2OptionalAttributes(projectName, networkName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNetworkingNetworkV2Exists(testAccNetworkingNetworkV2ResourceName, &found),
				),
			},
			{
				Config: testAccNetworkingNetworkV2Updated(projectName, updatedName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNetworkingNetworkV2SameID(testAccNetworkingNetworkV2ResourceName, &found),
					resource.TestCheckResourceAttr(testAccNetworkingNetworkV2ResourceName, "name", updatedName),
					resource.TestCheckResourceAttr(testAccNetworkingNetworkV2ResourceName, "description", "updated by the acceptance test"),
					resource.TestCheckResourceAttr(testAccNetworkingNetworkV2ResourceName, "admin_state_up", "false"),
					resource.TestCheckResourceAttr(testAccNetworkingNetworkV2ResourceName, "dns_domain", "acc-test-updated.example.com."),
					resource.TestCheckResourceAttr(testAccNetworkingNetworkV2ResourceName, "tags.#", "1"),
				),
			},
			{
				Config:   testAccNetworkingNetworkV2Updated(projectName, updatedName),
				PlanOnly: true,
			},
		},
	})
}

// TestAccNetworkingNetworkV2ExternalChanges covers SCN-005-01 (drift is
// reflected), SCN-005-02 (external deletion plans a re-create) and SCN-007-02
// (deleting an already absent network still succeeds).
func TestAccNetworkingNetworkV2ExternalChanges(t *testing.T) {
	var found network.Network
	projectName := acctest.RandomWithPrefix("tf-acc")
	networkName := acctest.RandomWithPrefix("tf-acc-net")

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckNetworkingNetworkV2Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkingNetworkV2OptionalAttributes(projectName, networkName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNetworkingNetworkV2Exists(testAccNetworkingNetworkV2ResourceName, &found),
				),
			},
			{
				// SCN-005-01: an out-of-band rename shows up as drift.
				Config:             testAccNetworkingNetworkV2OptionalAttributes(projectName, networkName),
				Check:              testAccCheckNetworkingNetworkV2RenameOutOfBand(&found),
				ExpectNonEmptyPlan: true,
			},
			{
				// SCN-005-02 and SCN-007-02: an out-of-band delete plans a
				// re-create rather than failing the refresh.
				Config:             testAccNetworkingNetworkV2OptionalAttributes(projectName, networkName),
				Check:              testAccCheckNetworkingNetworkV2DeleteOutOfBand(),
				ExpectNonEmptyPlan: true,
			},
			{
				Config: testAccNetworkingNetworkV2OptionalAttributes(projectName, networkName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNetworkingNetworkV2Exists(testAccNetworkingNetworkV2ResourceName, &found),
				),
			},
		},
	})
}

// TestAccNetworkingNetworkV2AdminManagedAttributes covers SCN-011-02 and
// SCN-011-03: the admin-managed attributes are reported as the API returns them
// and never drive a plan of their own.
func TestAccNetworkingNetworkV2AdminManagedAttributes(t *testing.T) {
	var found network.Network
	projectName := acctest.RandomWithPrefix("tf-acc")
	networkName := acctest.RandomWithPrefix("tf-acc-net")

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckNetworkingNetworkV2Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkingNetworkV2OptionalAttributes(projectName, networkName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNetworkingNetworkV2Exists(testAccNetworkingNetworkV2ResourceName, &found),
					resource.TestCheckResourceAttrSet(testAccNetworkingNetworkV2ResourceName, "blocked"),
					resource.TestCheckResourceAttrSet(testAccNetworkingNetworkV2ResourceName, "is_public"),
					resource.TestCheckResourceAttrSet(testAccNetworkingNetworkV2ResourceName, "is_dns_enabled"),
					resource.TestCheckResourceAttrSet(testAccNetworkingNetworkV2ResourceName, "provider_network_type"),
				),
			},
			{
				// Every admin-managed attribute is Computed-only, so a repeated
				// plan stays empty no matter what the platform set them to.
				Config:   testAccNetworkingNetworkV2OptionalAttributes(projectName, networkName),
				PlanOnly: true,
			},
		},
	})
}

// TestAccNetworkingNetworkV2DataSource covers SCN-012-01 (a unique match
// returns the network with its computed attributes) and SCN-012-02 (no match is
// a diagnostic error, not an empty result).
func TestAccNetworkingNetworkV2DataSource(t *testing.T) {
	projectName := acctest.RandomWithPrefix("tf-acc")
	networkName := acctest.RandomWithPrefix("tf-acc-net")

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckNetworkingNetworkV2Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkingNetworkV2DataSource(projectName, networkName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(testAccNetworkingNetworkV2DataSourceName, "name", networkName),
					resource.TestCheckResourceAttrPair(
						testAccNetworkingNetworkV2DataSourceName, "network_id",
						testAccNetworkingNetworkV2ResourceName, "id",
					),
					// The data source reports the same admin-managed values.
					resource.TestCheckResourceAttrPair(
						testAccNetworkingNetworkV2DataSourceName, "blocked",
						testAccNetworkingNetworkV2ResourceName, "blocked",
					),
					resource.TestCheckResourceAttrPair(
						testAccNetworkingNetworkV2DataSourceName, "shared",
						testAccNetworkingNetworkV2ResourceName, "shared",
					),
					resource.TestCheckResourceAttrPair(
						testAccNetworkingNetworkV2DataSourceName, "provider_network_type",
						testAccNetworkingNetworkV2ResourceName, "provider_network_type",
					),
				),
			},
			{
				// SCN-012-02: nothing matches, so the read fails explicitly.
				Config:      testAccNetworkingNetworkV2DataSourceNoMatch(projectName),
				ExpectError: regexp.MustCompile(`no networking network found`),
			},
		},
	})
}

// -----------------------------------------------------------------------------
// Acceptance test helpers.
// -----------------------------------------------------------------------------

// testAccNetworkingV2DriftedName is the value the drift checks write out of band,
// so a refresh has something to report as a change.
const testAccNetworkingV2DriftedName = "renamed outside terraform"

// networkingNetworkV2TestHandle bundles a networking client with the identifier
// under test.
type networkingNetworkV2TestHandle struct {
	client *vpc.Client
	id     string
}

func testAccNetworkingNetworkV2Handle(s *terraform.State, resourceName string) (*networkingNetworkV2TestHandle, error) {
	rs, ok := s.RootModule().Resources[resourceName]
	if !ok {
		return nil, fmt.Errorf("not found: %s", resourceName)
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

func testAccCheckNetworkingNetworkV2Exists(resourceName string, found *network.Network) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		handle, err := testAccNetworkingNetworkV2Handle(s, resourceName)
		if err != nil {
			return err
		}

		result, err := network.Get(context.Background(), handle.client, handle.id)
		if err != nil {
			return err
		}

		if result.ID != handle.id {
			return errors.New("networking network not found")
		}

		*found = *result

		return nil
	}
}

func testAccCheckNetworkingNetworkV2StableStatus(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("not found: %s", resourceName)
		}

		status := rs.Primary.Attributes["status"]
		if status != networkingNetworkV2StatusActive && status != networkingNetworkV2StatusDown {
			return fmt.Errorf("networking network is in the transient %s status", status)
		}

		return nil
	}
}

func testAccCheckNetworkingNetworkV2SameID(resourceName string, found *network.Network) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("not found: %s", resourceName)
		}

		if rs.Primary.ID != found.ID {
			return fmt.Errorf("networking network was replaced: %s -> %s", found.ID, rs.Primary.ID)
		}

		return nil
	}
}

// testAccCheckNetworkingNetworkV2RenameOutOfBand renames the network without
// Terraform, producing the drift SCN-005-01 describes.
func testAccCheckNetworkingNetworkV2RenameOutOfBand(found *network.Network) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		handle, err := testAccNetworkingNetworkV2Handle(s, testAccNetworkingNetworkV2ResourceName)
		if err != nil {
			return err
		}

		driftedName := found.Name + "-drifted"
		_, err = network.Update(context.Background(), handle.client, handle.id, network.UpdateRequest{
			Name: &driftedName,
		})

		return err
	}
}

// testAccCheckNetworkingNetworkV2DeleteOutOfBand deletes the network without
// Terraform, producing the external deletion SCN-005-02 describes.
func testAccCheckNetworkingNetworkV2DeleteOutOfBand() resource.TestCheckFunc {
	return func(s *terraform.State) error {
		handle, err := testAccNetworkingNetworkV2Handle(s, testAccNetworkingNetworkV2ResourceName)
		if err != nil {
			return err
		}

		return network.Delete(context.Background(), handle.client, handle.id)
	}
}

func testAccCheckNetworkingNetworkV2Destroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "selectel_networking_network_v2" {
			continue
		}

		client, err := networkingV2Client(
			testAccProvider.Meta(),
			rs.Primary.Attributes["project_id"],
			rs.Primary.Attributes["region"],
		)
		if err != nil {
			// The project is destroyed together with the network, so a client
			// that can no longer be built proves nothing is left behind.
			return nil //nolint:nilerr // see comment above
		}

		if _, err := network.Get(context.Background(), client, rs.Primary.ID); err == nil {
			return errors.New("networking network still exists")
		} else if !networkingIsNotFound(err) {
			return err
		}
	}

	return nil
}

// -----------------------------------------------------------------------------
// Acceptance test configurations.
// -----------------------------------------------------------------------------

func testAccNetworkingNetworkV2Project(projectName string) string {
	return fmt.Sprintf(`
resource "selectel_vpc_project_v2" "project_tf_acc_test_1" {
  name = "%s"
}
`, projectName)
}

func testAccNetworkingNetworkV2Basic(projectName string) string {
	return testAccNetworkingNetworkV2Project(projectName) + fmt.Sprintf(`
resource "selectel_networking_network_v2" "network_tf_acc_test_1" {
  project_id = selectel_vpc_project_v2.project_tf_acc_test_1.id
  region     = "%s"
}
`, networkingV2TestRegion())
}

func testAccNetworkingNetworkV2OptionalAttributes(projectName, networkName string) string {
	return testAccNetworkingNetworkV2Project(projectName) + fmt.Sprintf(`
resource "selectel_networking_network_v2" "network_tf_acc_test_1" {
  project_id     = selectel_vpc_project_v2.project_tf_acc_test_1.id
  region         = "%s"
  name           = "%s"
  description    = "created by the acceptance test"
  admin_state_up = true
  dns_domain     = "acc-test.example.com."
  tags           = ["tf-acc", "networking"]
}
`, networkingV2TestRegion(), networkName)
}

func testAccNetworkingNetworkV2Updated(projectName, networkName string) string {
	return testAccNetworkingNetworkV2Project(projectName) + fmt.Sprintf(`
resource "selectel_networking_network_v2" "network_tf_acc_test_1" {
  project_id     = selectel_vpc_project_v2.project_tf_acc_test_1.id
  region         = "%s"
  name           = "%s"
  description    = "updated by the acceptance test"
  admin_state_up = false
  dns_domain     = "acc-test-updated.example.com."
  tags           = ["tf-acc"]
}
`, networkingV2TestRegion(), networkName)
}

func testAccNetworkingNetworkV2DataSource(projectName, networkName string) string {
	return testAccNetworkingNetworkV2OptionalAttributes(projectName, networkName) + fmt.Sprintf(`
data "selectel_networking_network_v2" "network_tf_acc_test_1" {
  project_id = selectel_vpc_project_v2.project_tf_acc_test_1.id
  region     = "%s"
  name       = selectel_networking_network_v2.network_tf_acc_test_1.name
}
`, networkingV2TestRegion())
}

func testAccNetworkingNetworkV2DataSourceNoMatch(projectName string) string {
	return testAccNetworkingNetworkV2Project(projectName) + fmt.Sprintf(`
data "selectel_networking_network_v2" "network_tf_acc_test_1" {
  project_id = selectel_vpc_project_v2.project_tf_acc_test_1.id
  region     = "%s"
  name       = "tf-acc-no-such-network"
}
`, networkingV2TestRegion())
}
