package selectel

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/selectel/go-selvpcclient/v3/selvpcclient/resell/v2/projects"
	v1 "github.com/selectel/mks-go/pkg/v1"
	"github.com/selectel/mks-go/pkg/v1/cluster"
	"github.com/selectel/mks-go/pkg/v1/kubeoptions"
	"github.com/selectel/mks-go/pkg/v1/kubeversion"
)

func TestAccMKSClusterV1Basic(t *testing.T) {
	var (
		mksCluster cluster.View
		project    projects.Project
	)

	projectName := acctest.RandomWithPrefix("tf-acc")
	clusterName := acctest.RandomWithPrefix("tf-acc-cl")
	kubeVersion := testAccMKSClusterV1GetDefaultKubeVersion(t)
	maintenanceWindowStart := testAccMKSClusterV1GetMaintenanceWindowStart(12 * time.Hour)
	maintenanceWindowStartUpdated := testAccMKSClusterV1GetMaintenanceWindowStart(14 * time.Hour)

	defaultFeatureGates := testDefaultFeatureGates(t)
	defaultAdmissionControllers := testDefaultAdmissionControllers(t)
	featureGates := defaultFeatureGates[:1]
	featureGatesUpdate := defaultFeatureGates[1:2]
	admissionControllers := defaultAdmissionControllers[:1]
	admissionControllersUpdate := defaultAdmissionControllers[1:2]

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckVPCV2ProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccMKSClusterV1BasicWithKubeOptions(projectName, clusterName, kubeVersion, maintenanceWindowStart, featureGates, admissionControllers),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCV2ProjectExists("selectel_vpc_project_v2.project_tf_acc_test_1", &project),
					testAccCheckMKSClusterV1Exists("selectel_mks_cluster_v1.cluster_tf_acc_test_1", &mksCluster),
					resource.TestCheckResourceAttr("selectel_mks_cluster_v1.cluster_tf_acc_test_1", "name", clusterName),
					resource.TestCheckResourceAttr("selectel_mks_cluster_v1.cluster_tf_acc_test_1", "kube_version", kubeVersion),
					resource.TestCheckResourceAttr("selectel_mks_cluster_v1.cluster_tf_acc_test_1", "region", "ru-9"),
					resource.TestCheckResourceAttr("selectel_mks_cluster_v1.cluster_tf_acc_test_1", "enable_autorepair", "true"),
					resource.TestCheckResourceAttr("selectel_mks_cluster_v1.cluster_tf_acc_test_1", "enable_patch_version_auto_upgrade", "true"),
					resource.TestCheckResourceAttr("selectel_mks_cluster_v1.cluster_tf_acc_test_1", "maintenance_window_start", maintenanceWindowStart),
					resource.TestCheckResourceAttr("selectel_mks_cluster_v1.cluster_tf_acc_test_1", "status", "ACTIVE"),
					resource.TestCheckResourceAttr("selectel_mks_cluster_v1.cluster_tf_acc_test_1", "feature_gates.0", defaultFeatureGates[0]),
					resource.TestCheckResourceAttr("selectel_mks_cluster_v1.cluster_tf_acc_test_1", "admission_controllers.0", defaultAdmissionControllers[0]),
				),
			},
			{
				Config: testAccMKSClusterV1UpdateWithKubeOptions(projectName, clusterName, kubeVersion, maintenanceWindowStartUpdated, featureGatesUpdate, admissionControllersUpdate),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("selectel_mks_cluster_v1.cluster_tf_acc_test_1", "name", clusterName),
					resource.TestCheckResourceAttr("selectel_mks_cluster_v1.cluster_tf_acc_test_1", "kube_version", kubeVersion),
					resource.TestCheckResourceAttr("selectel_mks_cluster_v1.cluster_tf_acc_test_1", "region", "ru-9"),
					resource.TestCheckResourceAttr("selectel_mks_cluster_v1.cluster_tf_acc_test_1", "enable_autorepair", "false"),
					resource.TestCheckResourceAttr("selectel_mks_cluster_v1.cluster_tf_acc_test_1", "enable_patch_version_auto_upgrade", "false"),
					resource.TestCheckResourceAttr("selectel_mks_cluster_v1.cluster_tf_acc_test_1", "enable_pod_security_policy", "false"),
					resource.TestCheckResourceAttr("selectel_mks_cluster_v1.cluster_tf_acc_test_1", "maintenance_window_start", maintenanceWindowStartUpdated),
					resource.TestCheckResourceAttr("selectel_mks_cluster_v1.cluster_tf_acc_test_1", "status", "ACTIVE"),
					resource.TestCheckResourceAttr("selectel_mks_cluster_v1.cluster_tf_acc_test_1", "feature_gates.0", defaultFeatureGates[1]),
					resource.TestCheckResourceAttr("selectel_mks_cluster_v1.cluster_tf_acc_test_1", "admission_controllers.0", defaultAdmissionControllers[1]),
				),
			},
		},
	})
}

