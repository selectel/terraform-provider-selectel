package selectel

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	vpc "github.com/selectel/vpc-go/pkg/v2"
	"github.com/selectel/vpc-go/pkg/v2/floatingip"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testAccNetworkingFloatingIPAssociateV2ResourceName = "selectel_vpc_floatingip_associate_v2.association_tf_acc_test_1"

func TestNetworkingFloatingIPAssociateV2Schema(t *testing.T) {
	resourceSchema := resourceVPCFloatingIPAssociateV2().Schema

	required := []string{"project_id", "region", "floating_ip", "port_id"}
	for _, name := range required {
		assert.True(t, resourceSchema[name].Required, name)
	}
	assert.True(t, resourceSchema["fixed_ip"].Optional)
	assert.True(t, resourceSchema["fixed_ip"].Computed)

	assert.True(t, resourceSchema["project_id"].ForceNew)
	assert.True(t, resourceSchema["region"].ForceNew)
	assert.True(t, resourceSchema["floating_ip"].ForceNew)
	assert.False(t, resourceSchema["port_id"].ForceNew)
	assert.False(t, resourceSchema["fixed_ip"].ForceNew)
}

func TestNetworkingFloatingIPAssociateV2Result(t *testing.T) {
	const address = "203.0.113.10"

	t.Run("none", func(t *testing.T) {
		_, err := networkingFloatingIPAssociateV2Result(nil, address)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "no floating ip")
	})

	t.Run("one", func(t *testing.T) {
		id, err := networkingFloatingIPAssociateV2Result(
			[]floatingip.FloatingIP{{ID: "floating-ip-id"}}, address)
		require.NoError(t, err)
		assert.Equal(t, "floating-ip-id", id)
	})

	t.Run("many", func(t *testing.T) {
		_, err := networkingFloatingIPAssociateV2Result(
			[]floatingip.FloatingIP{{ID: "first"}, {ID: "second"}}, address)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "more than one")
	})
}

func TestNetworkingFloatingIPAssociateV2Flatten(t *testing.T) {
	resourceData := schema.TestResourceDataRaw(
		t,
		resourceVPCFloatingIPAssociateV2().Schema,
		map[string]interface{}{
			"project_id":  "project-id",
			"region":      "ru-3",
			"floating_ip": "203.0.113.1",
			"port_id":     "old-port",
		},
	)
	portID := "new-port"
	fixedIP := "192.168.199.50"

	flattenNetworkingFloatingIPAssociateV2(resourceData, &floatingip.FloatingIP{
		FloatingIPAddress: "203.0.113.10",
		PortID:            &portID,
		FixedIPAddress:    &fixedIP,
	})

	assert.Equal(t, "203.0.113.10", resourceData.Get("floating_ip"))
	assert.Equal(t, "new-port", resourceData.Get("port_id"))
	assert.Equal(t, "192.168.199.50", resourceData.Get("fixed_ip"))

	flattenNetworkingFloatingIPAssociateV2(resourceData, &floatingip.FloatingIP{
		FloatingIPAddress: "203.0.113.10",
	})

	assert.Empty(t, resourceData.Get("port_id"))
	assert.Empty(t, resourceData.Get("fixed_ip"))
}

// TestAccNetworkingFloatingIPAssociateV2Basic covers SCN-097-01, SCN-098-01
// and SCN-101-01. It is intentionally not run without TF_ACC=1.
func TestAccNetworkingFloatingIPAssociateV2Basic(t *testing.T) {
	projectID, region := networkingV2TestEnv()

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheckWithProjectIDAndRegion(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckNetworkingFloatingIPAssociateV2Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkingFloatingIPAssociateV2(projectID, region, "port_a", ""),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(testAccNetworkingFloatingIPAssociateV2ResourceName, "floating_ip"),
					resource.TestCheckResourceAttrSet(testAccNetworkingFloatingIPAssociateV2ResourceName, "fixed_ip"),
					resource.TestCheckResourceAttrPair(
						testAccNetworkingFloatingIPAssociateV2ResourceName, "port_id",
						"selectel_vpc_port_v2.port_a", "id",
					),
				),
			},
			{
				Config:   testAccNetworkingFloatingIPAssociateV2(projectID, region, "port_a", ""),
				PlanOnly: true,
			},
			{
				ResourceName:      testAccNetworkingFloatingIPAssociateV2ResourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

// TestAccNetworkingFloatingIPAssociateV2Update covers SCN-097-02,
// SCN-099-01 and SCN-099-02: selecting a fixed address and moving the same
// floating IP are in-place updates.
func TestAccNetworkingFloatingIPAssociateV2Update(t *testing.T) {
	projectID, region := networkingV2TestEnv()
	var originalID string

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheckWithProjectIDAndRegion(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckNetworkingFloatingIPAssociateV2Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkingFloatingIPAssociateV2(projectID, region, "port_a", "192.168.199.70"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNetworkingFloatingIPAssociateV2CaptureID(&originalID),
					resource.TestCheckResourceAttr(testAccNetworkingFloatingIPAssociateV2ResourceName, "fixed_ip", "192.168.199.70"),
				),
			},
			{
				Config: testAccNetworkingFloatingIPAssociateV2(projectID, region, "port_b", "192.168.199.71"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNetworkingFloatingIPAssociateV2ID(&originalID),
					resource.TestCheckResourceAttr(testAccNetworkingFloatingIPAssociateV2ResourceName, "fixed_ip", "192.168.199.71"),
					resource.TestCheckResourceAttrPair(
						testAccNetworkingFloatingIPAssociateV2ResourceName, "port_id",
						"selectel_vpc_port_v2.port_b", "id",
					),
				),
			},
		},
	})
}

