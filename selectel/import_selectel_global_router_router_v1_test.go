package selectel

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccGlobalRouterRouterV1ImportBasic(t *testing.T) {
	name := acctest.RandomWithPrefix("tf-acc") + "_router"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckGlobalRouterRouterV1Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGlobalRouterRouterV1Basic(name),
			},
			{
				ResourceName:            resourceGlobalRouterRouterName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"name"},
			},
		},
	})
}
