package selectel

import (
	"context"
	"fmt"
	"slices"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/stretchr/testify/assert"
)

func vpcTagsV2TestSet(tags ...string) *schema.Set {
	values := make([]any, len(tags))
	for i, tag := range tags {
		values[i] = tag
	}

	return schema.NewSet(schema.HashString, values)
}

func TestFlattenVPCTagsV2(t *testing.T) {
	t.Run("keeps tags that are present on the object", func(t *testing.T) {
		d := schema.TestResourceDataRaw(t, resourceVPCNetworkV2().Schema, map[string]any{
			"tags": []any{"managed"},
		})

		flattenVPCTagsV2(d, []string{"managed", "foreign"})

		assert.ElementsMatch(t, []any{"managed", "foreign"}, d.Get("all_tags").(*schema.Set).List())
		assert.ElementsMatch(t, []any{"managed"}, d.Get("tags").(*schema.Set).List())
	})

	t.Run("drops tags removed outside terraform", func(t *testing.T) {
		d := schema.TestResourceDataRaw(t, resourceVPCNetworkV2().Schema, map[string]any{
			"tags": []any{"managed", "removed-outside"},
		})

		flattenVPCTagsV2(d, []string{"managed", "foreign"})

		assert.ElementsMatch(t, []any{"managed"}, d.Get("tags").(*schema.Set).List())
	})
}

func TestExpandVPCUpdateTagsV2(t *testing.T) {
	d := schema.TestResourceDataRaw(t, resourceVPCNetworkV2().Schema, map[string]any{
		"tags": []any{"managed"},
	})
	d.Set("all_tags", []string{"foreign"})

	assert.ElementsMatch(t, []string{"managed", "foreign"}, expandVPCUpdateTagsV2(d))
}

type vpcTagsV2AccOperations interface {
	Get(ctx context.Context) ([]string, error)
	Add(ctx context.Context, tag string) error
}

func testAccCheckVPCTagsV2AddOutOfBand(
	operations func(*terraform.State) (vpcTagsV2AccOperations, error), tag string,
) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		ops, err := operations(s)
		if err != nil {
			return err
		}

		return ops.Add(context.Background(), tag)
	}
}

func testAccCheckVPCTagsV2Match(
	operations func(*terraform.State) (vpcTagsV2AccOperations, error), expected []string,
) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		ops, err := operations(s)
		if err != nil {
			return err
		}

		actual, err := ops.Get(context.Background())
		if err != nil {
			return err
		}

		actual = slices.Clone(actual)
		want := slices.Clone(expected)
		slices.Sort(actual)
		slices.Sort(want)

		if !slices.Equal(actual, want) {
			return fmt.Errorf("object has tags %v, want %v", actual, want)
		}

		return nil
	}
}

func TestMergeVPCTagsV2(t *testing.T) {
	t.Run("keeps foreign tags on managed tag change", func(t *testing.T) {
		merged := mergeVPCTagsV2(
			vpcTagsV2TestSet("one", "two", "foreign"),
			vpcTagsV2TestSet("one", "two"),
			vpcTagsV2TestSet("one", "three"),
		)

		assert.ElementsMatch(t, []string{"one", "three", "foreign"}, merged)
	})

	t.Run("keeps only foreign tags when all managed tags are removed", func(t *testing.T) {
		merged := mergeVPCTagsV2(
			vpcTagsV2TestSet("one", "foreign"),
			vpcTagsV2TestSet("one"),
			vpcTagsV2TestSet(),
		)

		assert.ElementsMatch(t, []string{"foreign"}, merged)
	})
}
