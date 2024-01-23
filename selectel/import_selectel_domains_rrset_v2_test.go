package selectel

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	domainsV2 "github.com/selectel/domains-go/pkg/v2"
)

func TestAccDomainsRrsetV2ImportBasic(t *testing.T) {
	testZoneName := fmt.Sprintf("%s.xyz.", acctest.RandomWithPrefix("tf-acc"))
	testRrsetName := fmt.Sprintf("%[1]s.%[2]s", acctest.RandomWithPrefix("tf-acc"), testZoneName)
	testRrsetType := domainsV2.TXT
	testRrsetTTL := 60
	testRrrsetContent := fmt.Sprintf("\"%[1]s\"", acctest.RandString(16))
	fullResourceName := fmt.Sprintf("selectel_domains_rrset_v2.%[1]s", resourceRrsetName)
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheckWithProjectID(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckDomainsV2ZoneDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDomainsRrsetV2WithZoneBasic(
					resourceRrsetName, testRrsetName, string(testRrsetType), testRrrsetContent, testRrsetTTL,
					resourceZoneName, testZoneName,
				),
			},
			{
				ImportStateIdFunc: getTestRrsetIDForImport,
				ResourceName:      fullResourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func getTestRrsetIDForImport(s *terraform.State) (string, error) {
	resourceZoneFullName := "selectel_domains_zone_v2.zone_tf_acc_test_1"
	resourceRrsetFullName := "selectel_domains_rrset_v2.rrset_tf_acc_test_1"
	resourceZone, ok := s.RootModule().Resources[resourceZoneFullName]
	if !ok {
		return "", fmt.Errorf("Not found zone: %s", resourceZoneFullName)
	}
	resourceRrset, ok := s.RootModule().Resources[resourceRrsetFullName]
	if !ok {
		return "", fmt.Errorf("Not found rrset: %s", resourceRrsetFullName)
	}

	return fmt.Sprintf("%s/%s/%s",
		resourceZone.Primary.Attributes["name"],
		resourceRrset.Primary.Attributes["name"],
		resourceRrset.Primary.Attributes["type"],
	), nil
}
