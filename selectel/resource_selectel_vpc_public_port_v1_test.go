package selectel

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	publicnetapi "github.com/selectel/public-net-api-go/pkg/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAccVPCPublicPortV1Basic(t *testing.T) {
	region := os.Getenv("INFRA_REGION")
	projectID := os.Getenv("INFRA_PROJECT_ID")

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheckWithProjectID(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccVPCPublicPortV1CheckDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCPublicPortV1(region, projectID),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("selectel_vpc_public_port_v1.port", "id"),
					resource.TestCheckResourceAttr("selectel_vpc_public_port_v1.port", "project_id", projectID),
					resource.TestCheckResourceAttrSet("selectel_vpc_public_port_v1.port", "network_id"),
					resource.TestCheckResourceAttrSet("selectel_vpc_public_port_v1.port", "ip_address"),
					resource.TestCheckResourceAttr("selectel_vpc_public_port_v1.port", "description", ""),
					resource.TestCheckResourceAttrSet("selectel_vpc_public_port_v1.port", "subnet"),
					resource.TestCheckResourceAttrSet("selectel_vpc_public_port_v1.port", "gateway"),
					resource.TestCheckResourceAttr("selectel_vpc_public_port_v1.port", "admin_state_up", "true"),
					resource.TestCheckResourceAttr("selectel_vpc_public_port_v1.port", "security_group_ids.#", "1"),
					testAccVPCPublicPortV1CheckExist("selectel_vpc_public_port_v1.port"),
				),
			},
			{
				Config: testAccVPCPublicPortV1WithAdminStateUp(region, projectID, true),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("selectel_vpc_public_port_v1.port", "id"),
					resource.TestCheckResourceAttr("selectel_vpc_public_port_v1.port", "project_id", projectID),
					resource.TestCheckResourceAttrSet("selectel_vpc_public_port_v1.port", "network_id"),
					resource.TestCheckResourceAttrSet("selectel_vpc_public_port_v1.port", "ip_address"),
					resource.TestCheckResourceAttr("selectel_vpc_public_port_v1.port", "description", ""),
					resource.TestCheckResourceAttrSet("selectel_vpc_public_port_v1.port", "subnet"),
					resource.TestCheckResourceAttrSet("selectel_vpc_public_port_v1.port", "gateway"),
					resource.TestCheckResourceAttr("selectel_vpc_public_port_v1.port", "admin_state_up", "true"),
					resource.TestCheckResourceAttr("selectel_vpc_public_port_v1.port", "security_group_ids.#", "1"),
				),
			},
			{
				Config: testAccVPCPublicPortV1WithAdminStateUp(region, projectID, false),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("selectel_vpc_public_port_v1.port", "id"),
					resource.TestCheckResourceAttr("selectel_vpc_public_port_v1.port", "project_id", projectID),
					resource.TestCheckResourceAttrSet("selectel_vpc_public_port_v1.port", "network_id"),
					resource.TestCheckResourceAttrSet("selectel_vpc_public_port_v1.port", "ip_address"),
					resource.TestCheckResourceAttr("selectel_vpc_public_port_v1.port", "description", ""),
					resource.TestCheckResourceAttrSet("selectel_vpc_public_port_v1.port", "subnet"),
					resource.TestCheckResourceAttrSet("selectel_vpc_public_port_v1.port", "gateway"),
					resource.TestCheckResourceAttr("selectel_vpc_public_port_v1.port", "admin_state_up", "false"),
					resource.TestCheckResourceAttr("selectel_vpc_public_port_v1.port", "security_group_ids.#", "1"),
				),
			},
			{
				Config: testAccVPCPublicPortV1WithDescription(region, projectID, ""),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("selectel_vpc_public_port_v1.port", "id"),
					resource.TestCheckResourceAttr("selectel_vpc_public_port_v1.port", "project_id", projectID),
					resource.TestCheckResourceAttrSet("selectel_vpc_public_port_v1.port", "network_id"),
					resource.TestCheckResourceAttrSet("selectel_vpc_public_port_v1.port", "ip_address"),
					resource.TestCheckResourceAttr("selectel_vpc_public_port_v1.port", "description", ""),
					resource.TestCheckResourceAttrSet("selectel_vpc_public_port_v1.port", "subnet"),
					resource.TestCheckResourceAttrSet("selectel_vpc_public_port_v1.port", "gateway"),
					resource.TestCheckResourceAttr("selectel_vpc_public_port_v1.port", "admin_state_up", "true"),
					resource.TestCheckResourceAttr("selectel_vpc_public_port_v1.port", "security_group_ids.#", "1"),
				),
			},
			{
				Config: testAccVPCPublicPortV1WithDescription(region, projectID, "tf-acc-test-desc"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("selectel_vpc_public_port_v1.port", "id"),
					resource.TestCheckResourceAttr("selectel_vpc_public_port_v1.port", "project_id", projectID),
					resource.TestCheckResourceAttrSet("selectel_vpc_public_port_v1.port", "network_id"),
					resource.TestCheckResourceAttrSet("selectel_vpc_public_port_v1.port", "ip_address"),
					resource.TestCheckResourceAttr("selectel_vpc_public_port_v1.port", "description", "tf-acc-test-desc"),
					resource.TestCheckResourceAttrSet("selectel_vpc_public_port_v1.port", "subnet"),
					resource.TestCheckResourceAttrSet("selectel_vpc_public_port_v1.port", "gateway"),
					resource.TestCheckResourceAttr("selectel_vpc_public_port_v1.port", "admin_state_up", "true"),
					resource.TestCheckResourceAttr("selectel_vpc_public_port_v1.port", "security_group_ids.#", "1"),
				),
			},
		},
	})
}

