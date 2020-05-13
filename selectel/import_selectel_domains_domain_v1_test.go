package selectel

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

func TestAccDomainsDomainV1ImportBasic(t *testing.T) {
	resourceName := "selectel_domains_domain_v1.domain_tf_acc_test_1"
	testDomainName := fmt.Sprintf("%s.xyz", acctest.RandomWithPrefix("tf-acc"))

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccSelectelPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckDomainsV1DomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDomainsDomainV1Basic(testDomainName),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}
