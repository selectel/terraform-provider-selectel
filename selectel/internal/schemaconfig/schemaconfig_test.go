package schemaconfig

import (
	"testing"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/stretchr/testify/assert"
)

func testResource() *schema.Resource {
	return &schema.Resource{Schema: map[string]*schema.Schema{
		"enable_snat": {Type: schema.TypeBool, Optional: true, Computed: true},
		"port_min":    {Type: schema.TypeInt, Optional: true, Computed: true},
	}}
}

func dataWithRawConfig(raw cty.Value) *schema.ResourceData {
	return testResource().Data(&terraform.InstanceState{RawConfig: raw})
}

func TestIsSet(t *testing.T) {
	t.Run("explicit false is set", func(t *testing.T) {
		d := dataWithRawConfig(cty.ObjectVal(map[string]cty.Value{
			"enable_snat": cty.False,
			"port_min":    cty.NullVal(cty.Number),
		}))
		assert.True(t, IsSet(d, "enable_snat"))
	})

	t.Run("explicit zero is set", func(t *testing.T) {
		d := dataWithRawConfig(cty.ObjectVal(map[string]cty.Value{
			"enable_snat": cty.NullVal(cty.Bool),
			"port_min":    cty.NumberIntVal(0),
		}))
		assert.True(t, IsSet(d, "port_min"))
	})

	t.Run("omitted attribute is not set", func(t *testing.T) {
		d := dataWithRawConfig(cty.ObjectVal(map[string]cty.Value{
			"enable_snat": cty.NullVal(cty.Bool),
			"port_min":    cty.NullVal(cty.Number),
		}))
		assert.False(t, IsSet(d, "enable_snat"))
		assert.False(t, IsSet(d, "port_min"))
	})

	t.Run("unknown value counts as set", func(t *testing.T) {
		d := dataWithRawConfig(cty.ObjectVal(map[string]cty.Value{
			"enable_snat": cty.UnknownVal(cty.Bool),
			"port_min":    cty.NullVal(cty.Number),
		}))
		assert.True(t, IsSet(d, "enable_snat"),
			"a value that is only known after apply was still written by the user")
	})

	t.Run("no raw config at all is not set", func(t *testing.T) {
		assert.False(t, IsSet(testResource().Data(nil), "enable_snat"))
		assert.False(t, IsSet(dataWithRawConfig(cty.NullVal(cty.EmptyObject)), "enable_snat"))
	})

	t.Run("attribute missing from the object is not set", func(t *testing.T) {
		d := dataWithRawConfig(cty.ObjectVal(map[string]cty.Value{
			"enable_snat": cty.True,
		}))
		assert.False(t, IsSet(d, "port_min"))
	})
}

func TestTestResourceData(t *testing.T) {
	sch := map[string]*schema.Schema{
		"enable_snat": {Type: schema.TypeBool, Optional: true, Computed: true},
		"port_min":    {Type: schema.TypeInt, Optional: true, Computed: true},
		"name":        {Type: schema.TypeString, Optional: true},
		"tags":        {Type: schema.TypeSet, Optional: true, Elem: &schema.Schema{Type: schema.TypeString}},
		"external_fixed_ip": {Type: schema.TypeList, Optional: true, Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"subnet_id":  {Type: schema.TypeString, Optional: true},
				"ip_address": {Type: schema.TypeString, Optional: true},
			},
		}},
	}

	t.Run("explicit zero values are both readable and set", func(t *testing.T) {
		d := TestResourceData(t, sch, map[string]any{
			"enable_snat": false,
			"port_min":    0,
		})

		assert.False(t, d.Get("enable_snat").(bool))
		assert.True(t, IsSet(d, "enable_snat"))
		assert.Equal(t, 0, d.Get("port_min").(int))
		assert.True(t, IsSet(d, "port_min"))
	})

	t.Run("omitted values are not set", func(t *testing.T) {
		d := TestResourceData(t, sch, map[string]any{"name": "router"})

		assert.Equal(t, "router", d.Get("name"))
		assert.True(t, IsSet(d, "name"))
		assert.False(t, IsSet(d, "enable_snat"))
		assert.False(t, IsSet(d, "port_min"))
		assert.False(t, IsSet(d, "tags"))
		assert.False(t, IsSet(d, "external_fixed_ip"))
	})

	t.Run("collections and nested blocks round trip", func(t *testing.T) {
		d := TestResourceData(t, sch, map[string]any{
			"tags": []any{"a", "b"},
			"external_fixed_ip": []any{
				map[string]any{"subnet_id": "subnet-1", "ip_address": "203.0.113.10"},
			},
		})

		assert.ElementsMatch(t, []any{"a", "b"}, d.Get("tags").(*schema.Set).List())
		assert.Equal(t, "203.0.113.10", d.Get("external_fixed_ip.0.ip_address"))
		assert.True(t, IsSet(d, "tags"))
		assert.True(t, IsSet(d, "external_fixed_ip"))
	})
}
