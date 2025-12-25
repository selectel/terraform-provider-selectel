package selectel

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccGlobalRouterStaticRouteV1ImportBasic(t *testing.T) {
	staticRouteName := acctest.RandomWithPrefix("tf-acc") + "_static_route"
	subnetName := acctest.RandomWithPrefix("tf-acc") + "_subent"
	networkName := acctest.RandomWithPrefix("tf-acc") + "_network"
	routerName := acctest.RandomWithPrefix("tf-acc") + "_router"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccSelectelPreCheck(t)
			testAccGlobalRouterVPCNetworkPreCheck(t)
			testAccGlobalRouterSubnetPreCheck(t)
			testAccGlobalRouterStaticRoutePreCheck(t)
		},
		ProviderFactories: testAccProvidersWithOpenStack,
		CheckDestroy:      testAccCheckGlobalRouterStaticRouteV1Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGlobalRouterStaticRouteV1Basic(
					routerName, networkName, subnetName, staticRouteName,
				),
			},
			{
				ResourceName:            resourceGlobalRouterStaticRouteName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"name"},
			},
		},
	})
}
