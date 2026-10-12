package schemadoc_test

import (
	"context"
	"maps"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	dsschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/defaults"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/terraform-providers/terraform-provider-selectel/selectel/internal/schemadoc"
)

func TestStringContract(t *testing.T) {
	ctx := context.Background()
	docs := schemadoc.Resource{Name: "network volume"}
	attribute := docs.String(schema.StringAttribute{
		Optional: true, Computed: true, Sensitive: true,
		Default:    schemadoc.StaticString("ab"),
		Validators: []validator.String{schemadoc.LengthAtLeast(1), schemadoc.LengthAtMost(4)},
		PlanModifiers: []planmodifier.String{
			schemadoc.RequiresReplace(), stringplanmodifier.UseStateForUnknown(),
		},
	}, schemadoc.Field{
		Definition: "Name of the network volume.",
		Behavior:   "Surrounding whitespace is rejected.",
		Concept:    schemadoc.Link{Title: "names", URL: "https://example.com/names"},
	})
	assert.Equal(t, "Name of the network volume. "+
		"The minimum length is `1` character. The maximum length is `4` characters. "+
		"Surrounding whitespace is rejected. The default value is `ab`. "+
		"Learn more about [names](https://example.com/names). Changing this creates a new network volume.",
		attribute.MarkdownDescription)
	assert.True(t, attribute.Optional && attribute.Computed && attribute.Sensitive)
	for _, tc := range []struct {
		value string
		valid bool
	}{{"", false}, {"abcd", true}, {"abcde", false}, {"яяяя", true}, {"яяяяя", false}} {
		var response validator.StringResponse
		for _, v := range attribute.Validators {
			v.ValidateString(ctx, validator.StringRequest{ConfigValue: types.StringValue(tc.value)}, &response)
		}
		assert.Equal(t, tc.valid, !response.Diagnostics.HasError(), "value %q", tc.value)
	}
	var defaultResponse defaults.StringResponse
	attribute.Default.DefaultString(ctx, defaults.StringRequest{}, &defaultResponse)
	require.False(t, defaultResponse.Diagnostics.HasError())
	assert.Equal(t, types.StringValue("ab"), defaultResponse.PlanValue)

	// A documented replacement must still replace on change and preserve a no-op plan.
	raw := tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{}}, map[string]tftypes.Value{})
	for _, value := range []string{"before", "after"} {
		var response planmodifier.StringResponse
		attribute.PlanModifiers[0].PlanModifyString(ctx, planmodifier.StringRequest{
			State: tfsdk.State{Raw: raw}, Plan: tfsdk.Plan{Raw: raw},
			StateValue: types.StringValue("before"), PlanValue: types.StringValue(value),
		}, &response)
		require.False(t, response.Diagnostics.HasError())
		assert.Equal(t, value != "before", response.RequiresReplace)
	}
}

func TestRangeAndDefaultFollowDeclarations(t *testing.T) {
	ctx := context.Background()
	docs := schemadoc.Resource{Name: "network volume"}
	for _, max := range []int64{10, 20} {
		attribute := docs.Int64(schema.Int64Attribute{
			Optional: true, Computed: true,
			Validators: []validator.Int64{schemadoc.Between(1, max)},
		}, schemadoc.Field{Definition: "Size in GiB."})
		assert.Contains(t, attribute.MarkdownDescription, "The available range is from `1` to `"+types.Int64Value(max).String()+"`.")
		for _, value := range []int64{0, 1, max, max + 1} {
			var response validator.Int64Response
			attribute.Validators[0].ValidateInt64(ctx, validator.Int64Request{ConfigValue: types.Int64Value(value)}, &response)
			assert.Equal(t, value >= 1 && value <= max, !response.Diagnostics.HasError())
		}
	}
	for _, value := range []bool{false, true} {
		attribute := docs.Bool(schema.BoolAttribute{
			Optional: true, Computed: true, Default: schemadoc.StaticBool(value),
		}, schemadoc.Field{Definition: "Enables online resizing."})
		assert.Contains(t, attribute.MarkdownDescription, "The default value is `"+types.BoolValue(value).String()+"`.")
		var response defaults.BoolResponse
		attribute.Default.DefaultBool(ctx, defaults.BoolRequest{}, &response)
		require.False(t, response.Diagnostics.HasError())
		assert.Equal(t, types.BoolValue(value), response.PlanValue)
		assert.False(t, strings.Contains(attribute.MarkdownDescription, "Changing this"))
	}
}