func TestAccMKSClusterV1Zonal(t *testing.T) {
	var (
		mksCluster cluster.View
		project    projects.Project
	)

	projectName := acctest.RandomWithPrefix("tf-acc")
	clusterName := acctest.RandomWithPrefix("tf-acc-cl")
	kubeVersion := testAccMKSClusterV1GetDefaultKubeVersion(t)
	maintenanceWindowStart := testAccMKSClusterV1GetMaintenanceWindowStart(12 * time.Hour)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckVPCV2ProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccMKSClusterV1Zonal(projectName, clusterName, kubeVersion, maintenanceWindowStart),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCV2ProjectExists("selectel_vpc_project_v2.project_tf_acc_test_1", &project),
					testAccCheckMKSClusterV1Exists("selectel_mks_cluster_v1.cluster_tf_acc_test_1", &mksCluster),
					resource.TestCheckResourceAttr("selectel_mks_cluster_v1.cluster_tf_acc_test_1", "name", clusterName),
					resource.TestCheckResourceAttr("selectel_mks_cluster_v1.cluster_tf_acc_test_1", "kube_version", kubeVersion),
					resource.TestCheckResourceAttr("selectel_mks_cluster_v1.cluster_tf_acc_test_1", "region", "ru-9"),
					resource.TestCheckResourceAttr("selectel_mks_cluster_v1.cluster_tf_acc_test_1", "enable_autorepair", "true"),
					resource.TestCheckResourceAttr("selectel_mks_cluster_v1.cluster_tf_acc_test_1", "enable_patch_version_auto_upgrade", "false"),
					resource.TestCheckResourceAttr("selectel_mks_cluster_v1.cluster_tf_acc_test_1", "zonal", "true"),
					resource.TestCheckResourceAttr("selectel_mks_cluster_v1.cluster_tf_acc_test_1", "private_kube_api", "false"),
					resource.TestCheckResourceAttr("selectel_mks_cluster_v1.cluster_tf_acc_test_1", "maintenance_window_start", maintenanceWindowStart),
					resource.TestCheckResourceAttr("selectel_mks_cluster_v1.cluster_tf_acc_test_1", "status", "ACTIVE"),
				),
			},
		},
	})
}

