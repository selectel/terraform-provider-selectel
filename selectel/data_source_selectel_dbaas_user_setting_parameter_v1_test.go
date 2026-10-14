package selectel

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/selectel/dbaas-go"
	"github.com/selectel/go-selvpcclient/v5/selvpcclient/resell/v2/projects"
)

func TestAccDBaaSUserSettingParametersV1Basic(t *testing.T) {
	var (
		dbaasUserSettingParameters []dbaas.UserSettingParameter
		project                    projects.Project
	)

	projectName := acctest.RandomWithPrefix("tf-acc")
	parameterName := "statement_timeout"
	parameterNameWithChoices := "default_transaction_isolation"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckVPCV2ProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDBaaSUserSettingParametersV1Basic(projectName, parameterName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCV2ProjectExists("selectel_vpc_project_v2.project_tf_acc_test_1", &project),
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
				Config: testAccDBaaSUserSettingParametersV1Basic(projectName, parameterNameWithChoices),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCV2ProjectExists("selectel_vpc_project_v2.project_tf_acc_test_1", &project),
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

func testAccDBaaSUserSettingParametersV1Basic(projectName, name string) string {
	return fmt.Sprintf(`
resource "selectel_vpc_project_v2" "project_tf_acc_test_1" {
  name        = "%s"
}

data "selectel_dbaas_datastore_type_v1" "dt_tf_acc_test_1" {
  project_id = "${selectel_vpc_project_v2.project_tf_acc_test_1.id}"
  region     = "ru-3"
  filter {
    engine = "postgresql"
    version = "12"
  }
}

data "selectel_dbaas_user_setting_parameter_v1" "user_setting_param_tf_acc_test_1" {
  project_id = "${selectel_vpc_project_v2.project_tf_acc_test_1.id}"
  region     = "ru-3"
  filter {
    datastore_type_id = "${data.selectel_dbaas_datastore_type_v1.dt_tf_acc_test_1.datastore_types[0].id}"
    name = "%s"
  }
}

output "user_setting" {
  value = data.selectel_dbaas_user_setting_parameter_v1.user_setting_param_tf_acc_test_1.user_setting_parameters[0]
}
`, projectName, name)
}
