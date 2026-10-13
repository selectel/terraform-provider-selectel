package schemaconfig

import (
	"fmt"
	"testing"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestResourceData(t *testing.T, sch map[string]*schema.Schema, raw map[string]any) *schema.ResourceData {
	t.Helper()

	d := schema.TestResourceDataRaw(t, sch, raw)
	d.SetId("raw-config")
	state := d.State()
	state.ID = ""
	delete(state.Attributes, "id")
	state.RawConfig = objectValue(sch, raw)

	return (&schema.Resource{Schema: sch}).Data(state)
}

func objectValue(sch map[string]*schema.Schema, raw map[string]any) cty.Value {
	attrs := make(map[string]cty.Value, len(sch))
	for key, s := range sch {
		if v, ok := raw[key]; ok {
			attrs[key] = attrValue(s, v)
		} else {
			attrs[key] = cty.NullVal(attrType(s))
		}
	}

	return cty.ObjectVal(attrs)
}

func attrType(s *schema.Schema) cty.Type {
	switch s.Type {
	case schema.TypeBool:
		return cty.Bool
	case schema.TypeInt, schema.TypeFloat:
		return cty.Number
	case schema.TypeString:
		return cty.String
	case schema.TypeList:
		return cty.List(elemType(s.Elem))
	case schema.TypeSet:
		return cty.Set(elemType(s.Elem))
	case schema.TypeMap:
		return cty.Map(elemType(s.Elem))
	case schema.TypeInvalid:
		panic(fmt.Sprintf("schemaconfig: unsupported schema type %s", s.Type))
	}

	panic(fmt.Sprintf("schemaconfig: unsupported schema type %s", s.Type))
}

func elemType(elem any) cty.Type {
	switch e := elem.(type) {
	case *schema.Schema:
		return attrType(e)
	case *schema.Resource:
		attrs := make(map[string]cty.Type, len(e.Schema))
		for key, s := range e.Schema {
			attrs[key] = attrType(s)
		}

		return cty.Object(attrs)
	default:
		return cty.String
	}
}

func attrValue(s *schema.Schema, v any) cty.Value {
	switch s.Type {
	case schema.TypeBool:
		return cty.BoolVal(v.(bool))
	case schema.TypeInt:
		return cty.NumberIntVal(int64(v.(int)))
	case schema.TypeFloat:
		return cty.NumberFloatVal(v.(float64))
	case schema.TypeString:
		return cty.StringVal(v.(string))
	case schema.TypeList:
		return collectionValue(s, v, cty.ListVal, cty.ListValEmpty)
	case schema.TypeSet:
		return collectionValue(s, v, cty.SetVal, cty.SetValEmpty)
	case schema.TypeMap:
		entries := make(map[string]cty.Value)
		for key, item := range v.(map[string]any) {
			entries[key] = cty.StringVal(fmt.Sprint(item))
		}
		if len(entries) == 0 {
			return cty.MapValEmpty(elemType(s.Elem))
		}

		return cty.MapVal(entries)
	case schema.TypeInvalid:
		panic(fmt.Sprintf("schemaconfig: unsupported schema type %s", s.Type))
	}

	panic(fmt.Sprintf("schemaconfig: unsupported schema type %s", s.Type))
}

func collectionValue(
	s *schema.Schema, v any,
	of func([]cty.Value) cty.Value, empty func(cty.Type) cty.Value,
) cty.Value {
	items := v.([]any)
	if len(items) == 0 {
		return empty(elemType(s.Elem))
	}

	values := make([]cty.Value, 0, len(items))
	for _, item := range items {
		values = append(values, elemValue(s.Elem, item))
	}

	return of(values)
}

func elemValue(elem any, v any) cty.Value {
	switch e := elem.(type) {
	case *schema.Schema:
		return attrValue(e, v)
	case *schema.Resource:
		return objectValue(e.Schema, v.(map[string]any))
	default:
		return cty.StringVal(fmt.Sprint(v))
	}
}