func TestAccMKSClusterV1PrivateKubeAPI(t *testing.T) {
	var (
		mksCluster cluster.View
		project    projects.Project
	)

	projectName := acctest.RandomWithPrefix("tf-acc")
	clusterName := acctest.RandomWithPrefix("tf-acc-cl")
	kubeVersion := testAccMKSClusterV1GetDefaultKubeVersion(t)
	maintenanceWindowStart := testAccMKSClusterV1GetMaintenanceWindowStart(12 * time.Hour)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckVPCV2ProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccMKSClusterV1PrivateKubeAPI(projectName, clusterName, kubeVersion, maintenanceWindowStart),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCV2ProjectExists("selectel_vpc_project_v2.project_tf_acc_test_1", &project),
					testAccCheckMKSClusterV1Exists("selectel_mks_cluster_v1.cluster_tf_acc_test_1", &mksCluster),
					resource.TestCheckResourceAttr("selectel_mks_cluster_v1.cluster_tf_acc_test_1", "name", clusterName),
					resource.TestCheckResourceAttr("selectel_mks_cluster_v1.cluster_tf_acc_test_1", "kube_version", kubeVersion),
					resource.TestCheckResourceAttr("selectel_mks_cluster_v1.cluster_tf_acc_test_1", "region", "ru-9"),
					resource.TestCheckResourceAttr("selectel_mks_cluster_v1.cluster_tf_acc_test_1", "enable_autorepair", "true"),
					resource.TestCheckResourceAttr("selectel_mks_cluster_v1.cluster_tf_acc_test_1", "enable_patch_version_auto_upgrade", "false"),
					resource.TestCheckResourceAttr("selectel_mks_cluster_v1.cluster_tf_acc_test_1", "zonal", "false"),
					resource.TestCheckResourceAttr("selectel_mks_cluster_v1.cluster_tf_acc_test_1", "private_kube_api", "true"),
					resource.TestCheckResourceAttr("selectel_mks_cluster_v1.cluster_tf_acc_test_1", "maintenance_window_start", maintenanceWindowStart),
					resource.TestCheckResourceAttr("selectel_mks_cluster_v1.cluster_tf_acc_test_1", "status", "ACTIVE"),
				),
			},
		},
	})
}

func testAccMKSClusterV1GetMaintenanceWindowStart(delay time.Duration) string {
	return time.Now().UTC().Add(delay).Format("15:04:00")
}

func testAccMKSClusterV1GetDefaultKubeVersion(t *testing.T) string {
	var (
		kubeVersion string
		project     projects.Project
	)
	projectName := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckVPCV2ProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCV2ProjectBasic(projectName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCV2ProjectExists("selectel_vpc_project_v2.project_tf_acc_test_1", &project),
					testAccCheckMKSClusterV1DefaultKubeVersion("selectel_vpc_project_v2.project_tf_acc_test_1", &kubeVersion),
				),
			},
		},
	})

	return kubeVersion
}

func testAccCheckMKSClusterV1DefaultKubeVersion(n string, kubeVersion *string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}
		if rs.Primary.ID == "" {
			return errors.New("no ID is set")
		}

		mksClient, err := newTestMKSClientWithParams(rs.Primary.ID, testRu3Region)
		if err != nil {
			return err
		}
		ctx := context.Background()

		defaultVersion, err := getDefaultKubeVersion(ctx, mksClient)
		if err != nil {
			return err
		}

		*kubeVersion = defaultVersion

		return nil
	}
}

func testAccCheckMKSClusterV1Exists(n string, mksCluster *cluster.View) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return errors.New("no ID is set")
		}

		ctx := context.Background()

		mksClient, err := newTestMKSClient(rs, testAccProvider)
		if err != nil {
			return err
		}

		foundCluster, _, err := cluster.Get(ctx, mksClient, rs.Primary.ID)
		if err != nil {
			return err
		}

		if foundCluster.ID != rs.Primary.ID {
			return errors.New("cluster not found")
		}

		*mksCluster = *foundCluster

		return nil
	}
}

func testAccMKSClusterV1Basic(projectName, clusterName, kubeVersion, maintenanceWindowStart string) string {
	return fmt.Sprintf(`
resource "selectel_vpc_project_v2" "project_tf_acc_test_1" {
  name        = "%s"
}
resource "selectel_mks_cluster_v1" "cluster_tf_acc_test_1" {
  name                     = "%s"
  kube_version             = "%s"
  project_id               = "${selectel_vpc_project_v2.project_tf_acc_test_1.id}"
  region                   = "ru-9"
  maintenance_window_start = "%s"
}`, projectName, clusterName, kubeVersion, maintenanceWindowStart)
}

