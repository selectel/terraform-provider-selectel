package selectel

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/selectel/go-selvpcclient/v4/selvpcclient/quotamanager/quotas"
	v1 "github.com/selectel/mks-go/pkg/v1"
	"github.com/selectel/mks-go/pkg/v1/kubeversion"
	"github.com/selectel/mks-go/pkg/v1/node"
	"github.com/selectel/mks-go/pkg/v1/nodegroup"
	"github.com/stretchr/testify/assert"
)

func newTestMKSClient(rs *terraform.ResourceState, testAccProvider *schema.Provider) (*v1.ServiceClient, error) {
	config := testAccProvider.Meta().(*Config)

	var projectID string
	var endpoint string

	if id, ok := rs.Primary.Attributes["project_id"]; ok {
		projectID = id
	}

	selvpcClient, err := config.GetSelVPCClientWithProjectScope(projectID)
	if err != nil {
		return nil, fmt.Errorf("can't get selvpc client for mks acc tests: %w", err)
	}

	if region, ok := rs.Primary.Attributes["region"]; ok {
		mksEndpoint, err := selvpcClient.Catalog.GetEndpoint(MKS, region)
		if err != nil {
			return nil, fmt.Errorf("can't get endpoint for mks acc tests: %w", err)
		}
		endpoint = mksEndpoint.URL
	}

	mksClient := v1.NewMKSClientV1(selvpcClient.GetXAuthToken(), endpoint)

	return mksClient, nil
}

func TestKubeVersionToMajorValid(t *testing.T) {
	tableTests := []struct {
		kubeVersion string
		expected    int
	}{
		{
			kubeVersion: "v1.15.3",
			expected:    1,
		},
		{
			kubeVersion: "v1.16",
			expected:    1,
		},
		{
			kubeVersion: "v1",
			expected:    1,
		},
		{
			kubeVersion: "2",
			expected:    2,
		},
		{
			kubeVersion: "v1.abc",
			expected:    1,
		},
		{
			kubeVersion: "1.-25",
			expected:    1,
		},
	}

	for _, test := range tableTests {
		actual, err := kubeVersionToMajor(test.kubeVersion)
		if err != nil {
			t.Error(err)
		}
		if actual != test.expected {
			t.Errorf("Expected %d kube version, but got: %d", test.expected, actual)
		}
	}
}

func TestKubeVersionToMajorInvalid(t *testing.T) {
	tableTests := []string{
		"",
		"v.12.a3",
		"abc",
		"abc.def",
		"-1.12.13",
		"-v.15.1",
		"-1",
	}

	for _, kubeVersion := range tableTests {
		actual, err := kubeVersionToMajor(kubeVersion)
		if err == nil {
			t.Error("Expected kube version parsing error but got nil")
		}
		if actual != 0 {
			t.Errorf("Expected empty kube version, but got: %d", actual)
		}
	}
}

func TestKubeVersionToMinorValid(t *testing.T) {
	tableTests := []struct {
		kubeVersion string
		expected    int
	}{
		{
			kubeVersion: "v1.15.3",
			expected:    15,
		},
		{
			kubeVersion: "v1.16.0",
			expected:    16,
		},
		{
			kubeVersion: "1.15",
			expected:    15,
		},
		{
			kubeVersion: "2.21.22.20",
			expected:    21,
		},
		{
			kubeVersion: "a.21",
			expected:    21,
		},
		{
			kubeVersion: "-2.21",
			expected:    21,
		},
	}

	for _, test := range tableTests {
		actual, err := kubeVersionToMinor(test.kubeVersion)
		if err != nil {
			t.Error(err)
		}
		if actual != test.expected {
			t.Errorf("Expected %d kube version, but got: %d", test.expected, actual)
		}
	}
}

func TestKubeVersionToMinorInvalid(t *testing.T) {
	tableTests := []string{
		"",
		"1.v12.a3",
		"v1.a",
		"abc",
		"abc.def",
		"1.-13.12",
		"1.-v.15",
	}

	for _, kubeVersion := range tableTests {
		actual, err := kubeVersionToMinor(kubeVersion)
		if err == nil {
			t.Error("Expected kube version parsing error but got nil")
		}
		if actual != 0 {
			t.Errorf("Expected empty kube version, but got: %d", actual)
		}
	}
}

