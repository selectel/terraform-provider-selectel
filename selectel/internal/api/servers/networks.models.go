package servers

import (
	"encoding/binary"
	"errors"
	"fmt"
	"net"
	"slices"
)

type (
	Network struct {
		UUID          string `json:"uuid"`
		TelematicType string `json:"telematics_type"`
	}

	Networks []*Network
)

func (n Networks) FilterByTelematicsTypeHosting() Networks {
	result := make(Networks, 0, len(n))

	for _, network := range n {
		if network.TelematicType == "HOSTING" {
			result = append(result, network)
		}
	}

	return result
}

type (
	Subnet struct {
		UUID           string   `json:"uuid"`
		NetworkUUID    string   `json:"network_uuid"`
		Subnet         string   `json:"subnet"`
		Gateway        net.IP   `json:"gateway"`
		Broadcast      net.IP   `json:"broadcast"`
		ReservedVRRPIP []net.IP `json:"reserved_vrrp_ip"`
		Free           int      `json:"free"`
	}
)

func (s *Subnet) ReservedVRRPIPAsStrings() []string {
	res := make([]string, 0, len(s.ReservedVRRPIP))
	for _, ip := range s.ReservedVRRPIP {
		res = append(res, ip.String())
	}

	return res
}

func (s *Subnet) GetFreeIP(reservedIPs ReservedIPs, isLocal bool) (net.IP, error) {
	baseIP, ipNet, err := net.ParseCIDR(s.Subnet)
	if err != nil {
		return nil, fmt.Errorf("error parsing subnet %s: %s", s.Subnet, err)
	}

	base := ipToUint32(baseIP.Mask(ipNet.Mask))

	ones, bits := ipNet.Mask.Size()
	total := uint32(1) << uint32(bits-ones) //nolint:gosec
	last := base + total

	if isLocal { // skip hidden gateway ip
		base++
	}

	for cur := base + 1; cur < last; cur++ {
		currentIP := uint32ToIP(cur)
		switch {
		case currentIP.Equal(s.Gateway):
		case currentIP.Equal(s.Broadcast):
		case slices.ContainsFunc(s.ReservedVRRPIP, func(ip net.IP) bool { // is reserved VRRP
			return currentIP.Equal(ip)
		}):
		case slices.ContainsFunc(reservedIPs, func(ip *ReservedIP) bool { // is reserved
			return s.NetworkUUID == ip.NetworkUUID && currentIP.Equal(ip.IP)
		}):
		default:
			return currentIP, nil
		}
	}

	return nil, errors.New("no free IP found")
}

func (s *Subnet) IsIncluding(ip string) (bool, error) {
	baseIP, ipNet, err := net.ParseCIDR(s.Subnet)
	if err != nil {
		return false, fmt.Errorf("error parsing subnet %s: %s", s.Subnet, err)
	}

	base := ipToUint32(baseIP.Mask(ipNet.Mask))

	ones, bits := ipNet.Mask.Size()
	total := uint32(1) << uint32(bits-ones) //nolint:gosec
	last := base + total - 1

	ipAddr := net.ParseIP(ip)
	if ipAddr == nil {
		return false, fmt.Errorf("invalid IP address: %s", ip)
	}

	currentIP := ipToUint32(ipAddr)

	return currentIP >= base && currentIP <= last, nil
}

type Subnets []*Subnet

func (s Subnets) FindBySubnet(subnet string) *Subnet {
	for _, sn := range s {
		if sn.Subnet == subnet {
			return sn
		}
	}

	return nil
}

// ipToUint32 converts a 4-byte net.IP to uint32.
func ipToUint32(ip net.IP) uint32 {
	return binary.BigEndian.Uint32(ip.To4())
}

// uint32ToIP converts uint32 back to net.IP (always 4 bytes).
func uint32ToIP(n uint32) net.IP {
	var b [4]byte
	binary.BigEndian.PutUint32(b[:], n)
	return b[:]
}

type (
	ReservedIP struct {
		IP          net.IP `json:"ip"`
		NetworkUUID string `json:"network_uuid"`
	}

	ReservedIPs []*ReservedIP
)
