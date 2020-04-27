package selectel

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetMKSClusterV1Endpoint(t *testing.T) {
	expectedEndpoints := map[string]string{
		ru1Region: ru1MKSClusterV1Endpoint,
		ru2Region: ru2MKSClusterV1Endpoint,
		ru3Region: ru3MKSClusterV1Endpoint,
		ru7Region: ru7MKSClusterV1Endpoint,
		ru8Region: ru8MKSClusterV1Endpoint,
	}

	for region, expected := range expectedEndpoints {
		actual := getMKSClusterV1Endpoint(region)
		assert.Equal(t, expected, actual)
	}
}

func TestKubeVersionToMinorValid(t *testing.T) {
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
		actual, err := kubeVersionToMinor(test.kubeVersion)
		if err != nil {
			t.Error(err)
		}
		if actual != test.expected {
			t.Errorf("Expected %s kube version, but got: %s", test.expected, actual)
		}
	}
}

func TestKubeVersionToMinorInvalid(t *testing.T) {
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
		actual, err := kubeVersionToMinor(kubeVersion)
		if err == nil {
			t.Error("Expected kube version parsing error but got nil")
		}
		if actual != "" {
			t.Errorf("Expected empty kube version, but got: %s", actual)
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

func TestGetLatestKubeVersionPatchVersionValid(t *testing.T) {
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

func TestGetLatestKubeVersionPatchVersionInvalid(t *testing.T) {
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