func TestKubeVersionToPatchValid(t *testing.T) {
	tableTests := []struct {
		kubeVersion string
		expected    int
	}{
		{
			kubeVersion: "v1.15.3",
			expected:    3,
		},
		{
			kubeVersion: "v1.16.0",
			expected:    0,
		},
		{
			kubeVersion: "1.15.2",
			expected:    2,
		},
		{
			kubeVersion: "2.21.22",
			expected:    22,
		},
	}

	for _, test := range tableTests {
		actual, err := kubeVersionToPatch(test.kubeVersion)
		if err != nil {
			t.Error(err)
		}
		if actual != test.expected {
			t.Errorf("Expected %d kube version, but got: %d", test.expected, actual)
		}
	}
}

func TestKubeVersionToPatchInvalid(t *testing.T) {
	tableTests := []string{
		"",
		"v.12.a3",
		"v1.a",
		"abc",
		"abc.def",
		"1.12.-13",
		"1.15.-v",
	}

	for _, kubeVersion := range tableTests {
		actual, err := kubeVersionToPatch(kubeVersion)
		if err == nil {
			t.Error("Expected kube version parsing error but got nil")
		}
		if actual != 0 {
			t.Errorf("Expected empty kube version, but got: %d", actual)
		}
	}
}

func TestCompareTwoKubeVersionsByPatchValid(t *testing.T) {
	tableTests := []struct {
		a, b, result string
	}{
		{
			a:      "v1.15.7",
			b:      "1.15.10",
			result: "1.15.10",
		},
		{
			a:      "1.16.7",
			b:      "v1.16.1",
			result: "1.16.7",
		},
		{
			a:      "1.18.1",
			b:      "1.19.0",
			result: "1.18.1",
		},
	}

	for _, test := range tableTests {
		actual, err := compareTwoKubeVersionsByPatch(test.a, test.b)
		if err != nil {
			t.Error(err)
		}
		if actual != test.result {
			t.Errorf("Expected %s kube version, but got: %s", test.result, actual)
		}
	}
}

func TestCompareTwoKubeVersionsByPatchInvalid(t *testing.T) {
	tableTests := []struct {
		a, b, result string
	}{
		{
			a: "",
			b: "v.12.a3",
		},
		{
			a: "v1.1.1",
			b: "abc",
		},
		{
			a: "1.12.-13",
			b: "v1.15.1",
		},
	}

	for _, test := range tableTests {
		actual, err := compareTwoKubeVersionsByPatch(test.a, test.b)
		if err == nil {
			t.Error("Expected kube version parsing error but got nil")
		}
		if actual != "" {
			t.Errorf("Expected empty kube version, but got: %s", actual)
		}
	}
}

func TestKubeVersionTrimToMinorValid(t *testing.T) {
	tableTests := []struct {
		kubeVersion,
		expected string
	}{
		{
			kubeVersion: "v1.15.3",
			expected:    "1.15",
		},
		{
			kubeVersion: "v1.16.0",
			expected:    "1.16",
		},
		{
			kubeVersion: "1.15.2",
			expected:    "1.15",
		},
		{
			kubeVersion: "v2.0.0-alpha",
			expected:    "2.0",
		},
		{
			kubeVersion: "1.15",
			expected:    "1.15",
		},
		{
			kubeVersion: "v1.17.18.19.20",
			expected:    "1.17",
		},
		{
			kubeVersion: "2.21.22.23.24",
			expected:    "2.21",
		},
	}

	for _, test := range tableTests {
		actual, err := kubeVersionTrimToMinor(test.kubeVersion)
		if err != nil {
			t.Error(err)
		}
		if actual != test.expected {
			t.Errorf("Expected %s kube version, but got: %s", test.expected, actual)
		}
	}
}

func TestKubeVersionTrimToMinorInvalid(t *testing.T) {
	tableTests := []string{
		"",
		"va.12.3",
		"v1.a.3",
		"v1.a",
		"abc",
		"abc.def",
		"-1.12.13",
		"1.-12",
		"v-1.15",
		"v1.-20.3",
	}

	for _, kubeVersion := range tableTests {
		actual, err := kubeVersionTrimToMinor(kubeVersion)
		if err == nil {
			t.Error("Expected kube version parsing error but got nil")
		}
		if actual != "" {
			t.Errorf("Expected empty kube version, but got: %s", actual)
		}
	}
}

