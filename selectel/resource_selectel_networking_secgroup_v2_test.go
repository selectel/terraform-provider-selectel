package selectel

import (
	"context"
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/selectel/vpc-go/pkg/v2/port"
	"github.com/selectel/vpc-go/pkg/v2/securitygroup"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testAccNetworkingSecGroupV2ResourceName   = "selectel_networking_secgroup_v2.secgroup_tf_acc_test_1"
	testAccNetworkingSecGroupV2DataSourceName = "data.selectel_networking_secgroup_v2.secgroup_tf_acc_test_1"
)

// -----------------------------------------------------------------------------
// Schema unit tests.
// -----------------------------------------------------------------------------

// TestNetworkingSecGroupV2Schema covers REQ-103, REQ-105 and REQ-106: what the
// caller owns, what replaces the group, and what is read-only.
func TestNetworkingSecGroupV2Schema(t *testing.T) {
	s := resourceNetworkingSecGroupV2().Schema

	t.Run("project_id, region and name are required", func(t *testing.T) {
		for _, name := range []string{"project_id", "region", "name"} {
			require.Contains(t, s, name)
			assert.True(t, s[name].Required)
		}
		// REQ-105: only project_id and region replace the group; name is updated.
		assert.False(t, s["name"].ForceNew)
	})

	// REQ-103: stateful needs Computed because its API default is true; an unset
	// attribute would otherwise land in the state as false and diff forever.
	t.Run("stateful is optional and computed", func(t *testing.T) {
		require.Contains(t, s, "stateful")
		assert.True(t, s["stateful"].Optional)
		assert.True(t, s["stateful"].Computed)
		assert.False(t, s["stateful"].ForceNew)
	})

	// REQ-103: description is Optional without Computed. Its API default is the
	// empty string, which is the zero value anyway, so Computed would buy nothing
	// and would stop the description from being cleared by omitting it — the
	// treatment every sibling capability already uses.
	t.Run("description is optional without computed", func(t *testing.T) {
		require.Contains(t, s, "description")
		assert.True(t, s["description"].Optional)
		assert.False(t, s["description"].Computed,
			"Computed would prevent clearing the description by removing it from the configuration")
		assert.False(t, s["description"].ForceNew)
	})

	// REQ-104: it only takes effect at create time.
	t.Run("delete_default_rules is optional, defaults to false and is ForceNew", func(t *testing.T) {
		require.Contains(t, s, "delete_default_rules")
		assert.True(t, s["delete_default_rules"].Optional)
		assert.Equal(t, false, s["delete_default_rules"].Default)
		assert.True(t, s["delete_default_rules"].ForceNew)
	})

	// REQ-105: exactly these three replace the group.
	t.Run("only the immutable attributes are ForceNew", func(t *testing.T) {
		forceNew := make([]string, 0)
		for name, attr := range s {
			if attr.ForceNew {
				forceNew = append(forceNew, name)
			}
		}

		assert.ElementsMatch(t, []string{"project_id", "region", "delete_default_rules"}, forceNew)
	})

	// REQ-106: shared is granted by an RBAC policy, not by this resource.
	// REQ-110: security_group_rules reports what is actually there.
	for _, name := range []string{"shared", "all_tags", "security_group_rules"} {
		t.Run(name+" is computed-only", func(t *testing.T) {
			require.Contains(t, s, name)
			assert.True(t, s[name].Computed)
			assert.False(t, s[name].Optional, name+" must not be a configuration argument")
			assert.False(t, s[name].Required)
		})
	}

	t.Run("tags are managed by terraform", func(t *testing.T) {
		require.Contains(t, s, "tags")
		assert.True(t, s["tags"].Optional)
		assert.False(t, s["tags"].ForceNew)
	})
}

