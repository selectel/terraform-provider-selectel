package selectel

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	vpc "github.com/selectel/vpc-go/pkg/v2"
	"github.com/selectel/vpc-go/pkg/v2/floatingip"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testAccVPCFloatingIPAssociateV2ResourceName = "selectel_vpc_floatingip_associate_v2.association_tf_acc_test_1"

func TestVPCFloatingIPAssociateV2Result(t *testing.T) {
	const address = "203.0.113.10"

	t.Run("none", func(t *testing.T) {
		_, err := vpcFloatingIPAssociateV2Result(nil, address)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "no floating ip")
	})

	t.Run("one", func(t *testing.T) {
		id, err := vpcFloatingIPAssociateV2Result(
			[]floatingip.FloatingIP{{ID: "floating-ip-id"}}, address)
		require.NoError(t, err)
		assert.Equal(t, "floating-ip-id", id)
	})

	t.Run("many", func(t *testing.T) {
		_, err := vpcFloatingIPAssociateV2Result(
			[]floatingip.FloatingIP{{ID: "first"}, {ID: "second"}}, address)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "more than one")
	})
}

func TestVPCFloatingIPAssociateV2Flatten(t *testing.T) {
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

	flattenVPCFloatingIPAssociateV2(resourceData, &floatingip.FloatingIP{
		FloatingIPAddress: "203.0.113.10",
		PortID:            &portID,
		FixedIPAddress:    &fixedIP,
	})

	assert.Equal(t, "203.0.113.10", resourceData.Get("floating_ip"))
	assert.Equal(t, "new-port", resourceData.Get("port_id"))
	assert.Equal(t, "192.168.199.50", resourceData.Get("fixed_ip"))

	flattenVPCFloatingIPAssociateV2(resourceData, &floatingip.FloatingIP{
		FloatingIPAddress: "203.0.113.10",
	})

	assert.Empty(t, resourceData.Get("port_id"))
	assert.Empty(t, resourceData.Get("fixed_ip"))
}

func TestAccVPCFloatingIPAssociateV2Basic(t *testing.T) {
	projectID, region := vpcV2TestEnv()

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheckWithProjectIDAndRegion(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckVPCFloatingIPAssociateV2Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCFloatingIPAssociateV2(projectID, region, "port_a", ""),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(testAccVPCFloatingIPAssociateV2ResourceName, "floating_ip"),
					resource.TestCheckResourceAttrSet(testAccVPCFloatingIPAssociateV2ResourceName, "fixed_ip"),
					resource.TestCheckResourceAttrPair(
						testAccVPCFloatingIPAssociateV2ResourceName, "port_id",
						"selectel_vpc_port_v2.port_a", "id",
					),
				),
			},
			{
				Config:   testAccVPCFloatingIPAssociateV2(projectID, region, "port_a", ""),
				PlanOnly: true,
			},
			{
				ResourceName:      testAccVPCFloatingIPAssociateV2ResourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccVPCFloatingIPAssociateV2Update(t *testing.T) {
	projectID, region := vpcV2TestEnv()
	var originalID string

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheckWithProjectIDAndRegion(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckVPCFloatingIPAssociateV2Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCFloatingIPAssociateV2(projectID, region, "port_a", "192.168.199.70"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCFloatingIPAssociateV2CaptureID(&originalID),
					resource.TestCheckResourceAttr(testAccVPCFloatingIPAssociateV2ResourceName, "fixed_ip", "192.168.199.70"),
				),
			},
			{
				Config: testAccVPCFloatingIPAssociateV2(projectID, region, "port_b", "192.168.199.71"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCFloatingIPAssociateV2ID(&originalID),
					resource.TestCheckResourceAttr(testAccVPCFloatingIPAssociateV2ResourceName, "fixed_ip", "192.168.199.71"),
					resource.TestCheckResourceAttrPair(
						testAccVPCFloatingIPAssociateV2ResourceName, "port_id",
						"selectel_vpc_port_v2.port_b", "id",
					),
				),
			},
		},
	})
}

func testAccCheckVPCFloatingIPAssociateV2CaptureID(target *string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		instance, ok := state.RootModule().Resources[testAccVPCFloatingIPAssociateV2ResourceName]
		if !ok || instance.Primary.ID == "" {
			return errors.New("VPC floating ip association has no ID")
		}
		*target = instance.Primary.ID

		return nil
	}
}

// Dereference at check time because an earlier step populates the ID.
func testAccCheckVPCFloatingIPAssociateV2ID(expected *string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		instance, ok := state.RootModule().Resources[testAccVPCFloatingIPAssociateV2ResourceName]
		if !ok {
			return fmt.Errorf("not found: %s", testAccVPCFloatingIPAssociateV2ResourceName)
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

func testAccCheckVPCFloatingIPAssociateV2Destroy(state *terraform.State) error {
	for _, instance := range state.RootModule().Resources {
		if instance.Type != "selectel_vpc_floatingip_associate_v2" {
			continue
		}
		client, err := vpcV2Client(
			testAccProvider.Meta(),
			instance.Primary.Attributes["project_id"],
			instance.Primary.Attributes["region"],
		)
		if err != nil {
			return err
		}
		found, err := floatingip.Get(context.Background(), client, instance.Primary.ID)
		if vpc.IsErrorClass(err, vpc.ErrorClassNotFound) {
			continue
		}
		if err != nil {
			return err
		}
		if found.PortID != nil {
			return errors.New("VPC floating ip is still associated")
		}
	}

	return nil
}

func testAccVPCFloatingIPAssociateV2(
	projectID, region, selectedPort, fixedIP string,
) string {
	fixedIPArgument := ""
	if fixedIP != "" {
		fixedIPArgument = fmt.Sprintf("\n  fixed_ip    = %q", fixedIP)
	}

	return testAccVPCPortV2Subnet(projectID, region) + fmt.Sprintf(`
resource "selectel_vpc_floatingip_v2" "floatingip_for_association" {
  project_id = local.project_id
  region     = "%[1]s"
}

data "selectel_vpc_network_v2" "external" {
  project_id = local.project_id
  region     = "%[1]s"
  external   = true
}

resource "selectel_vpc_router_v2" "router_tf_acc_test_1" {
  project_id          = local.project_id
  region              = "%[1]s"
  external_network_id = data.selectel_vpc_network_v2.external.network_id
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
