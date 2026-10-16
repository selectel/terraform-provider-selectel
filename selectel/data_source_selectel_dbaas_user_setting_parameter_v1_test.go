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

func TestAccDBaaSUserSettingParametersV1Basic(t *testing.T) {
	var dbaasUserSettingParameters []dbaas.UserSettingParameter

	projectID := os.Getenv("INFRA_PROJECT_ID")
	parameterName := "statement_timeout"
	parameterNameWithChoices := "default_transaction_isolation"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheckWithProjectID(t) },
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDBaaSUserSettingParametersV1Basic(projectID, parameterName),
				Check: resource.ComposeTestCheckFunc(
					testAccDBaaSUserSettingParametersV1Exists("data.selectel_dbaas_user_setting_parameter_v1.user_setting_param_tf_acc_test_1", &dbaasUserSettingParameters),
					resource.TestCheckResourceAttr("data.selectel_dbaas_user_setting_parameter_v1.user_setting_param_tf_acc_test_1", "user_setting_parameters.0.name", parameterName),
					resource.TestCheckResourceAttr("data.selectel_dbaas_user_setting_parameter_v1.user_setting_param_tf_acc_test_1", "user_setting_parameters.0.type", "int"),
					resource.TestCheckResourceAttr("data.selectel_dbaas_user_setting_parameter_v1.user_setting_param_tf_acc_test_1", "user_setting_parameters.0.unit", "ms"),
					resource.TestCheckResourceAttr("data.selectel_dbaas_user_setting_parameter_v1.user_setting_param_tf_acc_test_1", "user_setting_parameters.0.min", "0"),
					resource.TestCheckResourceAttr("data.selectel_dbaas_user_setting_parameter_v1.user_setting_param_tf_acc_test_1", "user_setting_parameters.0.max", "2147483647"),
					resource.TestCheckResourceAttr("data.selectel_dbaas_user_setting_parameter_v1.user_setting_param_tf_acc_test_1", "user_setting_parameters.0.choices.#", "0"),
					resource.TestCheckResourceAttr("data.selectel_dbaas_user_setting_parameter_v1.user_setting_param_tf_acc_test_1", "user_setting_parameters.0.apply_mechanism", "guc"),
					resource.TestCheckResourceAttr("data.selectel_dbaas_user_setting_parameter_v1.user_setting_param_tf_acc_test_1", "user_setting_parameters.0.is_available_for_customer", "true"),
					resource.TestCheckResourceAttr("data.selectel_dbaas_user_setting_parameter_v1.user_setting_param_tf_acc_test_1", "user_setting_parameters.0.is_changeable", "true"),
					resource.TestCheckResourceAttr("data.selectel_dbaas_user_setting_parameter_v1.user_setting_param_tf_acc_test_1", "user_setting_parameters.0.default_value", "0"),
					resource.TestCheckResourceAttr("data.selectel_dbaas_user_setting_parameter_v1.user_setting_param_tf_acc_test_1", "user_setting_parameters.0.can_be_empty", "true"),
				),
			},
			{
				Config: testAccDBaaSUserSettingParametersV1Basic(projectID, parameterNameWithChoices),
				Check: resource.ComposeTestCheckFunc(
					testAccDBaaSUserSettingParametersV1Exists("data.selectel_dbaas_user_setting_parameter_v1.user_setting_param_tf_acc_test_1", &dbaasUserSettingParameters),
					resource.TestCheckResourceAttr("data.selectel_dbaas_user_setting_parameter_v1.user_setting_param_tf_acc_test_1", "user_setting_parameters.0.name", parameterNameWithChoices),
					resource.TestCheckResourceAttr("data.selectel_dbaas_user_setting_parameter_v1.user_setting_param_tf_acc_test_1", "user_setting_parameters.0.type", "str"),
					resource.TestCheckResourceAttr("data.selectel_dbaas_user_setting_parameter_v1.user_setting_param_tf_acc_test_1", "user_setting_parameters.0.unit", ""),
					resource.TestCheckResourceAttr("data.selectel_dbaas_user_setting_parameter_v1.user_setting_param_tf_acc_test_1", "user_setting_parameters.0.min", ""),
					resource.TestCheckResourceAttr("data.selectel_dbaas_user_setting_parameter_v1.user_setting_param_tf_acc_test_1", "user_setting_parameters.0.max", ""),
					resource.TestCheckResourceAttr("data.selectel_dbaas_user_setting_parameter_v1.user_setting_param_tf_acc_test_1", "user_setting_parameters.0.choices.#", "4"),
					resource.TestCheckResourceAttr("data.selectel_dbaas_user_setting_parameter_v1.user_setting_param_tf_acc_test_1", "user_setting_parameters.0.choices.0", "serializable"),
					resource.TestCheckResourceAttr("data.selectel_dbaas_user_setting_parameter_v1.user_setting_param_tf_acc_test_1", "user_setting_parameters.0.choices.1", "repeatable read"),
					resource.TestCheckResourceAttr("data.selectel_dbaas_user_setting_parameter_v1.user_setting_param_tf_acc_test_1", "user_setting_parameters.0.choices.2", "read committed"),
					resource.TestCheckResourceAttr("data.selectel_dbaas_user_setting_parameter_v1.user_setting_param_tf_acc_test_1", "user_setting_parameters.0.choices.3", "read uncommitted"),
					resource.TestCheckResourceAttr("data.selectel_dbaas_user_setting_parameter_v1.user_setting_param_tf_acc_test_1", "user_setting_parameters.0.apply_mechanism", "guc"),
					resource.TestCheckResourceAttr("data.selectel_dbaas_user_setting_parameter_v1.user_setting_param_tf_acc_test_1", "user_setting_parameters.0.is_changeable", "true"),
					resource.TestCheckResourceAttr("data.selectel_dbaas_user_setting_parameter_v1.user_setting_param_tf_acc_test_1", "user_setting_parameters.0.default_value", "read committed"),
					resource.TestCheckResourceAttr("data.selectel_dbaas_user_setting_parameter_v1.user_setting_param_tf_acc_test_1", "user_setting_parameters.0.can_be_empty", "true"),
				),
			},
		},
	})
}

func testAccDBaaSUserSettingParametersV1Exists(n string, dbaasUserSettingParameters *[]dbaas.UserSettingParameter) resource.TestCheckFunc {
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

		userSettingParameters, err := dbaasClient.UserSettingParameters(ctx, nil)
		if err != nil {
			return err
		}

		*dbaasUserSettingParameters = userSettingParameters

		return nil
	}
}

func testAccDBaaSUserSettingParametersV1Basic(projectID, name string) string {
	return fmt.Sprintf(`
locals {
  project_id = "%s"
}

data "selectel_dbaas_datastore_type_v1" "dt_tf_acc_test_1" {
  project_id = local.project_id
  region     = "ru-3"
  filter {
    engine = "postgresql"
    version = "16"
  }
}

data "selectel_dbaas_user_setting_parameter_v1" "user_setting_param_tf_acc_test_1" {
  project_id = local.project_id
  region     = "ru-3"
  filter {
    datastore_type_id = "${data.selectel_dbaas_datastore_type_v1.dt_tf_acc_test_1.datastore_types[0].id}"
    name = "%s"
  }
}

output "user_setting" {
  value = data.selectel_dbaas_user_setting_parameter_v1.user_setting_param_tf_acc_test_1.user_setting_parameters[0]
}
`, projectID, name)
}
