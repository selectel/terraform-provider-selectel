package blockstorage

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/terraform-providers/terraform-provider-selectel/selectel/internal/schemadoc"
)

// dataConnection provides the identical configuration boundary to the three
// lookups; each lookup keeps its own schema, model, selection and state write.
type dataConnection struct{ connection }

func (d *dataConnection) Configure(
	_ context.Context,
	req datasource.ConfigureRequest,
	resp *datasource.ConfigureResponse,
) {
	resp.Diagnostics.Append(d.configure(req.ProviderData)...)
}

// filterString is the lookup filter of the Block Storage data sources: an exact
// match that, when omitted, returns the value of the found object.
func filterString(validators ...validator.String) schema.StringAttribute {
	return schema.StringAttribute{
		Optional: true, Computed: true,
		Validators: append([]validator.String{schemadoc.LengthAtLeast(1)}, validators...),
	}
}

// matchesFilter applies a filterString to a value of a found object; an omitted
// filter matches any value.
func matchesFilter(filter types.String, value string) bool {
	return filter.IsNull() || filter.ValueString() == value
}
