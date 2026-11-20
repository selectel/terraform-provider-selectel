package mks

import (
	"context"
	"testing"

	"github.com/selectel/mks-go/v2/pkg/mksclient"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestKubeVersionParts(t *testing.T) {
	major, err := kubeVersionToMajor("v1.31.8")
	require.NoError(t, err)
	assert.Equal(t, 1, major)
	major, err = kubeVersionToMinor("1.31.8")
	require.NoError(t, err)
	assert.Equal(t, 31, major)

	_, err = kubeVersionToMajor("x.31.8")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "major part is not an integer number")

	_, err = kubeVersionToMinor("1.31")
	require.NoError(t, err)
	_, err = kubeVersionToMinor("1")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "expected to have major and minor version parts")
	_, err = kubeVersionToMinor("1.b.8")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "minor part is not an integer number")

	_, err = kubeVersionToPatch("1.31")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "expected to have major, minor and patch version parts")
	_, err = kubeVersionToPatch("1.31.x")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "patch part is not an integer number")
}

func TestCompareTwoKubeVersionsByPatch(t *testing.T) {
	latest, err := compareTwoKubeVersionsByPatch("1.31.5", "1.31.9")
	require.NoError(t, err)
	assert.Equal(t, "1.31.9", latest)
	latest, err = compareTwoKubeVersionsByPatch("1.31.9", "1.31.5")
	require.NoError(t, err)
	assert.Equal(t, "1.31.9", latest)

	_, err = compareTwoKubeVersionsByPatch("1.31.9", "1.31")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unable to compare kube versions")
	assert.Contains(t, err.Error(), "expected to have major, minor and patch version parts")
}

func TestKubeVersionTrim(t *testing.T) {
	minor, err := kubeVersionTrimToMinor("1.31.8")
	require.NoError(t, err)
	assert.Equal(t, "1.31", minor)

	next, err := kubeVersionTrimToMinorIncremented("1.31.8")
	require.NoError(t, err)
	assert.Equal(t, "1.32", next)

	_, err = kubeVersionTrimToMinor("x")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "major part is not an integer number")
	_, err = kubeVersionTrimToMinorIncremented("1")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "expected to have major and minor version parts")

	trimmed, err := kubeVersionTrimToMinor("1.x.8")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "minor part is not an integer number")
	assert.Equal(t, "", trimmed)
}

func TestLatestKubePatchVersions(t *testing.T) {
	fake := newFakeAPI(t)
	client := fake.client()

	fake.fail(routeKubeVersions, 500)
	_, err := latestKubePatchVersions(context.Background(), client)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "error getting kube versions")

	fake.fail(routeKubeVersions, 0)
	fake.seedKubeVersions(
		mksclient.KubeVersionInfo{Version: new("1.31.8")},
		mksclient.KubeVersionInfo{Version: new("1.31.5")},
		mksclient.KubeVersionInfo{Version: new("1.32.3")},
	)
	got, err := latestKubePatchVersions(context.Background(), client)
	require.NoError(t, err)
	assert.Equal(t, map[string]string{"1.31": "1.31.8", "1.32": "1.32.3"}, got)
}