// TestAccNetworkingFloatingIPAssociateV2ExternalDetach covers SCN-100-01:
// refresh observes the detached state and plans to restore the configuration.
func TestAccNetworkingFloatingIPAssociateV2ExternalDetach(t *testing.T) {
	projectID, region := networkingV2TestEnv()

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheckWithProjectIDAndRegion(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckNetworkingFloatingIPAssociateV2Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkingFloatingIPAssociateV2(projectID, region, "port_a", ""),
			},
			{
				Config:             testAccNetworkingFloatingIPAssociateV2(projectID, region, "port_a", ""),
				Check:              testAccCheckNetworkingFloatingIPAssociateV2DetachOutOfBand(),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

// TestAccNetworkingFloatingIPAssociateV2ImportMissing covers SCN-101-02.
//
// The step carries a configuration even though nothing is applied: an import step
// needs the resource to be present in the configuration it imports into, and this
// is the first step, so there is no earlier one to inherit it from. The values are
// literals that are never used — the import fails while reading the identifier.
func TestAccNetworkingFloatingIPAssociateV2ImportMissing(t *testing.T) {
	projectID, region := networkingV2TestEnv()

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheckWithProjectIDAndRegion(t) },
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config:        testAccNetworkingFloatingIPAssociateV2Unmanaged(projectID, region),
				ResourceName:  testAccNetworkingFloatingIPAssociateV2ResourceName,
				ImportState:   true,
				ImportStateId: "00000000-0000-0000-0000-000000000000",
				// Read reports the identifier as absent, so the refusal comes from
				// Terraform rather than from the provider's own message.
				ExpectError: regexp.MustCompile(`Cannot import non-existent remote object`),
			},
		},
	})
}

func testAccCheckNetworkingFloatingIPAssociateV2CaptureID(target *string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		instance, ok := state.RootModule().Resources[testAccNetworkingFloatingIPAssociateV2ResourceName]
		if !ok || instance.Primary.ID == "" {
			return errors.New("networking floating ip association has no ID")
		}
		*target = instance.Primary.ID

		return nil
	}
}

// testAccCheckNetworkingFloatingIPAssociateV2ID takes a pointer because the checks
// of every step are built before the first one runs: comparing against a value
// captured here would compare against the empty string the earlier step had yet to
// fill in.
func testAccCheckNetworkingFloatingIPAssociateV2ID(expected *string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		instance, ok := state.RootModule().Resources[testAccNetworkingFloatingIPAssociateV2ResourceName]
		if !ok {
			return fmt.Errorf("not found: %s", testAccNetworkingFloatingIPAssociateV2ResourceName)
		}
		if *expected == "" {
			return errors.New("the association identifier of the earlier step was not captured")
		}
		if instance.Primary.ID != *expected {
			return fmt.Errorf("association was replaced: %s -> %s", *expected, instance.Primary.ID)
		}

		return nil
	}
}

