package selectel

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
)

func TestWithFrameworkDocsHints(t *testing.T) {
	s := resourceDocs{Name: "public port"}.withFrameworkDocsHints(map[string]schema.Attribute{
		"force_new": schema.StringAttribute{
			Required:      true,
			PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			Description:   "Force new field.",
		},
		"with_default": schema.BoolAttribute{
			Optional:    true,
			Computed:    true,
			Default:     booldefault.StaticBool(true),
			Description: "Bool field.",
		},
		"empty_default": schema.StringAttribute{
			Optional:    true,
			Computed:    true,
			Default:     stringdefault.StaticString(""),
			Description: "String field.",
		},
		"markdown": schema.StringAttribute{
			Required:            true,
			PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			Description:         "Markdown field.",
			MarkdownDescription: "**Markdown** field.",
		},
		"plain": schema.StringAttribute{
			Computed:    true,
			Description: "Computed field.",
		},
		"count": schema.Int64Attribute{
			Optional:      true,
			Computed:      true,
			Default:       int64default.StaticInt64(3),
			PlanModifiers: []planmodifier.Int64{int64planmodifier.RequiresReplace()},
			Description:   "Int field.",
		},
		"tags": schema.ListAttribute{
			ElementType:   types.StringType,
			Optional:      true,
			PlanModifiers: []planmodifier.List{listplanmodifier.RequiresReplace()},
			Description:   "List field.",
		},
		"zones": schema.ListAttribute{
			ElementType: types.StringType,
			Optional:    true,
			Computed:    true,
			Default: listdefault.StaticValue(types.ListValueMust(types.StringType, []attr.Value{
				types.StringValue("ru-1a"), types.StringValue("ru-1b"),
			})),
			Description: "Zones.",
		},
		"items": schema.ListNestedAttribute{
			Optional: true,
			NestedObject: schema.NestedAttributeObject{
				Attributes: map[string]schema.Attribute{
					"inner": schema.BoolAttribute{
						Optional:    true,
						Computed:    true,
						Default:     booldefault.StaticBool(false),
						Description: "Inner field.",
					},
				},
			},
		},
		"nested": schema.SingleNestedAttribute{
			Optional: true,
			Attributes: map[string]schema.Attribute{
				"inner": schema.StringAttribute{
					Optional:      true,
					PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplaceIfConfigured()},
					Description:   "Inner field.",
				},
			},
		},
	})

	assert.Equal(t, "Force new field. Changing this creates a new public port.", s["force_new"].GetDescription())
	assert.Equal(t, "Bool field. The default value is `true`.", s["with_default"].GetDescription())
	assert.Equal(t, "String field. The default value is an empty string.", s["empty_default"].GetDescription())
	assert.Equal(t, "Markdown field. Changing this creates a new public port.", s["markdown"].GetDescription())
	assert.Equal(t, "**Markdown** field. Changing this creates a new public port.", s["markdown"].GetMarkdownDescription())
	assert.Equal(t, "Computed field.", s["plain"].GetDescription())
	assert.Equal(t, "Inner field. Changing this creates a new public port.",
		s["nested"].(schema.SingleNestedAttribute).Attributes["inner"].GetDescription())

	blocks := resourceDocs{Name: "public port"}.withFrameworkBlockDocsHints(map[string]schema.Block{
		"block": schema.ListNestedBlock{
			PlanModifiers: []planmodifier.List{listplanmodifier.RequiresReplace()},
			Description:   "Block.",
			NestedObject: schema.NestedBlockObject{
				Blocks: map[string]schema.Block{
					"sub": schema.SingleNestedBlock{
						Attributes: map[string]schema.Attribute{
							"inner": schema.StringAttribute{
								Required:      true,
								PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
								Description:   "Inner field.",
							},
						},
					},
				},
			},
		},
	})

	assert.Equal(t, "Int field. Changing this creates a new public port. The default value is `3`.", s["count"].GetDescription())
	assert.Equal(t, "List field. Changing this creates a new public port.", s["tags"].GetDescription())
	assert.Equal(t, "Zones. The default value is `[\"ru-1a\",\"ru-1b\"]`.", s["zones"].GetDescription())
	assert.Equal(t, "Inner field. The default value is `false`.",
		s["items"].(schema.ListNestedAttribute).NestedObject.Attributes["inner"].GetDescription())

	block := blocks["block"].(schema.ListNestedBlock)
	assert.Equal(t, "Block. Changing this creates a new public port.", block.Description)
	assert.Equal(t, "Inner field. Changing this creates a new public port.",
		block.NestedObject.Blocks["sub"].(schema.SingleNestedBlock).Attributes["inner"].GetDescription())
}

func TestResourceDocsFrameworkIDSchemas(t *testing.T) {
	docs := resourceDocs{Name: "public port"}

	res := docs.idFrameworkResourceSchema()
	assert.True(t, res.Computed)
	assert.Equal(t, "Unique identifier of the public port.", res.Description)

	identity := docs.idFrameworkIdentitySchema("Copy it from the card.")
	assert.True(t, identity.RequiredForImport)
	assert.Equal(t,
		"Unique identifier of the public port, for example, `b311ce58-2658-46b5-b733-7a0f418703f2`. Copy it from the card.",
		identity.Description)
}

func TestFrameworkHelpersMatchSDKv2(t *testing.T) {
	docs := resourceDocs{Name: "public port"}
	hint := "Copy it from the card."

	assert.Equal(t, docs.regionResourceSchema().Description, docs.regionFrameworkResourceSchema().Description)
	assert.Equal(t, docs.regionIdentitySchema(hint).Description, docs.regionFrameworkIdentitySchema(hint).Description)
	assert.Equal(t, projectIDResourceSchema().Description, projectIDFrameworkResourceSchema().Description)
	assert.Equal(t, projectIDIdentitySchema().Description, projectIDFrameworkIdentitySchema().Description)
}
