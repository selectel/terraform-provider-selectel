package selectel

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
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
			},
		},
	})
}
