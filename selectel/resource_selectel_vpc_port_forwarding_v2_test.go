package selectel

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	vpc "github.com/selectel/vpc-go/pkg/v2"
	"github.com/selectel/vpc-go/pkg/v2/floatingip"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testAccVPCPortForwardingV2ResourceName = "selectel_vpc_port_forwarding_v2.pf_tf_acc_test_1"

func TestVPCPortForwardingV2ID(t *testing.T) {
	t.Run("splits a well-formed identifier", func(t *testing.T) {
		fipID, ruleID, err := parseVPCPortForwardingV2ID(
			"1a2b3c4d-5e6f-7081-92a3-b4c5d6e7f809/9f8e7d6c-5b4a-3928-1706-f5e4d3c2b1a0")

		require.NoError(t, err)
		assert.Equal(t, "1a2b3c4d-5e6f-7081-92a3-b4c5d6e7f809", fipID)
		assert.Equal(t, "9f8e7d6c-5b4a-3928-1706-f5e4d3c2b1a0", ruleID)
	})

	for _, c := range []struct{ name, id string }{
		{name: "no separator", id: "1a2b3c4d-5e6f"},
		{name: "empty floating ip", id: "/9f8e7d6c"},
		{name: "empty rule", id: "1a2b3c4d/"},
		{name: "three parts", id: "1a2b/3c4d/5e6f"},
		{name: "empty identifier", id: ""},
	} {
		t.Run("rejects "+c.name, func(t *testing.T) {
			_, _, err := parseVPCPortForwardingV2ID(c.id)

			require.Error(t, err)
			assert.Contains(t, err.Error(), "expected the format")
		})
	}
}

func TestVPCPortForwardingV2PortRangeValidation(t *testing.T) {
	for _, c := range []struct {
		value   string
		wantErr string
	}{
		{value: "80:90"},
		{value: "80:80"},
		{value: "1:65535"},
		{value: ""},
		{value: "80", wantErr: "expected <first>:<last>"},
		{value: "80:90:100", wantErr: "expected <first>:<last>"},
		{value: "abc:90", wantErr: "between 1 and 65535"},
		{value: "0:10", wantErr: "between 1 and 65535"},
		{value: "10:70000", wantErr: "between 1 and 65535"},
		{value: "90:80", wantErr: "must not exceed"},
	} {
		t.Run(c.value, func(t *testing.T) {
			_, errs := validateVPCPortForwardingV2PortRange(c.value, "internal_port_range")

			if c.wantErr == "" {
				assert.Empty(t, errs)

				return
			}

			require.NotEmpty(t, errs)
			assert.Contains(t, errs[0].Error(), c.wantErr)
		})
	}
}

