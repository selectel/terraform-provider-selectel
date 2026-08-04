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
	"github.com/selectel/vpc-go/pkg/v2/addressscope"
	"github.com/selectel/vpc-go/pkg/v2/subnetpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testAccNetworkingAddressScopeV2ResourceName   = "selectel_networking_addressscope_v2.addressscope_tf_acc_test_1"
	testAccNetworkingAddressScopeV2DataSourceName = "data.selectel_networking_addressscope_v2.addressscope_tf_acc_test_1"
)

// -----------------------------------------------------------------------------
// Schema unit tests.
// -----------------------------------------------------------------------------

// TestNetworkingAddressScopeV2Schema covers REQ-137 and REQ-140: what the caller
// owns, what replaces the scope, and that shared is read-only.
func TestNetworkingAddressScopeV2Schema(t *testing.T) {
	s := resourceNetworkingAddressScopeV2().Schema

	assert.Len(t, s, 5, "the address scope is fully described by these five attributes")

	t.Run("project_id and region are required and ForceNew", func(t *testing.T) {
		for _, name := range []string{"project_id", "region"} {
			require.Contains(t, s, name)
			assert.True(t, s[name].Required)
			assert.True(t, s[name].ForceNew)
		}
	})

	// REQ-139: name changes in place.
	t.Run("name is required and updatable", func(t *testing.T) {
		require.Contains(t, s, "name")
		assert.True(t, s["name"].Required)
		assert.False(t, s["name"].ForceNew)
	})

	// REQ-137: optional, ForceNew, defaults to 4.
	t.Run("ip_version is optional, ForceNew and defaults to 4", func(t *testing.T) {
		require.Contains(t, s, "ip_version")
		assert.True(t, s["ip_version"].Optional)
		assert.True(t, s["ip_version"].ForceNew)
		assert.Equal(t, 4, s["ip_version"].Default)
	})

	// REQ-141: no local validation — an unsupported value is the API's to reject.
	t.Run("ip_version is not validated locally", func(t *testing.T) {
		assert.Nil(t, s["ip_version"].ValidateFunc)
		assert.Nil(t, s["ip_version"].ValidateDiagFunc)
	})

	// REQ-137 SCN-137-03: shared cannot be configured at all.
	t.Run("shared is computed-only", func(t *testing.T) {
		require.Contains(t, s, "shared")
		assert.True(t, s["shared"].Computed)
		assert.False(t, s["shared"].Optional)
		assert.False(t, s["shared"].Required)
	})

	// REQ-140: exactly these three replace the scope.
	t.Run("only the immutable attributes are ForceNew", func(t *testing.T) {
		forceNew := make([]string, 0)
		for name, attr := range s {
			if attr.ForceNew {
				forceNew = append(forceNew, name)
			}
		}

		assert.ElementsMatch(t, []string{"project_id", "region", "ip_version"}, forceNew)
	})
}

// TestNetworkingAddressScopeV2Flatten covers REQ-139: the state reflects the
// remote shared value, whatever an admin set it to.
func TestNetworkingAddressScopeV2Flatten(t *testing.T) {
	for _, shared := range []bool{false, true} {
		t.Run(fmt.Sprintf("reflects shared=%t", shared), func(t *testing.T) {
			d := schema.TestResourceDataRaw(t, resourceNetworkingAddressScopeV2().Schema, map[string]interface{}{})

			flattenNetworkingAddressScopeV2(d, &addressscope.AddressScope{
				ID:        "scope-id",
				Name:      "scope-1",
				IPVersion: 6,
				Shared:    shared,
			})

			assert.Equal(t, "scope-1", d.Get("name"))
			assert.Equal(t, 6, d.Get("ip_version"))
			assert.Equal(t, shared, d.Get("shared"))
		})
	}
}

