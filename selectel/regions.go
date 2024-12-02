package selectel

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/selectel/go-selvpcclient/v4/selvpcclient"
)

func expandVPCV2Regions(rawRegions *schema.Set) []string {
	regions := rawRegions.List()

	expandedRegions := make([]string, len(regions))

	for i, region := range regions {
		expandedRegions[i] = region.(string)
	}

	return expandedRegions
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