// TestNetworkingSecGroupV2Flatten covers REQ-106 and REQ-110: the state reports
// what the API returned, including rules this resource never created.
func TestNetworkingSecGroupV2Flatten(t *testing.T) {
	d := schema.TestResourceDataRaw(t, resourceNetworkingSecGroupV2().Schema, map[string]interface{}{})

	flattenNetworkingSecGroupV2(d, &securitygroup.SecurityGroup{
		ID:          "sg-id",
		Name:        "web",
		Description: "web traffic",
		Stateful:    false,
		Shared:      true,
		Tags:        []string{"production"},
		SecurityGroupRules: []securitygroup.SecurityGroupRule{
			{ID: "rule-egress-v4"},
			{ID: "rule-egress-v6"},
			// An all-pass ingress rule the Neutron fork added itself.
			{ID: "rule-allpass-v4"},
		},
	})

	assert.Equal(t, "web", d.Get("name"))
	assert.Equal(t, "web traffic", d.Get("description"))
	assert.Equal(t, false, d.Get("stateful"))
	assert.Equal(t, true, d.Get("shared"))
	assert.Equal(t, []interface{}{"rule-egress-v4", "rule-egress-v6", "rule-allpass-v4"},
		d.Get("security_group_rules"))
}

// TestNetworkingSecGroupV2SetIDBeforeNonAtomicSteps covers SCN-104-03: the
// identifier has to be recorded before the default-rule deletion and the tag
// replacement, so a failure in either leaves the group tracked instead of
// orphaned.
//
// The order is asserted on the source, because a unit test cannot reach the
// point between the create call and its follow-up steps.
func TestNetworkingSecGroupV2SetIDBeforeNonAtomicSteps(t *testing.T) {
	file, err := parser.ParseFile(token.NewFileSet(),
		"resource_selectel_networking_secgroup_v2.go", nil, 0)
	require.NoError(t, err)

	var create *ast.FuncDecl
	for _, declaration := range file.Decls {
		function, ok := declaration.(*ast.FuncDecl)
		if ok && function.Name.Name == "resourceNetworkingSecGroupV2Create" {
			create = function

			break
		}
	}
	require.NotNil(t, create, "create function not found")

	// Positions of the three calls inside Create.
	var setIDPos, deleteRulesPos, tagsPos token.Pos
	ast.Inspect(create, func(node ast.Node) bool {
		call, ok := node.(*ast.CallExpr)
		if !ok {
			return true
		}
		switch fun := call.Fun.(type) {
		case *ast.SelectorExpr:
			if fun.Sel.Name == "SetId" {
				setIDPos = call.Pos()
			}
			if fun.Sel.Name == "Replace" {
				tagsPos = call.Pos()
			}
		case *ast.Ident:
			if fun.Name == "deleteNetworkingSecGroupV2DefaultRules" {
				deleteRulesPos = call.Pos()
			}
		}

		return true
	})

	require.NotZero(t, setIDPos, "Create must call d.SetId")
	require.NotZero(t, deleteRulesPos, "Create must delete the default rules")
	require.NotZero(t, tagsPos, "Create must replace the tags")

	assert.Less(t, int(setIDPos), int(deleteRulesPos),
		"d.SetId must run before the default rules are deleted")
	assert.Less(t, int(setIDPos), int(tagsPos),
		"d.SetId must run before the tags are replaced")
}

// TestNetworkingSecGroupV2DataSourceSelectors covers REQ-108.
func TestNetworkingSecGroupV2DataSourceSelectors(t *testing.T) {
	s := dataSourceNetworkingSecGroupV2().Schema

	selectors := make([]string, 0)
	for name, attr := range s {
		if attr.Optional || attr.Required {
			selectors = append(selectors, name)
		}
	}

	assert.ElementsMatch(t, []string{
		"project_id", "region", "secgroup_id", "name", "description", "stateful", "tags",
	}, selectors)

	// Returned, not selectable.
	for _, name := range []string{"shared", "all_tags", "security_group_rules"} {
		require.Contains(t, s, name)
		assert.True(t, s[name].Computed)
		assert.False(t, s[name].Optional, name+" must be returned, not selectable")
	}

	// delete_default_rules steers Create and has no meaning when reading.
	assert.NotContains(t, s, "delete_default_rules")
}

