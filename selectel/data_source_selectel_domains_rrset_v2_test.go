package selectel

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	domainsV2 "github.com/selectel/domains-go/pkg/v2"
)

func TestAccDomainsRRSetV2DataSourceBasic(t *testing.T) {
	testProjectName := acctest.RandomWithPrefix("tf-acc")
	testZoneName := fmt.Sprintf("%s.ru.", acctest.RandomWithPrefix("tf-acc"))
	testRRSetName := fmt.Sprintf("%[1]s.%[2]s", acctest.RandomWithPrefix("tf-acc"), testZoneName)
	testRRSetType := domainsV2.TXT
	testRRSetTTL := 60
	testRRSetContent := fmt.Sprintf("\"%[1]s\"", acctest.RandString(16))
	dataSourceRRSetName := fmt.Sprintf("data.selectel_domains_rrset_v2.%[1]s", resourceRRSetName)
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckDomainsV2RRSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDomainsRRSetV2DataSourceBasic(testProjectName, resourceRRSetName, testRRSetName, string(testRRSetType), testRRSetContent, testRRSetTTL, resourceZoneName, testZoneName),
				Check: resource.ComposeTestCheckFunc(
					testAccDomainsRRSetV2ID(dataSourceRRSetName),
					resource.TestCheckResourceAttr(dataSourceRRSetName, "name", testRRSetName),
					resource.TestCheckResourceAttr(dataSourceRRSetName, "type", string(testRRSetType)),
					resource.TestCheckResourceAttrSet(dataSourceRRSetName, "zone_id"),
				),
			},
		},
	})
}

func testAccDomainsRRSetV2DataSourceBasic(projectName, resourceRRSetName, rrsetName, rrsetType, rrsetContent string, ttl int, resourceZoneName, zoneName string) string {
	return fmt.Sprintf(`
	%[1]s
	data "selectel_domains_rrset_v2" %[2]q {
	  name = selectel_domains_rrset_v2.%[2]s.name
	  type = selectel_domains_rrset_v2.%[2]s.type
	  zone_id = selectel_domains_zone_v2.%[3]s.id
	  project_id = "${selectel_vpc_project_v2.project_tf_acc_test_1.id}"
	}
`, testAccDomainsRRSetV2WithZoneBasic(projectName, resourceRRSetName, rrsetName, rrsetType, rrsetContent, ttl, resourceZoneName, zoneName), resourceRRSetName, resourceZoneName)
}
