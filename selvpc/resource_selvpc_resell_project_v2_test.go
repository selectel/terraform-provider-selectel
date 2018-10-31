package selvpc

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"github.com/selectel/go-selvpcclient/selvpcclient/resell/v2/projects"
	"github.com/selectel/go-selvpcclient/selvpcclient/resell/v2/quotas"
	"github.com/stretchr/testify/assert"
)

func TestAccResellV2ProjectBasic(t *testing.T) {
	var project projects.Project
	projectName := acctest.RandomWithPrefix("tf-acc")
	projectNameUpdated := acctest.RandomWithPrefix("tf-acc-updated")
	projectCustomURL := acctest.RandomWithPrefix("tf-acc-url") + ".selvpc.ru"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccSelVPCPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckResellV2ProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccResellV2ProjectBasic(projectName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResellV2ProjectExists("selvpc_resell_project_v2.project_tf_acc_test_1", &project),
					resource.TestCheckResourceAttr("selvpc_resell_project_v2.project_tf_acc_test_1", "name", projectName),
				),
			},
			{
				Config: testAccResellV2ProjectUpdate1(projectName, projectCustomURL),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"selvpc_resell_project_v2.project_tf_acc_test_1", "name", projectName),
					resource.TestCheckResourceAttr(
						"selvpc_resell_project_v2.project_tf_acc_test_1", "custom_url", projectCustomURL),
					resource.TestCheckResourceAttr(
						"selvpc_resell_project_v2.project_tf_acc_test_1", "theme.color", "000000"),
					resource.TestCheckResourceAttr(
						"selvpc_resell_project_v2.project_tf_acc_test_1", "theme.logo", "fake.png"),
				),
			},
			{
				Config: testAccResellV2ProjectUpdate2(projectName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"selvpc_resell_project_v2.project_tf_acc_test_1", "name", projectName),
					resource.TestCheckResourceAttr(
						"selvpc_resell_project_v2.project_tf_acc_test_1", "custom_url", ""),
					resource.TestCheckResourceAttr(
						"selvpc_resell_project_v2.project_tf_acc_test_1", "theme.color", "FF0000"),
				),
			},
			{
				Config: testAccResellV2ProjectUpdate3(projectNameUpdated),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"selvpc_resell_project_v2.project_tf_acc_test_1", "name", projectNameUpdated),
					resource.TestCheckResourceAttr(
						"selvpc_resell_project_v2.project_tf_acc_test_1", "custom_url", ""),
					resource.TestCheckResourceAttr(
						"selvpc_resell_project_v2.project_tf_acc_test_1", "theme.color", "5D6D7E"),
					resource.TestCheckResourceAttr(
						"selvpc_resell_project_v2.project_tf_acc_test_1", "quotas.0.resource_name", "image_gigabytes"),
					resource.TestCheckResourceAttr(
						"selvpc_resell_project_v2.project_tf_acc_test_1", "quotas.0.resource_quotas.0.region", "ru-1"),
					resource.TestCheckResourceAttr(
						"selvpc_resell_project_v2.project_tf_acc_test_1", "quotas.0.resource_quotas.0.zone", ""),
					resource.TestCheckResourceAttr(
						"selvpc_resell_project_v2.project_tf_acc_test_1", "quotas.0.resource_quotas.0.value", "1"),
					resource.TestCheckResourceAttr(
						"selvpc_resell_project_v2.project_tf_acc_test_1", "quotas.1.resource_name", "volume_gigabytes_basic"),
					resource.TestCheckResourceAttr(
						"selvpc_resell_project_v2.project_tf_acc_test_1", "quotas.1.resource_quotas.0.region", "ru-1"),
					resource.TestCheckResourceAttr(
						"selvpc_resell_project_v2.project_tf_acc_test_1", "quotas.1.resource_quotas.0.zone", "ru-1a"),
					resource.TestCheckResourceAttr(
						"selvpc_resell_project_v2.project_tf_acc_test_1", "quotas.1.resource_quotas.0.value", "1"),
					resource.TestCheckResourceAttr(
						"selvpc_resell_project_v2.project_tf_acc_test_1", "quotas.1.resource_quotas.1.region", "ru-2"),
					resource.TestCheckResourceAttr(
						"selvpc_resell_project_v2.project_tf_acc_test_1", "quotas.1.resource_quotas.1.zone", "ru-2a"),
					resource.TestCheckResourceAttr(
						"selvpc_resell_project_v2.project_tf_acc_test_1", "quotas.1.resource_quotas.1.value", "2"),
				),
			},
		},
	})
}

