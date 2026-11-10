package selectel

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	vpc "github.com/selectel/vpc-go/pkg/v2"
	"github.com/selectel/vpc-go/pkg/v2/network"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testAccVPCNetworkV2ResourceName = "selectel_vpc_network_v2.network_tf_acc_test_1"

func vpcV2TestEnv() (projectID, region string) {
	return os.Getenv("INFRA_PROJECT_ID"), os.Getenv("INFRA_REGION")
}

func TestVPCNetworkV2DNSDomainValidation(t *testing.T) {
	validate := resourceVPCNetworkV2().Schema["dns_domain"].ValidateFunc
	require.NotNil(t, validate)

	t.Run("rejects a domain without a trailing dot", func(t *testing.T) {
		_, errs := validate("example.com", "dns_domain")

		require.NotEmpty(t, errs)
		assert.Contains(t, errs[0].Error(), "dot at the end")
	})

	for _, value := range []string{"", "example.com."} {
		t.Run("accepts "+value, func(t *testing.T) {
			_, errs := validate(value, "dns_domain")

			assert.Empty(t, errs)
		})
	}
}

func TestFlattenVPCNetworkV2Topology(t *testing.T) {
	d := schema.TestResourceDataRaw(t, resourceVPCNetworkV2().Schema, nil)
	found := &network.Network{
		Subnets:           []string{"subnet-1", "subnet-2"},
		AvailabilityZones: []string{"ru-3a", "ru-3b"},
	}

	flattenVPCNetworkV2(d, found)

	assert.Equal(t, []any{"subnet-1", "subnet-2"}, d.Get("subnets"))
	assert.Equal(t, []any{"ru-3a", "ru-3b"}, d.Get("availability_zones"))
}

func TestAccVPCNetworkV2Basic(t *testing.T) {
	var found network.Network
	projectID, region := vpcV2TestEnv()

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheckWithProjectIDAndRegion(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckVPCNetworkV2Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCNetworkV2Basic(projectID, region),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCNetworkV2Exists(testAccVPCNetworkV2ResourceName, &found),
					resource.TestCheckResourceAttr(testAccVPCNetworkV2ResourceName, "region", region),
					resource.TestCheckResourceAttr(testAccVPCNetworkV2ResourceName, "name", ""),
					resource.TestCheckResourceAttr(testAccVPCNetworkV2ResourceName, "admin_state_up", "true"),
					resource.TestCheckResourceAttrSet(testAccVPCNetworkV2ResourceName, "mtu"),
					resource.TestCheckResourceAttrSet(testAccVPCNetworkV2ResourceName, "shared"),
					resource.TestCheckResourceAttrSet(testAccVPCNetworkV2ResourceName, "blocked"),
					resource.TestCheckResourceAttrSet(testAccVPCNetworkV2ResourceName, "subnets.#"),
					resource.TestCheckResourceAttrSet(testAccVPCNetworkV2ResourceName, "availability_zones.#"),
				),
			},
			{
				Config:   testAccVPCNetworkV2Basic(projectID, region),
				PlanOnly: true,
			},
			{
				ResourceName:      testAccVPCNetworkV2ResourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config:             testAccVPCNetworkV2Basic(projectID, region),
				Check:              testAccCheckVPCV2DeleteOutOfBand(testAccVPCNetworkV2ResourceName, network.Delete),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccVPCNetworkV2Update(t *testing.T) {
	var found network.Network
	projectID, region := vpcV2TestEnv()
	networkName := acctest.RandomWithPrefix("tf-acc-net")
	updatedName := acctest.RandomWithPrefix("tf-acc-net-upd")

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheckWithProjectIDAndRegion(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckVPCNetworkV2Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCNetworkV2OptionalAttributes(projectID, region, networkName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCNetworkV2Exists(testAccVPCNetworkV2ResourceName, &found),
				),
			},
			{
				Config: testAccVPCNetworkV2Updated(projectID, region, updatedName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCNetworkV2SameID(testAccVPCNetworkV2ResourceName, &found),
					resource.TestCheckResourceAttr(testAccVPCNetworkV2ResourceName, "name", updatedName),
					resource.TestCheckResourceAttr(testAccVPCNetworkV2ResourceName, "description", "updated by the acceptance test"),
					resource.TestCheckResourceAttr(testAccVPCNetworkV2ResourceName, "admin_state_up", "false"),
					resource.TestCheckResourceAttr(testAccVPCNetworkV2ResourceName, "tags.#", "1"),
					resource.TestCheckResourceAttr(testAccVPCNetworkV2ResourceName, "all_tags.#", "1"),
				),
			},
			{
				Config:   testAccVPCNetworkV2Updated(projectID, region, updatedName),
				PlanOnly: true,
			},
		},
	})
}

