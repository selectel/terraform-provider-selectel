package selectel

import (
	"fmt"
	"net"
	"strconv"
	"strings"

	"github.com/selectel/go-selvpcclient/v4/selvpcclient"
	"github.com/selectel/go-selvpcclient/v4/selvpcclient/resell/v2/subnets"
)

func getPrefixLengthFromCIDR(cidr string) (int, error) {
	cidrParts := strings.Split(cidr, "/")
	if len(cidrParts) != 2 {
		return 0, fmt.Errorf("got invalid CIDR: %s", cidr)
	}

	prefixLenght, err := strconv.Atoi(cidrParts[1])
	if err != nil {
		return 0, fmt.Errorf("error reading prefix length from '%s': %s", cidrParts[1], err)
	}

	return prefixLenght, nil
}

func getIPVersionFromPrefixLength(prefixLength int) string {
	// Any subnet with prefix length larger than 32 is a IPv6 protocol subnet
	// and Selectel doesn't provide any IPv6 subnets with smaller prefix lengths.
	if prefixLength > 32 {
		return string(selvpcclient.IPv6)
	}

	return string(selvpcclient.IPv4)
}

// subnetsMapsFromStructs converts the provided subnets.Subnet to
// the slice of maps correspondingly to the resource's schema.
func subnetsMapsFromStructs(subnetsStructs []subnets.Subnet) []map[string]interface{} {
	associatedSubnets := make([]map[string]interface{}, len(subnetsStructs))

	for i, subnet := range subnetsStructs {
		associatedSubnets[i] = map[string]interface{}{
			"network_id":      subnet.NetworkID,
			"subnet_id":       subnet.SubnetID,
			"region":          subnet.Region,
			"cidr":            subnet.CIDR,
			"vlan_id":         subnet.VLANID,
			"project_id":      subnet.ProjectID,
			"vtep_ip_address": subnet.VTEPIPAddress,
		}
	}

	return associatedSubnets
}

// validatePrivateSubnet checks if the provided CIDR belongs to private IP ranges:
// 10.0.0.0/8, 172.16.0.0/12, or 192.168.0.0/16.
func validatePrivateSubnet(cidr string) error {
	_, network, err := net.ParseCIDR(cidr)
	if err != nil {
		return fmt.Errorf("invalid CIDR format: %s", err)
	}

	// Define private IP ranges
	privateRanges := []struct {
		Name string
		Net  *net.IPNet
	}{
		{
			Name: "10.0.0.0/8",
			Net:  &net.IPNet{IP: net.ParseIP("10.0.0.0"), Mask: net.CIDRMask(8, 32)},
		},
		{
			Name: "172.16.0.0/12",
			Net:  &net.IPNet{IP: net.ParseIP("172.16.0.0"), Mask: net.CIDRMask(12, 32)},
		},
		{
			Name: "192.168.0.0/16",
			Net:  &net.IPNet{IP: net.ParseIP("192.168.0.0"), Mask: net.CIDRMask(16, 32)},
		},
	}

	// Check if the network is within any of the private ranges
	for _, privateRange := range privateRanges {
		if privateRange.Net.Contains(network.IP) {
			return nil
		}
	}

	return fmt.Errorf("subnet %s does not belong to private IP ranges: 10.0.0.0/8, 172.16.0.0/12, or 192.168.0.0/16", cidr)
}

// IsPrivateSubnet validates if the provided CIDR belongs to private IP ranges.
func IsPrivateSubnet(cidr string) bool {
	return validatePrivateSubnet(cidr) == nil
}
