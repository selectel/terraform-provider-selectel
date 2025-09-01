package servers

import (
	"net"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNetworks_FilterByTelematicsTypeHosting(t *testing.T) {
	networks := Networks{
		&Network{UUID: "1", TelematicType: "HOSTING"},
		&Network{UUID: "2", TelematicType: "INET"},
		&Network{UUID: "3", TelematicType: "HOSTING"},
		&Network{UUID: "4", TelematicType: "INET"},
	}

	result := networks.FilterByTelematicsTypeHosting()

	require.Len(t, result, 2)
	require.Equal(t, "1", result[0].UUID)
	require.Equal(t, "3", result[1].UUID)
}

func TestSubnet_ReservedVRRPIPAsStrings(t *testing.T) {
	subnet := &Subnet{
		ReservedVRRPIP: []net.IP{
			net.ParseIP("192.168.1.1"),
			net.ParseIP("192.168.1.2"),
		},
	}

	result := subnet.ReservedVRRPIPAsStrings()

	require.Len(t, result, 2)
	require.Equal(t, "192.168.1.1", result[0])
	require.Equal(t, "192.168.1.2", result[1])
}

func TestSubnet_GetFreeIP(t *testing.T) {
	tests := []struct {
		name        string
		subnet      Subnet
		reservedIPs ReservedIPs
		isLocal     bool
		want        string
		wantErr     bool
	}{
		{
			name: "FreeIPAvailable",
			subnet: Subnet{
				NetworkUUID:    "net1",
				Subnet:         "192.168.1.0/29",
				Gateway:        net.ParseIP("192.168.1.1"),
				Broadcast:      net.ParseIP("192.168.1.7"),
				ReservedVRRPIP: []net.IP{net.ParseIP("192.168.1.2")},
			},
			reservedIPs: ReservedIPs{
				&ReservedIP{IP: net.ParseIP("192.168.1.3"), NetworkUUID: "net1"},
			},
			isLocal: false,
			want:    "192.168.1.4",
			wantErr: false,
		},
		{
			name: "NoFreeIP",
			subnet: Subnet{
				Subnet:         "192.168.1.0/30",
				Gateway:        net.ParseIP("192.168.1.1"),
				Broadcast:      net.ParseIP("192.168.1.3"),
				ReservedVRRPIP: []net.IP{net.ParseIP("192.168.1.2")},
			},
			reservedIPs: ReservedIPs{},
			isLocal:     false,
			want:        "",
			wantErr:     true,
		},
		{
			name: "InvalidSubnet",
			subnet: Subnet{
				Subnet: "invalid-subnet",
			},
			reservedIPs: ReservedIPs{},
			isLocal:     false,
			want:        "",
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.subnet.GetFreeIP(tt.reservedIPs, tt.isLocal)

			if tt.wantErr {
				require.Error(t, err)
				require.Nil(t, got)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.want, got.String())
			}
		})
	}
}

func TestSubnet_IsIncluding(t *testing.T) {
	tests := []struct {
		name    string
		subnet  Subnet
		ip      string
		want    bool
		wantErr bool
	}{
		{
			name:    "IPIncluded",
			subnet:  Subnet{Subnet: "192.168.1.0/24"},
			ip:      "192.168.1.100",
			want:    true,
			wantErr: false,
		},
		{
			name:    "IPNotIncluded",
			subnet:  Subnet{Subnet: "192.168.1.0/24"},
			ip:      "192.168.2.100",
			want:    false,
			wantErr: false,
		},
		{
			name:    "InvalidIP",
			subnet:  Subnet{Subnet: "192.168.1.0/24"},
			ip:      "invalid-ip",
			want:    false,
			wantErr: true,
		},
		{
			name:    "InvalidSubnet",
			subnet:  Subnet{Subnet: "invalid-subnet"},
			ip:      "192.168.1.100",
			want:    false,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.subnet.IsIncluding(tt.ip)

			if tt.wantErr {
				require.Error(t, err)
				require.False(t, got)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.want, got)
			}
		})
	}
}

func TestSubnets_FindBySubnet(t *testing.T) {
	subnets := Subnets{
		&Subnet{Subnet: "192.168.1.0/24"},
		&Subnet{Subnet: "192.168.2.0/24"},
	}

	tests := []struct {
		name   string
		subnet string
		want   *Subnet
	}{
		{
			name:   "SubnetFound",
			subnet: "192.168.2.0/24",
			want:   subnets[1],
		},
		{
			name:   "SubnetNotFound",
			subnet: "192.168.3.0/24",
			want:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := subnets.FindBySubnet(tt.subnet)

			require.Equal(t, tt.want, got)
		})
	}
}

func TestIPConversion(t *testing.T) {
	ip := net.ParseIP("192.168.1.1")
	uintIP := ipToUint32(ip)
	convertedIP := uint32ToIP(uintIP)

	require.Equal(t, ip.String(), convertedIP.String())
}
