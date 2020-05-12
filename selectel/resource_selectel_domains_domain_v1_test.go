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
)

func TestAccDomainsDomainV1Basic(t *testing.T) {
	var testDomain domain.View

	testDomainName := fmt.Sprintf("%s.xyz", acctest.RandomWithPrefix("tf-acc"))

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccSelectelPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckDomainsV1DomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDomainsDomainV1Basic(testDomainName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainsDomainV1Exists("selectel_domains_domain_v1.domain_tf_acc_test_1",
						&testDomain),
					resource.TestCheckResourceAttr("selectel_domains_domain_v1.domain_tf_acc_test_1",
						"name", testDomainName),
				),
			},
		},
	})
}

func testAccDomainsDomainV1Basic(domainName string) string {
	return fmt.Sprintf(`
resource "selectel_domains_domain_v1" "domain_tf_acc_test_1" {
  name = "%s"
}`, domainName)
}

func testAccCheckDomainsDomainV1Exists(n string, selectelDomain *domain.View) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return errors.New("no ID is set")
		}

		config := testAccProvider.Meta().(*Config)
		ctx := context.Background()

		domainsClientV1 := config.domainsV1Client()

		domainID, err := strconv.Atoi(rs.Primary.ID)
		if err != nil {
			return errParseDomainsDomainV1ID(rs.Primary.ID)
		}

		foundDomain, _, err := domain.GetByID(ctx, domainsClientV1, domainID)
		if err != nil {
			return err
		}

		*selectelDomain = *foundDomain

		return nil
	}
}

func testAccCheckDomainsV1DomainDestroy(s *terraform.State) error {
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
