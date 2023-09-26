package selectel

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/selectel/go-selvpcclient/v3/selvpcclient"
	"github.com/terraform-providers/terraform-provider-selectel/selectel/internal/hashcode"
)

func expandVPCV2Regions(rawRegions *schema.Set) []string {
	regions := rawRegions.List()

	expandedRegions := make([]string, len(regions))

	for i, region := range regions {
		expandedRegions[i] = region.(string)
	}

	return expandedRegions
}

// hashRegions is a hash function to use with the "regions" set.
func hashRegions(v interface{}) int {
	m := v.(map[string]interface{})
	return hashcode.String(fmt.Sprintf("%s-", m["region"].(string)))
}

func validateRegion(selvpcClient *selvpcclient.Client, serviceType string, region string) error {
	endpoints, err := selvpcClient.Catalog.GetEndpoints(serviceType)
	if err != nil {
		return fmt.Errorf("can't get endpoints for %s to validate region: %w", serviceType, err)
	}

	endpointRegions := make([]string, 0)

	for _, endpoint := range endpoints {
		if endpoint.Region == region {
			return nil
		}
		endpointRegions = append(endpointRegions, endpoint.RegionID)
	}

	return fmt.Errorf("region value must contain one of the values: %+q", endpointRegions)
}
