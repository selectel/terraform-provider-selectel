package mks

import (
	"context"
	"testing"

	"github.com/selectel/mks-go/v2/pkg/mksclient"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestKubeVersionParts(t *testing.T) {
	for _, tt := range []struct {
		version  string
		parse    func(string) (int, error)
		want     int
		wantPart string
	}{
		{version: "v1.31.8", parse: kubeVersionToMajor, want: 1},
		{version: "1.31.8", parse: kubeVersionToMinor, want: 31},
		{version: "x.31.8", parse: kubeVersionToMajor, wantPart: "major part is not an integer number"},
		{version: "1.b.8", parse: kubeVersionToMinor, wantPart: "minor part is not an integer number"},
		{version: "1.31", parse: kubeVersionToPatch, wantPart: "expected to have major, minor and patch version parts"},
		{version: "1.31.x", parse: kubeVersionToPatch, wantPart: "patch part is not an integer number"},
	} {
		t.Run(tt.version, func(t *testing.T) {
			got, err := tt.parse(tt.version)
			if tt.wantPart == "" {
				require.NoError(t, err)
				assert.Equal(t, tt.want, got)

				return
			}
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.wantPart)
		})
	}

	_, err := kubeVersionToMinor("1.31")
	assert.NoError(t, err)
	_, err = kubeVersionToMinor("1")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "expected to have major and minor version parts")
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

	_, err = kubeVersionTrimToMinorIncremented("1")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "expected to have major and minor version parts")

	_, err = kubeVersionTrimToMinor("x")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "major part is not an integer number")
	_, err = kubeVersionTrimToMinor("1.x.8")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "minor part is not an integer number")
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