func TestAccVPCPortForwardingV2Basic(t *testing.T) {
	var found floatingip.PortForwarding
	projectID, region := vpcV2TestEnv()

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheckWithProjectIDAndRegion(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckVPCPortForwardingV2Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCPortForwardingV2Basic(projectID, region, 8080, "tcp"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCPortForwardingV2Exists(&found),
					resource.TestCheckResourceAttr(testAccVPCPortForwardingV2ResourceName, "external_port", "8080"),
					resource.TestCheckResourceAttr(testAccVPCPortForwardingV2ResourceName, "external_port_range", "8080:8080"),
					resource.TestCheckResourceAttr(testAccVPCPortForwardingV2ResourceName, "internal_port_range", "80:80"),
					resource.TestCheckResourceAttr(testAccVPCPortForwardingV2ResourceName, "protocol", "tcp"),
					resource.TestCheckResourceAttrSet(testAccVPCPortForwardingV2ResourceName, "internal_ip_address"),
				),
			},
			{
				Config:   testAccVPCPortForwardingV2Basic(projectID, region, 8080, "tcp"),
				PlanOnly: true,
			},
			{
				ResourceName:      testAccVPCPortForwardingV2ResourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: testAccVPCPortForwardingV2ImportID,
			},
			{
				Config:             testAccVPCPortForwardingV2Basic(projectID, region, 8080, "tcp"),
				Check:              testAccCheckVPCPortForwardingV2DeleteOutOfBand(),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccVPCPortForwardingV2Update(t *testing.T) {
	var found floatingip.PortForwarding
	projectID, region := vpcV2TestEnv()

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheckWithProjectIDAndRegion(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckVPCPortForwardingV2Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCPortForwardingV2Ranges(projectID, region, "9090:9092", "tcp"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCPortForwardingV2Exists(&found),
					resource.TestCheckResourceAttr(testAccVPCPortForwardingV2ResourceName, "internal_port_range", "80:82"),
					resource.TestCheckResourceAttr(testAccVPCPortForwardingV2ResourceName, "external_port_range", "9090:9092"),
					resource.TestCheckResourceAttr(testAccVPCPortForwardingV2ResourceName, "internal_port", "0"),
					resource.TestCheckResourceAttr(testAccVPCPortForwardingV2ResourceName, "external_port", "0"),
					resource.TestCheckResourceAttr(testAccVPCPortForwardingV2ResourceName, "description", "updated by the acceptance test"),
					resource.TestCheckResourceAttrPair(
						testAccVPCPortForwardingV2ResourceName, "floatingip_id",
						"selectel_vpc_floatingip_v2.floatingip_tf_acc_test_1", "id",
					),
					resource.TestCheckResourceAttrPair(
						testAccVPCPortForwardingV2ResourceName, "internal_port_id",
						testAccVPCPortV2ResourceName, "id",
					),
				),
			},
			{
				Config: testAccVPCPortForwardingV2Basic(projectID, region, 8080, "tcp"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCPortForwardingV2SameID(&found),
					resource.TestCheckResourceAttr(testAccVPCPortForwardingV2ResourceName, "internal_port", "80"),
					resource.TestCheckResourceAttr(testAccVPCPortForwardingV2ResourceName, "external_port", "8080"),
					resource.TestCheckResourceAttr(testAccVPCPortForwardingV2ResourceName, "description", ""),
				),
			},
			{
				Config: testAccVPCPortForwardingV2Ranges(projectID, region, "9090:9092", "udp"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCPortForwardingV2SameID(&found),
					resource.TestCheckResourceAttr(testAccVPCPortForwardingV2ResourceName, "external_port_range", "9090:9092"),
					resource.TestCheckResourceAttr(testAccVPCPortForwardingV2ResourceName, "internal_port_range", "80:82"),
					resource.TestCheckResourceAttr(testAccVPCPortForwardingV2ResourceName, "external_port", "0"),
					resource.TestCheckResourceAttr(testAccVPCPortForwardingV2ResourceName, "internal_port", "0"),
					resource.TestCheckResourceAttr(testAccVPCPortForwardingV2ResourceName, "protocol", "udp"),
					resource.TestCheckResourceAttr(testAccVPCPortForwardingV2ResourceName, "description", "updated by the acceptance test"),
				),
			},
			{
				Config: testAccVPCPortForwardingV2Moved(projectID, region, 81, 8081),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCPortForwardingV2SameID(&found),
					resource.TestCheckResourceAttrPair(
						testAccVPCPortForwardingV2ResourceName, "internal_port_id",
						"selectel_vpc_port_v2.port_tf_acc_test_2", "id",
					),
					resource.TestCheckResourceAttr(testAccVPCPortForwardingV2ResourceName, "internal_ip_address", "192.168.199.61"),
					resource.TestCheckResourceAttr(testAccVPCPortForwardingV2ResourceName, "internal_port", "81"),
					resource.TestCheckResourceAttr(testAccVPCPortForwardingV2ResourceName, "external_port", "8081"),
					resource.TestCheckResourceAttr(testAccVPCPortForwardingV2ResourceName, "protocol", "tcp"),
				),
			},
			{
				Config: testAccVPCPortForwardingV2Ranges(projectID, region, "9090:9092", "udp"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCPortForwardingV2SameID(&found),
					resource.TestCheckResourceAttrPair(
						testAccVPCPortForwardingV2ResourceName, "internal_port_id",
						testAccVPCPortV2ResourceName, "id",
					),
					resource.TestCheckResourceAttr(testAccVPCPortForwardingV2ResourceName, "internal_ip_address", "192.168.199.60"),
					resource.TestCheckResourceAttr(testAccVPCPortForwardingV2ResourceName, "internal_port_range", "80:82"),
					resource.TestCheckResourceAttr(testAccVPCPortForwardingV2ResourceName, "external_port_range", "9090:9092"),
					resource.TestCheckResourceAttr(testAccVPCPortForwardingV2ResourceName, "protocol", "udp"),
				),
			},
			{
				Config:   testAccVPCPortForwardingV2Ranges(projectID, region, "9090:9092", "udp"),
				PlanOnly: true,
			},
		},
	})
}

func testAccVPCPortForwardingV2Handle(s *terraform.State) (*vpcNetworkV2TestHandle, string, error) {
	rs, ok := s.RootModule().Resources[testAccVPCPortForwardingV2ResourceName]
	if !ok {
		return nil, "", fmt.Errorf("not found: %s", testAccVPCPortForwardingV2ResourceName)
	}

	if rs.Primary.ID == "" {
		return nil, "", errors.New("no ID is set")
	}

	client, err := vpcV2Client(
		testAccProvider.Meta(),
		rs.Primary.Attributes["project_id"],
		rs.Primary.Attributes["region"],
	)
	if err != nil {
		return nil, "", err
	}

	return &vpcNetworkV2TestHandle{client: client, id: rs.Primary.ID},
		rs.Primary.Attributes["floatingip_id"], nil
}

func testAccVPCPortForwardingV2ImportID(s *terraform.State) (string, error) {
	rs, ok := s.RootModule().Resources[testAccVPCPortForwardingV2ResourceName]
	if !ok {
		return "", fmt.Errorf("not found: %s", testAccVPCPortForwardingV2ResourceName)
	}

	return rs.Primary.Attributes["floatingip_id"] + "/" + rs.Primary.ID, nil
}

