package selectel

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccVPCPublicPortV1Basic(t *testing.T) {
	region := os.Getenv("INFRA_REGION")
	projectID := os.Getenv("INFRA_PROJECT_ID")

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheckWithProjectID(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccVPCPublicPortV1CheckDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCPublicPortV1(region, projectID),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("selectel_vpc_public_port_v1.port", "id"),
					resource.TestCheckResourceAttr("selectel_vpc_public_port_v1.port", "project_id", projectID),
					resource.TestCheckResourceAttrSet("selectel_vpc_public_port_v1.port", "network_id"),
					resource.TestCheckResourceAttrSet("selectel_vpc_public_port_v1.port", "ip_address"),
					resource.TestCheckResourceAttr("selectel_vpc_public_port_v1.port", "description", ""),
					resource.TestCheckResourceAttrSet("selectel_vpc_public_port_v1.port", "subnet"),
					resource.TestCheckResourceAttrSet("selectel_vpc_public_port_v1.port", "gateway"),
					resource.TestCheckResourceAttr("selectel_vpc_public_port_v1.port", "admin_state_up", "true"),
					resource.TestCheckResourceAttr("selectel_vpc_public_port_v1.port", "security_group_ids.#", "1"),
					testAccVPCPublicPortV1CheckExist("selectel_vpc_public_port_v1.port"),
				),
			},
			{
				Config: testAccVPCPublicPortV1WithAdminStateUp(region, projectID, true),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("selectel_vpc_public_port_v1.port", "id"),
					resource.TestCheckResourceAttr("selectel_vpc_public_port_v1.port", "project_id", projectID),
					resource.TestCheckResourceAttrSet("selectel_vpc_public_port_v1.port", "network_id"),
					resource.TestCheckResourceAttrSet("selectel_vpc_public_port_v1.port", "ip_address"),
					resource.TestCheckResourceAttr("selectel_vpc_public_port_v1.port", "description", ""),
					resource.TestCheckResourceAttrSet("selectel_vpc_public_port_v1.port", "subnet"),
					resource.TestCheckResourceAttrSet("selectel_vpc_public_port_v1.port", "gateway"),
					resource.TestCheckResourceAttr("selectel_vpc_public_port_v1.port", "admin_state_up", "true"),
					resource.TestCheckResourceAttr("selectel_vpc_public_port_v1.port", "security_group_ids.#", "1"),
				),
			},
			{
				Config: testAccVPCPublicPortV1WithAdminStateUp(region, projectID, false),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("selectel_vpc_public_port_v1.port", "id"),
					resource.TestCheckResourceAttr("selectel_vpc_public_port_v1.port", "project_id", projectID),
					resource.TestCheckResourceAttrSet("selectel_vpc_public_port_v1.port", "network_id"),
					resource.TestCheckResourceAttrSet("selectel_vpc_public_port_v1.port", "ip_address"),
					resource.TestCheckResourceAttr("selectel_vpc_public_port_v1.port", "description", ""),
					resource.TestCheckResourceAttrSet("selectel_vpc_public_port_v1.port", "subnet"),
					resource.TestCheckResourceAttrSet("selectel_vpc_public_port_v1.port", "gateway"),
					resource.TestCheckResourceAttr("selectel_vpc_public_port_v1.port", "admin_state_up", "false"),
					resource.TestCheckResourceAttr("selectel_vpc_public_port_v1.port", "security_group_ids.#", "1"),
				),
			},
			{
				Config: testAccVPCPublicPortV1WithDescription(region, projectID, ""),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("selectel_vpc_public_port_v1.port", "id"),
					resource.TestCheckResourceAttr("selectel_vpc_public_port_v1.port", "project_id", projectID),
					resource.TestCheckResourceAttrSet("selectel_vpc_public_port_v1.port", "network_id"),
					resource.TestCheckResourceAttrSet("selectel_vpc_public_port_v1.port", "ip_address"),
					resource.TestCheckResourceAttr("selectel_vpc_public_port_v1.port", "description", ""),
					resource.TestCheckResourceAttrSet("selectel_vpc_public_port_v1.port", "subnet"),
					resource.TestCheckResourceAttrSet("selectel_vpc_public_port_v1.port", "gateway"),
					resource.TestCheckResourceAttr("selectel_vpc_public_port_v1.port", "admin_state_up", "true"),
					resource.TestCheckResourceAttr("selectel_vpc_public_port_v1.port", "security_group_ids.#", "1"),
				),
			},
			{
				Config: testAccVPCPublicPortV1WithDescription(region, projectID, "tf-acc-test-desc"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("selectel_vpc_public_port_v1.port", "id"),
					resource.TestCheckResourceAttr("selectel_vpc_public_port_v1.port", "project_id", projectID),
					resource.TestCheckResourceAttrSet("selectel_vpc_public_port_v1.port", "network_id"),
					resource.TestCheckResourceAttrSet("selectel_vpc_public_port_v1.port", "ip_address"),
					resource.TestCheckResourceAttr("selectel_vpc_public_port_v1.port", "description", "tf-acc-test-desc"),
					resource.TestCheckResourceAttrSet("selectel_vpc_public_port_v1.port", "subnet"),
					resource.TestCheckResourceAttrSet("selectel_vpc_public_port_v1.port", "gateway"),
					resource.TestCheckResourceAttr("selectel_vpc_public_port_v1.port", "admin_state_up", "true"),
					resource.TestCheckResourceAttr("selectel_vpc_public_port_v1.port", "security_group_ids.#", "1"),
				),
			},
		},
	})
}

