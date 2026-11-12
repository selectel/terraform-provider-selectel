package selectel

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccVPCV2ProjectImportBasic(t *testing.T) {
	resourceName := "selectel_vpc_project_v2.project_tf_acc_test_1"
	projectName := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckVPCV2ProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCV2ProjectBasic(projectName),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"all_quotas", "quotas"},
				ImportStateCheck:        testAccCheckVPCV2ProjectImportedQuotas,
			},
		},
	})
}

func testAccCheckVPCV2ProjectImportedQuotas(states []*terraform.InstanceState) error {
	if len(states) != 1 {
		return fmt.Errorf("expected 1 imported state, got %d", len(states))
	}

	count, ok := states[0].Attributes["quotas.#"]
	if !ok || count == "" || count == "0" {
		return fmt.Errorf("expected quotas to be set after import, got quotas.# = %q", count)
	}

	return nil
}