func TestKubeVersionTrimToMinorIncrementedValid(t *testing.T) {
	tableTests := []struct {
		kubeVersion,
		expected string
	}{
		{
			kubeVersion: "v1.15.3",
			expected:    "1.16",
		},
		{
			kubeVersion: "v1.16.0",
			expected:    "1.17",
		},
		{
			kubeVersion: "1.15.2",
			expected:    "1.16",
		},
		{
			kubeVersion: "v2.0.0-alpha",
			expected:    "2.1",
		},
		{
			kubeVersion: "1.18",
			expected:    "1.19",
		},
		{
			kubeVersion: "v1.17.18.19.20",
			expected:    "1.18",
		},
		{
			kubeVersion: "2.21.22.23.24",
			expected:    "2.22",
		},
	}

	for _, test := range tableTests {
		actual, err := kubeVersionTrimToMinorIncremented(test.kubeVersion)
		if err != nil {
			t.Error(err)
		}
		if actual != test.expected {
			t.Errorf("Expected %s kube version, but got: %s", test.expected, actual)
		}
	}
}

func TestMKSNodegroupV1ParseID(t *testing.T) {
	id := "5803b490-2d6b-418a-8645-eacda0f003c5/63ed5342-b22c-4c7a-9d41-c1fe4a142c13"

	expectedClusterID := "5803b490-2d6b-418a-8645-eacda0f003c5"
	expectedNodegroupID := "63ed5342-b22c-4c7a-9d41-c1fe4a142c13"

	actualClusterID, actualNodegroupID, err := mksNodegroupV1ParseID(id)

	assert.NoError(t, err)
	assert.Equal(t, expectedClusterID, actualClusterID)
	assert.Equal(t, expectedNodegroupID, actualNodegroupID)
}

func TestMKSNodegroupV1ParseIDErr(t *testing.T) {
	invalidIDs := []string{
		"63ed5342-b22c-4c7a-9d41-c1fe4a142c13",
		"63ed5342-b22c-4c7a-9d41-c1fe4a142c13/",
		"/63ed5342-b22c-4c7a-9d41-c1fe4a142c13",
		"uuid1/uuid2/uuid3",
		"",
	}

	for _, id := range invalidIDs {
		_, _, err := mksNodegroupV1ParseID(id)
		assert.EqualError(t, err, "got error parsing nodegroup ID: "+id)
	}
}

func TestFlattenMKSNodegroupV1Nodes(t *testing.T) {
	views := []*node.View{
		{
			ID:       "94838b31-9ae0-4a23-88ad-256e4f13d345",
			IP:       "198.51.100.101",
			Hostname: "first-node",
		},
		{
			ID:       "ab0128c4-4ba3-4522-85bc-df26eb73f54d",
			IP:       "198.51.100.102",
			Hostname: "second-node",
		},
	}

	expected := []map[string]interface{}{
		{
			"id":       "94838b31-9ae0-4a23-88ad-256e4f13d345",
			"ip":       "198.51.100.101",
			"hostname": "first-node",
		},
		{
			"id":       "ab0128c4-4ba3-4522-85bc-df26eb73f54d",
			"ip":       "198.51.100.102",
			"hostname": "second-node",
		},
	}
	actual := flattenMKSNodegroupV1Nodes(views)

	assert.Equal(t, expected, actual)
}

func TestFlattenMKSNodegroupV1Taints(t *testing.T) {
	views := []nodegroup.Taint{
		{
			Key:    "test-key-0",
			Value:  "test-value-0",
			Effect: nodegroup.NoScheduleEffect,
		},
		{
			Key:    "test-key-1",
			Value:  "test-value-1",
			Effect: nodegroup.NoExecuteEffect,
		},
		{
			Key:    "test-key-2",
			Value:  "test-value-2",
			Effect: nodegroup.PreferNoScheduleEffect,
		},
	}

	expected := []interface{}{
		map[string]interface{}{
			"key":    "test-key-0",
			"value":  "test-value-0",
			"effect": "NoSchedule",
		},
		map[string]interface{}{
			"key":    "test-key-1",
			"value":  "test-value-1",
			"effect": "NoExecute",
		},
		map[string]interface{}{
			"key":    "test-key-2",
			"value":  "test-value-2",
			"effect": "PreferNoSchedule",
		},
	}
	actual := flattenMKSNodegroupV1Taints(views)

	assert.Equal(t, expected, actual)
}

