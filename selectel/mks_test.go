package selectel

import (
	"testing"

	"github.com/selectel/mks-go/pkg/v1/node"
	"github.com/selectel/mks-go/pkg/v1/nodegroup"
	"github.com/stretchr/testify/assert"
)

func TestGetMKSClusterV1Endpoint(t *testing.T) {
	expectedEndpoints := map[string]string{
		ru1Region: ru1MKSClusterV1Endpoint,
		ru2Region: ru2MKSClusterV1Endpoint,
		ru3Region: ru3MKSClusterV1Endpoint,
		ru7Region: ru7MKSClusterV1Endpoint,
		ru8Region: ru8MKSClusterV1Endpoint,
		ru9Region: ru9MKSClusterV1Endpoint,
	}

	for region, expected := range expectedEndpoints {
		actual := getMKSClusterV1Endpoint(region)
		assert.Equal(t, expected, actual)
	}
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
