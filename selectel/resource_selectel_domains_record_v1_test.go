package selectel

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"github.com/selectel/domains-go/pkg/v1/domain"
	"github.com/selectel/domains-go/pkg/v1/record"
)

func TestAccDomainsRecordV1Basic(t *testing.T) {
	var (
		testDomain      domain.View
		testRecordA     record.View
		testRecordAAAA  record.View
		testRecordCNAME record.View
		testRecordTXT   record.View
		testRecordNS    record.View
		testRecordMX    record.View
		testRecordSRV   record.View
	)

	testDomainName := fmt.Sprintf("%s.xyz", acctest.RandomWithPrefix("tf-acc"))
	testRecordNameA := fmt.Sprintf("a.%s", testDomainName)
	testRecordNameAAAA := fmt.Sprintf("aaaa.%s", testDomainName)
	testRecordNameCNAME := fmt.Sprintf("cname.%s", testDomainName)
	testRecordNameTXT := fmt.Sprintf("txt.%s", testDomainName)
	testRecordNameNS := fmt.Sprintf("ns.%s", testDomainName)
	testRecordNameMX := fmt.Sprintf("mx.%s", testDomainName)
	testRecordNameSRV := fmt.Sprintf("srv.%s", testDomainName)

	//nolint
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccSelectelPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckDomainsRecordV1Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDomainsRecordV1Basic(
					testDomainName,
					testRecordNameA,
					testRecordNameAAAA,
					testRecordNameCNAME,
					testRecordNameTXT,
					testRecordNameNS,
					testRecordNameMX,
					testRecordNameSRV),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainsDomainV1Exists("selectel_domains_domain_v1.domain_tf_acc_test_1",
						&testDomain),
					// Record type A check
					testAccCheckDomainsRecordV1Exists("selectel_domains_record_v1.record_a_tf_acc_test_1",
						&testRecordA),
					resource.TestCheckResourceAttr("selectel_domains_record_v1.record_a_tf_acc_test_1",
						"name",
						testRecordNameA),
					resource.TestCheckResourceAttr("selectel_domains_record_v1.record_a_tf_acc_test_1",
						"type",
						"A"),
					resource.TestCheckResourceAttr("selectel_domains_record_v1.record_a_tf_acc_test_1",
						"content",
						"127.0.0.1"),
					resource.TestCheckResourceAttr("selectel_domains_record_v1.record_a_tf_acc_test_1",
						"ttl",
						"60"),
					// Record type AAAA check
					testAccCheckDomainsRecordV1Exists("selectel_domains_record_v1.record_aaaa_tf_acc_test_1",
						&testRecordAAAA),
					resource.TestCheckResourceAttr("selectel_domains_record_v1.record_aaaa_tf_acc_test_1",
						"name",
						testRecordNameAAAA),
					resource.TestCheckResourceAttr("selectel_domains_record_v1.record_aaaa_tf_acc_test_1",
						"type",
						"AAAA"),
					resource.TestCheckResourceAttr("selectel_domains_record_v1.record_aaaa_tf_acc_test_1",
						"content",
						"2400:cb00:2049:1::a29f:1804"),
					resource.TestCheckResourceAttr("selectel_domains_record_v1.record_aaaa_tf_acc_test_1",
						"ttl",
						"60"),
					// Record type CNAME check
					testAccCheckDomainsRecordV1Exists("selectel_domains_record_v1.record_cname_tf_acc_test_1",
						&testRecordCNAME),
					resource.TestCheckResourceAttr("selectel_domains_record_v1.record_cname_tf_acc_test_1",
						"name",
						testRecordNameCNAME),
					resource.TestCheckResourceAttr("selectel_domains_record_v1.record_cname_tf_acc_test_1",
						"type",
						"CNAME"),
					resource.TestCheckResourceAttr("selectel_domains_record_v1.record_cname_tf_acc_test_1",
						"content",
						"origin.com"),
					resource.TestCheckResourceAttr("selectel_domains_record_v1.record_cname_tf_acc_test_1",
						"ttl",
						"60"),
					// Record type TXT check
					testAccCheckDomainsRecordV1Exists("selectel_domains_record_v1.record_txt_tf_acc_test_1",
						&testRecordTXT),
					resource.TestCheckResourceAttr("selectel_domains_record_v1.record_txt_tf_acc_test_1",
						"name",
						testRecordNameTXT),
					resource.TestCheckResourceAttr("selectel_domains_record_v1.record_txt_tf_acc_test_1",
						"type",
						"TXT"),
					resource.TestCheckResourceAttr("selectel_domains_record_v1.record_txt_tf_acc_test_1",
						"content",
						"hello, world!"),
					resource.TestCheckResourceAttr("selectel_domains_record_v1.record_txt_tf_acc_test_1",
						"ttl",
						"60"),
					// Record type NS check
					testAccCheckDomainsRecordV1Exists("selectel_domains_record_v1.record_ns_tf_acc_test_1",
						&testRecordNS),
					resource.TestCheckResourceAttr("selectel_domains_record_v1.record_ns_tf_acc_test_1",
						"name",
						testRecordNameNS),
					resource.TestCheckResourceAttr("selectel_domains_record_v1.record_ns_tf_acc_test_1",
						"type",
						"NS"),
					resource.TestCheckResourceAttr("selectel_domains_record_v1.record_ns_tf_acc_test_1",
						"content",
						"ns.example.org"),
					resource.TestCheckResourceAttr("selectel_domains_record_v1.record_ns_tf_acc_test_1",
						"ttl",
						"60"),
					// Record type MX check
					testAccCheckDomainsRecordV1Exists("selectel_domains_record_v1.record_mx_tf_acc_test_1",
						&testRecordMX),
					resource.TestCheckResourceAttr("selectel_domains_record_v1.record_mx_tf_acc_test_1",
						"name",
						testRecordNameMX),
					resource.TestCheckResourceAttr("selectel_domains_record_v1.record_mx_tf_acc_test_1",
						"type",
						"MX"),
					resource.TestCheckResourceAttr("selectel_domains_record_v1.record_mx_tf_acc_test_1",
						"content",
						"mail1.example.org"),
					resource.TestCheckResourceAttr("selectel_domains_record_v1.record_mx_tf_acc_test_1",
						"ttl",
						"60"),
					resource.TestCheckResourceAttr("selectel_domains_record_v1.record_mx_tf_acc_test_1",
						"priority",
						"10"),
					// Record type SRV check
					testAccCheckDomainsRecordV1Exists("selectel_domains_record_v1.record_srv_tf_acc_test_1",
						&testRecordSRV),
					resource.TestCheckResourceAttr("selectel_domains_record_v1.record_srv_tf_acc_test_1",
						"name",
						testRecordNameSRV),
					resource.TestCheckResourceAttr("selectel_domains_record_v1.record_srv_tf_acc_test_1",
						"type",
						"SRV"),
					resource.TestCheckResourceAttr("selectel_domains_record_v1.record_srv_tf_acc_test_1",
						"target",
						"backupbox.example.com"),
					resource.TestCheckResourceAttr("selectel_domains_record_v1.record_srv_tf_acc_test_1",
						"ttl",
						"60"),
					resource.TestCheckResourceAttr("selectel_domains_record_v1.record_srv_tf_acc_test_1",
						"port",
						"5060"),
					resource.TestCheckResourceAttr("selectel_domains_record_v1.record_srv_tf_acc_test_1",
						"weight",
						"10"),
					resource.TestCheckResourceAttr("selectel_domains_record_v1.record_srv_tf_acc_test_1",
						"priority",
						"0"),
				),
			},
			{
				Config: testAccDomainsRecordV1Update(
					testDomainName,
					testRecordNameA,
					testRecordNameAAAA,
					testRecordNameCNAME,
					testRecordNameTXT,
					testRecordNameNS,
					testRecordNameMX,
					testRecordNameSRV),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainsDomainV1Exists("selectel_domains_domain_v1.domain_tf_acc_test_1",
						&testDomain),
					// Record type A check
					testAccCheckDomainsRecordV1Exists("selectel_domains_record_v1.record_a_tf_acc_test_1",
						&testRecordA),
					resource.TestCheckResourceAttr("selectel_domains_record_v1.record_a_tf_acc_test_1",
						"name",
						testRecordNameA),
					resource.TestCheckResourceAttr("selectel_domains_record_v1.record_a_tf_acc_test_1",
						"type",
						"A"),
					resource.TestCheckResourceAttr("selectel_domains_record_v1.record_a_tf_acc_test_1",
						"content",
						"10.10.10.10"),
					resource.TestCheckResourceAttr("selectel_domains_record_v1.record_a_tf_acc_test_1",
						"ttl",
						"120"),
					// Record type AAAA check
					testAccCheckDomainsRecordV1Exists("selectel_domains_record_v1.record_aaaa_tf_acc_test_1",
						&testRecordAAAA),
					resource.TestCheckResourceAttr("selectel_domains_record_v1.record_aaaa_tf_acc_test_1",
						"name",
						testRecordNameAAAA),
					resource.TestCheckResourceAttr("selectel_domains_record_v1.record_aaaa_tf_acc_test_1",
						"type",
						"AAAA"),
					resource.TestCheckResourceAttr("selectel_domains_record_v1.record_aaaa_tf_acc_test_1",
						"content",
						"2400:cb00:2049:1::a29f:1804"),
					resource.TestCheckResourceAttr("selectel_domains_record_v1.record_aaaa_tf_acc_test_1",
						"ttl",
						"60"),
					// Record type CNAME check
					testAccCheckDomainsRecordV1Exists("selectel_domains_record_v1.record_cname_tf_acc_test_1",
						&testRecordCNAME),
					resource.TestCheckResourceAttr("selectel_domains_record_v1.record_cname_tf_acc_test_1",
						"name",
						testRecordNameCNAME),
					resource.TestCheckResourceAttr("selectel_domains_record_v1.record_cname_tf_acc_test_1",
						"type",
						"CNAME"),
					resource.TestCheckResourceAttr("selectel_domains_record_v1.record_cname_tf_acc_test_1",
						"content",
						"origin.com"),
					resource.TestCheckResourceAttr("selectel_domains_record_v1.record_cname_tf_acc_test_1",
						"ttl",
						"60"),
					// Record type TXT check
					testAccCheckDomainsRecordV1Exists("selectel_domains_record_v1.record_txt_tf_acc_test_1",
						&testRecordTXT),
					resource.TestCheckResourceAttr("selectel_domains_record_v1.record_txt_tf_acc_test_1",
						"name",
						testRecordNameTXT),
					resource.TestCheckResourceAttr("selectel_domains_record_v1.record_txt_tf_acc_test_1",
						"type",
						"TXT"),
					resource.TestCheckResourceAttr("selectel_domains_record_v1.record_txt_tf_acc_test_1",
						"content",
						"hello, world!!!1"),
					resource.TestCheckResourceAttr("selectel_domains_record_v1.record_txt_tf_acc_test_1",
						"ttl",
						"60"),
					// Record type NS check
					testAccCheckDomainsRecordV1Exists("selectel_domains_record_v1.record_ns_tf_acc_test_1",
						&testRecordNS),
					resource.TestCheckResourceAttr("selectel_domains_record_v1.record_ns_tf_acc_test_1",
						"name",
						testRecordNameNS),
					resource.TestCheckResourceAttr("selectel_domains_record_v1.record_ns_tf_acc_test_1",
						"type",
						"NS"),
					resource.TestCheckResourceAttr("selectel_domains_record_v1.record_ns_tf_acc_test_1",
						"content",
						"ns.example.org"),
					resource.TestCheckResourceAttr("selectel_domains_record_v1.record_ns_tf_acc_test_1",
						"ttl",
						"60"),
					// Record type MX check
					testAccCheckDomainsRecordV1Exists("selectel_domains_record_v1.record_mx_tf_acc_test_1",
						&testRecordMX),
					resource.TestCheckResourceAttr("selectel_domains_record_v1.record_mx_tf_acc_test_1",
						"name",
						testRecordNameMX),
					resource.TestCheckResourceAttr("selectel_domains_record_v1.record_mx_tf_acc_test_1",
						"type",
						"MX"),
					resource.TestCheckResourceAttr("selectel_domains_record_v1.record_mx_tf_acc_test_1",
						"content",
						"mail.example.org"),
					resource.TestCheckResourceAttr("selectel_domains_record_v1.record_mx_tf_acc_test_1",
						"ttl",
						"60"),
					resource.TestCheckResourceAttr("selectel_domains_record_v1.record_mx_tf_acc_test_1",
						"priority",
						"10"),
					// Record type SRV check
					testAccCheckDomainsRecordV1Exists("selectel_domains_record_v1.record_srv_tf_acc_test_1",
						&testRecordSRV),
					resource.TestCheckResourceAttr("selectel_domains_record_v1.record_srv_tf_acc_test_1",
						"name",
						testRecordNameSRV),
					resource.TestCheckResourceAttr("selectel_domains_record_v1.record_srv_tf_acc_test_1",
						"type",
						"SRV"),
					resource.TestCheckResourceAttr("selectel_domains_record_v1.record_srv_tf_acc_test_1",
						"target",
						"backupbox.example.com"),
					resource.TestCheckResourceAttr("selectel_domains_record_v1.record_srv_tf_acc_test_1",
						"ttl",
						"120"),
					resource.TestCheckResourceAttr("selectel_domains_record_v1.record_srv_tf_acc_test_1",
						"port",
						"5061"),
					resource.TestCheckResourceAttr("selectel_domains_record_v1.record_srv_tf_acc_test_1",
						"weight",
						"20"),
					resource.TestCheckResourceAttr("selectel_domains_record_v1.record_srv_tf_acc_test_1",
						"priority",
						"10"),
				),
			},
		},
	})
}

