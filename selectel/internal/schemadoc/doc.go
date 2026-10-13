// Package schemadoc adds documentation to ordinary Framework attributes.
// Rules delegate runtime behavior to HashiCorp's implementations and expose
// documentation facts without parsing their human-readable descriptions.
package schemadoc

import (
	"fmt"
	"reflect"
	"strings"

	dsschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

// Resource describes the attributes of a resource. Name is the object noun used
// in generated sentences, such as "network volume". Location is the anchor of the
// product section on the product availability page, such as "network-volumes".
type Resource struct {
	Name     string
	Location string
}

// DataSource describes the attributes of a data source, as Resource does.
type DataSource struct {
	Name     string
	Location string
}

// Field contains information that cannot be inferred from the schema.
type Field struct {
	Definition string
	Behavior   string
	// Source says where to get the value of an ID or name argument.
	// Build it with FromDataSource, FromResource or FromOpenStackCLI.
	Source  string
	Concept Link
}

// Link renders "Learn more about [Title](URL)." or, with Topic,
// "Learn more about <Topic> in [Title](URL).". A Vendor link points outside the
// Selectel documentation and renders "For more information about <Topic>, see
// the [official <Vendor> documentation](URL)." (DOC-SCHEMA-CONCEPT).
type Link struct {
	Topic  string
	Title  string
	URL    string
	Vendor string
}

// Facts are the user-facing effects of a validator, default or plan modifier.
type Facts struct {
	Constraint       string
	Dependency       string
	Default          string
	ReplacesResource bool
}

func (f Facts) Documentation() Facts { return f }

// Documented is implemented by every rule that contributes to the description.
// Embedding Facts in a rule type implements it. Rules without it add nothing,
// and Lint reports them unless they are runtime-only.
type Documented interface {
	Documentation() Facts
}

// String, Int64, Bool and Map document the completed attribute. Call them
// after setting defaults, validators and plan modifiers.
func (r Resource) String(a schema.StringAttribute, field Field) schema.StringAttribute {
	a.MarkdownDescription = render(r.Name, field, rulesOf(a))
	return a
}

func (r Resource) Int64(a schema.Int64Attribute, field Field) schema.Int64Attribute {
	a.MarkdownDescription = render(r.Name, field, rulesOf(a))
	return a
}

func (r Resource) Bool(a schema.BoolAttribute, field Field) schema.BoolAttribute {
	a.MarkdownDescription = render(r.Name, field, rulesOf(a))
	return a
}

func (r Resource) Map(a schema.MapAttribute, field Field) schema.MapAttribute {
	a.MarkdownDescription = render(r.Name, field, rulesOf(a))
	return a
}

func (d DataSource) String(a dsschema.StringAttribute, field Field) dsschema.StringAttribute {
	a.MarkdownDescription = render(d.Name, field, rulesOf(a))
	return a
}

func (d DataSource) Int64(a dsschema.Int64Attribute, field Field) dsschema.Int64Attribute {
	a.MarkdownDescription = render(d.Name, field, rulesOf(a))
	return a
}

func (d DataSource) Bool(a dsschema.BoolAttribute, field Field) dsschema.BoolAttribute {
	a.MarkdownDescription = render(d.Name, field, rulesOf(a))
	return a
}

func (d DataSource) Map(a dsschema.MapAttribute, field Field) dsschema.MapAttribute {
	a.MarkdownDescription = render(d.Name, field, rulesOf(a))
	return a
}

// render orders sentences by the DOC-SCHEMA rules of the technical writers:
// definition, constraints, dependencies, behavior, source, default, link, replacement.
func render(noun string, field Field, rules []any) string {
	parts := []string{field.Definition}
	var dependencies []string
	var defaultText string
	var replace bool
	for _, rule := range rules {
		if rule, ok := rule.(Documented); ok {
			facts := rule.Documentation()
			parts = append(parts, facts.Constraint)
			dependencies = append(dependencies, facts.Dependency)
			if facts.Default != "" {
				defaultText = facts.Default
			}
			replace = replace || facts.ReplacesResource
		}
	}
	parts = append(parts, dependencies...)
	parts = append(parts, field.Behavior, field.Source, defaultText)
	if link := field.Concept; link.Vendor != "" {
		parts = append(parts, fmt.Sprintf("For more information about %s, see the [official %s documentation](%s).", link.Topic, link.Vendor, link.URL))
	} else if link.URL != "" {
		topic := ""
		if link.Topic != "" {
			topic = link.Topic + " in "
		}
		parts = append(parts, fmt.Sprintf("Learn more about %s[%s](%s).", topic, link.Title, link.URL))
	}
	if replace {
		parts = append(parts, replacementPrefix+noun+".")
	}
	nonempty := parts[:0]
	for _, part := range parts {
		if part != "" {
			nonempty = append(nonempty, part)
		}
	}

	return strings.Join(nonempty, " ")
}

const replacementPrefix = "Changing this creates a new "

const registry = "https://registry.terraform.io/providers/selectel/selectel/latest/docs/"

// FromDataSource names the data source that returns the value. what is the
// retrieved value as it reads in the sentence, for example "ID" or "name".
func FromDataSource(name, what string) string {
	return fmt.Sprintf("You can retrieve the %s with the [%s](%sdata-sources/%s) data source.",
		what, name, registry, strings.TrimPrefix(name, "selectel_"))
}

// FromResource names the resource that exports the value. On resource pages it
// takes priority over FromDataSource (DOC-SCHEMA-SOURCE).
func FromResource(name, what string) string {
	return fmt.Sprintf("You can retrieve the %s from the [%s](%sresources/%s) resource.",
		what, name, registry, strings.TrimPrefix(name, "selectel_"))
}

// FromOpenStackCLI is for values without a Selectel data source or resource.
// column is the column of the command output, for example "ID".
func FromOpenStackCLI(what, command, column string) string {
	return fmt.Sprintf("To get the %s, use the [OpenStack CLI](https://docs.selectel.ru/en/cloud-servers/tools/openstack-cli/) command `%s` and copy the `%s` field.",
		what, command, column)
}

// codeList joins values as an English list of inline code: `a`, `b`, and `c`.
func codeList(values []string) string {
	quoted := make([]string, len(values))
	for i, value := range values {
		quoted[i] = "`" + value + "`"
	}
	if len(quoted) == 2 {
		return quoted[0] + " and " + quoted[1]
	}
	if len(quoted) > 2 {
		return strings.Join(quoted[:len(quoted)-1], ", ") + ", and " + quoted[len(quoted)-1]
	}

	return strings.Join(quoted, "")
}

// rulesOf collects the validators, plan modifiers and default of an attribute.
// It reads the exported fields that every Framework attribute struct has, so
// it covers all attribute kinds of resources and data sources.
func rulesOf(attribute any) []any {
	fields := reflect.ValueOf(attribute)
	var rules []any
	for _, list := range []reflect.Value{fields.FieldByName("Validators"), fields.FieldByName("PlanModifiers")} {
		if list.IsValid() {
			for i := range list.Len() {
				rules = append(rules, list.Index(i).Interface())
			}
		}
	}
	if def := fields.FieldByName("Default"); def.IsValid() && !def.IsNil() {
		rules = append(rules, def.Interface())
	}

	return rules
}
