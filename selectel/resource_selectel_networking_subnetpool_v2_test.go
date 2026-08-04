package selectel

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/selectel/vpc-go/pkg/v2/subnetpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testAccNetworkingSubnetPoolV2ResourceName   = "selectel_networking_subnetpool_v2.subnetpool_tf_acc_test_1"
	testAccNetworkingSubnetPoolV2DataSourceName = "data.selectel_networking_subnetpool_v2.subnetpool_tf_acc_test_1"
)

// -----------------------------------------------------------------------------
// Schema unit tests.
// -----------------------------------------------------------------------------

// TestNetworkingSubnetPoolV2Schema covers REQ-029: which attributes are
// arguments, which are read-only, and which replace the resource.
func TestNetworkingSubnetPoolV2Schema(t *testing.T) {
	s := resourceNetworkingSubnetPoolV2().Schema

	t.Run("project_id and region are required and ForceNew", func(t *testing.T) {
		for _, name := range []string{"project_id", "region"} {
			require.Contains(t, s, name)
			assert.True(t, s[name].Required)
			assert.True(t, s[name].ForceNew)
		}
	})

	t.Run("name and prefixes are required and mutable", func(t *testing.T) {
		for _, name := range []string{"name", "prefixes"} {
			require.Contains(t, s, name)
			assert.True(t, s[name].Required)
			assert.False(t, s[name].ForceNew, name+" must be updatable in place")
		}
		assert.Equal(t, 1, s["prefixes"].MinItems, "prefixes must be non-empty")
	})

	// SCN-029-04 and SCN-029-05: shared and is_default are admin-managed, and
	// SCN-029-03: ip_version is derived by Neutron from prefixes.
	for _, name := range []string{
		"shared", "is_default", "ip_version",
		"created_at", "updated_at", "revision_number", "all_tags",
	} {
		t.Run(name+" is computed-only", func(t *testing.T) {
			require.Contains(t, s, name)
			assert.True(t, s[name].Computed)
			assert.False(t, s[name].Optional, name+" must not be a configuration argument")
			assert.False(t, s[name].Required)
		})
	}

	// REQ-031: exactly this set is updatable without replacement.
	t.Run("the updatable set matches the requirement", func(t *testing.T) {
		updatable := make([]string, 0)
		for name, attr := range s {
			if (attr.Optional || attr.Required) && !attr.ForceNew {
				updatable = append(updatable, name)
			}
		}

		assert.ElementsMatch(t, []string{
			"name", "prefixes", "description", "default_quota", "default_prefixlen",
			"min_prefixlen", "max_prefixlen", "address_scope_id", "tags",
		}, updatable)
	})

	// REQ-032: only project_id and region replace the pool.
	t.Run("only project_id and region are ForceNew", func(t *testing.T) {
		forceNew := make([]string, 0)
		for name, attr := range s {
			if attr.ForceNew {
				forceNew = append(forceNew, name)
			}
		}

		assert.ElementsMatch(t, []string{"project_id", "region"}, forceNew)
	})
}

// TestNetworkingSubnetPoolV2DataSourceSelectors covers REQ-036: every selector
// the requirement names is accepted, and the computed-only attributes are not
// offered as selectors.
func TestNetworkingSubnetPoolV2DataSourceSelectors(t *testing.T) {
	s := dataSourceNetworkingSubnetPoolV2().Schema

	selectors := make([]string, 0)
	for name, attr := range s {
		if attr.Optional || attr.Required {
			selectors = append(selectors, name)
		}
	}

	assert.ElementsMatch(t, []string{
		"project_id", "region", "subnetpool_id", "name", "description",
		"default_quota", "default_prefixlen", "min_prefixlen", "max_prefixlen",
		"address_scope_id", "ip_version", "shared", "is_default", "tags",
	}, selectors)

	for _, name := range []string{"prefixes", "all_tags", "created_at", "updated_at", "revision_number"} {
		require.Contains(t, s, name)
		assert.True(t, s[name].Computed)
		assert.False(t, s[name].Optional, name+" must be returned, not selectable")
	}
}