func TestExpandMKSNodegroupV1Labels(t *testing.T) {
	labels := map[string]interface{}{
		"label-key0": "label-value0",
		"label-key1": "label-value1",
		"label-key2": "label-value2",
	}
	expected := map[string]string{
		"label-key0": "label-value0",
		"label-key1": "label-value1",
		"label-key2": "label-value2",
	}
	actual := expandMKSNodegroupV1Labels(labels)

	assert.Equal(t, expected, actual)
}

func TestExpandMKSNodegroupV1Taints(t *testing.T) {
	taints := []interface{}{
		map[string]interface{}{
			"key":    "test-key-0",
			"value":  "test-value-0",
			"effect": "NoSchedule",
		},
		map[string]interface{}{
			"key":    "test-key-1",
			"value":  "test-value-1",
			"effect": "NoExecute",
		},
		map[string]interface{}{
			"key":    "test-key-2",
			"value":  "test-value-2",
			"effect": "PreferNoSchedule",
		},
	}

	expected := []nodegroup.Taint{
		{
			Key:    "test-key-0",
			Value:  "test-value-0",
			Effect: nodegroup.NoScheduleEffect,
		},
		{
			Key:    "test-key-1",
			Value:  "test-value-1",
			Effect: nodegroup.NoExecuteEffect,
		},
		{
			Key:    "test-key-2",
			Value:  "test-value-2",
			Effect: nodegroup.PreferNoScheduleEffect,
		},
	}
	actual := expandMKSNodegroupV1Taints(taints)

	assert.Equal(t, expected, actual)
}

func TestCompareTwoKubeVersionsByMinorValid(t *testing.T) {
	tableTests := []struct {
		a, b, result string
	}{
		{
			a:      "1.22.4",
			b:      "1.20.13",
			result: "1.22.4",
		},
		{
			a:      "1.20.13",
			b:      "1.21.7",
			result: "1.21.7",
		},
		{
			a:      "1.19",
			b:      "1.18.4",
			result: "1.19",
		},
	}

	for _, tt := range tableTests {
		actual, err := compareTwoKubeVersionsByMinor(tt.a, tt.b)
		if err != nil {
			t.Error(err)
		}
		if actual != tt.result {
			t.Errorf("Expected %s kube version, but got: %s", tt.result, actual)
		}
	}
}

func TestCompareTwoKubeVersionsByMinorInvalid(t *testing.T) {
	tableTests := []struct {
		a, b string
	}{
		{
			a: "",
			b: "v.12.a3",
		},
		{
			a: "v1.1.1",
			b: "abc",
		},
		{
			a: "1.-12.-13",
			b: "v1.15.1",
		},
	}

	for _, tt := range tableTests {
		actual, err := compareTwoKubeVersionsByMinor(tt.a, tt.b)
		if err == nil {
			t.Error("Expected kube version parsing error but got nil")
		}
		if actual != "" {
			t.Errorf("Expected empty kube version, but got: %s", actual)
		}
	}
}

func TestFlattenMKSKubeVersionsV1(t *testing.T) {
	versions := []*kubeversion.View{
		{
			Version:   "1.22.4",
			IsDefault: false,
		},
		{
			Version:   "1.21.7",
			IsDefault: true,
		},
		{
			Version:   "1.20.13",
			IsDefault: false,
		},
	}
	expected := []string{"1.22.4", "1.20.13", "1.21.7"}
	actual := flattenMKSKubeVersionsV1(versions)

	assert.ElementsMatch(t, expected, actual)
}

