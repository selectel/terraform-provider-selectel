package selectel

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/helper/hashcode"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
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
