package servers

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestOperatingSystems_FindOneByNameAndVersion(t *testing.T) {
	ops := OperatingSystems{
		&OperatingSystem{UUID: "1", Name: "Ubuntu", VersionValue: "20.04"},
		&OperatingSystem{UUID: "2", Name: "CentOS", VersionValue: "7"},
	}

	tests := []struct {
		name       string
		argName    string
		argVersion string
		want       *OperatingSystem
	}{
		{
			name:       "FoundUbuntu",
			argName:    "Ubuntu",
			argVersion: "20.04",
			want:       ops[0],
		},
		{
			name:       "NotFound",
			argName:    "Debian",
			argVersion: "1",
			want:       nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ops.FindOneByNameAndVersion(tt.argName, tt.argVersion)

			require.Equal(t, tt.want, got)
		})
	}
}

func TestOperatingSystems_FindOneByID(t *testing.T) {
	ops := OperatingSystems{
		&OperatingSystem{UUID: "1", Name: "Ubuntu"},
		&OperatingSystem{UUID: "2", Name: "CentOS"},
	}

	tests := []struct {
		name string
		arg  string
		want *OperatingSystem
	}{
		{
			name: "FoundByID_1",
			arg:  "1",
			want: &OperatingSystem{UUID: "1", Name: "Ubuntu"},
		},
		{
			name: "NotFound",
			arg:  "3",
			want: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ops.FindOneByID(tt.arg)

			require.Equal(t, tt.want, got)
		})
	}
}

func TestOperatingSystems_FindOneByArchAndVersionAndOs(t *testing.T) {
	ops := OperatingSystems{
		&OperatingSystem{UUID: "1", Arch: "x86_64", VersionValue: "20.04", OSValue: "ubuntu"},
		&OperatingSystem{UUID: "2", Arch: "arm64", VersionValue: "7", OSValue: "centos"},
	}

	tests := []struct {
		name       string
		argArch    string
		argVersion string
		argOSValue string
		want       *OperatingSystem
	}{
		{
			name:       "FoundUbuntu",
			argArch:    "x86_64",
			argVersion: "20.04",
			argOSValue: "ubuntu",
			want:       ops[0],
		},
		{
			name:       "NotFound",
			argArch:    "x86_64",
			argVersion: "18.04",
			argOSValue: "ubuntu",
			want:       nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ops.FindOneByArchAndVersionAndOs(tt.argArch, tt.argVersion, tt.argOSValue)

			require.Equal(t, tt.want, got)
		})
	}
}

func TestOperatingSystem_IsPrivateNetworkAvailable(t *testing.T) {
	tests := []struct {
		name string
		os   OperatingSystem
		want bool
	}{
		{
			name: "PrivateNetworkAvailable",
			os:   OperatingSystem{OSValue: "linux", TemplateVersion: "v2"},
			want: true,
		},
		{
			name: "PrivateNetworkUnavailable_Windows",
			os:   OperatingSystem{OSValue: "windows", TemplateVersion: "v2"},
			want: false,
		},
		{
			name: "PrivateNetworkUnavailable_OldTemplate",
			os:   OperatingSystem{OSValue: "linux", TemplateVersion: "v1"},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.os.IsPrivateNetworkAvailable()

			require.Equal(t, tt.want, got)
		})
	}
}
