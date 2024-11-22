package selectel

import (
	"bytes"
	"errors"
	"fmt"
	"net/url"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/selectel/go-selvpcclient/v4/selvpcclient/quotamanager/quotas"
	"github.com/selectel/go-selvpcclient/v4/selvpcclient/resell/v2/projects"
	resellQuotas "github.com/selectel/go-selvpcclient/v4/selvpcclient/resell/v2/quotas"
	"github.com/terraform-providers/terraform-provider-selectel/selectel/internal/hashcode"
)

// resourceVPCProjectV2QuotasOptsFromSet converts the provided quotaSet to
// the slice of quotas.QuotaOpts. It then can be used to make requests with
// quotas data.
func resourceVPCProjectV2QuotasOptsFromSet(
	quotaSet *schema.Set,
) (map[string]quotas.UpdateProjectQuotasOpts, error) {
	quotaSetLen := quotaSet.Len()
	if quotaSetLen == 0 {
		return nil, errors.New("got empty quotas")
	}

	bufferQuotas := map[string]map[string][]quotas.ResourceQuotaOpts{}

	// Iterate over each billing resource quotas map.
	for _, resourceQuotasData := range quotaSet.List() {
		var resourceNameRaw, resourceQuotasRaw interface{}
		var ok bool

		// Cast type of the current resource quotas map and check provided values.
		resourceQuotasMap := resourceQuotasData.(map[string]interface{})
		if resourceNameRaw, ok = resourceQuotasMap["resource_name"]; !ok {
			return nil, errors.New("resource_name value isn't provided")
		}
		if resourceQuotasRaw, ok = resourceQuotasMap["resource_quotas"]; !ok {
			return nil, errors.New("resource_quotas value isn't provided")
		}

		// Cast types of provided values and pre-populate slice of []quotas.ResourceQuotaOpts
		// in memory as we already know it's length.
		resourceName := resourceNameRaw.(string)
		resourceQuotasEntities := resourceQuotasRaw.(*schema.Set)

		// Populate bufferedMap with data from a single resourceQuotasMap's region zone and value.
		for _, resourceQuotasEntityRaw := range resourceQuotasEntities.List() {
			var (
				resourceQuotasEntityRegion string
				resourceQuotasEntityZone   string
				resourceQuotasEntityValue  int
			)
			resourceQuotasEntity := resourceQuotasEntityRaw.(map[string]interface{})
			if region, ok := resourceQuotasEntity["region"]; ok {
				resourceQuotasEntityRegion = region.(string)
			}
			if zone, ok := resourceQuotasEntity["zone"]; ok {
				resourceQuotasEntityZone = zone.(string)
			}
			if value, ok := resourceQuotasEntity["value"]; ok {
				resourceQuotasEntityValue = value.(int)
			}
			// Populate single entity of billing resource data with the region,
			// zone and value information.

			if _, ok := bufferQuotas[resourceQuotasEntityRegion]; !ok {
				bufferQuotas[resourceQuotasEntityRegion] = map[string][]quotas.ResourceQuotaOpts{}
			}

			bufferQuotas[resourceQuotasEntityRegion][resourceName] = append(
				bufferQuotas[resourceQuotasEntityRegion][resourceName], quotas.ResourceQuotaOpts{
					Zone:  &resourceQuotasEntityZone,
					Value: &resourceQuotasEntityValue,
				})
		}
	}

	quotasOpts := map[string]quotas.UpdateProjectQuotasOpts{}
	for region, bufferQuotaOpts := range bufferQuotas {
		if _, ok := quotasOpts[region]; !ok {
			quotasOpts[region] = quotas.UpdateProjectQuotasOpts{}
		}

		for resourceName, resourceQuotasOpts := range bufferQuotaOpts {
			quotaOpts := quotas.QuotaOpts{Name: resourceName, ResourceQuotasOpts: resourceQuotasOpts}
			regionalOpts := quotasOpts[region]
			regionalOpts.QuotasOpts = append(regionalOpts.QuotasOpts, quotaOpts)
			quotasOpts[region] = regionalOpts
		}
	}

	return quotasOpts, nil
}

