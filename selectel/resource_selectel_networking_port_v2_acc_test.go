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
	"github.com/selectel/vpc-go/pkg/v2/port"
)

const (
	testAccNetworkingPortV2ResourceName   = "selectel_networking_port_v2.port_tf_acc_test_1"
	testAccNetworkingPortV2DataSourceName = "data.selectel_networking_port_v2.port_tf_acc_test_1"
)

// TestAccNetworkingPortV2Basic covers SCN-037-01, SCN-037-02 (empty follow-up
// plan), SCN-046-01 (import) and SCN-045-01 (delete).
func TestAccNetworkingPortV2Basic(t *testing.T) {
	var found port.Port
	projectID, region := networkingV2TestEnv()

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheckWithProjectIDAndRegion(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckNetworkingPortV2Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkingPortV2Basic(projectID, region),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNetworkingPortV2Exists(&found),
					// The API assigns these when the configuration leaves them out.
					resource.TestCheckResourceAttrSet(testAccNetworkingPortV2ResourceName, "mac_address"),
					resource.TestCheckResourceAttrSet(testAccNetworkingPortV2ResourceName, "admin_state_up"),
					resource.TestCheckResourceAttrSet(testAccNetworkingPortV2ResourceName, "status"),
					// A DHCP subnet exists, so the API allocates one fixed IP and
					// the project's default security group.
					resource.TestCheckResourceAttr(testAccNetworkingPortV2ResourceName, "all_fixed_ips.#", "1"),
					resource.TestCheckResourceAttr(testAccNetworkingPortV2ResourceName, "all_security_group_ids.#", "1"),
					// Admin-managed attributes are reported, never configured.
					resource.TestCheckResourceAttrSet(testAccNetworkingPortV2ResourceName, "port_security_enabled"),
					resource.TestCheckResourceAttrSet(testAccNetworkingPortV2ResourceName, "dhcp_blocked"),
					resource.TestCheckResourceAttrSet(testAccNetworkingPortV2ResourceName, "blocked"),
				),
			},
			{
				Config:   testAccNetworkingPortV2Basic(projectID, region),
				PlanOnly: true,
			},
			{
				ResourceName:      testAccNetworkingPortV2ResourceName,
				ImportState:       true,
				ImportStateVerify: true,
				// The two negative flags exist only in the configuration.
				ImportStateVerifyIgnore: []string{"no_fixed_ip", "no_security_groups"},
			},
		},
	})
}

// TestAccNetworkingPortV2FixedIP covers SCN-038-01 and SCN-038-02.
func TestAccNetworkingPortV2FixedIP(t *testing.T) {
	var found port.Port
	projectID, region := networkingV2TestEnv()

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheckWithProjectIDAndRegion(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckNetworkingPortV2Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkingPortV2FixedIP(projectID, region, "192.168.199.50"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNetworkingPortV2Exists(&found),
					resource.TestCheckResourceAttr(testAccNetworkingPortV2ResourceName, "all_fixed_ips.#", "1"),
					resource.TestCheckResourceAttr(testAccNetworkingPortV2ResourceName, "all_fixed_ips.0.ip_address", "192.168.199.50"),
				),
			},
			{
				Config:   testAccNetworkingPortV2FixedIP(projectID, region, "192.168.199.50"),
				PlanOnly: true,
			},
			{
				// SCN-041-04: the address set changes in place.
				Config: testAccNetworkingPortV2FixedIP(projectID, region, "192.168.199.51"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNetworkingPortV2SameID(&found),
					resource.TestCheckResourceAttr(testAccNetworkingPortV2ResourceName, "all_fixed_ips.0.ip_address", "192.168.199.51"),
				),
			},
		},
	})
}

// TestAccNetworkingPortV2NoFixedIP covers SCN-038-02: no_fixed_ip creates the
// port without any address even though a DHCP subnet exists.
func TestAccNetworkingPortV2NoFixedIP(t *testing.T) {
	var found port.Port
	projectID, region := networkingV2TestEnv()

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheckWithProjectIDAndRegion(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckNetworkingPortV2Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkingPortV2NoFixedIP(projectID, region),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNetworkingPortV2Exists(&found),
					resource.TestCheckResourceAttr(testAccNetworkingPortV2ResourceName, "all_fixed_ips.#", "0"),
				),
			},
			{
				Config:   testAccNetworkingPortV2NoFixedIP(projectID, region),
				PlanOnly: true,
			},
		},
	})
}

