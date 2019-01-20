package selvpc

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform/helper/hashcode"
	"github.com/selectel/go-selvpcclient/selvpcclient"
	"github.com/selectel/go-selvpcclient/selvpcclient/resell/v2/subnets"
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
		associatedSubnets[i] = make(map[string]interface{})

		if subnet.NetworkID != "" {
			associatedSubnets[i]["network_id"] = subnet.NetworkID
		}
		if subnet.SubnetID != "" {
			associatedSubnets[i]["subnet_id"] = subnet.SubnetID
		}
		if subnet.Region != "" {
			associatedSubnets[i]["region"] = subnet.Region
		}
		if subnet.CIDR != "" {
			associatedSubnets[i]["cidr"] = subnet.CIDR
		}
		if subnet.VLANID != 0 {
			associatedSubnets[i]["vlan_id"] = subnet.VLANID
		}
		if subnet.ProjectID != "" {
			associatedSubnets[i]["project_id"] = subnet.ProjectID
		}
		if subnet.VTEPIPAddress != "" {
			associatedSubnets[i]["vtep_ip_address"] = subnet.VTEPIPAddress
		}
	}

	return associatedSubnets
}

// hashSubnets is a hash function to use with the "subnet" set.
func hashSubnets(v interface{}) int {
	m := v.(map[string]interface{})
	return hashcode.String(fmt.Sprintf("%s-", m["network_id"].(string)))
}
