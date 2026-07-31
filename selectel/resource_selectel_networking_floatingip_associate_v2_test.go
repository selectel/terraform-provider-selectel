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
	vpc "github.com/selectel/vpc-go/pkg/v2"
	"github.com/selectel/vpc-go/pkg/v2/floatingip"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testAccNetworkingFloatingIPAssociateV2ResourceName = "selectel_networking_floatingip_associate_v2.association_tf_acc_test_1"

func TestNetworkingFloatingIPAssociateV2Schema(t *testing.T) {
	resourceSchema := resourceNetworkingFloatingIPAssociateV2().Schema

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
		resourceNetworkingFloatingIPAssociateV2().Schema,
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
	pool := networkingFloatingIPV2TestPool(t)
	projectName := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckNetworkingFloatingIPAssociateV2Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkingFloatingIPAssociateV2(projectName, pool, "port_a", ""),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(testAccNetworkingFloatingIPAssociateV2ResourceName, "floating_ip"),
					resource.TestCheckResourceAttrSet(testAccNetworkingFloatingIPAssociateV2ResourceName, "fixed_ip"),
					resource.TestCheckResourceAttrPair(
						testAccNetworkingFloatingIPAssociateV2ResourceName, "port_id",
						"selectel_networking_port_v2.port_a", "id",
					),
				),
			},
			{
				Config:   testAccNetworkingFloatingIPAssociateV2(projectName, pool, "port_a", ""),
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
	pool := networkingFloatingIPV2TestPool(t)
	projectName := acctest.RandomWithPrefix("tf-acc")
	var originalID string

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckNetworkingFloatingIPAssociateV2Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkingFloatingIPAssociateV2(projectName, pool, "port_a", "192.168.199.70"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNetworkingFloatingIPAssociateV2CaptureID(&originalID),
					resource.TestCheckResourceAttr(testAccNetworkingFloatingIPAssociateV2ResourceName, "fixed_ip", "192.168.199.70"),
				),
			},
			{
				Config: testAccNetworkingFloatingIPAssociateV2(projectName, pool, "port_b", "192.168.199.71"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNetworkingFloatingIPAssociateV2ID(originalID),
					resource.TestCheckResourceAttr(testAccNetworkingFloatingIPAssociateV2ResourceName, "fixed_ip", "192.168.199.71"),
					resource.TestCheckResourceAttrPair(
						testAccNetworkingFloatingIPAssociateV2ResourceName, "port_id",
						"selectel_networking_port_v2.port_b", "id",
					),
				),
			},
		},
	})
}

// TestAccNetworkingFloatingIPAssociateV2ExternalDetach covers SCN-100-01:
// refresh observes the detached state and plans to restore the configuration.
func TestAccNetworkingFloatingIPAssociateV2ExternalDetach(t *testing.T) {
	pool := networkingFloatingIPV2TestPool(t)
	projectName := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckNetworkingFloatingIPAssociateV2Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkingFloatingIPAssociateV2(projectName, pool, "port_a", ""),
			},
			{
				Config:             testAccNetworkingFloatingIPAssociateV2(projectName, pool, "port_a", ""),
				Check:              testAccCheckNetworkingFloatingIPAssociateV2DetachOutOfBand(),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

// TestAccNetworkingFloatingIPAssociateV2ImportMissing covers SCN-101-02.
func TestAccNetworkingFloatingIPAssociateV2ImportMissing(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheck(t) },
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				ResourceName:  testAccNetworkingFloatingIPAssociateV2ResourceName,
				ImportState:   true,
				ImportStateId: "00000000-0000-0000-0000-000000000000",
				ExpectError:   regexp.MustCompile(`error getting networking floating ip association`),
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

func testAccCheckNetworkingFloatingIPAssociateV2ID(expected string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		instance, ok := state.RootModule().Resources[testAccNetworkingFloatingIPAssociateV2ResourceName]
		if !ok {
			return fmt.Errorf("not found: %s", testAccNetworkingFloatingIPAssociateV2ResourceName)
		}
		if instance.Primary.ID != expected {
			return fmt.Errorf("association was replaced: %s -> %s", expected, instance.Primary.ID)
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
		if instance.Type != "selectel_networking_floatingip_associate_v2" {
			continue
		}
		client, err := networkingV2Client(
			testAccProvider.Meta(),
			instance.Primary.Attributes["project_id"],
			instance.Primary.Attributes["region"],
		)
		if err != nil {
			return nil //nolint:nilerr // project cleanup removed all networking objects
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

func testAccNetworkingFloatingIPAssociateV2(
	projectName, pool, selectedPort, fixedIP string,
) string {
	fixedIPArgument := ""
	if fixedIP != "" {
		fixedIPArgument = fmt.Sprintf("\n  fixed_ip    = %q", fixedIP)
	}

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

resource "selectel_networking_floatingip_v2" "floatingip_for_association" {
  project_id = selectel_vpc_project_v2.project_tf_acc_test_1.id
  region     = "%s"
  pool       = "%s"
}

resource "selectel_networking_floatingip_associate_v2" "association_tf_acc_test_1" {
  project_id  = selectel_vpc_project_v2.project_tf_acc_test_1.id
  region      = "%s"
  floating_ip = selectel_networking_floatingip_v2.floatingip_for_association.address
  port_id     = selectel_networking_port_v2.%s.id%s
}
`,
		networkingV2TestRegion(),
		networkingV2TestRegion(),
		networkingV2TestRegion(),
		pool,
		networkingV2TestRegion(),
		selectedPort,
		fixedIPArgument,
	)
}