// TestNetworkingAddressScopeV2DataSourceSelectors covers REQ-144.
func TestNetworkingAddressScopeV2DataSourceSelectors(t *testing.T) {
	s := dataSourceNetworkingAddressScopeV2().Schema

	selectors := make([]string, 0)
	for name, attr := range s {
		if attr.Optional || attr.Required {
			selectors = append(selectors, name)
		}
	}

	// shared is selectable on the data source even though the resource cannot
	// set it: finding the shared scopes an admin created is a legitimate read.
	assert.ElementsMatch(t, []string{
		"project_id", "region", "name", "ip_version", "shared",
	}, selectors)
}

// -----------------------------------------------------------------------------
// Acceptance tests against a staging environment (TF_ACC=1).
// -----------------------------------------------------------------------------

// TestAccNetworkingAddressScopeV2Basic covers SCN-137-01, SCN-138-01, SCN-142-01
// and SCN-143-01.
func TestAccNetworkingAddressScopeV2Basic(t *testing.T) {
	var found addressscope.AddressScope
	projectID, region := networkingV2TestEnv()

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheckWithProjectIDAndRegion(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckNetworkingAddressScopeV2Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkingAddressScopeV2Basic(projectID, region, "scope-1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNetworkingAddressScopeV2Exists(&found),
					// SCN-137-01: ip_version defaults to 4 and shared comes from
					// the API.
					resource.TestCheckResourceAttr(testAccNetworkingAddressScopeV2ResourceName, "ip_version", "4"),
					resource.TestCheckResourceAttrSet(testAccNetworkingAddressScopeV2ResourceName, "shared"),
				),
			},
			{
				// SCN-138-01: the follow-up plan is empty.
				Config:   testAccNetworkingAddressScopeV2Basic(projectID, region, "scope-1"),
				PlanOnly: true,
			},
			{
				// SCN-142-01: import by the identifier alone.
				ResourceName:      testAccNetworkingAddressScopeV2ResourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

// TestAccNetworkingAddressScopeV2IPVersion6 covers SCN-137-02 and SCN-140-01.
func TestAccNetworkingAddressScopeV2IPVersion6(t *testing.T) {
	var found addressscope.AddressScope
	projectID, region := networkingV2TestEnv()

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheckWithProjectIDAndRegion(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckNetworkingAddressScopeV2Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkingAddressScopeV2Version(projectID, region, 6),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNetworkingAddressScopeV2Exists(&found),
					resource.TestCheckResourceAttr(testAccNetworkingAddressScopeV2ResourceName, "ip_version", "6"),
				),
			},
			{
				// SCN-140-01: a different ip_version replaces the scope.
				Config: testAccNetworkingAddressScopeV2Version(projectID, region, 4),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNetworkingAddressScopeV2Replaced(&found),
					resource.TestCheckResourceAttr(testAccNetworkingAddressScopeV2ResourceName, "ip_version", "4"),
				),
			},
		},
	})
}

// TestAccNetworkingAddressScopeV2Update covers SCN-139-01 and SCN-139-04: the
// rename happens in place and does not disturb shared.
func TestAccNetworkingAddressScopeV2Update(t *testing.T) {
	var found addressscope.AddressScope
	projectID, region := networkingV2TestEnv()

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheckWithProjectIDAndRegion(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckNetworkingAddressScopeV2Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkingAddressScopeV2Basic(projectID, region, "scope-1"),
				Check:  testAccCheckNetworkingAddressScopeV2Exists(&found),
			},
			{
				Config: testAccNetworkingAddressScopeV2Basic(projectID, region, "scope-1-renamed"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNetworkingAddressScopeV2SameID(&found),
					resource.TestCheckResourceAttr(testAccNetworkingAddressScopeV2ResourceName, "name", "scope-1-renamed"),
					// SCN-139-04: shared is untouched by the rename.
					testAccCheckNetworkingAddressScopeV2SharedUnchanged(&found),
				),
			},
		},
	})
}

