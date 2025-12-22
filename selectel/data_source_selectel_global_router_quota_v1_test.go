package selectel

import (
	"errors"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccGlobalRouterQuotaV1DataSourceBasic(t *testing.T) {
	dataGRQuotaName := "gr_quota_1"
	testQuotaName := "subnets"
	testScope := "account_id"
	testScopeValue := "12345"
	dataSourceQuotaName := fmt.Sprintf("data.selectel_global_router_quota_v1.%[1]s", dataGRQuotaName)
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheck(t) },
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccGlobalRouterQuotaV1DataSourceBasic(dataGRQuotaName, testQuotaName, testScope, testScopeValue),
				Check: resource.ComposeTestCheckFunc(
					testAccQuotaV1Exists(dataSourceQuotaName),
					resource.TestCheckResourceAttrSet(dataSourceQuotaName, "name"),
					// computed opts
					resource.TestCheckResourceAttrSet(dataSourceQuotaName, "scope"),
					resource.TestCheckResourceAttr(dataSourceQuotaName, "scope_value", ""),
					resource.TestCheckResourceAttrSet(dataSourceQuotaName, "limit"),
				),
			},
		},
	})
}

func testAccGlobalRouterQuotaV1DataSourceBasic(resourceName, quotaName string, scope string, scopeValue string) string {
	return fmt.Sprintf(`
data "selectel_global_router_quota_v1" %[1]q {
  name=%[2]q
  scope=%[3]q
  scope_value=%[4]q
}`, resourceName, quotaName, scope, scopeValue)
}

func testAccQuotaV1Exists(name string) resource.TestCheckFunc {
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
