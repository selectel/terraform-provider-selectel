package selectel

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccCloudPrivateDNSServiceV1Basic(t *testing.T) {
	region := os.Getenv("INFRA_REGION")
	projectID := os.Getenv("INFRA_PROJECT_ID")

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheckWithProjectID(t) },
		ProviderFactories: testAccProvidersWithOpenStack,
		CheckDestroy:      testAccCloudPrivateDNSServiceV1Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCloudPrivateDNSServiceV1Basic(region, projectID),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"selectel_private_dns_service_v1.service", "high_availability", "true",
					),
					resource.TestCheckResourceAttr(
						"selectel_private_dns_service_v1.service", "project_id", os.Getenv("INFRA_PROJECT_ID"),
					),
					resource.TestCheckResourceAttr("selectel_private_dns_service_v1.service", "addresses.#", "2"),
					testAccCloudPrivateDNSServiceV1Exists("selectel_private_dns_service_v1.service"),
				),
			},
		},
	})
}

func testAccCloudPrivateDNSServiceV1Basic(region, projectID string) string {
	return fmt.Sprintf(`
provider openstack {
	tenant_id = "%s"
}

resource "openstack_networking_network_v2" "network_one" {
 	region = "%s"
  	name = "network_one"
}

resource "openstack_networking_subnet_v2" "subnet" {
  network_id = openstack_networking_network_v2.network_one.id
  cidr       = "192.168.1.0/24"
  ip_version = 4
  enable_dhcp = false
  name = "subnet"
}

resource "selectel_private_dns_service_v1" "service" {
    region = "%s"
    project_id = "%s"
    network_id = openstack_networking_network_v2.network_one.id
}
`, projectID, region, region, projectID)
}

func testAccCloudPrivateDNSServiceV1Destroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "selectel_private_dns_service_v1" {
			continue
		}
		client, err := newTestPrivateDNSClient(rs, testAccProvider)
		if err != nil {
			return err
		}
		_, err = client.GetService(context.Background(), rs.Primary.ID)
		if err == nil {
			return errors.New("service still exists")
		}
	}

	return nil
}

func testAccCloudPrivateDNSServiceV1Exists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}
		if rs.Primary.ID == "" {
			return errors.New("no ID is set")
		}

		ctx := context.Background()

		client, err := newTestPrivateDNSClient(rs, testAccProvider)
		if err != nil {
			return err
		}

		_, err = client.GetService(ctx, rs.Primary.ID)
		if err != nil {
			return err
		}

		return nil
	}
}