// TestAccNetworkingAddressScopeV2InvalidIPVersion covers SCN-141-02: the value is
// sent unvalidated and the API's rejection becomes the diagnostic.
func TestAccNetworkingAddressScopeV2InvalidIPVersion(t *testing.T) {
	projectID, region := networkingV2TestEnv()

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheckWithProjectIDAndRegion(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckNetworkingAddressScopeV2Destroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccNetworkingAddressScopeV2Version(projectID, region, 5),
				ExpectError: regexp.MustCompile(`(?i)error creating|invalid|bad request`),
			},
		},
	})
}

// TestAccNetworkingAddressScopeV2ExternalChanges covers SCN-138-02 and SCN-138-03.
func TestAccNetworkingAddressScopeV2ExternalChanges(t *testing.T) {
	var found addressscope.AddressScope
	projectID, region := networkingV2TestEnv()

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheckWithProjectIDAndRegion(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckNetworkingAddressScopeV2Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkingAddressScopeV2Basic(projectID, region, "scope-1"),
				Check:  testAccCheckNetworkingAddressScopeV2Exists(&found),
			},
			{
				Config:             testAccNetworkingAddressScopeV2Basic(projectID, region, "scope-1"),
				Check:              testAccCheckNetworkingAddressScopeV2RenameOutOfBand(),
				ExpectNonEmptyPlan: true,
			},
			{
				Config:             testAccNetworkingAddressScopeV2Basic(projectID, region, "scope-1"),
				Check:              testAccCheckNetworkingAddressScopeV2DeleteOutOfBand(),
				ExpectNonEmptyPlan: true,
			},
			{
				Config: testAccNetworkingAddressScopeV2Basic(projectID, region, "scope-1"),
				Check:  testAccCheckNetworkingAddressScopeV2Exists(&found),
			},
		},
	})
}

// TestAccNetworkingAddressScopeV2DeleteConflict covers SCN-143-02: a subnet pool
// referencing the scope makes the delete a conflict.
//
// The pool is created through the SDK rather than as a resource, because it has
// to outlive the scope for the destroy to hit the conflict — an HCL pool that
// references the scope would be destroyed first and there would be no conflict.
func TestAccNetworkingAddressScopeV2DeleteConflict(t *testing.T) {
	projectID, region := networkingV2TestEnv()
	referencing := &networkingAddressScopeV2TestPool{}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheckWithProjectIDAndRegion(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckNetworkingAddressScopeV2Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkingAddressScopeV2Basic(projectID, region, "scope-1"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(testAccNetworkingAddressScopeV2ResourceName, "id"),
					testAccCheckNetworkingAddressScopeV2CreatePool(referencing),
				),
			},
			{
				// The scope is still referenced, so the destroy is refused.
				Config:      testAccNetworkingAddressScopeV2Basic(projectID, region, "scope-1"),
				Destroy:     true,
				ExpectError: regexp.MustCompile(`(?i)conflict|in use`),
			},
			{
				// Remove the reference so the final destroy — and CheckDestroy —
				// can succeed.
				Config: testAccNetworkingAddressScopeV2Basic(projectID, region, "scope-1"),
				Check:  testAccCheckNetworkingAddressScopeV2DeletePool(referencing),
			},
		},
	})
}

// TestAccNetworkingAddressScopeV2DataSource covers SCN-144-01, SCN-144-02 and
// SCN-144-03.
func TestAccNetworkingAddressScopeV2DataSource(t *testing.T) {
	projectID, region := networkingV2TestEnv()
	// The project is shared, so the name the data source searches on must be
	// unique to this run.
	scopeName := acctest.RandomWithPrefix("tf-acc-scope")

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheckWithProjectIDAndRegion(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckNetworkingAddressScopeV2Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkingAddressScopeV2DataSource(projectID, region, scopeName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(
						testAccNetworkingAddressScopeV2DataSourceName, "id",
						testAccNetworkingAddressScopeV2ResourceName, "id",
					),
					// Every attribute comes from the found object.
					resource.TestCheckResourceAttrPair(
						testAccNetworkingAddressScopeV2DataSourceName, "ip_version",
						testAccNetworkingAddressScopeV2ResourceName, "ip_version",
					),
					resource.TestCheckResourceAttrPair(
						testAccNetworkingAddressScopeV2DataSourceName, "shared",
						testAccNetworkingAddressScopeV2ResourceName, "shared",
					),
				),
			},
			{
				Config:      testAccNetworkingAddressScopeV2DataSourceNoMatch(projectID, region),
				ExpectError: regexp.MustCompile(`no networking address scope found`),
			},
			{
				// SCN-144-03: two scopes match a selector that is not precise.
				Config:      testAccNetworkingAddressScopeV2DataSourceAmbiguous(projectID, region),
				ExpectError: regexp.MustCompile(`networking address scopes found`),
			},
		},
	})
}