func TestAccVPCNetworkV2TagsOutOfBand(t *testing.T) {
	var found network.Network
	projectID, region := vpcV2TestEnv()
	networkName := acctest.RandomWithPrefix("tf-acc-net")

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheckWithProjectIDAndRegion(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckVPCNetworkV2Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCNetworkV2OptionalAttributes(projectID, region, networkName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCNetworkV2Exists(testAccVPCNetworkV2ResourceName, &found),
					testAccCheckVPCTagsV2AddOutOfBand(testAccVPCNetworkV2TagOperations, "foreign"),
				),
			},
			{
				Config: testAccVPCNetworkV2Updated(projectID, region, networkName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCNetworkV2SameID(testAccVPCNetworkV2ResourceName, &found),
					resource.TestCheckResourceAttr(testAccVPCNetworkV2ResourceName, "tags.#", "1"),
					resource.TestCheckTypeSetElemAttr(testAccVPCNetworkV2ResourceName, "all_tags.*", "foreign"),
					testAccCheckVPCTagsV2Match(testAccVPCNetworkV2TagOperations, []string{"tf-acc", "foreign"}),
				),
			},
			{
				Config:   testAccVPCNetworkV2Updated(projectID, region, networkName),
				PlanOnly: true,
			},
		},
	})
}

const testAccVPCV2DriftedName = "renamed outside terraform"

type vpcNetworkV2TestHandle struct {
	client *vpc.Client
	id     string
}

func testAccVPCNetworkV2Handle(s *terraform.State, resourceName string) (*vpcNetworkV2TestHandle, error) {
	rs, ok := s.RootModule().Resources[resourceName]
	if !ok {
		return nil, fmt.Errorf("not found: %s", resourceName)
	}

	if rs.Primary.ID == "" {
		return nil, errors.New("no ID is set")
	}

	client, err := vpcV2Client(
		testAccProvider.Meta(),
		rs.Primary.Attributes["project_id"],
		rs.Primary.Attributes["region"],
	)
	if err != nil {
		return nil, err
	}

	return &vpcNetworkV2TestHandle{client: client, id: rs.Primary.ID}, nil
}

func testAccCheckVPCV2DeleteOutOfBand(
	resourceName string, remove func(context.Context, *vpc.Client, string) error,
) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		handle, err := testAccVPCNetworkV2Handle(s, resourceName)
		if err != nil {
			return err
		}

		return remove(context.Background(), handle.client, handle.id)
	}
}

func testAccCheckVPCNetworkV2Exists(resourceName string, found *network.Network) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		handle, err := testAccVPCNetworkV2Handle(s, resourceName)
		if err != nil {
			return err
		}

		result, err := network.Get(context.Background(), handle.client, handle.id)
		if err != nil {
			return err
		}

		if result.ID != handle.id {
			return errors.New("VPC network not found")
		}

		*found = *result

		return nil
	}
}

func testAccVPCNetworkV2TagOperations(s *terraform.State) (vpcTagsV2AccOperations, error) {
	handle, err := testAccVPCNetworkV2Handle(s, testAccVPCNetworkV2ResourceName)
	if err != nil {
		return nil, err
	}

	return network.TagOperations(handle.client, handle.id), nil
}

func testAccCheckVPCNetworkV2SameID(resourceName string, found *network.Network) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("not found: %s", resourceName)
		}

		if rs.Primary.ID != found.ID {
			return fmt.Errorf("VPC network was replaced: %s -> %s", found.ID, rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckVPCNetworkV2Destroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "selectel_vpc_network_v2" {
			continue
		}

		client, err := vpcV2Client(
			testAccProvider.Meta(),
			rs.Primary.Attributes["project_id"],
			rs.Primary.Attributes["region"],
		)
		if err != nil {
			return err
		}

		if _, err := network.Get(context.Background(), client, rs.Primary.ID); err == nil {
			return errors.New("VPC network still exists")
		} else if !vpc.IsErrorClass(err, vpc.ErrorClassNotFound) {
			return err
		}
	}

	return nil
}

func testAccVPCV2Base(projectID string) string {
	return fmt.Sprintf(`
locals {
  project_id = %q
}
`, projectID)
}

func testAccVPCNetworkV2Basic(projectID, region string) string {
	return testAccVPCV2Base(projectID) + fmt.Sprintf(`
resource "selectel_vpc_network_v2" "network_tf_acc_test_1" {
  project_id = local.project_id
  region     = "%s"
}
`, region)
}

func testAccVPCNetworkV2OptionalAttributes(projectID, region, networkName string) string {
	return testAccVPCV2Base(projectID) + fmt.Sprintf(`
resource "selectel_vpc_network_v2" "network_tf_acc_test_1" {
  project_id     = local.project_id
  region         = "%s"
  name           = "%s"
  description    = "created by the acceptance test"
  admin_state_up = true
  tags           = ["tf-acc", "VPC"]
}
`, region, networkName)
}

func testAccVPCNetworkV2Updated(projectID, region, networkName string) string {
	return testAccVPCV2Base(projectID) + fmt.Sprintf(`
resource "selectel_vpc_network_v2" "network_tf_acc_test_1" {
  project_id     = local.project_id
  region         = "%s"
  name           = "%s"
  description    = "updated by the acceptance test"
  admin_state_up = false
  tags           = ["tf-acc"]
}
`, region, networkName)
}
