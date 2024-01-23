package selectel

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccDomainsZoneV2ImportBasic(t *testing.T) {
	resourceName := "zone_tf_acc_test_1"
	fullResourceName := fmt.Sprintf("selectel_domains_zone_v2.%[1]s", resourceName)
	testZoneName := fmt.Sprintf("%s.xyz.", acctest.RandomWithPrefix("tf-acc"))

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheckWithProjectID(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckDomainsV2ZoneDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDomainsZoneV2Basic(resourceName, testZoneName),
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
