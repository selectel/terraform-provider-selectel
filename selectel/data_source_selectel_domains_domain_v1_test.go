package selectel

import (
	"errors"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccDomainsDomainV1DataSourceBasic(t *testing.T) {
	testDomainName := fmt.Sprintf("%s.xyz", acctest.RandomWithPrefix("tf-acc"))

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckDomainsV1DomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDomainsDomainV1Basic(testDomainName),
			},
			{
				Config: testAccDomainsDomainV1DataSourceBasic(testDomainName),
				Check: resource.ComposeTestCheckFunc(
					testAccDomainsDomainV1DataSourceID("data.selectel_domains_domain_v1.domain_tf_acc_test_1"),
					resource.TestCheckResourceAttr("data.selectel_domains_domain_v1.domain_tf_acc_test_1", "name", testDomainName),
				),
			},
		},
	})
}

func testAccDomainsDomainV1DataSourceID(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("can't find domain data source: %s", name)
		}

		if rs.Primary.ID == "" {
			return errors.New("domain data source ID not set")
		}

		return nil
	}
}

func testAccDomainsDomainV1DataSourceBasic(name string) string {
	return fmt.Sprintf(`
	%s

	data "selectel_domains_domain_v1" "domain_tf_acc_test_1" {
	  name = "${selectel_domains_domain_v1.domain_tf_acc_test_1.name}"
	}
`, testAccDomainsDomainV1Basic(name))
}
