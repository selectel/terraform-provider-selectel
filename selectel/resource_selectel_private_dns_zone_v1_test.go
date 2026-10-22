package selectel

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	privatedns "github.com/selectel/private-dns-go/pkg/v1"
	"github.com/stretchr/testify/assert"
)

func TestAccCloudPrivateDNSZoneV1Basic(t *testing.T) {
	region := os.Getenv("INFRA_REGION")
	projectID := os.Getenv("INFRA_PROJECT_ID")

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheckWithProjectID(t) },
		ProviderFactories: testAccProvidersWithOpenStack,
		CheckDestroy:      testAccCloudPrivateDNSZoneV1Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCloudPrivateDNSZoneV1Basic(region, 1800, projectID),
				Check: resource.ComposeTestCheckFunc(
					testAccCloudPrivateDNSZoneV1Exists("selectel_private_dns_zone_v1.zone"),
					resource.TestCheckResourceAttr(
						"selectel_private_dns_zone_v1.zone", "domain", "example.com.",
					),
					resource.TestCheckResourceAttr(
						"selectel_private_dns_zone_v1.zone", "ttl", "1800",
					),
				),
			},
			{
				Config: testAccCloudPrivateDNSZoneV1Basic(region, 1900, projectID),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"selectel_private_dns_zone_v1.zone", "ttl", "1900",
					),
				),
			},
			{
				Config: testAccCloudPrivateDNSZoneV1WithRecords(region, 1900, projectID),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"selectel_private_dns_zone_v1.zone", "records.0.domain", "sub.example.com.",
					),
					resource.TestCheckResourceAttr(
						"selectel_private_dns_zone_v1.zone", "records.0.ttl", "10",
					),
					resource.TestCheckResourceAttr(
						"selectel_private_dns_zone_v1.zone", "records.0.type", "A",
					),
				),
			},
			{
				Config: testAccCloudPrivateDNSZoneV1Basic(region, 1900, projectID),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckNoResourceAttr(
						"selectel_private_dns_zone_v1.zone", "records.0",
					),
				),
			},
		},
	})
}

func TestAccCloudPrivateDNSZoneV1RecordsOrder(t *testing.T) {
	region := os.Getenv("INFRA_REGION")
	projectID := os.Getenv("INFRA_PROJECT_ID")

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheckWithProjectID(t) },
		ProviderFactories: testAccProvidersWithOpenStack,
		CheckDestroy:      testAccCloudPrivateDNSZoneV1Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCloudPrivateDNSZoneV1WithOrderedRecords(region, projectID, "z", "a", "m"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("selectel_private_dns_zone_v1.zone", "records.0.domain", "z.example.com."),
					resource.TestCheckResourceAttr("selectel_private_dns_zone_v1.zone", "records.1.domain", "a.example.com."),
					resource.TestCheckResourceAttr("selectel_private_dns_zone_v1.zone", "records.2.domain", "m.example.com."),
				),
			},
			{
				Config: testAccCloudPrivateDNSZoneV1WithOrderedRecords(region, projectID, "z", "k", "a", "m"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("selectel_private_dns_zone_v1.zone", "records.1.domain", "k.example.com."),
					resource.TestCheckResourceAttr("selectel_private_dns_zone_v1.zone", "records.#", "4"),
				),
			},
		},
	})
}

func testAccCloudPrivateDNSZoneV1WithOrderedRecords(region, projectID string, names ...string) string {
	var records strings.Builder
	for i, name := range names {
		records.WriteString(fmt.Sprintf(`
	records {
		domain = "%s.example.com."
		type = "A"
		ttl = 60
		values = ["192.168.0.%d"]
	}
`, name, i+1))
	}

	return fmt.Sprintf(`
resource "selectel_private_dns_zone_v1" "zone" {
	region = "%s"
	project_id = "%s"
	domain = "example.com."
	ttl = 1800
%s}
`, region, projectID, records.String())
}

func testAccCloudPrivateDNSZoneV1Basic(region string, ttl int, projectID string) string {
	return fmt.Sprintf(`
resource "selectel_private_dns_zone_v1" "zone" {
    region = "%s"
    project_id = "%s"
	domain = "example.com."
	ttl = %d
}
`, region, projectID, ttl)
}

func testAccCloudPrivateDNSZoneV1WithRecords(region string, ttl int, projectID string) string {
	return fmt.Sprintf(`
resource "selectel_private_dns_zone_v1" "zone" {
    region = "%s"
    project_id = "%s"
	domain = "example.com."
	ttl = %d
	records {
		domain = "sub.example.com."
		type = "A"
		ttl = 10
		values = [
			"192.168.0.1",
		]
	}
}
`, region, projectID, ttl)
}

func testAccCloudPrivateDNSZoneV1Destroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "selectel_private_dns_zone_v1" {
			continue
		}
		client, err := newTestPrivateDNSClient(rs, testAccProvider)
		if err != nil {
			return err
		}
		_, err = client.GetZone(context.Background(), rs.Primary.ID)
		if err == nil {
			return errors.New("zone still exists")
		}
	}

	return nil
}

