package schemadiff

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRecomputeOnChange(t *testing.T) {
	res := &schema.Resource{
		Schema: map[string]*schema.Schema{
			"tags": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"name": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"all_tags": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
		CustomizeDiff: RecomputeOnChange("all_tags", "tags"),
	}

	state := &terraform.InstanceState{
		ID: "id",
		Attributes: map[string]string{
			"id":         "id",
			"name":       "one",
			"tags.#":     "1",
			"tags.0":     "managed",
			"all_tags.#": "2",
			"all_tags.0": "managed",
			"all_tags.1": "foreign",
		},
	}

	t.Run("a changed trigger makes the set unknown", func(t *testing.T) {
		diff, err := res.Diff(context.Background(), state, terraform.NewResourceConfigRaw(map[string]any{
			"name": "one",
			"tags": []any{"managed", "added"},
		}), nil)

		require.NoError(t, err)
		require.NotNil(t, diff)
		require.Contains(t, diff.Attributes, "all_tags.#")
		assert.True(t, diff.Attributes["all_tags.#"].NewComputed)
	})

	t.Run("an unrelated change leaves the set alone", func(t *testing.T) {
		diff, err := res.Diff(context.Background(), state, terraform.NewResourceConfigRaw(map[string]any{
			"name": "two",
			"tags": []any{"managed"},
		}), nil)

		require.NoError(t, err)
		require.NotNil(t, diff)
		assert.NotContains(t, diff.Attributes, "all_tags.#")
	})
}
