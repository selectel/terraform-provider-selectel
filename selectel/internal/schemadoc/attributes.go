package schemadoc

import (
	"context"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	dsschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

// Attributes shared by resources and data sources. A resource and a data source
// with the same noun get identical descriptions.

var projectIDField = Field{
	Definition: "Unique identifier of the associated project, for example, `a07abc12310546f1b9291ab3013a7d75`.",
	Source:     FromResource("selectel_vpc_project_v2", "project ID"),
	Concept: Link{
		Title: "Projects",
		URL:   "https://docs.selectel.ru/en/control-panel-actions/projects/about-projects/",
	},
}

func idField(noun string) Field {
	return Field{Definition: "Unique identifier of the " + noun + "."}
}

// LocationLink points to the product availability page, at the product section
// when anchor is set.
func LocationLink(topic, anchor string) Link {
	url := "https://docs.selectel.ru/en/infrastructure/product-availability-by-location/"
	if anchor != "" {
		url += "#" + anchor
	}

	return Link{Topic: topic, Title: "Product availability by location", URL: url}
}

func regionField(noun, anchor string) Field {
	return Field{
		Definition: "Pool where the " + noun + " is located, for example, `ru-6`.",
		Concept:    LocationLink("available pools", anchor),
	}
}

// ID is the computed identifier of the object.
func (r Resource) ID() schema.StringAttribute {
	return r.String(schema.StringAttribute{
		Computed:      true,
		PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
	}, idField(r.Name))
}

// ProjectID is the project of the object. Changing it replaces the object.
func (r Resource) ProjectID() schema.StringAttribute { return r.scope(projectIDField) }

// Region is the pool of the object. Changing it replaces the object.
func (r Resource) Region() schema.StringAttribute { return r.scope(regionField(r.Name, r.Location)) }

func (r Resource) scope(field Field) schema.StringAttribute {
	return r.String(schema.StringAttribute{
		Required:      true,
		Validators:    []validator.String{NonEmptyTrimmed()},
		PlanModifiers: []planmodifier.String{RequiresReplace()},
	}, field)
}

// ID is the identifier of the found object.
func (d DataSource) ID() dsschema.StringAttribute {
	return d.String(dsschema.StringAttribute{Computed: true}, idField(d.Name))
}

// ProjectID is the project to search in.
func (d DataSource) ProjectID() dsschema.StringAttribute { return d.scope(projectIDField) }

// Region is the pool to search in.
func (d DataSource) Region() dsschema.StringAttribute {
	return d.scope(regionField(d.Name, d.Location))
}

func (d DataSource) scope(field Field) dsschema.StringAttribute {
	return d.String(dsschema.StringAttribute{
		Required:   true,
		Validators: []validator.String{NonEmptyTrimmed()},
	}, field)
}

// Timeouts are the operation timeouts used when the timeouts block omits them.
// A zero duration leaves the operation out of the block.
type Timeouts struct {
	Create, Read, Update, Delete time.Duration
}

// Timeouts is the documented timeouts block. Read the configured values with
// the Framework timeouts.Value and pass the same Timeouts as their defaults.
func (r Resource) Timeouts(ctx context.Context, defaults Timeouts) schema.Block {
	block := timeouts.Block(ctx, timeouts.Opts{
		Create: defaults.Create != 0,
		Read:   defaults.Read != 0,
		Update: defaults.Update != 0,
		Delete: defaults.Delete != 0,
	}).(schema.SingleNestedBlock)
	block.MarkdownDescription = "Timeouts for " + r.Name + " operations."
	for name, duration := range map[string]time.Duration{
		"create": defaults.Create, "read": defaults.Read, "update": defaults.Update, "delete": defaults.Delete,
	} {
		attribute, ok := block.Attributes[name].(schema.StringAttribute)
		if !ok {
			continue
		}
		for i, rule := range attribute.Validators {
			attribute.Validators[i] = DescribeString(rule, "The format is a duration, for example, `30s` or `10m`.")
		}
		// The default is applied at runtime, not by a Terraform attribute default.
		block.Attributes[name] = r.String(attribute, Field{
			Definition: "Timeout for the " + name + " operation.",
			Behavior:   "If omitted, the timeout is `" + shortDuration(duration) + "`.",
		})
	}

	return block
}

// shortDuration drops zero minutes and seconds: 10m0s becomes 10m, 1h0m0s becomes 1h.
func shortDuration(duration time.Duration) string {
	text := duration.String()
	if strings.HasSuffix(text, "m0s") {
		text = strings.TrimSuffix(text, "0s")
	}
	if strings.HasSuffix(text, "h0m") {
		text = strings.TrimSuffix(text, "0m")
	}

	return text
}