func testAccCloudPrivateDNSZoneV1Exists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}
		if rs.Primary.ID == "" {
			return errors.New("no ID is set")
		}

		ctx := context.Background()

		client, err := newTestPrivateDNSClient(rs, testAccProvider)
		if err != nil {
			return err
		}

		_, err = client.GetZone(ctx, rs.Primary.ID)
		if err != nil {
			return err
		}

		return nil
	}
}

func TestOrderPrivateDNSRecordsLike(t *testing.T) {
	rec := func(recType, domain string) *privatedns.Record {
		return &privatedns.Record{Type: recType, Domain: domain, TTL: 60, Values: []string{"v"}}
	}
	prior := func(keys ...[2]string) []any {
		res := make([]any, 0, len(keys))
		for _, k := range keys {
			res = append(res, map[string]any{"type": k[0], "domain": k[1]})
		}

		return res
	}
	keysOf := func(records []*privatedns.Record) [][2]string {
		res := make([][2]string, 0, len(records))
		for _, r := range records {
			res = append(res, [2]string{r.Type, r.Domain})
		}

		return res
	}

	tests := []struct {
		name  string
		api   []*privatedns.Record
		prior []any
		want  [][2]string
	}{
		{
			name:  "api order differs from state order, state order wins",
			api:   []*privatedns.Record{rec("CNAME", "c."), rec("A", "a."), rec("CNAME", "b.")},
			prior: prior([2]string{"A", "a."}, [2]string{"CNAME", "b."}, [2]string{"CNAME", "c."}),
			want:  [][2]string{{"A", "a."}, {"CNAME", "b."}, {"CNAME", "c."}},
		},
		{
			name:  "same domain with different types is matched by type and domain",
			api:   []*privatedns.Record{rec("TXT", "a."), rec("A", "a.")},
			prior: prior([2]string{"A", "a."}, [2]string{"TXT", "a."}),
			want:  [][2]string{{"A", "a."}, {"TXT", "a."}},
		},
		{
			name:  "records missing from state go to the end sorted by type and domain",
			api:   []*privatedns.Record{rec("CNAME", "z."), rec("A", "b."), rec("A", "a."), rec("A", "x.")},
			prior: prior([2]string{"A", "x."}),
			want:  [][2]string{{"A", "x."}, {"A", "a."}, {"A", "b."}, {"CNAME", "z."}},
		},
		{
			name:  "records absent from api are dropped",
			api:   []*privatedns.Record{rec("A", "b.")},
			prior: prior([2]string{"A", "a."}, [2]string{"A", "b."}),
			want:  [][2]string{{"A", "b."}},
		},
		{
			name:  "empty state gives sorted order",
			api:   []*privatedns.Record{rec("CNAME", "a."), rec("A", "b."), rec("A", "a.")},
			prior: nil,
			want:  [][2]string{{"A", "a."}, {"A", "b."}, {"CNAME", "a."}},
		},
		{
			name:  "empty api gives empty result",
			api:   nil,
			prior: prior([2]string{"A", "a."}),
			want:  [][2]string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := orderPrivateDNSRecordsLike(tt.api, tt.prior)
			assert.Equal(t, tt.want, keysOf(got))
		})
	}
}

func TestOrderPrivateDNSRecordsLikeDoesNotMutateInput(t *testing.T) {
	api := []*privatedns.Record{
		{Type: "A", Domain: "b."},
		{Type: "A", Domain: "a."},
	}
	prior := []any{map[string]any{"type": "A", "domain": "a."}}

	_ = orderPrivateDNSRecordsLike(api, prior)

	assert.Equal(t, "b.", api[0].Domain)
	assert.Equal(t, "a.", api[1].Domain)
}

func TestFillPrivateDNSZoneV1DataKeepsConfiguredRecordsOrder(t *testing.T) {
	d := schema.TestResourceDataRaw(t, resourcePrivateDNSZoneV1().Schema, map[string]any{
		"domain": "example.com.",
		"records": []any{
			map[string]any{"domain": "z.example.com.", "type": "A", "ttl": 60, "values": []any{"10.0.0.1"}},
			map[string]any{"domain": "a.example.com.", "type": "A", "ttl": 60, "values": []any{"10.0.0.2"}},
			map[string]any{"domain": "m.example.com.", "type": "CNAME", "ttl": 60, "values": []any{"z.example.com."}},
		},
	})

	zone := &privatedns.ZoneDetails{
		Records: []*privatedns.Record{
			{Type: "CNAME", Domain: "m.example.com.", TTL: 60, Values: []string{"z.example.com."}},
			{Type: "A", Domain: "a.example.com.", TTL: 60, Values: []string{"10.0.0.2"}},
			{Type: "A", Domain: "z.example.com.", TTL: 60, Values: []string{"10.0.0.1"}},
		},
	}
	zone.Domain = "example.com."

	fillPrivateDNSZoneV1Data(zone, d)

	got := make([]string, 0, 3)
	for _, rec := range d.Get("records").([]any) {
		got = append(got, rec.(map[string]any)["domain"].(string))
	}
	assert.Equal(t, []string{"z.example.com.", "a.example.com.", "m.example.com."}, got)
}
