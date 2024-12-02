package selectel

import (
	"errors"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/selectel/go-selvpcclient/v4/selvpcclient/quotamanager/quotas"
	"github.com/selectel/go-selvpcclient/v4/selvpcclient/resell/v2/projects"
	"github.com/stretchr/testify/assert"
)

func TestAccVPCV2ProjectBasic(t *testing.T) {
	var project projects.Project
	projectName := acctest.RandomWithPrefix("tf-acc")
	projectNameUpdated := acctest.RandomWithPrefix("tf-acc-updated")
	projectCustomURL := acctest.RandomWithPrefix("tf-acc-url") + ".selvpc.ru"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckVPCV2ProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCV2ProjectBasic(projectName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCV2ProjectExists("selectel_vpc_project_v2.project_tf_acc_test_1", &project),
					resource.TestCheckResourceAttr("selectel_vpc_project_v2.project_tf_acc_test_1", "name", projectName),
				),
			},
			{
				Config: testAccVPCV2ProjectUpdate1(projectName, projectCustomURL),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"selectel_vpc_project_v2.project_tf_acc_test_1", "name", projectName),
					resource.TestCheckResourceAttr(
						"selectel_vpc_project_v2.project_tf_acc_test_1", "custom_url", projectCustomURL),
					resource.TestCheckResourceAttr(
						"selectel_vpc_project_v2.project_tf_acc_test_1", "theme.color", "000000"),
					resource.TestCheckResourceAttr(
						"selectel_vpc_project_v2.project_tf_acc_test_1", "theme.logo", "fake.png"),
				),
			},
			{
				Config: testAccVPCV2ProjectUpdate2(projectName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"selectel_vpc_project_v2.project_tf_acc_test_1", "name", projectName),
					resource.TestCheckResourceAttr(
						"selectel_vpc_project_v2.project_tf_acc_test_1", "custom_url", ""),
					resource.TestCheckResourceAttr(
						"selectel_vpc_project_v2.project_tf_acc_test_1", "theme.color", "FF0000"),
				),
			},
			{
				Config: testAccVPCV2ProjectUpdate3(projectNameUpdated),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"selectel_vpc_project_v2.project_tf_acc_test_1", "name", projectNameUpdated),
					resource.TestCheckResourceAttr(
						"selectel_vpc_project_v2.project_tf_acc_test_1", "custom_url", ""),
					resource.TestCheckResourceAttr(
						"selectel_vpc_project_v2.project_tf_acc_test_1", "theme.color", "5D6D7E"),
					resource.TestCheckResourceAttr(
						"selectel_vpc_project_v2.project_tf_acc_test_1", "quotas.#", "2"),
				),
			},
		},
	})
}

func TestAccVPCV2ProjectWithSpecificQuotas(t *testing.T) {
	projectName := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckVPCV2ProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCV2ProjectWithSpecificQuotas(projectName),
				Check: resource.TestCheckResourceAttr(
					"selectel_vpc_project_v2.project_tf_acc_test_2", "quotas.#", "2"),
			},
		},
	})
}

func testAccCheckVPCV2ProjectDestroy(s *terraform.State) error {
	config := testAccProvider.Meta().(*Config)
	selvpcClient, err := config.GetSelVPCClient()
	if err != nil {
		return fmt.Errorf("can't get selvpc client for test project object: %w", err)
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "selectel_vpc_project_v2" {
			continue
		}

		_, _, err := projects.Get(selvpcClient, rs.Primary.ID)
		if err == nil {
			return errors.New("project still exists")
		}
	}

	return nil
}

func testAccCheckVPCV2ProjectExists(n string, project *projects.Project) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return errors.New("no ID is set")
		}

		config := testAccProvider.Meta().(*Config)
		selvpcClient, err := config.GetSelVPCClient()
		if err != nil {
			return fmt.Errorf("can't get selvpc client for test project object: %w", err)
		}

		foundProject, _, err := projects.Get(selvpcClient, rs.Primary.ID)
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

func testAccVPCV2ProjectBasic(name string) string {
	return fmt.Sprintf(`
resource "selectel_vpc_project_v2" "project_tf_acc_test_1" {
  name = "%s"
}`, name)
}

func testAccVPCV2ProjectWithSpecificQuotas(name string) string {
	return fmt.Sprintf(`
resource "selectel_vpc_project_v2" "project_tf_acc_test_2" {
  name = "%s"
  quotas {
    resource_name = "compute_cores"
    resource_quotas {
      region = "ru-1"
      zone = "ru-1b"
      value = 4
    }
    resource_quotas {
      region = "ru-2"
      zone = "ru-2b"
      value = 6
    }
  }
  quotas {
    resource_name = "volume_gigabytes_basic"
    resource_quotas {
      region = "ru-2"
      zone = "ru-2a"
      value = 2
    }
  }
}`, name)
}

