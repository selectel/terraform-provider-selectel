package schemaconfig

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func IsSet(d *schema.ResourceData, key string) bool {
	raw := d.GetRawConfig()
	if raw.IsNull() || !raw.IsKnown() || !raw.Type().IsObjectType() || !raw.Type().HasAttribute(key) {
		return false
	}

	return !raw.GetAttr(key).IsNull()
}