// TestAccNetworkingPortV2SecurityGroups covers SCN-039-01, SCN-039-02,
// SCN-041-02 and SCN-041-03.
func TestAccNetworkingPortV2SecurityGroups(t *testing.T) {
	var found port.Port
	projectID, region := networkingV2TestEnv()

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheckWithProjectIDAndRegion(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckNetworkingPortV2Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkingPortV2NoSecurityGroups(projectID, region),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNetworkingPortV2Exists(&found),
					resource.TestCheckResourceAttr(testAccNetworkingPortV2ResourceName, "all_security_group_ids.#", "0"),
				),
			},
			{
				Config:   testAccNetworkingPortV2NoSecurityGroups(projectID, region),
				PlanOnly: true,
			},
		},
	})
}

// TestAccNetworkingPortV2ExtendedAttributes covers SCN-040-01, SCN-040-03 and
// the binding.vnic_type part of SCN-041-05.
func TestAccNetworkingPortV2ExtendedAttributes(t *testing.T) {
	var found port.Port
	projectID, region := networkingV2TestEnv()
	portName := acctest.RandomWithPrefix("tf-acc-port")

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheckWithProjectIDAndRegion(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckNetworkingPortV2Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkingPortV2Extended(projectID, region, portName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNetworkingPortV2Exists(&found),
					resource.TestCheckResourceAttr(testAccNetworkingPortV2ResourceName, "name", portName),
					resource.TestCheckResourceAttr(testAccNetworkingPortV2ResourceName, "allowed_address_pairs.#", "1"),
					resource.TestCheckResourceAttr(testAccNetworkingPortV2ResourceName, "extra_dhcp_option.#", "2"),
					resource.TestCheckResourceAttr(testAccNetworkingPortV2ResourceName, "binding.0.vnic_type", "normal"),
					resource.TestCheckResourceAttr(testAccNetworkingPortV2ResourceName, "tags.#", "2"),
				),
			},
			{
				Config:   testAccNetworkingPortV2Extended(projectID, region, portName),
				PlanOnly: true,
			},
		},
	})
}

// TestAccNetworkingPortV2Update covers SCN-041-01 and SCN-041-05.
func TestAccNetworkingPortV2Update(t *testing.T) {
	var found port.Port
	projectID, region := networkingV2TestEnv()
	portName := acctest.RandomWithPrefix("tf-acc-port")
	updatedName := acctest.RandomWithPrefix("tf-acc-port-upd")

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheckWithProjectIDAndRegion(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckNetworkingPortV2Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkingPortV2Extended(projectID, region, portName),
				Check:  testAccCheckNetworkingPortV2Exists(&found),
			},
			{
				Config: testAccNetworkingPortV2ExtendedUpdated(projectID, region, updatedName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNetworkingPortV2SameID(&found),
					resource.TestCheckResourceAttr(testAccNetworkingPortV2ResourceName, "name", updatedName),
					resource.TestCheckResourceAttr(testAccNetworkingPortV2ResourceName, "description", "updated by the acceptance test"),
					resource.TestCheckResourceAttr(testAccNetworkingPortV2ResourceName, "admin_state_up", "false"),
					resource.TestCheckResourceAttr(testAccNetworkingPortV2ResourceName, "allowed_address_pairs.#", "2"),
					resource.TestCheckResourceAttr(testAccNetworkingPortV2ResourceName, "extra_dhcp_option.#", "1"),
					resource.TestCheckResourceAttr(testAccNetworkingPortV2ResourceName, "tags.#", "1"),
				),
			},
			{
				Config:   testAccNetworkingPortV2ExtendedUpdated(projectID, region, updatedName),
				PlanOnly: true,
			},
		},
	})
}

// TestAccNetworkingPortV2ExternalChanges covers SCN-043-01, SCN-043-02 and
// SCN-045-02.
func TestAccNetworkingPortV2ExternalChanges(t *testing.T) {
	var found port.Port
	projectID, region := networkingV2TestEnv()
	portName := acctest.RandomWithPrefix("tf-acc-port")

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheckWithProjectIDAndRegion(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckNetworkingPortV2Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkingPortV2Named(projectID, region, portName),
				Check:  testAccCheckNetworkingPortV2Exists(&found),
			},
			{
				Config:             testAccNetworkingPortV2Named(projectID, region, portName),
				Check:              testAccCheckNetworkingPortV2RenameOutOfBand(&found),
				ExpectNonEmptyPlan: true,
			},
			{
				Config:             testAccNetworkingPortV2Named(projectID, region, portName),
				Check:              testAccCheckNetworkingPortV2DeleteOutOfBand(),
				ExpectNonEmptyPlan: true,
			},
			{
				Config: testAccNetworkingPortV2Named(projectID, region, portName),
				Check:  testAccCheckNetworkingPortV2Exists(&found),
			},
		},
	})
}

