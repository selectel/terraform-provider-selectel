package mks

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClusterKubeVersionKeepsState(t *testing.T) {
	for _, tt := range []struct {
		name   string
		prior  string
		actual string
		want   string
	}{
		{name: "no prior, after import", prior: "", actual: "1.30.5", want: "1.30.5"},
		{name: "same version", prior: "1.30.3", actual: "1.30.3", want: "1.30.3"},
		{name: "newer patch keeps the configured one", prior: "1.30.3", actual: "1.30.5", want: "1.30.3"},
		{name: "older patch is drift", prior: "1.30.5", actual: "1.30.3", want: "1.30.3"},
		{name: "newer minor is drift", prior: "1.30.5", actual: "1.31.1", want: "1.31.1"},
		{name: "newer major is drift", prior: "1.30.5", actual: "2.0.1", want: "2.0.1"},
		{name: "older minor is drift", prior: "1.31.1", actual: "1.30.5", want: "1.30.5"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, kubeVersionKeepsState(tt.prior, tt.actual))
		})
	}
}

func TestClusterValidateKubeVersionChange(t *testing.T) {
	for _, tt := range []struct {
		name    string
		current string
		desired string
		wantErr string
	}{
		{name: "skipped minor", current: "1.31.8", desired: "1.33.1", wantErr: "must be upgraded one minor version at a time"},
		{name: "major change", current: "1.31.8", desired: "2.31.1", wantErr: "can't be upgraded to version"},
		{name: "downgrade", current: "1.32.8", desired: "1.31.5", wantErr: "can't be downgraded to version"},
		{name: "patch downgrade", current: "1.31.8", desired: "1.31.5", wantErr: "can't be downgraded to version"},
		{name: "one minor up", current: "1.31.8", desired: "1.32.3"},
		{name: "patch up", current: "1.31.8", desired: "1.31.9"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			err := validateKubeVersionChange(tt.current, tt.desired)
			if tt.wantErr == "" {
				assert.NoError(t, err)
			} else {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr)
			}
		})
	}
}