func TestParseMKSKubeVersionsV1Latest(t *testing.T) {
	versions := []*kubeversion.View{
		{
			Version:   "1.22.4",
			IsDefault: false,
		},
		{
			Version:   "1.21.7",
			IsDefault: true,
		},
		{
			Version:   "1.20.13",
			IsDefault: false,
		},
	}
	expected := "1.22.4"
	actual, err := parseMKSKubeVersionsV1Latest(versions)

	assert.NoError(t, err)
	assert.Equal(t, expected, actual)
}

func TestParseMKSKubeVersionsV1Default(t *testing.T) {
	versions := []*kubeversion.View{
		{
			Version:   "1.22.4",
			IsDefault: false,
		},
		{
			Version:   "1.21.7",
			IsDefault: true,
		},
		{
			Version:   "1.20.13",
			IsDefault: false,
		},
	}
	expected := "1.21.7"
	actual := parseMKSKubeVersionsV1Default(versions)

	assert.Equal(t, expected, actual)
}

func TestCheckQuotasForClusterErrRegional(t *testing.T) {
	testQuotas := []*quotas.Quota{
		{
			Name: "mks_cluster_regional",
			ResourceQuotasEntities: []quotas.ResourceQuotaEntity{
				{
					Zone:  "",
					Value: 10,
					Used:  10,
				},
			},
		},
	}

	err := checkQuotasForCluster(testQuotas, false)

	assert.Error(t, err)
	assert.Equal(t, "not enough quota to create regional k8s cluster", err.Error())
}

func TestCheckQuotasForClusterErrZonal(t *testing.T) {
	testQuotas := []*quotas.Quota{
		{
			Name: "mks_cluster_zonal",
			ResourceQuotasEntities: []quotas.ResourceQuotaEntity{
				{
					Zone:  "",
					Value: 10,
					Used:  10,
				},
			},
		},
	}

	err := checkQuotasForCluster(testQuotas, true)

	assert.Error(t, err)
	assert.Equal(t, "not enough quota to create zonal k8s cluster", err.Error())
}

func TestCheckQuotasForRegionalClusterErrUnableToCheck(t *testing.T) {
	var testQuotas []*quotas.Quota

	err := checkQuotasForCluster(testQuotas, false)

	assert.Error(t, err)
	assert.Equal(t, "unable to find regional k8s cluster quotas", err.Error())
}

func TestCheckQuotasForZonalClusterErrUnableToCheck(t *testing.T) {
	var testQuotas []*quotas.Quota

	err := checkQuotasForCluster(testQuotas, true)

	assert.Error(t, err)
	assert.Equal(t, "unable to find zonal k8s cluster quotas", err.Error())
}

func TestCheckQuotasForClusterOkRegional(t *testing.T) {
	testQuotas := []*quotas.Quota{
		{
			Name: "mks_cluster_regional",
			ResourceQuotasEntities: []quotas.ResourceQuotaEntity{
				{
					Zone:  "",
					Value: 10,
					Used:  0,
				},
			},
		},
	}

	assert.NoError(t, checkQuotasForCluster(testQuotas, false))
}

func TestCheckQuotasForClusterOkZonal(t *testing.T) {
	testQuotas := []*quotas.Quota{
		{
			Name: "mks_cluster_zonal",
			ResourceQuotasEntities: []quotas.ResourceQuotaEntity{
				{
					Zone:  "",
					Value: 10,
					Used:  0,
				},
			},
		},
	}

	assert.NoError(t, checkQuotasForCluster(testQuotas, true))
}

var testQuotasFull = []*quotas.Quota{
	{
		Name: "compute_cores",
		ResourceQuotasEntities: []quotas.ResourceQuotaEntity{
			{
				Zone:  "ru-9a",
				Value: 10,
				Used:  10,
			},
		},
	},
	{
		Name: "compute_ram",
		ResourceQuotasEntities: []quotas.ResourceQuotaEntity{
			{
				Zone:  "ru-9a",
				Value: 10,
				Used:  10,
			},
		},
	},
	{
		Name: "volume_gigabytes_fast",
		ResourceQuotasEntities: []quotas.ResourceQuotaEntity{
			{
				Zone:  "ru-9a",
				Value: 10,
				Used:  10,
			},
		},
	},
	{
		Name: "volume_gigabytes_basic",
		ResourceQuotasEntities: []quotas.ResourceQuotaEntity{
			{
				Zone:  "ru-9a",
				Value: 10,
				Used:  10,
			},
		},
	},
	{
		Name: "volume_gigabytes_universal",
		ResourceQuotasEntities: []quotas.ResourceQuotaEntity{
			{
				Zone:  "ru-9a",
				Value: 10,
				Used:  10,
			},
		},
	},
	{
		Name: "volume_gigabytes_local",
		ResourceQuotasEntities: []quotas.ResourceQuotaEntity{
			{
				Zone:  "ru-9a",
				Value: 10,
				Used:  10,
			},
		},
	},
}