func testAccMKSClusterV1BasicWithKubeOptions(projectName, clusterName, kubeVersion, maintenanceWindowStart string, featureGates, admissionControllers []string) string {
	flatFeatureGates := flatStringsListWithQuotes(featureGates)
	flatAdmissionControllers := flatStringsListWithQuotes(admissionControllers)

	return fmt.Sprintf(`
resource "selectel_vpc_project_v2" "project_tf_acc_test_1" {
  name        = "%s"
}
resource "selectel_mks_cluster_v1" "cluster_tf_acc_test_1" {
  name                     = "%s"
  kube_version             = "%s"
  project_id               = "${selectel_vpc_project_v2.project_tf_acc_test_1.id}"
  region                   = "ru-9"
  maintenance_window_start = "%s"
  feature_gates            = [%s]
  admission_controllers    = [%s]
}`, projectName, clusterName, kubeVersion, maintenanceWindowStart, flatFeatureGates, flatAdmissionControllers)
}

func testAccMKSClusterV1UpdateWithKubeOptions(projectName, clusterName, kubeVersion, maintenanceWindowStart string, featureGates, admissionControllers []string) string {
	flatFeatureGates := flatStringsListWithQuotes(featureGates)
	flatAdmissionControllers := flatStringsListWithQuotes(admissionControllers)

	return fmt.Sprintf(`
resource "selectel_vpc_project_v2" "project_tf_acc_test_1" {
  name        = "%s"
}
resource "selectel_mks_cluster_v1" "cluster_tf_acc_test_1" {
  name         = "%s"
  kube_version = "%s"
  project_id                        = "${selectel_vpc_project_v2.project_tf_acc_test_1.id}"
  region                            = "ru-9"
  maintenance_window_start          = "%s"
  enable_autorepair                 = false
  enable_patch_version_auto_upgrade = false
  enable_pod_security_policy        = false
  feature_gates                     = [%s]
  admission_controllers             = [%s]
}`, projectName, clusterName, kubeVersion, maintenanceWindowStart, flatFeatureGates, flatAdmissionControllers)
}

func testAccMKSClusterV1Zonal(projectName, clusterName, kubeVersion, maintenanceWindowStart string) string {
	return fmt.Sprintf(`
 resource "selectel_vpc_project_v2" "project_tf_acc_test_1" {
   name        = "%s"
 }
 resource "selectel_mks_cluster_v1" "cluster_tf_acc_test_1" {
   name                              = "%s"
   kube_version                      = "%s"
   project_id                        = "${selectel_vpc_project_v2.project_tf_acc_test_1.id}"
   region                            = "ru-9"
   maintenance_window_start          = "%s"
   enable_patch_version_auto_upgrade = false
   zonal                             = true
 }`, projectName, clusterName, kubeVersion, maintenanceWindowStart)
}

func testAccMKSClusterV1PrivateKubeAPI(projectName, clusterName, kubeVersion, maintenanceWindowStart string) string {
	return fmt.Sprintf(`
 resource "selectel_vpc_project_v2" "project_tf_acc_test_1" {
   name        = "%s"
 }
 resource "selectel_mks_cluster_v1" "cluster_tf_acc_test_1" {
   name                              = "%s"
   kube_version                      = "%s"
   project_id                        = "${selectel_vpc_project_v2.project_tf_acc_test_1.id}"
   region                            = "ru-9"
   maintenance_window_start          = "%s"
   enable_patch_version_auto_upgrade = false
   zonal                             = false
   private_kube_api                  = true
 }`, projectName, clusterName, kubeVersion, maintenanceWindowStart)
}

func testDefaultFeatureGates(t *testing.T) []string {
	var project projects.Project
	featureGates := make([]string, 0)
	projectName := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckVPCV2ProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCV2ProjectBasic(projectName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCV2ProjectExists("selectel_vpc_project_v2.project_tf_acc_test_1", &project),
					testAccCheckMKSClusterV1DefaultKubeVersionFeatureGates("selectel_vpc_project_v2.project_tf_acc_test_1", &featureGates),
				),
			},
		},
	})

	return featureGates
}