// -----------------------------------------------------------------------------
// Acceptance tests against a staging environment (TF_ACC=1).
// -----------------------------------------------------------------------------

// TestAccNetworkingSubnetPoolV2Basic covers SCN-029-01, SCN-030-01 (empty
// follow-up plan), SCN-034-01 (import) and SCN-035-01 (delete).
func TestAccNetworkingSubnetPoolV2Basic(t *testing.T) {
	var found subnetpool.SubnetPool
	projectID, region := networkingV2TestEnv()
	poolName := acctest.RandomWithPrefix("tf-acc-pool")

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheckWithProjectIDAndRegion(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckNetworkingSubnetPoolV2Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkingSubnetPoolV2Basic(projectID, region, poolName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNetworkingSubnetPoolV2Exists(&found),
					resource.TestCheckResourceAttr(testAccNetworkingSubnetPoolV2ResourceName, "name", poolName),
					resource.TestCheckResourceAttr(testAccNetworkingSubnetPoolV2ResourceName, "prefixes.#", "1"),
					// The API fills the prefix lengths, the quota and the version.
					resource.TestCheckResourceAttrSet(testAccNetworkingSubnetPoolV2ResourceName, "default_prefixlen"),
					resource.TestCheckResourceAttrSet(testAccNetworkingSubnetPoolV2ResourceName, "min_prefixlen"),
					resource.TestCheckResourceAttrSet(testAccNetworkingSubnetPoolV2ResourceName, "max_prefixlen"),
					resource.TestCheckResourceAttr(testAccNetworkingSubnetPoolV2ResourceName, "ip_version", "4"),
					// Admin-managed attributes are reported, never configured.
					resource.TestCheckResourceAttrSet(testAccNetworkingSubnetPoolV2ResourceName, "shared"),
					resource.TestCheckResourceAttrSet(testAccNetworkingSubnetPoolV2ResourceName, "is_default"),
					resource.TestCheckResourceAttrSet(testAccNetworkingSubnetPoolV2ResourceName, "revision_number"),
				),
			},
			{
				Config:   testAccNetworkingSubnetPoolV2Basic(projectID, region, poolName),
				PlanOnly: true,
			},
			{
				ResourceName:      testAccNetworkingSubnetPoolV2ResourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

// TestAccNetworkingSubnetPoolV2OptionalAttributes covers SCN-029-02.
func TestAccNetworkingSubnetPoolV2OptionalAttributes(t *testing.T) {
	var found subnetpool.SubnetPool
	projectID, region := networkingV2TestEnv()
	poolName := acctest.RandomWithPrefix("tf-acc-pool")

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheckWithProjectIDAndRegion(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckNetworkingSubnetPoolV2Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkingSubnetPoolV2Full(projectID, region, poolName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNetworkingSubnetPoolV2Exists(&found),
					resource.TestCheckResourceAttr(testAccNetworkingSubnetPoolV2ResourceName, "description", "created by the acceptance test"),
					resource.TestCheckResourceAttr(testAccNetworkingSubnetPoolV2ResourceName, "default_quota", "8"),
					resource.TestCheckResourceAttr(testAccNetworkingSubnetPoolV2ResourceName, "default_prefixlen", "26"),
					resource.TestCheckResourceAttr(testAccNetworkingSubnetPoolV2ResourceName, "min_prefixlen", "24"),
					resource.TestCheckResourceAttr(testAccNetworkingSubnetPoolV2ResourceName, "max_prefixlen", "28"),
					resource.TestCheckResourceAttr(testAccNetworkingSubnetPoolV2ResourceName, "tags.#", "2"),
					resource.TestCheckResourceAttr(testAccNetworkingSubnetPoolV2ResourceName, "all_tags.#", "2"),
				),
			},
			{
				Config:   testAccNetworkingSubnetPoolV2Full(projectID, region, poolName),
				PlanOnly: true,
			},
		},
	})
}

