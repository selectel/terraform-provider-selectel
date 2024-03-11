package selectel

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

const resourceZoneName = "zone_tf_acc_test_1"

func TestAccDomainsZoneV2Basic(t *testing.T) {
	projectName := acctest.RandomWithPrefix("tf-acc")
	testZoneName := fmt.Sprintf("%s.xyz.", acctest.RandomWithPrefix("tf-acc"))
	resourceZoneName := "zone_tf_acc_test_1"
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckDomainsV2ZoneDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDomainsZoneV2Basic(projectName, resourceZoneName, testZoneName),
				Check: resource.ComposeTestCheckFunc(
					testAccDomainsZoneV2Exists(fmt.Sprintf("selectel_domains_zone_v2.%[1]s", resourceZoneName)),
					resource.TestCheckResourceAttr(fmt.Sprintf("selectel_domains_zone_v2.%[1]s", resourceZoneName), "name", testZoneName),
				),
			},
		},
	})
}

func testAccDomainsZoneV2Basic(projectName, resourceName, zoneName string) string {
	return fmt.Sprintf(`
		resource "selectel_vpc_project_v2" "project_tf_acc_test_1" {
			name = %[1]q
		}
		resource "selectel_domains_zone_v2" %[2]q {
			name = %[3]q
			project_id = selectel_vpc_project_v2.project_tf_acc_test_1.id
		}`, projectName, resourceName, zoneName)
}

func testAccCheckDomainsV2ZoneDestroy(s *terraform.State) error {
	ctx := context.Background()

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "selectel_domains_zone_v2" {
			continue
		}

		zoneID := rs.Primary.ID
		client, err := getDomainsV2ClientTest(rs, testAccProvider)
		if err != nil {
			return err
		}
		_, err = client.GetZone(ctx, zoneID, nil)
		if err == nil {
			return errors.New("domain still exists")
		}
	}

	return nil
}

func testAccDomainsZoneV2Exists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("can't find zone: %s", name)
		}

		zoneID := rs.Primary.ID
		if zoneID == "" {
			return errors.New("zone ID not set in tf state")
		}
		client, err := getDomainsV2ClientTest(rs, testAccProvider)
		if err != nil {
			return err
		}
		ctx := context.Background()
		_, err = client.GetZone(ctx, zoneID, nil)
		if err != nil {
			return errors.New("zone in api not found")
		}

		return nil
	}
}
