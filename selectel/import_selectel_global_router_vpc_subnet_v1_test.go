package selectel

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccGlobalRouterVPCSubnetV1ImportBasic(t *testing.T) {
	subnetName := acctest.RandomWithPrefix("tf-acc") + "_subnet"
	networkName := acctest.RandomWithPrefix("tf-acc") + "_network"
	routerName := acctest.RandomWithPrefix("tf-acc") + "_router"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccSelectelPreCheck(t)
			testAccGlobalRouterVPCNetworkPreCheck(t)
			testAccGlobalRouterSubnetPreCheck(t)
		},
		ProviderFactories: testAccProvidersWithOpenStack,
		CheckDestroy:      testAccCheckGlobalRouterVPCSubnetV1Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGlobalRouterVPCSubnetV1Basic(
					routerName, networkName, subnetName,
				),
			},
			{
				ResourceName:            resourceGlobalRouterVPCSubnetName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"name"},
			},
		},
	})
}