// TestAccNetworkingPortV2DataSource covers SCN-047-01 … SCN-047-03, including the
// two partial-match selectors.
func TestAccNetworkingPortV2DataSource(t *testing.T) {
	projectID, region := networkingV2TestEnv()
	portName := acctest.RandomWithPrefix("tf-acc-port")

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheckWithProjectIDAndRegion(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckNetworkingPortV2Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkingPortV2DataSource(projectID, region, portName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(testAccNetworkingPortV2DataSourceName, "name", portName),
					resource.TestCheckResourceAttrPair(
						testAccNetworkingPortV2DataSourceName, "port_id",
						testAccNetworkingPortV2ResourceName, "id",
					),
					resource.TestCheckResourceAttrPair(
						testAccNetworkingPortV2DataSourceName, "blocked",
						testAccNetworkingPortV2ResourceName, "blocked",
					),
				),
			},
			{
				Config:      testAccNetworkingPortV2DataSourceNoMatch(projectID, region),
				ExpectError: regexp.MustCompile(`no networking port found`),
			},
		},
	})
}

// -----------------------------------------------------------------------------
// Helpers.
// -----------------------------------------------------------------------------

func testAccNetworkingPortV2Handle(s *terraform.State) (*networkingNetworkV2TestHandle, error) {
	rs, ok := s.RootModule().Resources[testAccNetworkingPortV2ResourceName]
	if !ok {
		return nil, fmt.Errorf("not found: %s", testAccNetworkingPortV2ResourceName)
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

func testAccCheckNetworkingPortV2Exists(found *port.Port) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		handle, err := testAccNetworkingPortV2Handle(s)
		if err != nil {
			return err
		}

		result, err := port.Get(context.Background(), handle.client, handle.id)
		if err != nil {
			return err
		}

		*found = *result

		return nil
	}
}

func testAccCheckNetworkingPortV2SameID(found *port.Port) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[testAccNetworkingPortV2ResourceName]
		if !ok {
			return fmt.Errorf("not found: %s", testAccNetworkingPortV2ResourceName)
		}

		if rs.Primary.ID != found.ID {
			return fmt.Errorf("networking port was replaced: %s -> %s", found.ID, rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckNetworkingPortV2RenameOutOfBand(found *port.Port) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		handle, err := testAccNetworkingPortV2Handle(s)
		if err != nil {
			return err
		}

		driftedName := found.Name + "-drifted"
		_, err = port.Update(context.Background(), handle.client, handle.id, port.UpdateRequest{
			Name: &driftedName,
		})

		return err
	}
}

func testAccCheckNetworkingPortV2DeleteOutOfBand() resource.TestCheckFunc {
	return func(s *terraform.State) error {
		handle, err := testAccNetworkingPortV2Handle(s)
		if err != nil {
			return err
		}

		return port.Delete(context.Background(), handle.client, handle.id)
	}
}

func testAccCheckNetworkingPortV2Destroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "selectel_networking_port_v2" {
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

		if _, err := port.Get(context.Background(), client, rs.Primary.ID); err == nil {
			return errors.New("networking port still exists")
		} else if !networkingIsNotFound(err) {
			return err
		}
	}

	return nil
}

// -----------------------------------------------------------------------------
// Configurations.
// -----------------------------------------------------------------------------

// testAccNetworkingPortV2Subnet gives the port a network with a DHCP subnet.
func testAccNetworkingPortV2Subnet(projectID, region string) string {
	return testAccNetworkingSubnetV2Basic(projectID, region, "192.168.199.0/24")
}

func testAccNetworkingPortV2Basic(projectID, region string) string {
	return testAccNetworkingPortV2Subnet(projectID, region) + fmt.Sprintf(`
resource "selectel_networking_port_v2" "port_tf_acc_test_1" {
  project_id = local.project_id
  region     = "%s"
  network_id = selectel_networking_network_v2.network_tf_acc_test_1.id

  depends_on = [selectel_networking_subnet_v2.subnet_tf_acc_test_1]
}
`, region)
}

func testAccNetworkingPortV2Named(projectID, region, portName string) string {
	return testAccNetworkingPortV2Subnet(projectID, region) + fmt.Sprintf(`
resource "selectel_networking_port_v2" "port_tf_acc_test_1" {
  project_id = local.project_id
  region     = "%s"
  network_id = selectel_networking_network_v2.network_tf_acc_test_1.id
  name       = "%s"

  depends_on = [selectel_networking_subnet_v2.subnet_tf_acc_test_1]
}
`, region, portName)
}

