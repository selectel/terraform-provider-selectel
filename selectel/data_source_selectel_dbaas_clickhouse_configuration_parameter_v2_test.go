package selectel

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	dbaas_v2_ch "github.com/selectel/dbaas-go/v2/clickhouse"
	"github.com/selectel/go-selvpcclient/v4/selvpcclient/resell/v2/projects"
)

func TestAccDBaaSV2ClickhouseConfigurationParametersBasic(t *testing.T) {
	var (
		dbaasConfigurationParameters []dbaas_v2_ch.DatastoreConfigurationParameterResponse
		project                      projects.Project
	)

	projectName := acctest.RandomWithPrefix("tf-acc")
	datastoreTypeEngine := "clickhouse"
	datastoreTypeVersion := "26.3.12.3"
	parameterName := "server_settings.async_insert_threads"
	parameterNameWithChoices := "merge_tree_settings.deduplicate_merge_projection_mode"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckVPCV2ProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDBaaSV2ClickhouseConfigurationParametersBasic(projectName, datastoreTypeEngine, datastoreTypeVersion, parameterName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCV2ProjectExists("selectel_vpc_project_v2.project_tf_acc_test_1", &project),
					testAccDBaaSV2ClickhouseConfigurationParametersExists("data.selectel_dbaas_clickhouse_configuration_parameter_v2.configuration_param_tf_acc_test_1", &dbaasConfigurationParameters),
					resource.TestCheckResourceAttr("data.selectel_dbaas_clickhouse_configuration_parameter_v2.configuration_param_tf_acc_test_1", "configuration_parameters.0.name", parameterName),
					resource.TestCheckResourceAttr("data.selectel_dbaas_clickhouse_configuration_parameter_v2.configuration_param_tf_acc_test_1", "configuration_parameters.0.type", "int"),
					resource.TestCheckResourceAttr("data.selectel_dbaas_clickhouse_configuration_parameter_v2.configuration_param_tf_acc_test_1", "configuration_parameters.0.min", "0"),
					resource.TestCheckResourceAttr("data.selectel_dbaas_clickhouse_configuration_parameter_v2.configuration_param_tf_acc_test_1", "configuration_parameters.0.max", "18446744073709551615"),
					resource.TestCheckResourceAttr("data.selectel_dbaas_clickhouse_configuration_parameter_v2.configuration_param_tf_acc_test_1", "configuration_parameters.0.default_value", "16"),
					resource.TestCheckResourceAttr("data.selectel_dbaas_clickhouse_configuration_parameter_v2.configuration_param_tf_acc_test_1", "configuration_parameters.0.choices.#", "0"),
					resource.TestCheckResourceAttr("data.selectel_dbaas_clickhouse_configuration_parameter_v2.configuration_param_tf_acc_test_1", "configuration_parameters.0.invalid_values.#", "0"),
					resource.TestCheckResourceAttr("data.selectel_dbaas_clickhouse_configuration_parameter_v2.configuration_param_tf_acc_test_1", "configuration_parameters.0.is_restart_required", "true"),
					resource.TestCheckResourceAttr("data.selectel_dbaas_clickhouse_configuration_parameter_v2.configuration_param_tf_acc_test_1", "configuration_parameters.0.is_changeable", "true"),
					resource.TestCheckResourceAttr("data.selectel_dbaas_clickhouse_configuration_parameter_v2.configuration_param_tf_acc_test_1", "configuration_parameters.0.is_multiple_choice_available", "false"),
					resource.TestCheckResourceAttr("data.selectel_dbaas_clickhouse_configuration_parameter_v2.configuration_param_tf_acc_test_1", "configuration_parameters.0.can_be_empty", "false"),
				),
			},
			{
				Config: testAccDBaaSV2ClickhouseConfigurationParametersBasic(projectName, datastoreTypeEngine, datastoreTypeVersion, parameterNameWithChoices),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCV2ProjectExists("selectel_vpc_project_v2.project_tf_acc_test_1", &project),
					testAccDBaaSV2ClickhouseConfigurationParametersExists("data.selectel_dbaas_clickhouse_configuration_parameter_v2.configuration_param_tf_acc_test_1", &dbaasConfigurationParameters),
					resource.TestCheckResourceAttr("data.selectel_dbaas_clickhouse_configuration_parameter_v2.configuration_param_tf_acc_test_1", "configuration_parameters.0.name", parameterNameWithChoices),
					resource.TestCheckResourceAttr("data.selectel_dbaas_clickhouse_configuration_parameter_v2.configuration_param_tf_acc_test_1", "configuration_parameters.0.type", "str"),
					resource.TestCheckResourceAttr("data.selectel_dbaas_clickhouse_configuration_parameter_v2.configuration_param_tf_acc_test_1", "configuration_parameters.0.min", ""),
					resource.TestCheckResourceAttr("data.selectel_dbaas_clickhouse_configuration_parameter_v2.configuration_param_tf_acc_test_1", "configuration_parameters.0.max", ""),
					resource.TestCheckResourceAttr("data.selectel_dbaas_clickhouse_configuration_parameter_v2.configuration_param_tf_acc_test_1", "configuration_parameters.0.default_value", "throw"),
					resource.TestCheckResourceAttr("data.selectel_dbaas_clickhouse_configuration_parameter_v2.configuration_param_tf_acc_test_1", "configuration_parameters.0.choices.#", "4"),
					resource.TestCheckResourceAttr("data.selectel_dbaas_clickhouse_configuration_parameter_v2.configuration_param_tf_acc_test_1", "configuration_parameters.0.choices.0", "ignore"),
					resource.TestCheckResourceAttr("data.selectel_dbaas_clickhouse_configuration_parameter_v2.configuration_param_tf_acc_test_1", "configuration_parameters.0.choices.1", "throw"),
					resource.TestCheckResourceAttr("data.selectel_dbaas_clickhouse_configuration_parameter_v2.configuration_param_tf_acc_test_1", "configuration_parameters.0.choices.2", "drop"),
					resource.TestCheckResourceAttr("data.selectel_dbaas_clickhouse_configuration_parameter_v2.configuration_param_tf_acc_test_1", "configuration_parameters.0.choices.3", "rebuild"),
					resource.TestCheckResourceAttr("data.selectel_dbaas_clickhouse_configuration_parameter_v2.configuration_param_tf_acc_test_1", "configuration_parameters.0.invalid_values.#", "0"),
					resource.TestCheckResourceAttr("data.selectel_dbaas_clickhouse_configuration_parameter_v2.configuration_param_tf_acc_test_1", "configuration_parameters.0.is_restart_required", "true"),
					resource.TestCheckResourceAttr("data.selectel_dbaas_clickhouse_configuration_parameter_v2.configuration_param_tf_acc_test_1", "configuration_parameters.0.is_changeable", "true"),
					resource.TestCheckResourceAttr("data.selectel_dbaas_clickhouse_configuration_parameter_v2.configuration_param_tf_acc_test_1", "configuration_parameters.0.is_multiple_choice_available", "false"),
					resource.TestCheckResourceAttr("data.selectel_dbaas_clickhouse_configuration_parameter_v2.configuration_param_tf_acc_test_1", "configuration_parameters.0.can_be_empty", "false"),
				),
			},
		},
	})
}

