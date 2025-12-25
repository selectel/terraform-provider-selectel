package selectel

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccGlobalRouterDedicatedNetworkV1ImportBasic(t *testing.T) {
	networkName := acctest.RandomWithPrefix("tf-acc") + "_network"
	routerName := acctest.RandomWithPrefix("tf-acc") + "_router"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccSelectelPreCheck(t)
			testAccGlobalRouterDedicatedNetworkPreCheck(t)
		},
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckGlobalRouterDedicatedNetworkV1Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGlobalRouterDedicatedNetworkV1Basic(routerName, networkName),
			},
			{
				ResourceName:            resourceGlobalRouterDedicatedNetworkName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"name"},
			},
		},
	})
}