// -----------------------------------------------------------------------------
// Acceptance tests against a staging environment (TF_ACC=1).
// -----------------------------------------------------------------------------

// TestAccNetworkingSecGroupV2Basic covers SCN-103-01, SCN-103-02, SCN-104-02,
// SCN-107-01 and SCN-109-01.
func TestAccNetworkingSecGroupV2Basic(t *testing.T) {
	var found securitygroup.SecurityGroup
	projectID, region := networkingV2TestEnv()

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheckWithProjectIDAndRegion(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckNetworkingSecGroupV2Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkingSecGroupV2Basic(projectID, region, "secgroup-1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNetworkingSecGroupV2Exists(&found),
					// SCN-103-01: the API defaults land in the state.
					resource.TestCheckResourceAttrSet(testAccNetworkingSecGroupV2ResourceName, "stateful"),
					resource.TestCheckResourceAttrSet(testAccNetworkingSecGroupV2ResourceName, "shared"),
					// SCN-104-02: the default rules the API created are untouched.
					testAccCheckNetworkingSecGroupV2RuleCountAtLeast(1),
				),
			},
			{
				// SCN-103-02: the follow-up plan is empty.
				Config:   testAccNetworkingSecGroupV2Basic(projectID, region, "secgroup-1"),
				PlanOnly: true,
			},
			{
				// SCN-107-01: import by the identifier alone.
				ResourceName:      testAccNetworkingSecGroupV2ResourceName,
				ImportState:       true,
				ImportStateVerify: true,
				// delete_default_rules steers Create only and has no remote
				// counterpart to import.
				ImportStateVerifyIgnore: []string{"delete_default_rules"},
			},
		},
	})
}

// TestAccNetworkingSecGroupV2DeleteDefaultRules covers SCN-104-01: the rules the
// API created with the group are removed right after it is created.
func TestAccNetworkingSecGroupV2DeleteDefaultRules(t *testing.T) {
	var found securitygroup.SecurityGroup
	projectID, region := networkingV2TestEnv()

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheckWithProjectIDAndRegion(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckNetworkingSecGroupV2Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkingSecGroupV2DeleteDefaultRules(projectID, region),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNetworkingSecGroupV2Exists(&found),
					resource.TestCheckResourceAttr(testAccNetworkingSecGroupV2ResourceName, "security_group_rules.#", "0"),
					testAccCheckNetworkingSecGroupV2RuleCount(0),
				),
			},
			{
				Config:   testAccNetworkingSecGroupV2DeleteDefaultRules(projectID, region),
				PlanOnly: true,
			},
		},
	})
}

// TestAccNetworkingSecGroupV2Update covers SCN-105-01 and SCN-105-03.
func TestAccNetworkingSecGroupV2Update(t *testing.T) {
	var found securitygroup.SecurityGroup
	projectID, region := networkingV2TestEnv()

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheckWithProjectIDAndRegion(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckNetworkingSecGroupV2Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkingSecGroupV2Described(projectID, region, "secgroup-1", "web traffic", true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNetworkingSecGroupV2Exists(&found),
					resource.TestCheckResourceAttr(testAccNetworkingSecGroupV2ResourceName, "stateful", "true"),
					resource.TestCheckResourceAttr(testAccNetworkingSecGroupV2ResourceName, "tags.#", "2"),
				),
			},
			{
				// SCN-105-01: name, description, stateful and tags all change in
				// place, and the group keeps its identifier.
				Config: testAccNetworkingSecGroupV2Described(projectID, region, "secgroup-1-renamed", "changed", false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNetworkingSecGroupV2SameID(&found),
					resource.TestCheckResourceAttr(testAccNetworkingSecGroupV2ResourceName, "name", "secgroup-1-renamed"),
					resource.TestCheckResourceAttr(testAccNetworkingSecGroupV2ResourceName, "description", "changed"),
					resource.TestCheckResourceAttr(testAccNetworkingSecGroupV2ResourceName, "stateful", "false"),
				),
			},
			{
				// SCN-105-03: delete_default_rules replaces the group.
				Config: testAccNetworkingSecGroupV2DeleteDefaultRules(projectID, region),
				Check:  testAccCheckNetworkingSecGroupV2Replaced(&found),
			},
		},
	})
}