func TestEnumInDataSource(t *testing.T) {
	// Adding a format changes both the accepted configuration and its documentation.
	for _, formats := range [][]string{{"raw"}, {"raw", "qcow2"}, {"raw", "qcow2", "vmdk"}} {
		attribute := schemadoc.DataSource{Name: "image"}.String(dsschema.StringAttribute{
			Optional: true, Validators: []validator.String{schemadoc.OneOf(formats...)},
		}, schemadoc.Field{Definition: "Image format."})
		if len(formats) == 1 {
			assert.Contains(t, attribute.MarkdownDescription, "The only available value is `raw`.")
		}
		if len(formats) == 3 {
			assert.Contains(t, attribute.MarkdownDescription, "`raw`, `qcow2`, and `vmdk`.")
		}
		var response validator.StringResponse
		attribute.Validators[0].ValidateString(context.Background(), validator.StringRequest{
			ConfigValue: types.StringValue("qcow2"),
		}, &response)
		assert.Equal(t, len(formats) >= 2, !response.Diagnostics.HasError())
		assert.Equal(t, len(formats) >= 2, strings.Contains(attribute.MarkdownDescription, "`qcow2`"))
	}
}

func TestSourceLinks(t *testing.T) {
	assert.Equal(t, "You can retrieve the ID with the [selectel_compute_volume_v3]"+
		"(https://registry.terraform.io/providers/selectel/selectel/latest/docs/data-sources/compute_volume_v3) data source.",
		schemadoc.FromDataSource("selectel_compute_volume_v3", "ID"))
	assert.Equal(t, "You can retrieve the ID from the [selectel_vpc_project_v2]"+
		"(https://registry.terraform.io/providers/selectel/selectel/latest/docs/resources/vpc_project_v2) resource.",
		schemadoc.FromResource("selectel_vpc_project_v2", "ID"))
}

func TestTimeoutsDocumentDefaults(t *testing.T) {
	block := schemadoc.Resource{Name: "network volume"}.Timeouts(context.Background(), schemadoc.Timeouts{
		Create: 10 * time.Minute, Delete: 90 * time.Minute,
	}).(schema.SingleNestedBlock)
	assert.ElementsMatch(t, []string{"create", "delete"}, slices.Collect(maps.Keys(block.Attributes)))
	assert.Contains(t, block.Attributes["create"].GetMarkdownDescription(), "If omitted, the timeout is `10m`.")
	assert.Contains(t, block.Attributes["delete"].GetMarkdownDescription(), "If omitted, the timeout is `1h30m`.")
	assert.Empty(t, schemadoc.Lint(schema.Schema{Blocks: map[string]schema.Block{"timeouts": block}}))
}

func TestLint(t *testing.T) {
	docs := schemadoc.Resource{Name: "network volume"}
	documented := docs.String(schema.StringAttribute{
		Required:      true,
		Validators:    []validator.String{schemadoc.LengthAtMost(255)},
		PlanModifiers: []planmodifier.String{schemadoc.RequiresReplace(), stringplanmodifier.UseStateForUnknown()},
	}, schemadoc.Field{Definition: "Name."})
	stale := documented
	stale.Validators = []validator.String{schemadoc.LengthAtMost(128)}
	plain := schema.StringAttribute{
		Optional: true, Description: "Name.",
		Validators: []validator.String{schemadoc.OneOf("a")},
	}
	undocumented := docs.String(schema.StringAttribute{
		Optional: true, Validators: []validator.String{stringvalidator.LengthAtMost(1)},
	}, schemadoc.Field{Definition: "Name."})
	nested := schema.SetNestedAttribute{Computed: true, NestedObject: schema.NestedAttributeObject{
		Attributes: map[string]schema.Attribute{"inner": undocumented},
	}}

	assert.Equal(t, []string{
		"nested.inner: stringvalidator.lengthAtMostValidator has no documentation facts; use a schemadoc rule or DescribeString",
		"plain: description lacks \"The only available value is `a`.\"",
		"stale: description lacks \"The maximum length is `128` characters.\"",
		"undocumented: stringvalidator.lengthAtMostValidator has no documentation facts; use a schemadoc rule or DescribeString",
	}, schemadoc.Lint(schema.Schema{Attributes: map[string]schema.Attribute{
		"documented": documented, "stale": stale, "plain": plain, "undocumented": undocumented, "nested": nested,
	}}))
}

func TestLintSkipsSchemaWithoutSchemadoc(t *testing.T) {
	assert.Empty(t, schemadoc.Lint(schema.Schema{Attributes: map[string]schema.Attribute{
		"name": schema.StringAttribute{
			Required: true, Description: "Name.",
			PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
		},
	}}))
}
