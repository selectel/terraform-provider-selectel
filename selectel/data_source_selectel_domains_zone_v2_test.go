package selectel

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDomainsZoneV2DataSourceBasic(t *testing.T) {
	testProjectName := acctest.RandomWithPrefix("tf-acc")
	testZoneName := fmt.Sprintf("%s.ru.", acctest.RandomWithPrefix("tf-acc"))
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckDomainsV2ZoneDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDomainsZoneV2DataSourceBasic(testProjectName, resourceZoneName, testZoneName),
				Check: resource.ComposeTestCheckFunc(
					testAccDomainsZoneV2Exists(fmt.Sprintf("data.selectel_domains_zone_v2.%[1]s", resourceZoneName)),
					resource.TestCheckResourceAttr(fmt.Sprintf("data.selectel_domains_zone_v2.%[1]s", resourceZoneName), "name", testZoneName),
				),
			},
		},
	})
}

func testAccDomainsZoneV2DataSourceBasic(projectName, resourceName, zoneName string) string {
	return fmt.Sprintf(`
	%[1]s
	data "selectel_domains_zone_v2" %[2]q {
	  name = selectel_domains_zone_v2.%[2]s.name
	  project_id = "${selectel_vpc_project_v2.project_tf_acc_test_1.id}"
	}
`, testAccDomainsZoneV2Basic(projectName, resourceName, zoneName), resourceName, zoneName)
}