// TestAccNetworkingSecGroupV2ExternalChanges covers SCN-106-01 and SCN-106-02.
func TestAccNetworkingSecGroupV2ExternalChanges(t *testing.T) {
	var found securitygroup.SecurityGroup
	projectID, region := networkingV2TestEnv()

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheckWithProjectIDAndRegion(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckNetworkingSecGroupV2Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkingSecGroupV2Basic(projectID, region, "secgroup-1"),
				Check:  testAccCheckNetworkingSecGroupV2Exists(&found),
			},
			{
				Config:             testAccNetworkingSecGroupV2Basic(projectID, region, "secgroup-1"),
				Check:              testAccCheckNetworkingSecGroupV2RenameOutOfBand(),
				ExpectNonEmptyPlan: true,
			},
			{
				Config:             testAccNetworkingSecGroupV2Basic(projectID, region, "secgroup-1"),
				Check:              testAccCheckNetworkingSecGroupV2DeleteOutOfBand(),
				ExpectNonEmptyPlan: true,
			},
			{
				Config: testAccNetworkingSecGroupV2Basic(projectID, region, "secgroup-1"),
				Check:  testAccCheckNetworkingSecGroupV2Exists(&found),
			},
		},
	})
}

// TestAccNetworkingSecGroupV2DeleteConflict covers SCN-109-03: a group attached
// to a port is not deleted.
//
// The port is created through the SDK, because it has to outlive the group for
// the destroy to hit the conflict.
func TestAccNetworkingSecGroupV2DeleteConflict(t *testing.T) {
	projectID, region := networkingV2TestEnv()
	attached := &networkingSecGroupV2TestPort{}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheckWithProjectIDAndRegion(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckNetworkingSecGroupV2Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkingSecGroupV2WithNetwork(projectID, region),
				Check:  testAccCheckNetworkingSecGroupV2AttachPort(attached),
			},
			{
				Config:      testAccNetworkingSecGroupV2WithNetwork(projectID, region),
				Destroy:     true,
				ExpectError: regexp.MustCompile(`(?i)conflict|in use`),
			},
			{
				// Detach so the final destroy can succeed.
				Config: testAccNetworkingSecGroupV2WithNetwork(projectID, region),
				Check:  testAccCheckNetworkingSecGroupV2DeletePort(attached),
			},
		},
	})
}

// TestAccNetworkingSecGroupV2DataSource covers SCN-108-01, SCN-108-02 and
// SCN-108-03.
func TestAccNetworkingSecGroupV2DataSource(t *testing.T) {
	projectID, region := networkingV2TestEnv()
	// The project is shared, so the name the data source searches on must be
	// unique to this run.
	secGroupName := acctest.RandomWithPrefix("tf-acc-secgroup")

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheckWithProjectIDAndRegion(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckNetworkingSecGroupV2Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkingSecGroupV2DataSource(projectID, region, secGroupName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(
						testAccNetworkingSecGroupV2DataSourceName, "secgroup_id",
						testAccNetworkingSecGroupV2ResourceName, "id",
					),
					resource.TestCheckResourceAttrPair(
						testAccNetworkingSecGroupV2DataSourceName, "shared",
						testAccNetworkingSecGroupV2ResourceName, "shared",
					),
					resource.TestCheckResourceAttrPair(
						testAccNetworkingSecGroupV2DataSourceName, "security_group_rules.#",
						testAccNetworkingSecGroupV2ResourceName, "security_group_rules.#",
					),
				),
			},
			{
				Config:      testAccNetworkingSecGroupV2DataSourceNoMatch(projectID, region),
				ExpectError: regexp.MustCompile(`no networking security group found`),
			},
		},
	})
}

// -----------------------------------------------------------------------------
// Helpers.
// -----------------------------------------------------------------------------

