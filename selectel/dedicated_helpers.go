package selectel

import (
	"fmt"
	"net"
	"net/netip"
)

// validatePrivateSubnet checks if the provided CIDR belongs to private IP ranges:
// 10.0.0.0/8, 172.16.0.0/12, or 192.168.0.0/16.
func validatePrivateSubnet(cidr string) error {
	_, network, err := net.ParseCIDR(cidr)
	if err != nil {
		return fmt.Errorf("invalid CIDR format: %s", err)
	}

	ipAddr, ok := netip.AddrFromSlice(network.IP)
	if !ok {
		return fmt.Errorf("invalid IP address: %s", network.IP)
	}

	if ipAddr.IsPrivate() {
		return nil
	}

	return fmt.Errorf("subnet %s does not belong to private IP ranges: 10.0.0.0/8, 172.16.0.0/12, or 192.168.0.0/16", cidr)
}