func TestCheckQuotasForNodegroupErrCPU(t *testing.T) {
	testNodegroupOpts := nodegroup.CreateOpts{
		Count:            1,
		CPUs:             1,
		RAMMB:            0,
		VolumeType:       "fast.ru-9a",
		VolumeGB:         0,
		AvailabilityZone: "ru-9a",
	}

	err := checkQuotasForNodegroup(testQuotasFull, &testNodegroupOpts)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not enough CPU quota to create nodes")
}

func TestCheckQuotasForNodegroupErrRAM(t *testing.T) {
	testNodegroupOpts := nodegroup.CreateOpts{
		Count:            1,
		CPUs:             0,
		RAMMB:            1,
		VolumeType:       "fast.ru-9a",
		VolumeGB:         0,
		AvailabilityZone: "ru-9a",
	}

	err := checkQuotasForNodegroup(testQuotasFull, &testNodegroupOpts)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not enough RAM quota to create nodes")
}

func TestCheckQuotasForNodegroupErrFastVolume(t *testing.T) {
	testNodegroupOpts := nodegroup.CreateOpts{
		Count:            1,
		CPUs:             0,
		RAMMB:            0,
		VolumeType:       "fast.ru-9a",
		VolumeGB:         1,
		AvailabilityZone: "ru-9a",
	}

	err := checkQuotasForNodegroup(testQuotasFull, &testNodegroupOpts)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not enough fast volume quota to create nodes")
}

func TestCheckQuotasForNodegroupErrBasicVolume(t *testing.T) {
	testNodegroupOpts := nodegroup.CreateOpts{
		Count:            1,
		CPUs:             0,
		RAMMB:            0,
		VolumeType:       "basic.ru-9a",
		VolumeGB:         1,
		AvailabilityZone: "ru-9a",
	}

	err := checkQuotasForNodegroup(testQuotasFull, &testNodegroupOpts)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not enough basic volume quota to create nodes")
}

func TestCheckQuotasForNodegroupErrUniversalVolume(t *testing.T) {
	testNodegroupOpts := nodegroup.CreateOpts{
		Count:            1,
		CPUs:             0,
		RAMMB:            0,
		VolumeType:       "universal.ru-9a",
		VolumeGB:         1,
		AvailabilityZone: "ru-9a",
	}

	err := checkQuotasForNodegroup(testQuotasFull, &testNodegroupOpts)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not enough universal volume quota to create nodes")
}

func TestCheckQuotasForNodegroupErrLocalVolume(t *testing.T) {
	testNodegroupOpts := nodegroup.CreateOpts{
		Count:            1,
		CPUs:             0,
		RAMMB:            0,
		LocalVolume:      true,
		VolumeGB:         1,
		AvailabilityZone: "ru-9a",
	}

	err := checkQuotasForNodegroup(testQuotasFull, &testNodegroupOpts)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not enough local volume quota to create nodes")
}

func TestCheckQuotasForNodegroupErrUnableToFindCPUQuota(t *testing.T) {
	testQuotas := []*quotas.Quota{
		{
			Name: "volume_gigabytes_universal",
			ResourceQuotasEntities: []quotas.ResourceQuotaEntity{
				{
					Zone:  "ru-9a",
					Value: 10,
					Used:  0,
				},
			},
		},
		{
			Name: "compute_ram",
			ResourceQuotasEntities: []quotas.ResourceQuotaEntity{
				{
					Zone:  "ru-9a",
					Value: 10,
					Used:  0,
				},
			},
		},
	}
	testNodegroupOpts := nodegroup.CreateOpts{
		Count:            1,
		CPUs:             2,
		RAMMB:            2,
		VolumeGB:         2,
		VolumeType:       "universal.ru-9a",
		AvailabilityZone: "ru-9a",
	}

	err := checkQuotasForNodegroup(testQuotas, &testNodegroupOpts)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unable to find CPU quota")
}