func testAccCheckNetworkingSecGroupV2Exists(found *securitygroup.SecurityGroup) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		handle, err := testAccNetworkingNetworkV2Handle(s, testAccNetworkingSecGroupV2ResourceName)
		if err != nil {
			return err
		}

		result, err := securitygroup.Get(context.Background(), handle.client, handle.id)
		if err != nil {
			return err
		}

		*found = *result

		return nil
	}
}

func testAccCheckNetworkingSecGroupV2SameID(found *securitygroup.SecurityGroup) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[testAccNetworkingSecGroupV2ResourceName]
		if !ok {
			return fmt.Errorf("not found: %s", testAccNetworkingSecGroupV2ResourceName)
		}

		if rs.Primary.ID != found.ID {
			return fmt.Errorf("networking security group was replaced: %s -> %s", found.ID, rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckNetworkingSecGroupV2Replaced(found *securitygroup.SecurityGroup) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[testAccNetworkingSecGroupV2ResourceName]
		if !ok {
			return fmt.Errorf("not found: %s", testAccNetworkingSecGroupV2ResourceName)
		}

		if rs.Primary.ID == found.ID {
			return fmt.Errorf("networking security group %s was updated in place instead of replaced", found.ID)
		}

		found.ID = rs.Primary.ID

		return nil
	}
}

func testAccCheckNetworkingSecGroupV2RuleCount(expected int) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		handle, err := testAccNetworkingNetworkV2Handle(s, testAccNetworkingSecGroupV2ResourceName)
		if err != nil {
			return err
		}

		found, err := securitygroup.Get(context.Background(), handle.client, handle.id)
		if err != nil {
			return err
		}

		if len(found.SecurityGroupRules) != expected {
			return fmt.Errorf("the group has %d rules, expected %d",
				len(found.SecurityGroupRules), expected)
		}

		return nil
	}
}

// testAccCheckNetworkingSecGroupV2RuleCountAtLeast is used where the exact count
// depends on the region: the fork adds all-pass rules only where
// security_group_all_pass is enabled, which the API does not expose.
func testAccCheckNetworkingSecGroupV2RuleCountAtLeast(minimum int) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		handle, err := testAccNetworkingNetworkV2Handle(s, testAccNetworkingSecGroupV2ResourceName)
		if err != nil {
			return err
		}

		found, err := securitygroup.Get(context.Background(), handle.client, handle.id)
		if err != nil {
			return err
		}

		if len(found.SecurityGroupRules) < minimum {
			return fmt.Errorf("the group has %d rules, expected at least %d",
				len(found.SecurityGroupRules), minimum)
		}

		return nil
	}
}

func testAccCheckNetworkingSecGroupV2RenameOutOfBand() resource.TestCheckFunc {
	return func(s *terraform.State) error {
		handle, err := testAccNetworkingNetworkV2Handle(s, testAccNetworkingSecGroupV2ResourceName)
		if err != nil {
			return err
		}

		drifted := testAccNetworkingV2DriftedName
		_, err = securitygroup.Update(context.Background(), handle.client, handle.id,
			securitygroup.UpdateRequest{Name: &drifted})

		return err
	}
}

func testAccCheckNetworkingSecGroupV2DeleteOutOfBand() resource.TestCheckFunc {
	return func(s *terraform.State) error {
		handle, err := testAccNetworkingNetworkV2Handle(s, testAccNetworkingSecGroupV2ResourceName)
		if err != nil {
			return err
		}

		return securitygroup.Delete(context.Background(), handle.client, handle.id)
	}
}

// networkingSecGroupV2TestPort carries the port created out of band between test
// steps, so it can outlive the group under test.
type networkingSecGroupV2TestPort struct {
	id string
}

func testAccCheckNetworkingSecGroupV2AttachPort(created *networkingSecGroupV2TestPort) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		handle, err := testAccNetworkingNetworkV2Handle(s, testAccNetworkingSecGroupV2ResourceName)
		if err != nil {
			return err
		}

		network, ok := s.RootModule().Resources["selectel_networking_network_v2.network_tf_acc_test_1"]
		if !ok {
			return errors.New("not found: selectel_networking_network_v2.network_tf_acc_test_1")
		}

		result, err := port.Create(context.Background(), handle.client, port.CreateRequest{
			NetworkID:      network.Primary.ID,
			SecurityGroups: &[]string{handle.id},
		})
		if err != nil {
			return err
		}

		created.id = result.ID

		return nil
	}
}

