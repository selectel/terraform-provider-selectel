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

func TestAccDomainsZoneV2DataSourceBasic(t *testing.T) {
	testZoneName := fmt.Sprintf("%s.ru.", acctest.RandomWithPrefix("tf-acc"))
	resourceZoneName := "zone_tf_acc_test_1"
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheckWithProjectID(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckDomainsV2ZoneDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDomainsZoneV2DataSourceBasic(resourceZoneName, testZoneName),
				Check: resource.ComposeTestCheckFunc(
					testAccDomainsZoneV2Exists(fmt.Sprintf("data.selectel_domains_zone_v2.%[1]s", resourceZoneName)),
					resource.TestCheckResourceAttr(fmt.Sprintf("data.selectel_domains_zone_v2.%[1]s", resourceZoneName), "name", testZoneName),
				),
			},
		},
	})
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

		meta := testAccProvider.Meta()
		client, err := getDomainsV2Client(meta)
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

func testAccDomainsZoneV2DataSourceBasic(resourceName, zoneName string) string {
	return fmt.Sprintf(`
	%[1]s

	data "selectel_domains_zone_v2" %[2]q {
	  name = selectel_domains_zone_v2.%[2]s.name
	}
`, testAccDomainsZoneV2Basic(resourceName, zoneName), resourceName, zoneName)
}