// TestAccNetworkingSubnetPoolV2Update covers SCN-031-01 … SCN-031-04: every
// mutable attribute changes in place, keeping the identifier.
func TestAccNetworkingSubnetPoolV2Update(t *testing.T) {
	var found subnetpool.SubnetPool
	projectID, region := networkingV2TestEnv()
	poolName := acctest.RandomWithPrefix("tf-acc-pool")
	updatedName := acctest.RandomWithPrefix("tf-acc-pool-upd")

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheckWithProjectIDAndRegion(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckNetworkingSubnetPoolV2Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkingSubnetPoolV2Full(projectID, region, poolName),
				Check:  testAccCheckNetworkingSubnetPoolV2Exists(&found),
			},
			{
				Config: testAccNetworkingSubnetPoolV2Updated(projectID, region, updatedName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNetworkingSubnetPoolV2SameID(&found),
					resource.TestCheckResourceAttr(testAccNetworkingSubnetPoolV2ResourceName, "name", updatedName),
					resource.TestCheckResourceAttr(testAccNetworkingSubnetPoolV2ResourceName, "description", "updated by the acceptance test"),
					resource.TestCheckResourceAttr(testAccNetworkingSubnetPoolV2ResourceName, "default_quota", "16"),
					// SCN-031-02: the prefix list grew without replacement.
					resource.TestCheckResourceAttr(testAccNetworkingSubnetPoolV2ResourceName, "prefixes.#", "2"),
					// SCN-031-03: the prefix length range moved.
					resource.TestCheckResourceAttr(testAccNetworkingSubnetPoolV2ResourceName, "default_prefixlen", "27"),
					resource.TestCheckResourceAttr(testAccNetworkingSubnetPoolV2ResourceName, "min_prefixlen", "25"),
					resource.TestCheckResourceAttr(testAccNetworkingSubnetPoolV2ResourceName, "max_prefixlen", "29"),
					// SCN-031-04: the tag set changed.
					resource.TestCheckResourceAttr(testAccNetworkingSubnetPoolV2ResourceName, "tags.#", "1"),
					resource.TestCheckResourceAttr(testAccNetworkingSubnetPoolV2ResourceName, "all_tags.#", "1"),
				),
			},
			{
				Config:   testAccNetworkingSubnetPoolV2Updated(projectID, region, updatedName),
				PlanOnly: true,
			},
		},
	})
}

// TestAccNetworkingSubnetPoolV2ExternalChanges covers SCN-033-01, SCN-033-02 and
// SCN-035-02.
func TestAccNetworkingSubnetPoolV2ExternalChanges(t *testing.T) {
	var found subnetpool.SubnetPool
	projectID, region := networkingV2TestEnv()
	poolName := acctest.RandomWithPrefix("tf-acc-pool")

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheckWithProjectIDAndRegion(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckNetworkingSubnetPoolV2Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkingSubnetPoolV2Basic(projectID, region, poolName),
				Check:  testAccCheckNetworkingSubnetPoolV2Exists(&found),
			},
			{
				// SCN-033-01: an out-of-band rename shows up as drift.
				Config:             testAccNetworkingSubnetPoolV2Basic(projectID, region, poolName),
				Check:              testAccCheckNetworkingSubnetPoolV2RenameOutOfBand(&found),
				ExpectNonEmptyPlan: true,
			},
			{
				// SCN-033-02 and SCN-035-02: an out-of-band delete plans a
				// re-create instead of failing.
				Config:             testAccNetworkingSubnetPoolV2Basic(projectID, region, poolName),
				Check:              testAccCheckNetworkingSubnetPoolV2DeleteOutOfBand(),
				ExpectNonEmptyPlan: true,
			},
			{
				Config: testAccNetworkingSubnetPoolV2Basic(projectID, region, poolName),
				Check:  testAccCheckNetworkingSubnetPoolV2Exists(&found),
			},
		},
	})
}

// TestAccNetworkingSubnetPoolV2InUse covers SCN-035-03: a pool a subnet still
// allocates from cannot be deleted, and the resource stays in the state.
func TestAccNetworkingSubnetPoolV2InUse(t *testing.T) {
	projectID, region := networkingV2TestEnv()
	poolName := acctest.RandomWithPrefix("tf-acc-pool")

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheckWithProjectIDAndRegion(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckNetworkingSubnetPoolV2Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkingSubnetPoolV2WithSubnet(projectID, region, poolName),
			},
			{
				// Removing only the pool from the configuration, while the subnet
				// allocated from it stays, must fail.
				Config:      testAccNetworkingSubnetPoolV2SubnetOnly(projectID, region),
				ExpectError: regexp.MustCompile(`error deleting networking subnetpool`),
			},
		},
	})
}