func testAccCheckNetworkingSecGroupV2DeletePort(created *networkingSecGroupV2TestPort) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if created.id == "" {
			return errors.New("no port was created to delete")
		}

		handle, err := testAccNetworkingNetworkV2Handle(s, testAccNetworkingSecGroupV2ResourceName)
		if err != nil {
			return err
		}

		return port.Delete(context.Background(), handle.client, created.id)
	}
}

func testAccCheckNetworkingSecGroupV2Destroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "selectel_networking_secgroup_v2" {
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

		if _, err := securitygroup.Get(context.Background(), client, rs.Primary.ID); err == nil {
			return errors.New("networking security group still exists")
		} else if !networkingIsNotFound(err) {
			return err
		}
	}

	return nil
}

// -----------------------------------------------------------------------------
// Configurations.
// -----------------------------------------------------------------------------

func testAccNetworkingSecGroupV2Basic(projectID, region, name string) string {
	return testAccNetworkingV2Base(projectID) + fmt.Sprintf(`
resource "selectel_networking_secgroup_v2" "secgroup_tf_acc_test_1" {
  project_id = local.project_id
  region     = "%s"
  name       = "%s"
}
`, region, name)
}

func testAccNetworkingSecGroupV2Described(projectID, region, name, description string, stateful bool) string {
	return testAccNetworkingV2Base(projectID) + fmt.Sprintf(`
resource "selectel_networking_secgroup_v2" "secgroup_tf_acc_test_1" {
  project_id  = local.project_id
  region      = "%s"
  name        = "%s"
  description = "%s"
  stateful    = %t
  tags        = ["tf-acc", "networking"]
}
`, region, name, description, stateful)
}

func testAccNetworkingSecGroupV2DeleteDefaultRules(projectID, region string) string {
	return testAccNetworkingV2Base(projectID) + fmt.Sprintf(`
resource "selectel_networking_secgroup_v2" "secgroup_tf_acc_test_1" {
  project_id           = local.project_id
  region               = "%s"
  name                 = "secgroup-tf-acc-test-1"
  delete_default_rules = true
}
`, region)
}

// testAccNetworkingSecGroupV2WithNetwork adds the network and subnet the
// out-of-band port needs.
func testAccNetworkingSecGroupV2WithNetwork(projectID, region string) string {
	return testAccNetworkingSecGroupV2Basic(projectID, region, "secgroup-1") + fmt.Sprintf(`
resource "selectel_networking_network_v2" "network_tf_acc_test_1" {
  project_id = local.project_id
  region     = "%[1]s"
  name       = "network-tf-acc-test-1"
}

resource "selectel_networking_subnet_v2" "subnet_tf_acc_test_1" {
  project_id = local.project_id
  region     = "%[1]s"
  network_id = selectel_networking_network_v2.network_tf_acc_test_1.id
  cidr       = "192.168.199.0/24"
}
`, region)
}

func testAccNetworkingSecGroupV2DataSource(projectID, region, name string) string {
	return testAccNetworkingSecGroupV2Basic(projectID, region, name) + fmt.Sprintf(`
data "selectel_networking_secgroup_v2" "secgroup_tf_acc_test_1" {
  project_id = local.project_id
  region     = "%s"
  name       = selectel_networking_secgroup_v2.secgroup_tf_acc_test_1.name
}
`, region)
}

func testAccNetworkingSecGroupV2DataSourceNoMatch(projectID, region string) string {
	return testAccNetworkingV2Base(projectID) + fmt.Sprintf(`
data "selectel_networking_secgroup_v2" "secgroup_tf_acc_test_1" {
  project_id = local.project_id
  region     = "%s"
  name       = "tf-acc-no-such-secgroup"
}
`, region)
}