func TestAccVPCPublicPortV1Complex(t *testing.T) {
	region := os.Getenv("INFRA_REGION")
	projectID := os.Getenv("INFRA_PROJECT_ID")

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheckWithProjectID(t) },
		ProviderFactories: testAccProvidersWithOpenStack,
		CheckDestroy:      testAccVPCPublicPortV1CheckDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCPublicPortV1WithSecurityGroups(region, projectID),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("selectel_vpc_public_port_v1.port", "id"),
					resource.TestCheckResourceAttr("selectel_vpc_public_port_v1.port", "project_id", projectID),
					resource.TestCheckResourceAttrSet("selectel_vpc_public_port_v1.port", "network_id"),
					resource.TestCheckResourceAttrSet("selectel_vpc_public_port_v1.port", "ip_address"),
					resource.TestCheckResourceAttr("selectel_vpc_public_port_v1.port", "description", ""),
					resource.TestCheckResourceAttrSet("selectel_vpc_public_port_v1.port", "subnet"),
					resource.TestCheckResourceAttrSet("selectel_vpc_public_port_v1.port", "gateway"),
					resource.TestCheckResourceAttr("selectel_vpc_public_port_v1.port", "admin_state_up", "true"),
					resource.TestCheckResourceAttr("selectel_vpc_public_port_v1.port", "security_group_ids.#", "1"),
					resource.TestCheckResourceAttrPair(
						"selectel_vpc_public_port_v1.port",
						"security_group_ids.0",
						"openstack_networking_secgroup_v2.sg",
						"id",
					),
				),
			},
		},
	})
}

func testAccVPCPublicPortV1(region, projectID string) string {
	return fmt.Sprintf(`
resource "selectel_vpc_public_port_v1" "port" {
  region      = %q
  project_id  = %q
}`, region, projectID)
}

func testAccVPCPublicPortV1WithDescription(region, projectID, description string) string {
	return fmt.Sprintf(`
resource "selectel_vpc_public_port_v1" "port" {
    region      = %q
    project_id  = %q
    description = %q
}`, region, projectID, description)
}

func testAccVPCPublicPortV1WithAdminStateUp(region, projectID string, adminStateUp bool) string {
	return fmt.Sprintf(`
resource "selectel_vpc_public_port_v1" "port" {
  region         = %q
  project_id     = %q
  admin_state_up = %v
}`, region, projectID, adminStateUp)
}

func testAccVPCPublicPortV1WithSecurityGroups(region, projectID string) string {
	return fmt.Sprintf(`
provider openstack {
    tenant_id = %q
}

resource "openstack_networking_secgroup_v2" "sg" {
    region = %q
    name   = "tf-acc-test-sg"
}

resource "selectel_vpc_public_port_v1" "port" {
    region             = %q
    project_id         = %q
    security_group_ids = [openstack_networking_secgroup_v2.sg.id]
}`, projectID, region, region, projectID)
}

func testAccVPCPublicPortV1CheckExist(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource %q not found in Terraform state", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("resource %q has no ID", resourceName)
		}

		client, err := newTestPublicNetAPI(rs, testAccProvider)
		if err != nil {
			return err
		}

		if _, err := client.GetPort(context.Background(), rs.Primary.ID); err != nil {
			return fmt.Errorf("failed to get port %q: %w", rs.Primary.ID, err)
		}

		return nil
	}
}

func testAccVPCPublicPortV1CheckDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "selectel_vpc_public_port_v1" {
			continue
		}

		client, err := newTestPublicNetAPI(rs, testAccProvider)
		if err != nil {
			return err
		}

		_, err = client.GetPort(context.Background(), rs.Primary.ID)
		if err == nil {
			return fmt.Errorf("selectel_vpc_public_port_v1.port %q still exists", rs.Primary.ID)
		}

		if isNotFound(err) {

			return err
		}
	}

	return nil
}