func TestAccVPCPublicPortV1Complex(t *testing.T) {
	region := os.Getenv("INFRA_REGION")
	projectID := os.Getenv("INFRA_PROJECT_ID")

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheckWithProjectID(t) },
		ProviderFactories: testAccProvidersWithOpenStack,
		CheckDestroy:      testAccVPCPublicPortV1CheckDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCPublicPortV1WithSecurityGroups(region, projectID),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("selectel_vpc_public_port_v1.port", "id"),
					resource.TestCheckResourceAttr("selectel_vpc_public_port_v1.port", "project_id", projectID),
					resource.TestCheckResourceAttrSet("selectel_vpc_public_port_v1.port", "network_id"),
					resource.TestCheckResourceAttrSet("selectel_vpc_public_port_v1.port", "ip_address"),
					resource.TestCheckResourceAttr("selectel_vpc_public_port_v1.port", "description", ""),
					resource.TestCheckResourceAttrSet("selectel_vpc_public_port_v1.port", "subnet"),
					resource.TestCheckResourceAttrSet("selectel_vpc_public_port_v1.port", "gateway"),
					resource.TestCheckResourceAttr("selectel_vpc_public_port_v1.port", "admin_state_up", "true"),
					resource.TestCheckResourceAttr("selectel_vpc_public_port_v1.port", "security_group_ids.#", "1"),
					resource.TestCheckResourceAttrPair(
						"selectel_vpc_public_port_v1.port",
						"security_group_ids.0",
						"openstack_networking_secgroup_v2.sg",
						"id",
					),
				),
			},
		},
	})
}

func testAccVPCPublicPortV1(region, projectID string) string {
	return fmt.Sprintf(`
resource "selectel_vpc_public_port_v1" "port" {
  region      = %q
  project_id  = %q
}`, region, projectID)
}

func testAccVPCPublicPortV1WithDescription(region, projectID, description string) string {
	return fmt.Sprintf(`
resource "selectel_vpc_public_port_v1" "port" {
    region      = %q
    project_id  = %q
    description = %q
}`, region, projectID, description)
}

func testAccVPCPublicPortV1WithAdminStateUp(region, projectID string, adminStateUp bool) string {
	return fmt.Sprintf(`
resource "selectel_vpc_public_port_v1" "port" {
  region         = %q
  project_id     = %q
  admin_state_up = %v
}`, region, projectID, adminStateUp)
}

func testAccVPCPublicPortV1WithSecurityGroups(region, projectID string) string {
	return fmt.Sprintf(`
provider openstack {
    tenant_id = %q
}

resource "openstack_networking_secgroup_v2" "sg" {
    region = %q
    name   = "tf-acc-test-sg"
}

resource "selectel_vpc_public_port_v1" "port" {
    region             = %q
    project_id         = %q
    security_group_ids = [openstack_networking_secgroup_v2.sg.id]
}`, projectID, region, region, projectID)
}

