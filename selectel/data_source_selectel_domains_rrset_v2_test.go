package selectel

import (
	"errors"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	domainsV2 "github.com/selectel/domains-go/pkg/v2"
)

const resourceRrsetName = "rrset_tf_acc_test_1"

func TestAccDomainsRrsetV2DataSourceBasic(t *testing.T) {
	testZoneName := fmt.Sprintf("%s.ru.", acctest.RandomWithPrefix("tf-acc"))
	testRrsetName := fmt.Sprintf("%[1]s.%[2]s", acctest.RandomWithPrefix("tf-acc"), testZoneName)
	testRrsetType := domainsV2.TXT
	testRrsetTTL := 60
	testRrrsetContent := fmt.Sprintf("\"%[1]s\"", acctest.RandString(16))
	dataSourceRrrsetName := fmt.Sprintf("data.selectel_domains_rrset_v2.%[1]s", resourceRrsetName)
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheckWithProjectID(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckDomainsV2RrsetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDomainsRrsetV2DataSourceBasic(resourceRrsetName, testRrsetName, string(testRrsetType), testRrrsetContent, testRrsetTTL, resourceZoneName, testZoneName),
				Check: resource.ComposeTestCheckFunc(
					testAccDomainsRrsetV2ID(dataSourceRrrsetName),
					resource.TestCheckResourceAttr(dataSourceRrrsetName, "name", testRrsetName),
					resource.TestCheckResourceAttr(dataSourceRrrsetName, "type", string(testRrsetType)),
					resource.TestCheckResourceAttrSet(dataSourceRrrsetName, "zone_id"),
				),
			},
		},
	})
}

func testAccDomainsRrsetV2ID(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("can't find rrset: %s", name)
		}

		if rs.Primary.ID == "" {
			return errors.New("rrset data source ID not set")
		}

		return nil
	}
}

func testAccDomainsRrsetV2DataSourceBasic(resourceRrsetName, rrsetName, rrsetType, rrsetContent string, ttl int, resourceZoneName, zoneName string) string {
	return fmt.Sprintf(`
	%[1]s

	%[2]s

	data "selectel_domains_rrset_v2" %[3]q {
	  name = selectel_domains_rrset_v2.%[3]s.name
	  type = selectel_domains_rrset_v2.%[3]s.type
	  zone_id = selectel_domains_zone_v2.%[4]s.id
	}
`, testAccDomainsZoneV2Basic(resourceZoneName, zoneName), testAccDomainsRrsetV2Basic(resourceRrsetName, rrsetName, rrsetType, rrsetContent, ttl, resourceZoneName), resourceRrsetName, resourceZoneName)
}