func testAccCheckNetworkingFloatingIPAssociateV2DetachOutOfBand() resource.TestCheckFunc {
	return func(state *terraform.State) error {
		instance, ok := state.RootModule().Resources[testAccNetworkingFloatingIPAssociateV2ResourceName]
		if !ok {
			return fmt.Errorf("not found: %s", testAccNetworkingFloatingIPAssociateV2ResourceName)
		}
		client, err := networkingV2Client(
			testAccProvider.Meta(),
			instance.Primary.Attributes["project_id"],
			instance.Primary.Attributes["region"],
		)
		if err != nil {
			return err
		}
		_, err = floatingip.Update(context.Background(), client, instance.Primary.ID, floatingip.UpdateRequest{
			PortID: vpc.Null[string](),
		})

		return err
	}
}

func testAccCheckNetworkingFloatingIPAssociateV2Destroy(state *terraform.State) error {
	for _, instance := range state.RootModule().Resources {
		if instance.Type != "selectel_vpc_floatingip_associate_v2" {
			continue
		}
		client, err := networkingV2Client(
			testAccProvider.Meta(),
			instance.Primary.Attributes["project_id"],
			instance.Primary.Attributes["region"],
		)
		if err != nil {
			return err
		}
		found, err := floatingip.Get(context.Background(), client, instance.Primary.ID)
		if networkingIsNotFound(err) {
			continue
		}
		if err != nil {
			return err
		}
		if found.PortID != nil {
			return errors.New("networking floating ip is still associated")
		}
	}

	return nil
}

// testAccNetworkingFloatingIPAssociateV2Unmanaged declares the resource without
// creating anything: it exists so an import step has an address to import into.
func testAccNetworkingFloatingIPAssociateV2Unmanaged(projectID, region string) string {
	return testAccNetworkingV2Base(projectID) + fmt.Sprintf(`
resource "selectel_vpc_floatingip_associate_v2" "association_tf_acc_test_1" {
  project_id  = local.project_id
  region      = "%s"
  floating_ip = "203.0.113.1"
  port_id     = "00000000-0000-0000-0000-000000000000"
}
`, region)
}

func testAccNetworkingFloatingIPAssociateV2(
	projectID, region, selectedPort, fixedIP string,
) string {
	fixedIPArgument := ""
	if fixedIP != "" {
		fixedIPArgument = fmt.Sprintf("\n  fixed_ip    = %q", fixedIP)
	}

	return testAccNetworkingPortV2Subnet(projectID, region) + fmt.Sprintf(`
resource "selectel_vpc_floatingip_v2" "floatingip_for_association" {
  project_id = local.project_id
  region     = "%[1]s"
}

data "selectel_networking_floatingip_v2" "floatingip_for_association" {
  project_id    = local.project_id
  region        = "%[1]s"
  floatingip_id = selectel_vpc_floatingip_v2.floatingip_for_association.id
}

resource "selectel_vpc_router_v2" "router_tf_acc_test_1" {
  project_id          = local.project_id
  region              = "%[1]s"
  external_network_id = data.selectel_networking_floatingip_v2.floatingip_for_association.floating_network_id
}

resource "selectel_vpc_router_interface_v2" "interface_tf_acc_test_1" {
  project_id = local.project_id
  region     = "%[1]s"
  router_id  = selectel_vpc_router_v2.router_tf_acc_test_1.id
  subnet_id  = selectel_vpc_network_subnet_v2.subnet_tf_acc_test_1.id
}

resource "selectel_vpc_port_v2" "port_a" {
  project_id = local.project_id
  region     = "%[1]s"
  network_id = selectel_vpc_network_v2.network_tf_acc_test_1.id

  fixed_ip {
    subnet_id  = selectel_vpc_network_subnet_v2.subnet_tf_acc_test_1.id
    ip_address = "192.168.199.70"
  }
}

resource "selectel_vpc_port_v2" "port_b" {
  project_id = local.project_id
  region     = "%[1]s"
  network_id = selectel_vpc_network_v2.network_tf_acc_test_1.id

  fixed_ip {
    subnet_id  = selectel_vpc_network_subnet_v2.subnet_tf_acc_test_1.id
    ip_address = "192.168.199.71"
  }
}

resource "selectel_vpc_floatingip_associate_v2" "association_tf_acc_test_1" {
  project_id  = local.project_id
  region      = "%[1]s"
  floating_ip = selectel_vpc_floatingip_v2.floatingip_for_association.floating_ip_address
  port_id     = selectel_vpc_port_v2.%[2]s.id%[3]s

  # The route to the external network is what makes the association possible, and
  # nothing else in this resource refers to it.
  depends_on = [selectel_vpc_router_interface_v2.interface_tf_acc_test_1]
}
`,
		region,
		selectedPort,
		fixedIPArgument,
	)
}