// TestAccNetworkingSubnetPoolV2DataSource covers SCN-036-01 … SCN-036-03.
func TestAccNetworkingSubnetPoolV2DataSource(t *testing.T) {
	projectID, region := networkingV2TestEnv()
	poolName := acctest.RandomWithPrefix("tf-acc-pool")

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheckWithProjectIDAndRegion(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckNetworkingSubnetPoolV2Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkingSubnetPoolV2DataSource(projectID, region, poolName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(testAccNetworkingSubnetPoolV2DataSourceName, "name", poolName),
					resource.TestCheckResourceAttrPair(
						testAccNetworkingSubnetPoolV2DataSourceName, "subnetpool_id",
						testAccNetworkingSubnetPoolV2ResourceName, "id",
					),
					// SCN-036-02: the unique match fills every remaining attribute.
					resource.TestCheckResourceAttrSet(testAccNetworkingSubnetPoolV2DataSourceName, "created_at"),
					resource.TestCheckResourceAttrSet(testAccNetworkingSubnetPoolV2DataSourceName, "revision_number"),
					resource.TestCheckResourceAttrPair(
						testAccNetworkingSubnetPoolV2DataSourceName, "prefixes.#",
						testAccNetworkingSubnetPoolV2ResourceName, "prefixes.#",
					),
				),
			},
			{
				// SCN-036-01: zero matches is a diagnostic error.
				Config:      testAccNetworkingSubnetPoolV2DataSourceNoMatch(projectID, region),
				ExpectError: regexp.MustCompile(`no networking subnetpool found`),
			},
		},
	})
}

// -----------------------------------------------------------------------------
// Acceptance test helpers.
// -----------------------------------------------------------------------------

