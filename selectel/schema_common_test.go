package selectel

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"
)

func TestWithDocsHints(t *testing.T) {
	s := withDocsHints("public port", map[string]*schema.Schema{
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
