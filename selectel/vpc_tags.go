package selectel

import (
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/terraform-providers/terraform-provider-selectel/selectel/internal/convert"
)

func flattenVPCTagsV2(d *schema.ResourceData, tags []string) {
	if err := d.Set("all_tags", tags); err != nil {
		log.Print(errSettingComplexAttr("all_tags", err))
	}

	allTags := d.Get("all_tags").(*schema.Set)
	desiredTags := d.Get("tags").(*schema.Set)
	actualTags := allTags.Intersection(desiredTags)

	if !actualTags.Equal(desiredTags) {
		if err := d.Set("tags", convert.ToStringSlice(actualTags.List())); err != nil {
			log.Print(errSettingComplexAttr("tags", err))
		}
	}
}

func expandVPCUpdateTagsV2(d *schema.ResourceData) []string {
	oldTagsRaw, newTagsRaw := d.GetChange("tags")

	return mergeVPCTagsV2(
		d.Get("all_tags").(*schema.Set),
		oldTagsRaw.(*schema.Set),
		newTagsRaw.(*schema.Set),
	)
}

func mergeVPCTagsV2(allTags, oldTags, newTags *schema.Set) []string {
	return convert.ToStringSlice(allTags.Difference(oldTags).Union(newTags).List())
}
