package selectel

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccGlobalRouterDedicatedSubnetV1ImportBasic(t *testing.T) {
	subnetName := acctest.RandomWithPrefix("tf-acc") + "_subnet"
	networkName := acctest.RandomWithPrefix("tf-acc") + "_network"
	routerName := acctest.RandomWithPrefix("tf-acc") + "_router"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccSelectelPreCheck(t)
			testAccGlobalRouterDedicatedNetworkPreCheck(t)
			testAccGlobalRouterSubnetPreCheck(t)
		},
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckGlobalRouterDedicatedSubnetV1Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGlobalRouterDedicatedSubnetV1Basic(
					routerName, networkName, subnetName,
				),
			},
			{
				ResourceName:            resourceGlobalRouterDedicatedSubnetName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"name"},
			},
		},
	})
}
