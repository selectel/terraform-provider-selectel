package selvpc

import (
	"errors"
	"fmt"

	"github.com/hashicorp/terraform/helper/hashcode"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/selectel/go-selvpcclient/selvpcclient/resell/v2/crossregionsubnets"
	"github.com/selectel/go-selvpcclient/selvpcclient/resell/v2/subnets"
)

func expandResellV2Regions(rawRegions *schema.Set) []string {
	regions := rawRegions.List()

	expandedRegions := make([]string, len(regions))

	for i, region := range regions {
		expandedRegions[i] = region.(string)
	}

	return expandedRegions
}

// expandResellV2CrossRegionOpts converts the provided rawRegions structure to
// the slice of crossregionsubnets.CrossRegionOpt.
func expandResellV2CrossRegionOpts(rawRegions *schema.Set) ([]crossregionsubnets.CrossRegionOpt, error) {
	rawRegionsLen := rawRegions.Len()
	if rawRegionsLen == 0 {
		return nil, errors.New("got empty regions")
	}

	// Pre-allocate memory for expandedCrossRegionOpts slice.
	expandedCrossRegionOpts := make([]crossregionsubnets.CrossRegionOpt, rawRegionsLen)

	// Iterate over each value in rawRegions and add it to expandedCrossRegionOpts slice.
	for i, rawRegion := range rawRegions.List() {
		var region string
		mapRegion := rawRegion.(map[string]interface{})
		if value, ok := mapRegion["region"]; ok {
			region = value.(string)
		}
		expandedCrossRegionOpts[i] = crossregionsubnets.CrossRegionOpt{
			Region: region,
		}
	}

	return expandedCrossRegionOpts, nil
}

// regionsMapsFromSubnetsStructs converts the provided subnets.Subnet to
// the slice of maps with associated regions correspondingly to the resource's schema.
func regionsMapsFromSubnetsStructs(subnetsStructs []subnets.Subnet) []map[string]interface{} {
	associatedRegions := make([]map[string]interface{}, len(subnetsStructs))

	for i, subnet := range subnetsStructs {
		associatedRegions[i] = make(map[string]interface{})

		if subnet.Region != "" {
			associatedRegions[i]["region"] = subnet.Region
		}
	}

	return associatedRegions
}

// hashRegions is a hash function to use with the "regions" set.
func hashRegions(v interface{}) int {
	m := v.(map[string]interface{})
	return hashcode.String(fmt.Sprintf("%s-", m["region"].(string)))
}
