package selectel

import (
	"github.com/hashicorp/terraform/helper/schema"
)

func expandVPCV2Regions(rawRegions *schema.Set) []string {
	regions := rawRegions.List()

	expandedRegions := make([]string, len(regions))

	for i, region := range regions {
		expandedRegions[i] = region.(string)
	}

	return expandedRegions
}