func TestAccResellV2ProjectAutoQuotas(t *testing.T) {
	var project projects.Project
	projectName := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccSelVPCPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckResellV2ProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccResellV2ProjectAutoQuotas(projectName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResellV2ProjectExists("selvpc_resell_project_v2.project_tf_acc_test_1", &project),
					resource.TestCheckResourceAttr("selvpc_resell_project_v2.project_tf_acc_test_1", "name", projectName),
					resource.TestCheckResourceAttrSet("selvpc_resell_project_v2.project_tf_acc_test_1", "all_quotas.0.resource_quotas.0.region"),
					resource.TestCheckResourceAttrSet("selvpc_resell_project_v2.project_tf_acc_test_1", "all_quotas.0.resource_quotas.0.value"),
					resource.TestCheckResourceAttrSet("selvpc_resell_project_v2.project_tf_acc_test_1", "all_quotas.0.resource_quotas.0.used"),
					resource.TestCheckResourceAttrSet("selvpc_resell_project_v2.project_tf_acc_test_1", "all_quotas.1.resource_quotas.0.region"),
					resource.TestCheckResourceAttrSet("selvpc_resell_project_v2.project_tf_acc_test_1", "all_quotas.1.resource_quotas.0.value"),
					resource.TestCheckResourceAttrSet("selvpc_resell_project_v2.project_tf_acc_test_1", "all_quotas.1.resource_quotas.0.used"),
					resource.TestCheckResourceAttrSet("selvpc_resell_project_v2.project_tf_acc_test_1", "all_quotas.2.resource_quotas.0.region"),
					resource.TestCheckResourceAttrSet("selvpc_resell_project_v2.project_tf_acc_test_1", "all_quotas.2.resource_quotas.0.value"),
					resource.TestCheckResourceAttrSet("selvpc_resell_project_v2.project_tf_acc_test_1", "all_quotas.2.resource_quotas.0.used"),
					resource.TestCheckResourceAttrSet("selvpc_resell_project_v2.project_tf_acc_test_1", "all_quotas.3.resource_quotas.0.region"),
					resource.TestCheckResourceAttrSet("selvpc_resell_project_v2.project_tf_acc_test_1", "all_quotas.3.resource_quotas.0.value"),
					resource.TestCheckResourceAttrSet("selvpc_resell_project_v2.project_tf_acc_test_1", "all_quotas.3.resource_quotas.0.used"),
					resource.TestCheckResourceAttrSet("selvpc_resell_project_v2.project_tf_acc_test_1", "all_quotas.4.resource_quotas.0.region"),
					resource.TestCheckResourceAttrSet("selvpc_resell_project_v2.project_tf_acc_test_1", "all_quotas.4.resource_quotas.0.value"),
					resource.TestCheckResourceAttrSet("selvpc_resell_project_v2.project_tf_acc_test_1", "all_quotas.4.resource_quotas.0.used"),
					resource.TestCheckResourceAttrSet("selvpc_resell_project_v2.project_tf_acc_test_1", "all_quotas.5.resource_quotas.0.region"),
					resource.TestCheckResourceAttrSet("selvpc_resell_project_v2.project_tf_acc_test_1", "all_quotas.5.resource_quotas.0.value"),
					resource.TestCheckResourceAttrSet("selvpc_resell_project_v2.project_tf_acc_test_1", "all_quotas.5.resource_quotas.0.used"),
					resource.TestCheckResourceAttrSet("selvpc_resell_project_v2.project_tf_acc_test_1", "all_quotas.6.resource_quotas.0.region"),
					resource.TestCheckResourceAttrSet("selvpc_resell_project_v2.project_tf_acc_test_1", "all_quotas.6.resource_quotas.0.value"),
					resource.TestCheckResourceAttrSet("selvpc_resell_project_v2.project_tf_acc_test_1", "all_quotas.6.resource_quotas.0.used"),
					resource.TestCheckResourceAttrSet("selvpc_resell_project_v2.project_tf_acc_test_1", "all_quotas.7.resource_quotas.0.region"),
					resource.TestCheckResourceAttrSet("selvpc_resell_project_v2.project_tf_acc_test_1", "all_quotas.7.resource_quotas.0.value"),
					resource.TestCheckResourceAttrSet("selvpc_resell_project_v2.project_tf_acc_test_1", "all_quotas.7.resource_quotas.0.used"),
				),
			},
		},
	})
}

func testAccCheckResellV2ProjectDestroy(s *terraform.State) error {
	config := testAccProvider.Meta().(*Config)
	resellV2Client := config.resellV2Client()
	ctx := context.Background()

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "selvpc_resell_project_v2" {
			continue
		}

		_, _, err := projects.Get(ctx, resellV2Client, rs.Primary.ID)
		if err == nil {
			return errors.New("project still exists")
		}
	}

	return nil
}

func testAccCheckResellV2ProjectExists(n string, project *projects.Project) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return errors.New("no ID is set")
		}

		config := testAccProvider.Meta().(*Config)
		resellV2Client := config.resellV2Client()
		ctx := context.Background()

		foundProject, _, err := projects.Get(ctx, resellV2Client, rs.Primary.ID)
		if err != nil {
			return err
		}

		if foundProject.ID != rs.Primary.ID {
			return errors.New("project not found")
		}

		*project = *foundProject

		return nil
	}
}

func testAccResellV2ProjectBasic(name string) string {
	return fmt.Sprintf(`
resource "selvpc_resell_project_v2" "project_tf_acc_test_1" {
  name = "%s"
}`, name)
}