func testAccCheckMKSClusterV1DefaultKubeVersionFeatureGates(n string, featureGates *[]string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}
		if rs.Primary.ID == "" {
			return errors.New("no ID is set")
		}

		mksClient, err := newTestMKSClientWithParams(rs.Primary.ID, testRu3Region)
		if err != nil {
			return err
		}
		ctx := context.Background()

		defaultVersion, err := getDefaultKubeVersion(ctx, mksClient)
		if err != nil {
			return err
		}

		allFeatureGates, _, err := kubeoptions.ListFeatureGates(ctx, mksClient)
		if err != nil {
			return err
		}
		kubeVersionMinor, err := kubeVersionTrimToMinor(defaultVersion)
		if err != nil {
			return err
		}

		for _, item := range allFeatureGates {
			if kubeVersionMinor == item.KubeVersion {
				*featureGates = item.Names
			}
		}

		return nil
	}
}

func testDefaultAdmissionControllers(t *testing.T) []string {
	var project projects.Project
	admissionContollers := make([]string, 0)
	projectName := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckVPCV2ProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCV2ProjectBasic(projectName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCV2ProjectExists("selectel_vpc_project_v2.project_tf_acc_test_1", &project),
					testAccCheckMKSClusterV1DefaultKubeVersionAdmissionControllers("selectel_vpc_project_v2.project_tf_acc_test_1", &admissionContollers),
				),
			},
		},
	})

	return admissionContollers
}

func testAccCheckMKSClusterV1DefaultKubeVersionAdmissionControllers(n string, admissionControllers *[]string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}
		if rs.Primary.ID == "" {
			return errors.New("no ID is set")
		}

		mksClient, err := newTestMKSClientWithParams(rs.Primary.ID, testRu3Region)
		if err != nil {
			return err
		}
		ctx := context.Background()

		defaultVersion, err := getDefaultKubeVersion(ctx, mksClient)
		if err != nil {
			return err
		}

		allAdmissionControllers, _, err := kubeoptions.ListAdmissionControllers(ctx, mksClient)
		if err != nil {
			return err
		}
		kubeVersionMinor, err := kubeVersionTrimToMinor(defaultVersion)
		if err != nil {
			return err
		}

		for _, item := range allAdmissionControllers {
			if kubeVersionMinor == item.KubeVersion {
				*admissionControllers = item.Names
			}
		}

		return nil
	}
}

func newTestMKSClientWithParams(projectID, region string) (*v1.ServiceClient, error) {
	config := testAccProvider.Meta().(*Config)

	selvpcClient, err := config.GetSelVPCClientWithProjectScope(projectID)
	if err != nil {
		return nil, fmt.Errorf("can't get selvpc client for mks acc tests: %w", err)
	}
	endpoint, err := selvpcClient.Catalog.GetEndpoint(MKS, region)
	if err != nil {
		return nil, fmt.Errorf("can't get endpoint for mks acc tests: %w", err)
	}

	mksClient := v1.NewMKSClientV1(selvpcClient.GetXAuthToken(), endpoint.URL)

	return mksClient, nil
}

func getDefaultKubeVersion(ctx context.Context, mksClient *v1.ServiceClient) (string, error) {
	kubeVersions, _, err := kubeversion.List(ctx, mksClient)
	if err != nil {
		return "", err
	}

	for _, version := range kubeVersions {
		if version.IsDefault {
			return version.Version, nil
		}
	}

	return "", fmt.Errorf("default kube version is not found")
}

func flatStringsListWithQuotes(list []string) string {
	var builder strings.Builder
	for _, item := range list {
		builder.WriteString(`"`)
		builder.WriteString(item)
		builder.WriteString(`",`)
	}

	return builder.String()
}
