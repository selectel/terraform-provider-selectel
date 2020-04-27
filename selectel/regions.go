package selectel

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/helper/hashcode"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

const (
	ru1Region = "ru-1"
	ru2Region = "ru-2"
	ru3Region = "ru-3"
	ru7Region = "ru-7"
	ru8Region = "ru-8"
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

func validateRegion(region string) error {
	valid := map[string]struct{}{
		ru1Region: {},
		ru2Region: {},
		ru3Region: {},
		ru7Region: {},
		ru8Region: {},
	}

	_, isValid := valid[region]
	if !isValid {
		return fmt.Errorf("region is invalid: %s", region)
	}

	return nil
}