func testAccResellV2ProjectUpdate1(name, customURL string) string {
	return fmt.Sprintf(`
resource "selvpc_resell_project_v2" "project_tf_acc_test_1" {
  name       = "%s"
  custom_url = "%s"
  theme {
    color = "000000"
    logo  = "fake.png"
  }
}`, name, customURL)
}

func testAccResellV2ProjectUpdate2(name string) string {
	return fmt.Sprintf(`
resource "selvpc_resell_project_v2" "project_tf_acc_test_1" {
  name       = "%s"
  theme {
    color = "FF0000"
  }
}`, name)
}

func testAccResellV2ProjectUpdate3(name string) string {
	return fmt.Sprintf(`
resource "selvpc_resell_project_v2" "project_tf_acc_test_1" {
  name = "%s"
  theme {
    color = "5D6D7E"
  }
  quotas = [
    {
      resource_name = "image_gigabytes"
      resource_quotas = [
        {
          region = "ru-1"
          value = 1
        }
      ]
    },
    {
      resource_name = "volume_gigabytes_basic"
      resource_quotas = [
        {
          region = "ru-1"
          zone = "ru-1a"
          value = 1
        },
        {
          region = "ru-2"
          zone = "ru-2a"
          value = 2
        }
      ]
    }
  ]
}`, name)
}

func testAccResellV2ProjectAutoQuotas(name string) string {
	return fmt.Sprintf(`
resource "selvpc_resell_project_v2" "project_tf_acc_test_1" {
  name        = "%s"
  auto_quotas = true
}`, name)
}

func TestResourceResellProjectV2QuotasOptsFromList(t *testing.T) {
	quotasList := []interface{}{
		map[string]interface{}{
			"resource_name": "volume_gigabytes_fast",
			"resource_quotas": []interface{}{
				map[string]interface{}{
					"region": "ru-3",
					"zone":   "ru-3a",
					"value":  100,
				},
			},
		},
	}
	expectedResourceQuotaValue := 100
	expectedQuotasOpts := []quotas.QuotaOpts{
		{
			Name: "volume_gigabytes_fast",
			ResourceQuotasOpts: []quotas.ResourceQuotaOpts{
				{
					Region: "ru-3",
					Zone:   "ru-3a",
					Value:  &expectedResourceQuotaValue,
				},
			},
		},
	}

	actualQuotaOpts, err := resourceResellProjectV2QuotasOptsFromList(quotasList)

	assert.Empty(t, err)
	assert.Equal(t, expectedQuotasOpts, actualQuotaOpts)
}

func TestResourceResellProjectV2QuotasOptsFromListNoName(t *testing.T) {
	quotasList := []interface{}{
		map[string]interface{}{
			"resource_quotas": []interface{}{
				map[string]interface{}{
					"region": "ru-3",
					"zone":   "ru-3a",
					"value":  100,
				},
			},
		},
	}

	quotaOpts, err := resourceResellProjectV2QuotasOptsFromList(quotasList)

	assert.Empty(t, quotaOpts)
	assert.EqualError(t, err, "resource_name value isn't provided")
}

func TestResourceResellProjectV2QuotasOptsFromListNoQuotas(t *testing.T) {
	quotasList := []interface{}{
		map[string]interface{}{
			"resource_name": "volume_gigabytes_fast",
		},
	}

	quotaOpts, err := resourceResellProjectV2QuotasOptsFromList(quotasList)

	assert.Empty(t, quotaOpts)
	assert.EqualError(t, err, "resource_quotas value isn't provided")
}

func TestResourceResellProjectV2QuotasToMap(t *testing.T) {
	quotasOpts := []quotas.Quota{
		{
			Name: "compute_cores",
			ResourceQuotasEntities: []quotas.ResourceQuotaEntity{
				{
					Region: "ru-2",
					Value:  12,
					Used:   2,
				},
			},
		},
	}
	expectedQuotasMap := map[string]interface{}{
		"compute_cores": []map[string]interface{}{
			{
				"region": "ru-2",
				"zone":   "",
				"value":  12,
				"used":   2,
			},
		},
	}

	actualQuotasMap := resourceResellProjectV2QuotasToMap(quotasOpts)

	assert.Equal(t, expectedQuotasMap, actualQuotasMap)
}

func TestResourceProjectV2UpdateThemeOptsFromMap(t *testing.T) {
	themeOptsMap := map[string]interface{}{
		"color": "FF0000",
		"logo":  "fake.png",
	}
	expectedColor := "FF0000"
	expectedLogo := "fake.png"
	expectedThemeUpdateOpts := &projects.ThemeUpdateOpts{
		Color: &expectedColor,
		Logo:  &expectedLogo,
	}

	actualThemeUpdateOpts := resourceProjectV2UpdateThemeOptsFromMap(themeOptsMap)

	assert.Equal(t, expectedThemeUpdateOpts, actualThemeUpdateOpts)
}

func TestResourceResellProjectV2URLWithoutSchema(t *testing.T) {
	customURL := "https://my-url.selvpc.ru"
	expectedURL := "my-url.selvpc.ru"

	actualURL, err := resourceResellProjectV2URLWithoutSchema(customURL)

	assert.Empty(t, err)
	assert.Equal(t, expectedURL, actualURL)
}
