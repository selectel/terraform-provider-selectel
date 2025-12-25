package selectel

import (
	"errors"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccGlobalRouterZoneGroupV1DataSourceBasic(t *testing.T) {
	dataGRZoneGroupName := "gr_zone_group_1"
	testZoneGroupName := "public_rf"
	dataSourceZoneGroupName := fmt.Sprintf("data.selectel_global_router_zone_group_v1.%[1]s", dataGRZoneGroupName)
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheck(t) },
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccGlobalRouterZoneGroupV1DataSourceBasic(dataGRZoneGroupName, testZoneGroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccZoneGroupV1Exists(dataSourceZoneGroupName),
					resource.TestCheckResourceAttr(dataSourceZoneGroupName, "name", testZoneGroupName),
					// computed opts
					resource.TestCheckResourceAttr(dataSourceZoneGroupName, "description", ""),
					resource.TestCheckResourceAttrSet(dataSourceZoneGroupName, "created_at"),
					resource.TestCheckResourceAttr(dataSourceZoneGroupName, "updated_at", ""),
				),
			},
		},
	})
}

func testAccGlobalRouterZoneGroupV1DataSourceBasic(resourceName, zoneGroupName string) string {
	return fmt.Sprintf(`
data "selectel_global_router_zone_group_v1" %[1]q {
  name=%[2]q
}`, resourceName, zoneGroupName)
}

func testAccZoneGroupV1Exists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return errors.New("no ID is set")
		}

		return nil
	}
}