// resourceVPCProjectV2QuotasToSet converts the provided quotas.Quota slice
// to a nested complex set structure correspondingly to the resource's schema.
func resourceVPCProjectV2QuotasToSet(quotasStructures []resellQuotas.Quota) *schema.Set {
	quotaSet := &schema.Set{
		F: quotasHashSetFunc(),
	}

	// Iterate over each billing resource quota.
	for _, quota := range quotasStructures {
		// For each billing resource populate corresponding resourceQuotasSet that
		// contain quota data (region, zone and value).
		resourceQuotasSet := &schema.Set{
			F: resourceQuotasHashSetFunc(),
		}
		for _, resourceQuotasEntity := range quota.ResourceQuotasEntities {
			resourceQuotasSet.Add(map[string]interface{}{
				"region": resourceQuotasEntity.Region,
				"zone":   resourceQuotasEntity.Zone,
				"value":  resourceQuotasEntity.Value,
				"used":   resourceQuotasEntity.Used,
			})
		}

		// Populate single quota element.
		quotaSet.Add(map[string]interface{}{
			"resource_name":   quota.Name,
			"resource_quotas": resourceQuotasSet,
		})
	}

	return quotaSet
}

// resourceProjectV2UpdateThemeOptsFromMap converts the provided themeOptsMap to
// the *project.ThemeUpdateOpts.
// It can be used to make requests with project theme parameters.
func resourceProjectV2UpdateThemeOptsFromMap(themeOptsMap map[string]interface{}) *projects.ThemeUpdateOpts {
	themeUpdateOpts := &projects.ThemeUpdateOpts{}

	var themeColor, themeLogo string
	if color, ok := themeOptsMap["color"]; ok {
		themeColor = color.(string)
	}
	if logo, ok := themeOptsMap["logo"]; ok {
		themeLogo = logo.(string)
	}
	themeUpdateOpts.Color = &themeColor
	themeUpdateOpts.Logo = &themeLogo

	return themeUpdateOpts
}

// resourceVPCProjectV2URLWithoutSchema strips the scheme part from project URL.
func resourceVPCProjectV2URLWithoutSchema(customURL string) (string, error) {
	var customURLWithoutSchema string

	if customURL != "" {
		u, err := url.Parse(customURL)
		if err != nil {
			return "", err
		}
		customURLWithoutSchema = u.Hostname()
	}

	return customURLWithoutSchema, nil
}

// quotasSchema returns *schema.Resource from the "quotas" attribute.
func quotasSchema() *schema.Resource {
	return resourceVPCProjectV2().Schema["quotas"].Elem.(*schema.Resource)
}

// quotasSchema returns *schema.Resource from the "resource_quotas" attribute.
func resourceQuotasSchema() *schema.Resource {
	return quotasSchema().Schema["resource_quotas"].Elem.(*schema.Resource)
}

// quotasHashSetFunc returns schema.SchemaSetFunc that can be used to
// create a new schema.Set for the "quotas" or "all_quotas" attributes.
func quotasHashSetFunc() schema.SchemaSetFunc {
	return schema.HashResource(quotasSchema())
}

// resourceQuotasHashSetFunc returns schema.SchemaSetFunc that can be used to
// create a new schema.Set for the "resource_quotas" attribute.
func resourceQuotasHashSetFunc() schema.SchemaSetFunc {
	return schema.HashResource(resourceQuotasSchema())
}

// hashResourceQuotas is a hash function to use with the "resource_quotas" set.
func hashResourceQuotas(v interface{}) int {
	var buf bytes.Buffer
	m := v.(map[string]interface{})
	if m["region"] != "" {
		buf.WriteString(fmt.Sprintf("%s-", m["region"].(string)))
	}
	if m["zone"] != "" {
		buf.WriteString(fmt.Sprintf("%s-", m["zone"].(string)))
	}

	return hashcode.String(buf.String())
}

func flattenVPCProjectV2Theme(theme projects.Theme) map[string]string {
	if theme.Logo == "" && theme.Color == "" {
		return nil
	}

	m := make(map[string]string)
	if theme.Color != "" {
		m["color"] = theme.Color
	}
	if theme.Logo != "" {
		m["logo"] = theme.Logo
	}

	return m
}