func testAccVPCV2ProjectUpdate1(name, customURL string) string {
	return fmt.Sprintf(`
resource "selectel_vpc_project_v2" "project_tf_acc_test_1" {
  name       = "%s"
  custom_url = "%s"
  theme = {
    color = "000000"
    logo  = "fake.png"
  }
}`, name, customURL)
}

func testAccVPCV2ProjectUpdate2(name string) string {
	return fmt.Sprintf(`
resource "selectel_vpc_project_v2" "project_tf_acc_test_1" {
  name       = "%s"
  theme = {
    color = "FF0000"
  }
}`, name)
}

func testAccVPCV2ProjectUpdate3(name string) string {
	return fmt.Sprintf(`
resource "selectel_vpc_project_v2" "project_tf_acc_test_1" {
  name = "%s"
  theme = {
    color = "5D6D7E"
  }
  quotas {
    resource_name = "image_gigabytes"
    resource_quotas {
      region = "ru-1"
      value = 1
    }
  }
  quotas {
    resource_name = "volume_gigabytes_basic"
    resource_quotas {
      region = "ru-1"
      zone = "ru-1a"
      value = 1
    }
    resource_quotas {
      region = "ru-2"
      zone = "ru-2a"
      value = 2
    }
  }
}`, name)
}

func TestResourceVPCProjectV2QuotasOptsFromSet(t *testing.T) {
	region := "ru-3"
	zone := "ru-3a"

	quotaSet := &schema.Set{
		F: quotasHashSetFunc(),
	}
	resourceQuotas := &schema.Set{
		F: resourceQuotasHashSetFunc(),
	}
	resourceQuotas.Add(map[string]interface{}{
		"region": region,
		"zone":   zone,
		"value":  100,
	})
	quotaSet.Add(map[string]interface{}{
		"resource_name":   "volume_gigabytes_fast",
		"resource_quotas": resourceQuotas,
	})

	expectedResourceQuotaValue := 100
	expectedQuotasOpts := map[string]quotas.UpdateProjectQuotasOpts{
		region: {
			QuotasOpts: []quotas.QuotaOpts{
				{
					Name: "volume_gigabytes_fast",
					ResourceQuotasOpts: []quotas.ResourceQuotaOpts{
						{
							Zone:  &zone,
							Value: &expectedResourceQuotaValue,
						},
					},
				},
			},
		},
	}

	actualQuotaOpts, err := resourceVPCProjectV2QuotasOptsFromSet(quotaSet)

	assert.Empty(t, err)
	assert.Equal(t, expectedQuotasOpts, actualQuotaOpts)
}

func TestResourceVPCProjectV2QuotasOptsFromListNoName(t *testing.T) {
	quotaSet := &schema.Set{
		F: quotasHashSetFunc(),
	}
	resourceQuotas := &schema.Set{
		F: resourceQuotasHashSetFunc(),
	}
	resourceQuotas.Add(map[string]interface{}{
		"region": "ru-3",
		"zone":   "ru-3a",
		"value":  100,
	})
	quotaSet.Add(map[string]interface{}{
		"resource_quotas": resourceQuotas,
	})

	quotaOpts, err := resourceVPCProjectV2QuotasOptsFromSet(quotaSet)

	assert.Empty(t, quotaOpts)
	assert.EqualError(t, err, "resource_name value isn't provided")
}

func TestResourceVPCProjectV2QuotasOptsFromListNoQuotas(t *testing.T) {
	quotaSet := schema.NewSet(
		schema.HashResource(resourceVPCProjectV2().Schema["quotas"].Elem.(*schema.Resource)),
		[]interface{}{
			map[string]interface{}{
				"resource_name": "volume_gigabytes_fast",
			},
		})

	quotaOpts, err := resourceVPCProjectV2QuotasOptsFromSet(quotaSet)

	assert.Empty(t, quotaOpts)
	assert.EqualError(t, err, "resource_quotas value isn't provided")
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

func TestResourceVPCProjectV2URLWithoutSchema(t *testing.T) {
	customURL := "https://my-url.selvpc.ru"
	expectedURL := "my-url.selvpc.ru"

	actualURL, err := resourceVPCProjectV2URLWithoutSchema(customURL)

	assert.Empty(t, err)
	assert.Equal(t, expectedURL, actualURL)
}
