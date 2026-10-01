package selectel

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"
)

func TestWithDocsHints(t *testing.T) {
	s := resourceDocs{Name: "public port"}.withDocsHints(map[string]*schema.Schema{
		"force_new": {
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
			Description: "Force new field.",
		},
		"with_default": {
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     true,
			Description: "Bool field.",
		},
		"empty_default": {
			Type:        schema.TypeString,
			Optional:    true,
			Default:     "",
			Description: "String field.",
		},
		"plain": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "Computed field.",
		},
		"nested": {
			Type:     schema.TypeList,
			Optional: true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"inner": {
						Type:        schema.TypeString,
						Optional:    true,
						ForceNew:    true,
						Description: "Inner field.",
					},
				},
			},
		},
	})

	assert.Equal(t, "Force new field. Changing this creates a new public port.", s["force_new"].Description)
	assert.Equal(t, "Bool field. The default value is `true`.", s["with_default"].Description)
	assert.Equal(t, "String field. The default value is an empty string.", s["empty_default"].Description)
	assert.Equal(t, "Computed field.", s["plain"].Description)
	assert.Equal(t, "Inner field. Changing this creates a new public port.",
		s["nested"].Elem.(*schema.Resource).Schema["inner"].Description)
}

func TestWithDocsHintsDependencies(t *testing.T) {
	s := resourceDocs{Name: "port"}.withDocsHints(map[string]*schema.Schema{
		"conflicts": {
			Type:          schema.TypeString,
			Optional:      true,
			ForceNew:      true,
			ConflictsWith: []string{"a", "b", "c"},
			Description:   "Conflicting field.",
		},
		"pair": {
			Type:         schema.TypeString,
			Optional:     true,
			ExactlyOneOf: []string{"pair", "other"},
			Description:  "Pair field.",
		},
		"triple": {
			Type:         schema.TypeString,
			Optional:     true,
			ExactlyOneOf: []string{"triple", "x", "y"},
			Description:  "Triple field.",
		},
		"least": {
			Type:         schema.TypeString,
			Optional:     true,
			AtLeastOneOf: []string{"least", "x"},
			Description:  "Least field.",
		},
		"required": {
			Type:         schema.TypeString,
			Optional:     true,
			RequiredWith: []string{"protocol", "port_range_min"},
			Default:      "tcp",
			Description:  "Required field.",
		},
	})

	assert.Equal(t, "Conflicting field. Conflicts with `a`, `b`, and `c`. Changing this creates a new port.",
		s["conflicts"].Description)
	assert.Equal(t, "Pair field. Specify either `pair` or `other`.", s["pair"].Description)
	assert.Equal(t, "Triple field. Exactly one of `triple`, `x`, and `y` must be specified.", s["triple"].Description)
	assert.Equal(t, "Least field. At least one of `least` and `x` must be specified.", s["least"].Description)
	assert.Equal(t, "Required field. Requires `protocol` and `port_range_min`. The default value is `tcp`.",
		s["required"].Description)
}

func TestResourceDocsIDSchemas(t *testing.T) {
	docs := resourceDocs{Name: "public port"}

	res := docs.idResourceSchema()
	assert.True(t, res.Computed)
	assert.Equal(t, "Unique identifier of the public port.", res.Description)

	identity := docs.idIdentitySchema(exampleResourceID, "Copy it from the card.")
	assert.True(t, identity.RequiredForImport)
	assert.Equal(t,
		"Unique identifier of the public port, for example, `b311ce58-2658-46b5-b733-7a0f418703f2`. Copy it from the card.",
		identity.Description)
}