func testAccNetworkingPortV2FixedIP(projectID, region, address string) string {
	return testAccNetworkingPortV2Subnet(projectID, region) + fmt.Sprintf(`
resource "selectel_networking_port_v2" "port_tf_acc_test_1" {
  project_id = local.project_id
  region     = "%s"
  network_id = selectel_networking_network_v2.network_tf_acc_test_1.id

  fixed_ip {
    subnet_id  = selectel_networking_subnet_v2.subnet_tf_acc_test_1.id
    ip_address = "%s"
  }
}
`, region, address)
}

func testAccNetworkingPortV2NoFixedIP(projectID, region string) string {
	return testAccNetworkingPortV2Subnet(projectID, region) + fmt.Sprintf(`
resource "selectel_networking_port_v2" "port_tf_acc_test_1" {
  project_id  = local.project_id
  region      = "%s"
  network_id  = selectel_networking_network_v2.network_tf_acc_test_1.id
  no_fixed_ip = true

  depends_on = [selectel_networking_subnet_v2.subnet_tf_acc_test_1]
}
`, region)
}

func testAccNetworkingPortV2NoSecurityGroups(projectID, region string) string {
	return testAccNetworkingPortV2Subnet(projectID, region) + fmt.Sprintf(`
resource "selectel_networking_port_v2" "port_tf_acc_test_1" {
  project_id         = local.project_id
  region             = "%s"
  network_id         = selectel_networking_network_v2.network_tf_acc_test_1.id
  no_security_groups = true

  depends_on = [selectel_networking_subnet_v2.subnet_tf_acc_test_1]
}
`, region)
}

func testAccNetworkingPortV2Extended(projectID, region, portName string) string {
	return testAccNetworkingPortV2Subnet(projectID, region) + fmt.Sprintf(`
resource "selectel_networking_port_v2" "port_tf_acc_test_1" {
  project_id     = local.project_id
  region         = "%s"
  network_id     = selectel_networking_network_v2.network_tf_acc_test_1.id
  name           = "%s"
  description    = "created by the acceptance test"
  admin_state_up = true
  tags           = ["tf-acc", "networking"]

  fixed_ip {
    subnet_id = selectel_networking_subnet_v2.subnet_tf_acc_test_1.id
  }

  allowed_address_pairs {
    ip_address = "192.168.199.100"
  }

  extra_dhcp_option {
    name  = "tftp-server"
    value = "192.168.199.2"
  }

  extra_dhcp_option {
    name  = "bootfile-name"
    value = "pxelinux.0"
  }

  binding {
    vnic_type = "normal"
  }
}
`, region, portName)
}

func testAccNetworkingPortV2ExtendedUpdated(projectID, region, portName string) string {
	return testAccNetworkingPortV2Subnet(projectID, region) + fmt.Sprintf(`
resource "selectel_networking_port_v2" "port_tf_acc_test_1" {
  project_id     = local.project_id
  region         = "%s"
  network_id     = selectel_networking_network_v2.network_tf_acc_test_1.id
  name           = "%s"
  description    = "updated by the acceptance test"
  admin_state_up = false
  tags           = ["tf-acc"]

  fixed_ip {
    subnet_id = selectel_networking_subnet_v2.subnet_tf_acc_test_1.id
  }

  allowed_address_pairs {
    ip_address = "192.168.199.100"
  }

  allowed_address_pairs {
    ip_address = "192.168.199.101"
  }

  extra_dhcp_option {
    name  = "tftp-server"
    value = "192.168.199.3"
  }

  binding {
    vnic_type = "normal"
  }
}
`, region, portName)
}

func testAccNetworkingPortV2DataSource(projectID, region, portName string) string {
	return testAccNetworkingPortV2Named(projectID, region, portName) + fmt.Sprintf(`
data "selectel_networking_port_v2" "port_tf_acc_test_1" {
  project_id = local.project_id
  region     = "%s"
  name       = selectel_networking_port_v2.port_tf_acc_test_1.name
}
`, region)
}

func testAccNetworkingPortV2DataSourceNoMatch(projectID, region string) string {
	return testAccNetworkingPortV2Subnet(projectID, region) + fmt.Sprintf(`
data "selectel_networking_port_v2" "port_tf_acc_test_1" {
  project_id = local.project_id
  region     = "%s"
  name       = "tf-acc-no-such-port"
}
`, region)
}
