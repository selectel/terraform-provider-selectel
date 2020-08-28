package selectel

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDomainsRecordV1ImportBasic(t *testing.T) {
	resourceName := "selectel_domains_record_v1.record_a_tf_acc_test_1"
	testDomainName := fmt.Sprintf("%s.xyz", acctest.RandomWithPrefix("tf-acc"))
	testRecordName := fmt.Sprintf("a.%s", testDomainName)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckDomainsV1DomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDomainsRecordV1BasicSingle(testDomainName, testRecordName),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"domain_id"},
			},
		},
	})
}

func testAccDomainsRecordV1BasicSingle(domainName, recordName string) string {
	return fmt.Sprintf(`
resource "selectel_domains_domain_v1" "domain_tf_acc_test_1" {
  name = "%s"
}

resource "selectel_domains_record_v1" "record_a_tf_acc_test_1" {
  domain_id = selectel_domains_domain_v1.domain_tf_acc_test_1.id
  name = "%s"
  type = "A"
  content = "127.0.0.1"
  ttl  = 60
}
`, domainName, recordName)
}