func testAccNetworkingSubnetPoolV2Handle(s *terraform.State) (*networkingNetworkV2TestHandle, error) {
	rs, ok := s.RootModule().Resources[testAccNetworkingSubnetPoolV2ResourceName]
	if !ok {
		return nil, fmt.Errorf("not found: %s", testAccNetworkingSubnetPoolV2ResourceName)
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

func testAccCheckNetworkingSubnetPoolV2Exists(found *subnetpool.SubnetPool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		handle, err := testAccNetworkingSubnetPoolV2Handle(s)
		if err != nil {
			return err
		}

		result, err := subnetpool.Get(context.Background(), handle.client, handle.id)
		if err != nil {
			return err
		}

		*found = *result

		return nil
	}
}

func testAccCheckNetworkingSubnetPoolV2SameID(found *subnetpool.SubnetPool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[testAccNetworkingSubnetPoolV2ResourceName]
		if !ok {
			return fmt.Errorf("not found: %s", testAccNetworkingSubnetPoolV2ResourceName)
		}

		if rs.Primary.ID != found.ID {
			return fmt.Errorf("networking subnetpool was replaced: %s -> %s", found.ID, rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckNetworkingSubnetPoolV2RenameOutOfBand(found *subnetpool.SubnetPool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		handle, err := testAccNetworkingSubnetPoolV2Handle(s)
		if err != nil {
			return err
		}

		driftedName := found.Name + "-drifted"
		_, err = subnetpool.Update(context.Background(), handle.client, handle.id, subnetpool.UpdateRequest{
			Name: &driftedName,
		})

		return err
	}
}

func testAccCheckNetworkingSubnetPoolV2DeleteOutOfBand() resource.TestCheckFunc {
	return func(s *terraform.State) error {
		handle, err := testAccNetworkingSubnetPoolV2Handle(s)
		if err != nil {
			return err
		}

		return subnetpool.Delete(context.Background(), handle.client, handle.id)
	}
}

func testAccCheckNetworkingSubnetPoolV2Destroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "selectel_networking_subnetpool_v2" {
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

		if _, err := subnetpool.Get(context.Background(), client, rs.Primary.ID); err == nil {
			return errors.New("networking subnetpool still exists")
		} else if !networkingIsNotFound(err) {
			return err
		}
	}

	return nil
}

// -----------------------------------------------------------------------------
// Acceptance test configurations.
// -----------------------------------------------------------------------------

func testAccNetworkingSubnetPoolV2Basic(projectID, region, poolName string) string {
	return testAccNetworkingV2Base(projectID) + fmt.Sprintf(`
resource "selectel_networking_subnetpool_v2" "subnetpool_tf_acc_test_1" {
  project_id = local.project_id
  region     = "%s"
  name       = "%s"
  prefixes   = ["192.168.200.0/24"]
}
`, region, poolName)
}

func testAccNetworkingSubnetPoolV2Full(projectID, region, poolName string) string {
	return testAccNetworkingV2Base(projectID) + fmt.Sprintf(`
resource "selectel_networking_subnetpool_v2" "subnetpool_tf_acc_test_1" {
  project_id        = local.project_id
  region            = "%s"
  name              = "%s"
  description       = "created by the acceptance test"
  prefixes          = ["192.168.200.0/24"]
  default_quota     = 8
  default_prefixlen = 26
  min_prefixlen     = 24
  max_prefixlen     = 28
  tags              = ["tf-acc", "networking"]
}
`, region, poolName)
}

func testAccNetworkingSubnetPoolV2Updated(projectID, region, poolName string) string {
	return testAccNetworkingV2Base(projectID) + fmt.Sprintf(`
resource "selectel_networking_subnetpool_v2" "subnetpool_tf_acc_test_1" {
  project_id        = local.project_id
  region            = "%s"
  name              = "%s"
  description       = "updated by the acceptance test"
  prefixes          = ["192.168.200.0/24", "192.168.201.0/24"]
  default_quota     = 16
  default_prefixlen = 27
  min_prefixlen     = 25
  max_prefixlen     = 29
  tags              = ["tf-acc"]
}
`, region, poolName)
}

func testAccNetworkingSubnetPoolV2WithSubnet(projectID, region, poolName string) string {
	return testAccNetworkingSubnetPoolV2Basic(projectID, region, poolName) + fmt.Sprintf(`
resource "selectel_networking_network_v2" "network_tf_acc_test_1" {
  project_id = local.project_id
  region     = "%s"
  name       = "tf-acc-pool-parent"
}

resource "selectel_networking_subnet_v2" "subnet_tf_acc_test_1" {
  project_id    = local.project_id
  region        = "%s"
  network_id    = selectel_networking_network_v2.network_tf_acc_test_1.id
  subnetpool_id = selectel_networking_subnetpool_v2.subnetpool_tf_acc_test_1.id
  prefix_length = 26
}
`, region, region)
}

func testAccNetworkingSubnetPoolV2SubnetOnly(projectID, region string) string {
	return testAccNetworkingV2Base(projectID) + fmt.Sprintf(`
resource "selectel_networking_network_v2" "network_tf_acc_test_1" {
  project_id = local.project_id
  region     = "%s"
  name       = "tf-acc-pool-parent"
}
`, region)
}

func testAccNetworkingSubnetPoolV2DataSource(projectID, region, poolName string) string {
	return testAccNetworkingSubnetPoolV2Basic(projectID, region, poolName) + fmt.Sprintf(`
data "selectel_networking_subnetpool_v2" "subnetpool_tf_acc_test_1" {
  project_id = local.project_id
  region     = "%s"
  name       = selectel_networking_subnetpool_v2.subnetpool_tf_acc_test_1.name
}
`, region)
}

func testAccNetworkingSubnetPoolV2DataSourceNoMatch(projectID, region string) string {
	return testAccNetworkingV2Base(projectID) + fmt.Sprintf(`
data "selectel_networking_subnetpool_v2" "subnetpool_tf_acc_test_1" {
  project_id = local.project_id
  region     = "%s"
  name       = "tf-acc-no-such-pool"
}
`, region)
}
