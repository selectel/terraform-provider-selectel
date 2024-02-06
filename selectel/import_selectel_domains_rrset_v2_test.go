package selectel

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	domainsV2 "github.com/selectel/domains-go/pkg/v2"
)

func TestAccDomainsRRSetV2ImportBasic(t *testing.T) {
	projectName := acctest.RandomWithPrefix("tf-acc")
	testZoneName := fmt.Sprintf("%s.xyz.", acctest.RandomWithPrefix("tf-acc"))
	testRRSetName := fmt.Sprintf("%[1]s.%[2]s", acctest.RandomWithPrefix("tf-acc"), testZoneName)
	testRRSetType := domainsV2.TXT
	testRRSetTTL := 60
	testRRSetContent := fmt.Sprintf("\"%[1]s\"", acctest.RandString(16))
	fullResourceName := fmt.Sprintf("selectel_domains_rrset_v2.%[1]s", resourceRRSetName)
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckVPCV2ProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDomainsRRSetV2WithZoneBasic(
					projectName,
					resourceRRSetName, testRRSetName, string(testRRSetType), testRRSetContent, testRRSetTTL,
					resourceZoneName, testZoneName,
				),
				Check: testAccCheckSelectelImportEnv(fullResourceName),
			},
			{
				ImportStateIdFunc: getTestRRSetIDForImport,
				ResourceName:      fullResourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func getTestRRSetIDForImport(s *terraform.State) (string, error) {
	resourceZoneFullName := "selectel_domains_zone_v2.zone_tf_acc_test_1"
	resourceRRSetFullName := "selectel_domains_rrset_v2.rrset_tf_acc_test_1"
	resourceZone, ok := s.RootModule().Resources[resourceZoneFullName]
	if !ok {
		return "", fmt.Errorf("Not found zone: %s", resourceZoneFullName)
	}
	resourceRRSet, ok := s.RootModule().Resources[resourceRRSetFullName]
	if !ok {
		return "", fmt.Errorf("Not found rrset: %s", resourceRRSetFullName)
	}

	return fmt.Sprintf("%s/%s/%s",
		resourceZone.Primary.Attributes["name"],
		resourceRRSet.Primary.Attributes["name"],
		resourceRRSet.Primary.Attributes["type"],
	), nil
}