func TestCheckQuotasForNodegroupErrUnableToFindRAMQuota(t *testing.T) {
	testQuotas := []*quotas.Quota{
		{
			Name: "volume_gigabytes_universal",
			ResourceQuotasEntities: []quotas.ResourceQuotaEntity{
				{
					Zone:  "ru-9a",
					Value: 10,
					Used:  0,
				},
			},
		},
		{
			Name: "compute_cores",
			ResourceQuotasEntities: []quotas.ResourceQuotaEntity{
				{
					Zone:  "ru-9a",
					Value: 10,
					Used:  0,
				},
			},
		},
	}
	testNodegroupOpts := nodegroup.CreateOpts{
		Count:            1,
		CPUs:             2,
		RAMMB:            2,
		VolumeGB:         2,
		VolumeType:       "universal.ru-9a",
		AvailabilityZone: "ru-9a",
	}

	err := checkQuotasForNodegroup(testQuotas, &testNodegroupOpts)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unable to find RAM quota")
}

func TestCheckQuotasForNodegroupErrUnableToFindVolumeQuota(t *testing.T) {
	testQuotas := []*quotas.Quota{
		{
			Name: "compute_ram",
			ResourceQuotasEntities: []quotas.ResourceQuotaEntity{
				{
					Zone:  "ru-9a",
					Value: 10,
					Used:  0,
				},
			},
		},
		{
			Name: "compute_cores",
			ResourceQuotasEntities: []quotas.ResourceQuotaEntity{
				{
					Zone:  "ru-9a",
					Value: 10,
					Used:  0,
				},
			},
		},
	}
	testNodegroupOpts := nodegroup.CreateOpts{
		Count:            1,
		CPUs:             2,
		RAMMB:            2,
		VolumeGB:         2,
		VolumeType:       "universal.ru-9a",
		AvailabilityZone: "ru-9a",
	}

	err := checkQuotasForNodegroup(testQuotas, &testNodegroupOpts)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unable to find volume quota")
}

func TestCheckQuotasForNodegroupErrUnableToCheckCPUQuota(t *testing.T) {
	testQuotas := []*quotas.Quota{
		{
			Name: "compute_ram",
			ResourceQuotasEntities: []quotas.ResourceQuotaEntity{
				{
					Zone:  "ru-9b",
					Value: 10,
					Used:  0,
				},
			},
		},
		{
			Name: "compute_cores",
			ResourceQuotasEntities: []quotas.ResourceQuotaEntity{
				{
					Zone:  "ru-9a",
					Value: 10,
					Used:  0,
				},
			},
		},
		{
			Name: "volume_gigabytes_universal",
			ResourceQuotasEntities: []quotas.ResourceQuotaEntity{
				{
					Zone:  "ru-9b",
					Value: 10,
					Used:  0,
				},
			},
		},
	}

	testNodegroupOpts := nodegroup.CreateOpts{
		Count:            1,
		CPUs:             2,
		RAMMB:            2,
		VolumeGB:         2,
		VolumeType:       "universal.ru-9b",
		AvailabilityZone: "ru-9b",
	}

	err := checkQuotasForNodegroup(testQuotas, &testNodegroupOpts)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unable to check CPU quota for a nodegroup")
}

func TestCheckQuotasForNodegroupErrUnableToCheckRAMQuota(t *testing.T) {
	testQuotas := []*quotas.Quota{
		{
			Name: "compute_ram",
			ResourceQuotasEntities: []quotas.ResourceQuotaEntity{
				{
					Zone:  "ru-9a",
					Value: 10,
					Used:  0,
				},
			},
		},
		{
			Name: "compute_cores",
			ResourceQuotasEntities: []quotas.ResourceQuotaEntity{
				{
					Zone:  "ru-9b",
					Value: 10,
					Used:  0,
				},
			},
		},
		{
			Name: "volume_gigabytes_universal",
			ResourceQuotasEntities: []quotas.ResourceQuotaEntity{
				{
					Zone:  "ru-9b",
					Value: 10,
					Used:  0,
				},
			},
		},
	}
	testNodegroupOpts := nodegroup.CreateOpts{
		Count:            1,
		CPUs:             2,
		RAMMB:            2,
		VolumeGB:         2,
		VolumeType:       "universal.ru-9b",
		AvailabilityZone: "ru-9b",
	}

	err := checkQuotasForNodegroup(testQuotas, &testNodegroupOpts)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unable to check RAM quota for a nodegroup")
}