func testAccVPCPublicPortV1CheckExist(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource %q not found in Terraform state", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("resource %q has no ID", resourceName)
		}

		client, err := newTestPublicNetAPI(rs, testAccProvider)
		if err != nil {
			return err
		}

		if _, err := client.GetPort(context.Background(), rs.Primary.ID); err != nil {
			return fmt.Errorf("failed to get port %q: %w", rs.Primary.ID, err)
		}

		return nil
	}
}

func testAccVPCPublicPortV1CheckDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "selectel_vpc_public_port_v1" {
			continue
		}

		client, err := newTestPublicNetAPI(rs, testAccProvider)
		if err != nil {
			return err
		}

		_, err = client.GetPort(context.Background(), rs.Primary.ID)
		if err == nil {
			return fmt.Errorf("selectel_vpc_public_port_v1.port %q still exists", rs.Primary.ID)
		}

		if isNotFound(err) {
			return err
		}
	}

	return nil
}

func TestResourceVPCPublicPortV1ImportStateByIdentity(t *testing.T) {
	r := resourceVPCPublicPortV1()
	require.NotNil(t, r.Identity, "resource must declare an identity schema")

	d := schema.TestResourceDataWithIdentityRaw(t, r.SchemaMap(), r.Identity.SchemaMap(), map[string]string{
		"id":         "b311ce58-2658-46b5-b733-7a0f418703f2",
		"project_id": "3d3bb1bd-bc57-4b6d-9a6f-2f2a1c25d3d4",
		"region":     "ru-6",
	})

	result, err := resourceVPCPublicPortV1ImportState(context.Background(), d, &Config{})
	require.NoError(t, err)
	require.Len(t, result, 1)

	assert.Equal(t, "b311ce58-2658-46b5-b733-7a0f418703f2", result[0].Id())
	assert.Equal(t, "3d3bb1bd-bc57-4b6d-9a6f-2f2a1c25d3d4", result[0].Get("project_id"))
	assert.Equal(t, "ru-6", result[0].Get("region"))
}

func TestResourceVPCPublicPortV1ImportStateByID(t *testing.T) {
	r := resourceVPCPublicPortV1()
	d := r.TestResourceData()
	d.SetId("b311ce58-2658-46b5-b733-7a0f418703f2")

	result, err := resourceVPCPublicPortV1ImportState(context.Background(), d, &Config{
		ProjectID: "3d3bb1bd-bc57-4b6d-9a6f-2f2a1c25d3d4",
		Region:    "ru-6",
	})
	require.NoError(t, err)
	require.Len(t, result, 1)

	assert.Equal(t, "b311ce58-2658-46b5-b733-7a0f418703f2", result[0].Id())
	assert.Equal(t, "3d3bb1bd-bc57-4b6d-9a6f-2f2a1c25d3d4", result[0].Get("project_id"))
	assert.Equal(t, "ru-6", result[0].Get("region"))
}

func TestResourceVPCPublicPortV1ImportStateByIDWithoutProjectScope(t *testing.T) {
	r := resourceVPCPublicPortV1()
	d := r.TestResourceData()
	d.SetId("b311ce58-2658-46b5-b733-7a0f418703f2")

	_, err := resourceVPCPublicPortV1ImportState(context.Background(), d, &Config{})
	assert.Error(t, err)
}

func TestFillVPCPublicPortDataSetsIdentity(t *testing.T) {
	r := resourceVPCPublicPortV1()
	d := r.TestResourceData()
	d.SetId("b311ce58-2658-46b5-b733-7a0f418703f2")
	require.NoError(t, d.Set("region", "ru-6"))

	description := "port"
	fillVPCPublicPortData(&publicnetapi.Port{
		ID:          "b311ce58-2658-46b5-b733-7a0f418703f2",
		ProjectID:   "3d3bb1bd-bc57-4b6d-9a6f-2f2a1c25d3d4",
		Description: &description,
	}, d)

	identity, err := d.Identity()
	require.NoError(t, err)

	assert.Equal(t, "b311ce58-2658-46b5-b733-7a0f418703f2", identity.Get("id"))
	assert.Equal(t, "3d3bb1bd-bc57-4b6d-9a6f-2f2a1c25d3d4", identity.Get("project_id"))
	assert.Equal(t, "ru-6", identity.Get("region"))
}