// -----------------------------------------------------------------------------
// Helpers.
// -----------------------------------------------------------------------------

func testAccCheckNetworkingAddressScopeV2Exists(found *addressscope.AddressScope) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		handle, err := testAccNetworkingNetworkV2Handle(s, testAccNetworkingAddressScopeV2ResourceName)
		if err != nil {
			return err
		}

		result, err := addressscope.Get(context.Background(), handle.client, handle.id)
		if err != nil {
			return err
		}

		*found = *result

		return nil
	}
}

func testAccCheckNetworkingAddressScopeV2SameID(found *addressscope.AddressScope) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[testAccNetworkingAddressScopeV2ResourceName]
		if !ok {
			return fmt.Errorf("not found: %s", testAccNetworkingAddressScopeV2ResourceName)
		}

		if rs.Primary.ID != found.ID {
			return fmt.Errorf("networking address scope was replaced: %s -> %s", found.ID, rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckNetworkingAddressScopeV2Replaced(found *addressscope.AddressScope) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[testAccNetworkingAddressScopeV2ResourceName]
		if !ok {
			return fmt.Errorf("not found: %s", testAccNetworkingAddressScopeV2ResourceName)
		}

		if rs.Primary.ID == found.ID {
			return fmt.Errorf("networking address scope %s was updated in place instead of replaced", found.ID)
		}

		found.ID = rs.Primary.ID

		return nil
	}
}

// testAccCheckNetworkingAddressScopeV2SharedUnchanged asserts the rename left the
// admin-managed shared value alone.
func testAccCheckNetworkingAddressScopeV2SharedUnchanged(found *addressscope.AddressScope) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		handle, err := testAccNetworkingNetworkV2Handle(s, testAccNetworkingAddressScopeV2ResourceName)
		if err != nil {
			return err
		}

		result, err := addressscope.Get(context.Background(), handle.client, handle.id)
		if err != nil {
			return err
		}

		if result.Shared != found.Shared {
			return fmt.Errorf("shared changed from %t to %t during a rename", found.Shared, result.Shared)
		}

		return nil
	}
}

func testAccCheckNetworkingAddressScopeV2RenameOutOfBand() resource.TestCheckFunc {
	return func(s *terraform.State) error {
		handle, err := testAccNetworkingNetworkV2Handle(s, testAccNetworkingAddressScopeV2ResourceName)
		if err != nil {
			return err
		}

		drifted := testAccNetworkingV2DriftedName
		_, err = addressscope.Update(context.Background(), handle.client, handle.id,
			addressscope.UpdateRequest{Name: &drifted})

		return err
	}
}

func testAccCheckNetworkingAddressScopeV2DeleteOutOfBand() resource.TestCheckFunc {
	return func(s *terraform.State) error {
		handle, err := testAccNetworkingNetworkV2Handle(s, testAccNetworkingAddressScopeV2ResourceName)
		if err != nil {
			return err
		}

		return addressscope.Delete(context.Background(), handle.client, handle.id)
	}
}

// networkingAddressScopeV2TestPool carries the identifier of the subnet pool
// created out of band between test steps.
type networkingAddressScopeV2TestPool struct {
	id string
}

