package selectel

import (
	"errors"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccGlobalRouterZoneV1DataSourceBasic(t *testing.T) {
	dataGRZoneName := "gr_zone_1"
	testZoneName := "ru-1"
	testService := "vpc"
	dataSourceZoneName := fmt.Sprintf("data.selectel_global_router_zone_v1.%[1]s", dataGRZoneName)
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheck(t) },
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccGlobalRouterZoneV1DataSourceBasic(dataGRZoneName, testZoneName, testService),
				Check: resource.ComposeTestCheckFunc(
					testAccZoneV1Exists(dataSourceZoneName),
					resource.TestCheckResourceAttr(dataSourceZoneName, "name", testZoneName),
					// computed opts
					resource.TestCheckResourceAttr(dataSourceZoneName, "service", testService),
					resource.TestCheckResourceAttrSet(dataSourceZoneName, "visible_name"),
					resource.TestCheckResourceAttr(dataSourceZoneName, "enable", "true"),
					resource.TestCheckResourceAttr(dataSourceZoneName, "allow_create", "true"),
					resource.TestCheckResourceAttr(dataSourceZoneName, "allow_update", "true"),
					resource.TestCheckResourceAttr(dataSourceZoneName, "allow_delete", "true"),
					resource.TestCheckResourceAttrSet(dataSourceZoneName, "created_at"),
					resource.TestCheckResourceAttrSet(dataSourceZoneName, "updated_at"),
					resource.TestCheckResourceAttr(dataSourceZoneName, "options", ""),
					// inner groups
					resource.TestCheckResourceAttrSet(dataSourceZoneName, "groups.0.id"),
					resource.TestCheckResourceAttr(dataSourceZoneName, "groups.0.name", "public_rf"),
					resource.TestCheckResourceAttr(dataSourceZoneName, "groups.0.description", ""),
					resource.TestCheckResourceAttrSet(dataSourceZoneName, "groups.0.created_at"),
					resource.TestCheckResourceAttrSet(dataSourceZoneName, "groups.0.updated_at"),
				),
			},
		},
	})
}

func testAccGlobalRouterZoneV1DataSourceBasic(resourceName, zoneName string, service string) string {
	return fmt.Sprintf(`
data "selectel_global_router_zone_v1" %[1]q {
  name=%[2]q
  service=%[3]q
}`, resourceName, zoneName, service)
}

func testAccZoneV1Exists(name string) resource.TestCheckFunc {
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
