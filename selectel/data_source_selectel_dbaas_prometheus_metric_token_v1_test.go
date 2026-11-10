package selectel

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/selectel/dbaas-go"
)

func TestAccDBaaSDataSourcePrometheusMetricTokenV1Basic(t *testing.T) {
	var dbaasTokens []dbaas.PrometheusMetricToken

	projectID := os.Getenv("INFRA_PROJECT_ID")

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheckWithProjectID(t) },
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDBaaSDataSourcePrometheusMetricTokenV1Basic(projectID),
				Check: resource.ComposeTestCheckFunc(
					testAccDBaaSPrometheusMetricTokensV1Exists("data.selectel_dbaas_prometheus_metric_token_v1.prometheus_metric_token_tf_acc_test_1", &dbaasTokens),
				),
			},
		},
	})
}

func testAccDBaaSPrometheusMetricTokensV1Exists(n string, dbaasTokens *[]dbaas.PrometheusMetricToken) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		ctx := context.Background()

		dbaasClient, err := newTestDBaaSClient(ctx, rs, testAccProvider)
		if err != nil {
			return err
		}

		tokens, err := dbaasClient.PrometheusMetricTokens(ctx)
		if err != nil {
			return err
		}

		*dbaasTokens = tokens

		return nil
	}
}

func testAccDBaaSDataSourcePrometheusMetricTokenV1Basic(projectID string) string {
	return fmt.Sprintf(`
locals {
  project_id = "%s"
}

data "selectel_dbaas_prometheus_metric_token_v1" "prometheus_metric_token_tf_acc_test_1" {
  project_id = local.project_id
  region     = "ru-3"
}
`, projectID)
}
