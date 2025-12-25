package selectel

import (
	"errors"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccGlobalRouterServiceV1DataSourceBasic(t *testing.T) {
	dataGRServiceName := "gr_service_1"
	testServiceName := "vpc"
	dataSourceServiceName := fmt.Sprintf("data.selectel_global_router_service_v1.%[1]s", dataGRServiceName)
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheck(t) },
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccGlobalRouterServiceV1DataSourceBasic(dataGRServiceName, testServiceName),
				Check: resource.ComposeTestCheckFunc(
					testAccServiceV1Exists(dataSourceServiceName),
					resource.TestCheckResourceAttr(dataSourceServiceName, "name", testServiceName),
					// computed opts
					resource.TestCheckResourceAttrSet(dataSourceServiceName, "created_at"),
					resource.TestCheckResourceAttrSet(dataSourceServiceName, "extension"),
				),
			},
		},
	})
}

func testAccGlobalRouterServiceV1DataSourceBasic(resourceName, serviceName string) string {
	return fmt.Sprintf(`
data "selectel_global_router_service_v1" %[1]q {
  name=%[2]q
}`, resourceName, serviceName)
}

func testAccServiceV1Exists(name string) resource.TestCheckFunc {
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
