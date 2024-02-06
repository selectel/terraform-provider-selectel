package selectel

import (
	"context"
	"errors"
	"fmt"
	"log"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	domainsV2 "github.com/selectel/domains-go/pkg/v2"
)

func TestAccDomainsRRSetV2Basic(t *testing.T) {
	projectName := acctest.RandomWithPrefix("tf-acc")
	testZoneName := fmt.Sprintf("%s.ru.", acctest.RandomWithPrefix("tf-acc"))
	testRRSetName := fmt.Sprintf("%[1]s.%[2]s", acctest.RandomWithPrefix("tf-acc"), testZoneName)
	testRRSetType := domainsV2.TXT
	testRRSetTTL := 60
	testRRSetContent := fmt.Sprintf("\"%[1]s\"", acctest.RandString(16))
	resourceZoneName := "zone_tf_acc_test_1"
	resourceRRSetName := "rrset_tf_acc_test_1"
	dataSourceRRSetName := fmt.Sprintf("selectel_domains_rrset_v2.%[1]s", resourceRRSetName)
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckDomainsV2ZoneDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDomainsRRSetV2WithZoneBasic(projectName, resourceRRSetName, testRRSetName, string(testRRSetType), testRRSetContent, testRRSetTTL, resourceZoneName, testZoneName),
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

func testAccDomainsRRSetV2WithZoneBasic(projectName, resourceRRSetName, rrsetName, rrsetType, rrsetContent string, ttl int, resourceZoneName, zoneName string) string {
	return fmt.Sprintf(`
	%[7]s

	resource "selectel_domains_rrset_v2" %[1]q {
		name = %[2]q
		type = %[3]q
		ttl = %[4]d
		zone_id = selectel_domains_zone_v2.%[6]s.id
		project_id = "${selectel_vpc_project_v2.project_tf_acc_test_1.id}"
		records {
			content = %[5]q
			disabled = false
		}
	}`, resourceRRSetName, rrsetName, rrsetType, ttl, rrsetContent, resourceZoneName, testAccDomainsZoneV2Basic(projectName, resourceZoneName, zoneName))
}

func testAccDomainsRRSetV2Basic(resourceRRSetName, rrsetName, rrsetType, rrsetContent string, ttl int, resourceZoneName string) string {
	return fmt.Sprintf(`
	resource "selectel_domains_rrset_v2" %[1]q {
		name = %[2]q
		type = %[3]q
		ttl = %[4]d
		zone_id = selectel_domains_zone_v2.%[5]s.id
		project_id = "${selectel_vpc_project_v2.project_tf_acc_test_1.id}"
		records {
			content = %[6]q
			disabled = false
		}
	}`, resourceRRSetName, rrsetName, rrsetType, ttl, resourceZoneName, rrsetContent)
}

func testAccCheckDomainsV2RRSetDestroy(s *terraform.State) error {
	ctx := context.Background()

	for _, rs := range s.RootModule().Resources {
		log.Printf("RT: %s", rs.Type)
		if rs.Type != "selectel_domains_rrset_v2" {
			continue
		}

		zoneID := rs.Primary.Attributes["zone_id"]
		rrsetID := rs.Primary.ID
		client, err := getDomainsV2ClientTest(rs, testAccProvider)
		if err != nil {
			return err
		}
		_, err = client.GetRRSet(ctx, zoneID, rrsetID)
		if err == nil {
			return errors.New("rrset still exists")
		}
	}

	return nil
}