func testAccDomainsRecordV1Basic(
	domainName,
	recordNameA,
	recordNameAAAA,
	recordNameCNAME,
	recordNameTXT,
	recordNameNS,
	recordNameMX,
	recordNameSRV string) string {
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

resource "selectel_domains_record_v1" "record_aaaa_tf_acc_test_1" {
  domain_id = selectel_domains_domain_v1.domain_tf_acc_test_1.id
  name = "%s"
  type = "AAAA"
  content = "2400:cb00:2049:1::a29f:1804"
  ttl  = 60
}

resource "selectel_domains_record_v1" "record_cname_tf_acc_test_1" {
  domain_id = selectel_domains_domain_v1.domain_tf_acc_test_1.id
  name = "%s"
  type = "CNAME"
  content = "origin.com"
  ttl = 60
}

resource "selectel_domains_record_v1" "record_txt_tf_acc_test_1" {
  domain_id = selectel_domains_domain_v1.domain_tf_acc_test_1.id
  name = "%s"
  type = "TXT"
  content = "hello, world!"
  ttl = 60
}


resource "selectel_domains_record_v1" "record_ns_tf_acc_test_1" {
  domain_id = selectel_domains_domain_v1.domain_tf_acc_test_1.id
  name = "%s"
  type = "NS"
  content = "ns.example.org"
  ttl = 60
}

resource "selectel_domains_record_v1" "record_mx_tf_acc_test_1" {
  domain_id = selectel_domains_domain_v1.domain_tf_acc_test_1.id
  name = "%s"
  type = "MX"
  content = "mail1.example.org"
  ttl = 60
  priority = 10
}

resource "selectel_domains_record_v1" "record_srv_tf_acc_test_1" {
  domain_id = selectel_domains_domain_v1.domain_tf_acc_test_1.id
  name = "%s"
  type = "SRV"
  target = "backupbox.example.com"
  ttl = 60
  priority = 0
  weight = 10
  port = 5060
}
`,
		domainName,
		recordNameA,
		recordNameAAAA,
		recordNameCNAME,
		recordNameTXT,
		recordNameNS,
		recordNameMX,
		recordNameSRV)
}

func testAccDomainsRecordV1Update(
	domainName,
	recordNameA,
	recordNameAAAA,
	recordNameCNAME,
	recordNameTXT,
	recordNameNS,
	recordNameMX,
	recordNameSRV string) string {
	return fmt.Sprintf(`
resource "selectel_domains_domain_v1" "domain_tf_acc_test_1" {
  name = "%s"
}

resource "selectel_domains_record_v1" "record_a_tf_acc_test_1" {
  domain_id = selectel_domains_domain_v1.domain_tf_acc_test_1.id
  name = "%s"
  type = "A"
  content = "10.10.10.10"
  ttl  = 120
}

resource "selectel_domains_record_v1" "record_aaaa_tf_acc_test_1" {
  domain_id = selectel_domains_domain_v1.domain_tf_acc_test_1.id
  name = "%s"
  type = "AAAA"
  content = "2400:cb00:2049:1::a29f:1804"
  ttl  = 60
}

resource "selectel_domains_record_v1" "record_cname_tf_acc_test_1" {
  domain_id = selectel_domains_domain_v1.domain_tf_acc_test_1.id
  name = "%s"
  type = "CNAME"
  content = "origin.com"
  ttl = 60
}

resource "selectel_domains_record_v1" "record_txt_tf_acc_test_1" {
  domain_id = selectel_domains_domain_v1.domain_tf_acc_test_1.id
  name = "%s"
  type = "TXT"
  content = "hello, world!!!1"
  ttl = 60
}


resource "selectel_domains_record_v1" "record_ns_tf_acc_test_1" {
  domain_id = selectel_domains_domain_v1.domain_tf_acc_test_1.id
  name = "%s"
  type = "NS"
  content = "ns.example.org"
  ttl = 60
}

resource "selectel_domains_record_v1" "record_mx_tf_acc_test_1" {
  domain_id = selectel_domains_domain_v1.domain_tf_acc_test_1.id
  name = "%s"
  type = "MX"
  content = "mail.example.org"
  ttl = 60
  priority = 10
}

resource "selectel_domains_record_v1" "record_srv_tf_acc_test_1" {
  domain_id = selectel_domains_domain_v1.domain_tf_acc_test_1.id
  name = "%s"
  type = "SRV"
  target = "backupbox.example.com"
  ttl = 120
  priority = 10
  weight = 20
  port = 5061
}
`,
		domainName,
		recordNameA,
		recordNameAAAA,
		recordNameCNAME,
		recordNameTXT,
		recordNameNS,
		recordNameMX,
		recordNameSRV)
}

func testAccCheckDomainsRecordV1Exists(n string, domainRecord *record.View) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}
		if rs.Primary.ID == "" {
			return errors.New("no ID is set")
		}

		domainID, recordID, err := domainsV1ParseDomainRecordIDsPair(rs.Primary.ID)
		if err != nil {
			return errParseDomainsDomainRecordV1IDsPair(rs.Primary.ID)
		}

		config := testAccProvider.Meta().(*Config)
		client := config.domainsV1Client()
		ctx := context.Background()

		foundRecord, _, err := record.Get(ctx, client, domainID, recordID)
		if err != nil {
			return err
		}

		if foundRecord.ID != recordID {
			return errors.New("record not found")
		}

		*domainRecord = *foundRecord

		return nil
	}
}

func testAccCheckDomainsRecordV1Destroy(s *terraform.State) error {
	config := testAccProvider.Meta().(*Config)
	domainsClientV1 := config.domainsV1Client()
	ctx := context.Background()

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "selectel_domains_domain_v1" {
			continue
		}

		domainID, err := strconv.Atoi(rs.Primary.ID)
		if err != nil {
			return errParseDomainsDomainV1ID(rs.Primary.ID)
		}

		_, _, err = domain.GetByID(ctx, domainsClientV1, domainID)
		if err == nil {
			return errors.New("domain still exists")
		}
	}

	return nil
}