func testAccDBaaSV2ClickhouseConfigurationParametersExists(n string, dbaasConfigurationParameters *[]dbaas_v2_ch.DatastoreConfigurationParameterResponse) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		ctx := context.Background()

		dbaasClient, err := newTestDBaaSV2Client(ctx, rs, testAccProvider)
		if err != nil {
			return err
		}

		configurationParameters, err := dbaasClient.ClickHouse.GetDatastoreConfigurationParameters(ctx)
		if err != nil {
			return err
		}

		*dbaasConfigurationParameters = configurationParameters

		return nil
	}
}

func testAccDBaaSV2ClickhouseConfigurationParametersBasic(projectName, engine, version, name string) string {
	return fmt.Sprintf(`
resource "selectel_vpc_project_v2" "project_tf_acc_test_1" {
  name        = "%s"
}

data "selectel_dbaas_datastore_type_v2" "dt" {
  project_id = "${selectel_vpc_project_v2.project_tf_acc_test_1.id}"
  region     = "ru-1"
  filter {
    engine = "%s"
    version = "%s"
  }
}

data "selectel_dbaas_clickhouse_configuration_parameter_v2" "configuration_param_tf_acc_test_1" {
  project_id = "${selectel_vpc_project_v2.project_tf_acc_test_1.id}"
  region     = "ru-1"
  filter {
    datastore_type_id = "${data.selectel_dbaas_datastore_type_v2.dt.datastore_types[0].id}"
    name = "%s"
  }
}

output "config" {
  value = data.selectel_dbaas_clickhouse_configuration_parameter_v2.configuration_param_tf_acc_test_1.configuration_parameters[0]
}
`, projectName, engine, version, name)
}