func testAccCheckNetworkingAddressScopeV2CreatePool(pool *networkingAddressScopeV2TestPool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		handle, err := testAccNetworkingNetworkV2Handle(s, testAccNetworkingAddressScopeV2ResourceName)
		if err != nil {
			return err
		}

		name := "subnetpool-tf-acc-test-1"
		created, err := subnetpool.Create(context.Background(), handle.client, subnetpool.CreateRequest{
			Name:           &name,
			Prefixes:       []string{"192.168.200.0/24"},
			AddressScopeID: &handle.id,
		})
		if err != nil {
			return err
		}

		pool.id = created.ID

		return nil
	}
}

func testAccCheckNetworkingAddressScopeV2DeletePool(pool *networkingAddressScopeV2TestPool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if pool.id == "" {
			return errors.New("no subnet pool was created to delete")
		}

		handle, err := testAccNetworkingNetworkV2Handle(s, testAccNetworkingAddressScopeV2ResourceName)
		if err != nil {
			return err
		}

		return subnetpool.Delete(context.Background(), handle.client, pool.id)
	}
}

func testAccCheckNetworkingAddressScopeV2Destroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "selectel_networking_addressscope_v2" {
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

		if _, err := addressscope.Get(context.Background(), client, rs.Primary.ID); err == nil {
			return errors.New("networking address scope still exists")
		} else if !networkingIsNotFound(err) {
			return err
		}
	}

	return nil
}

// -----------------------------------------------------------------------------
// Configurations.
// -----------------------------------------------------------------------------

func testAccNetworkingAddressScopeV2Basic(projectID, region, name string) string {
	return testAccNetworkingV2Base(projectID) + fmt.Sprintf(`
resource "selectel_networking_addressscope_v2" "addressscope_tf_acc_test_1" {
  project_id = local.project_id
  region     = "%s"
  name       = "%s"
}
`, region, name)
}

func testAccNetworkingAddressScopeV2Version(projectID, region string, ipVersion int) string {
	return testAccNetworkingV2Base(projectID) + fmt.Sprintf(`
resource "selectel_networking_addressscope_v2" "addressscope_tf_acc_test_1" {
  project_id = local.project_id
  region     = "%s"
  name       = "scope-tf-acc-test-1"
  ip_version = %d
}
`, region, ipVersion)
}

func testAccNetworkingAddressScopeV2DataSource(projectID, region, name string) string {
	return testAccNetworkingAddressScopeV2Basic(projectID, region, name) + fmt.Sprintf(`
data "selectel_networking_addressscope_v2" "addressscope_tf_acc_test_1" {
  project_id = local.project_id
  region     = "%s"
  name       = selectel_networking_addressscope_v2.addressscope_tf_acc_test_1.name
}
`, region)
}

func testAccNetworkingAddressScopeV2DataSourceNoMatch(projectID, region string) string {
	return testAccNetworkingV2Base(projectID) + fmt.Sprintf(`
data "selectel_networking_addressscope_v2" "addressscope_tf_acc_test_1" {
  project_id = local.project_id
  region     = "%s"
  name       = "tf-acc-no-such-scope"
}
`, region)
}

// testAccNetworkingAddressScopeV2DataSourceAmbiguous creates two scopes of the
// same ip_version and searches on that alone.
func testAccNetworkingAddressScopeV2DataSourceAmbiguous(projectID, region string) string {
	return testAccNetworkingAddressScopeV2Basic(projectID, region, "scope-1") + fmt.Sprintf(`
resource "selectel_networking_addressscope_v2" "addressscope_tf_acc_test_2" {
  project_id = local.project_id
  region     = "%[1]s"
  name       = "scope-2"
}

data "selectel_networking_addressscope_v2" "addressscope_tf_acc_test_1" {
  project_id = local.project_id
  region     = "%[1]s"
  ip_version = 4

  depends_on = [
    selectel_networking_addressscope_v2.addressscope_tf_acc_test_1,
    selectel_networking_addressscope_v2.addressscope_tf_acc_test_2,
  ]
}
`, region)
}
