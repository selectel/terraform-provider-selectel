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

func TestAccCloudPrivateDNSZoneV1Basic(t *testing.T) {
	region := os.Getenv("INFRA_REGION")
	projectID := os.Getenv("INFRA_PROJECT_ID")

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheckWithProjectID(t) },
		ProviderFactories: testAccProvidersWithOpenStack,
		CheckDestroy:      testAccCloudPrivateDNSZoneV1Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCloudPrivateDNSZoneV1Basic(region, 1800, projectID),
				Check: resource.ComposeTestCheckFunc(
					testAccCloudPrivateDNSZoneV1Exists("selectel_private_dns_zone_v1.zone"),
					resource.TestCheckResourceAttr(
						"selectel_private_dns_zone_v1.zone", "domain", "example.com.",
					),
					resource.TestCheckResourceAttr(
						"selectel_private_dns_zone_v1.zone", "ttl", "1800",
					),
				),
			},
			{
				Config: testAccCloudPrivateDNSZoneV1Basic(region, 1900, projectID),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"selectel_private_dns_zone_v1.zone", "ttl", "1900",
					),
				),
			},
			{
				Config: testAccCloudPrivateDNSZoneV1WithRecords(region, 1900, projectID),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"selectel_private_dns_zone_v1.zone", "records.0.domain", "sub.example.com.",
					),
					resource.TestCheckResourceAttr(
						"selectel_private_dns_zone_v1.zone", "records.0.ttl", "10",
					),
					resource.TestCheckResourceAttr(
						"selectel_private_dns_zone_v1.zone", "records.0.type", "A",
					),
				),
			},
			{
				Config: testAccCloudPrivateDNSZoneV1Basic(region, 1900, projectID),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckNoResourceAttr(
						"selectel_private_dns_zone_v1.zone", "records.0",
					),
				),
			},
		},
	})
}

func testAccCloudPrivateDNSZoneV1Basic(region string, ttl int, projectID string) string {
	return fmt.Sprintf(`
resource "selectel_private_dns_zone_v1" "zone" {
    region = "%s"
    project_id = "%s"
	domain = "example.com."
	ttl = %d
}
`, region, projectID, ttl)
}

func testAccCloudPrivateDNSZoneV1WithRecords(region string, ttl int, projectID string) string {
	return fmt.Sprintf(`
resource "selectel_private_dns_zone_v1" "zone" {
    region = "%s"
    project_id = "%s"
	domain = "example.com."
	ttl = %d
	records {
		domain = "sub.example.com."
		type = "A"
		ttl = 10
		values = [
			"192.168.0.1",
		]
	}
}
`, region, projectID, ttl)
}

func testAccCloudPrivateDNSZoneV1Destroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "selectel_private_dns_zone_v1" {
			continue
		}
		client, err := newTestPrivateDNSClient(rs, testAccProvider)
		if err != nil {
			return err
		}
		_, err = client.GetZone(context.Background(), rs.Primary.ID)
		if err == nil {
			return errors.New("zone still exists")
		}
	}

	return nil
}

func testAccCloudPrivateDNSZoneV1Exists(n string) resource.TestCheckFunc {
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

		_, err = client.GetZone(ctx, rs.Primary.ID)
		if err != nil {
			return err
		}

		return nil
	}
}
