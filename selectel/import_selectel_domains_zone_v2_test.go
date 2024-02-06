package selectel

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccDomainsZoneV2ImportBasic(t *testing.T) {
	projectID := os.Getenv("SEL_PROJECT_ID")
	fullResourceName := fmt.Sprintf("selectel_domains_zone_v2.%[1]s", resourceZoneName)
	testZoneName := fmt.Sprintf("%s.xyz.", acctest.RandomWithPrefix("tf-acc"))
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheckWithProjectID(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckDomainsV2ZoneDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDomainsZoneV2WithoutProjectBasic(projectID, resourceZoneName, testZoneName),
			},
			{
				ResourceName:      fullResourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: getTestZoneIDForImport,
			},
		},
	})
}

func getTestZoneIDForImport(s *terraform.State) (string, error) {
	resourceZoneFullName := "selectel_domains_zone_v2.zone_tf_acc_test_1"
	resourceZone, ok := s.RootModule().Resources[resourceZoneFullName]
	if !ok {
		return "", fmt.Errorf("Not found zone: %s", resourceZoneFullName)
	}

	return resourceZone.Primary.Attributes["name"], nil
}

func testAccDomainsZoneV2WithoutProjectBasic(projectID, resourceName, zoneName string) string {
	return fmt.Sprintf(`
		resource "selectel_domains_zone_v2" %[2]q {
			name = %[3]q
			project_id = %[1]q
		}`, projectID, resourceName, zoneName)
}