func TestCheckQuotasForNodegroupErrUnableToCheckVolumeQuota(t *testing.T) {
	testQuotas := []*quotas.Quota{
		{
			Name: "compute_ram",
			ResourceQuotasEntities: []quotas.ResourceQuotaEntity{
				{
					Zone:  "ru-9b",
					Value: 10,
					Used:  0,
				},
			},
		},
		{
			Name: "compute_cores",
			ResourceQuotasEntities: []quotas.ResourceQuotaEntity{
				{
					Zone:  "ru-9b",
					Value: 10,
					Used:  0,
				},
			},
		},
		{
			Name: "volume_gigabytes_universal",
			ResourceQuotasEntities: []quotas.ResourceQuotaEntity{
				{
					Zone:  "ru-9a",
					Value: 10,
					Used:  0,
				},
			},
		},
	}
	testNodegroupOpts := nodegroup.CreateOpts{
		Count:            1,
		CPUs:             2,
		RAMMB:            2,
		VolumeGB:         2,
		VolumeType:       "universal.ru-9b",
		AvailabilityZone: "ru-9b",
	}

	err := checkQuotasForNodegroup(testQuotas, &testNodegroupOpts)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unable to check volume quota for a nodegroup")
}

func TestCheckQuotasForNodegroupErrInvalidVolumeType(t *testing.T) {
	testQuotas := []*quotas.Quota{
		{
			Name: "compute_ram",
			ResourceQuotasEntities: []quotas.ResourceQuotaEntity{
				{
					Zone:  "ru-9b",
					Value: 10,
					Used:  0,
				},
			},
		},
		{
			Name: "compute_cores",
			ResourceQuotasEntities: []quotas.ResourceQuotaEntity{
				{
					Zone:  "ru-9b",
					Value: 10,
					Used:  0,
				},
			},
		},
		{
			Name: "volume_gigabytes_universal",
			ResourceQuotasEntities: []quotas.ResourceQuotaEntity{
				{
					Zone:  "ru-9a",
					Value: 10,
					Used:  0,
				},
			},
		},
	}
	testNodegroupOpts := nodegroup.CreateOpts{
		VolumeGB:         2,
		VolumeType:       "invalid.ru-9b",
		AvailabilityZone: "ru-9b",
	}

	err := checkQuotasForNodegroup(testQuotas, &testNodegroupOpts)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "expected 'fast.<zone>', 'universal.<zone>' or 'basic.<zone>' volume type, got")
}

func TestCheckQuotasForNodegroupOk(t *testing.T) {
	testQuotas := []*quotas.Quota{
		{
			Name: "compute_ram",
			ResourceQuotasEntities: []quotas.ResourceQuotaEntity{
				{
					Zone:  "ru-9a",
					Value: 10,
					Used:  0,
				},
			},
		},
		{
			Name: "compute_cores",
			ResourceQuotasEntities: []quotas.ResourceQuotaEntity{
				{
					Zone:  "ru-9a",
					Value: 10,
					Used:  0,
				},
			},
		},
		{
			Name: "volume_gigabytes_universal",
			ResourceQuotasEntities: []quotas.ResourceQuotaEntity{
				{
					Zone:  "ru-9a",
					Value: 10,
					Used:  0,
				},
			},
		},
	}
	testNodegroupOpts := nodegroup.CreateOpts{
		Count:            1,
		CPUs:             2,
		RAMMB:            2,
		VolumeGB:         2,
		VolumeType:       "universal.ru-9a",
		AvailabilityZone: "ru-9a",
	}

	assert.NoError(t, checkQuotasForNodegroup(testQuotas, &testNodegroupOpts))
}