func testAccCheckVPCPortForwardingV2Exists(found *floatingip.PortForwarding) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		handle, floatingIPID, err := testAccVPCPortForwardingV2Handle(s)
		if err != nil {
			return err
		}

		result, err := floatingip.GetPortForwarding(context.Background(), handle.client, floatingIPID, handle.id)
		if err != nil {
			return err
		}

		*found = *result

		return nil
	}
}

func testAccCheckVPCPortForwardingV2DeleteOutOfBand() resource.TestCheckFunc {
	return func(s *terraform.State) error {
		handle, floatingIPID, err := testAccVPCPortForwardingV2Handle(s)
		if err != nil {
			return err
		}

		return floatingip.DeletePortForwarding(context.Background(), handle.client, floatingIPID, handle.id)
	}
}

func testAccCheckVPCPortForwardingV2SameID(found *floatingip.PortForwarding) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[testAccVPCPortForwardingV2ResourceName]
		if !ok {
			return fmt.Errorf("not found: %s", testAccVPCPortForwardingV2ResourceName)
		}

		if rs.Primary.ID != found.ID {
			return fmt.Errorf("port forwarding was replaced: %s -> %s", found.ID, rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckVPCPortForwardingV2Destroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "selectel_vpc_port_forwarding_v2" {
			continue
		}

		client, err := vpcV2Client(
			testAccProvider.Meta(),
			rs.Primary.Attributes["project_id"],
			rs.Primary.Attributes["region"],
		)
		if err != nil {
			return err
		}

		_, err = floatingip.GetPortForwarding(
			context.Background(), client, rs.Primary.Attributes["floatingip_id"], rs.Primary.ID)
		if err == nil {
			return errors.New("VPC port forwarding still exists")
		} else if !vpc.IsErrorClass(err, vpc.ErrorClassNotFound) {
			return err
		}
	}

	return nil
}

func testAccVPCPortForwardingV2Base(projectID, region string) string {
	return testAccVPCPortV2FixedIP(projectID, region, "192.168.199.60") + fmt.Sprintf(`
resource "selectel_vpc_floatingip_v2" "floatingip_tf_acc_test_1" {
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

resource "selectel_vpc_port_v2" "port_tf_acc_test_2" {
  project_id = local.project_id
  region     = "%[1]s"
  network_id = selectel_vpc_network_v2.network_tf_acc_test_1.id

  fixed_ip {
    subnet_id  = selectel_vpc_network_subnet_v2.subnet_tf_acc_test_1.id
    ip_address = "192.168.199.61"
  }
}
`, region)
}

func testAccVPCPortForwardingV2Moved(projectID, region string, internalPort, externalPort int) string {
	return testAccVPCPortForwardingV2Base(projectID, region) + fmt.Sprintf(`
resource "selectel_vpc_port_forwarding_v2" "pf_tf_acc_test_1" {
  project_id          = local.project_id
  region              = "%s"
  floatingip_id       = selectel_vpc_floatingip_v2.floatingip_tf_acc_test_1.id
  internal_port_id    = selectel_vpc_port_v2.port_tf_acc_test_2.id
  internal_ip_address = "192.168.199.61"
  internal_port       = %d
  external_port       = %d
  protocol            = "tcp"

  depends_on = [selectel_vpc_router_interface_v2.interface_tf_acc_test_1]
}
`, region, internalPort, externalPort)
}

func testAccVPCPortForwardingV2Basic(projectID, region string, externalPort int, protocol string) string {
	return testAccVPCPortForwardingV2Base(projectID, region) + fmt.Sprintf(`
resource "selectel_vpc_port_forwarding_v2" "pf_tf_acc_test_1" {
  project_id          = local.project_id
  region              = "%s"
  floatingip_id       = selectel_vpc_floatingip_v2.floatingip_tf_acc_test_1.id
  internal_port_id    = selectel_vpc_port_v2.port_tf_acc_test_1.id
  internal_ip_address = "192.168.199.60"
  internal_port       = 80
  external_port       = %d
  protocol            = "%s"

  # The route to the external network is what makes the association possible, and
  # nothing else in this resource refers to it.
  depends_on = [selectel_vpc_router_interface_v2.interface_tf_acc_test_1]
}
`, region, externalPort, protocol)
}

func testAccVPCPortForwardingV2Ranges(projectID, region, externalPortRange, protocol string) string {
	return testAccVPCPortForwardingV2Base(projectID, region) + fmt.Sprintf(`
resource "selectel_vpc_port_forwarding_v2" "pf_tf_acc_test_1" {
  project_id          = local.project_id
  region              = "%s"
  floatingip_id       = selectel_vpc_floatingip_v2.floatingip_tf_acc_test_1.id
  internal_port_id    = selectel_vpc_port_v2.port_tf_acc_test_1.id
  internal_ip_address = "192.168.199.60"
  internal_port_range = "80:82"
  external_port_range = "%s"
  protocol            = "%s"
  description         = "updated by the acceptance test"

  # The route to the external network is what makes the association possible, and
  # nothing else in this resource refers to it.
  depends_on = [selectel_vpc_router_interface_v2.interface_tf_acc_test_1]
}
`, region, externalPortRange, protocol)
}
