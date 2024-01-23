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

func TestAccDomainsRrsetV2Basic(t *testing.T) {
	testZoneName := fmt.Sprintf("%s.ru.", acctest.RandomWithPrefix("tf-acc"))
	testRrsetName := fmt.Sprintf("%[1]s.%[2]s", acctest.RandomWithPrefix("tf-acc"), testZoneName)
	testRrsetType := domainsV2.TXT
	testRrsetTTL := 60
	testRrrsetContent := fmt.Sprintf("\"%[1]s\"", acctest.RandString(16))
	resourceZoneName := "zone_tf_acc_test_1"
	resourceRrsetName := "rrset_tf_acc_test_1"
	dataSourceRrrsetName := fmt.Sprintf("selectel_domains_rrset_v2.%[1]s", resourceRrsetName)
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheckWithProjectID(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckDomainsV2ZoneDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDomainsRrsetV2WithZoneBasic(resourceRrsetName, testRrsetName, string(testRrsetType), testRrrsetContent, testRrsetTTL, resourceZoneName, testZoneName),
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

func testAccDomainsRrsetV2WithZoneBasic(resourceRrsetName, rrsetName, rrsetType, rrsetContent string, ttl int, resourceZoneName, zoneName string) string {
	return fmt.Sprintf(`
	%[7]s

	resource "selectel_domains_rrset_v2" %[1]q {
		name = %[2]q
		type = %[3]q
		ttl = %[4]d
		zone_id = selectel_domains_zone_v2.%[6]s.id
		records {
			content = %[5]q
			disabled = false
		}
	}`, resourceRrsetName, rrsetName, rrsetType, ttl, rrsetContent, resourceZoneName, testAccDomainsZoneV2Basic(resourceZoneName, zoneName))
}

func testAccDomainsRrsetV2Basic(resourceRrsetName, rrsetName, rrsetType, rrsetContent string, ttl int, resourceZoneName string) string {
	return fmt.Sprintf(`
	resource "selectel_domains_rrset_v2" %[1]q {
		name = %[2]q
		type = %[3]q
		ttl = %[4]d
		zone_id = selectel_domains_zone_v2.%[5]s.id
		records {
			content = %[6]q
			disabled = false
		}
	}`, resourceRrsetName, rrsetName, rrsetType, ttl, resourceZoneName, rrsetContent)
}

func testAccCheckDomainsV2RrsetDestroy(s *terraform.State) error {
	meta := testAccProvider.Meta()
	client, err := getDomainsV2Client(meta)
	if err != nil {
		return err
	}

	ctx := context.Background()

	for _, rs := range s.RootModule().Resources {
		log.Printf("RT: %s", rs.Type)
		if rs.Type != "selectel_domains_rrset_v2" {
			continue
		}

		zoneID := rs.Primary.Attributes["zone_id"]
		rrsetID := rs.Primary.ID

		_, err = client.GetRRSet(ctx, zoneID, rrsetID)
		if err == nil {
			return errors.New("rrset still exists")
		}
	}

	return nil
}
