package selectel

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/resource/identityschema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/defaults"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
)

func (r resourceDocs) idFrameworkResourceSchema() schema.StringAttribute {
	return schema.StringAttribute{
		Computed:      true,
		Description:   r.idDescription() + ".",
		PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
	}
}

func (r resourceDocs) idFrameworkIdentitySchema(controlPanelHint string) identityschema.StringAttribute {
	return identityschema.StringAttribute{
		RequiredForImport: true,
		Description:       fmt.Sprintf("%s, for example, `%s`. %s", r.idDescription(), exampleResourceID, controlPanelHint),
	}
}

func (r resourceDocs) regionFrameworkResourceSchema() schema.StringAttribute {
	return schema.StringAttribute{
		Required:      true,
		Description:   r.regionDescription() + " " + regionLearnMore,
		PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
	}
}

func (r resourceDocs) regionFrameworkIdentitySchema(controlPanelHint string) identityschema.StringAttribute {
	return identityschema.StringAttribute{
		RequiredForImport: true,
		Description:       r.regionDescription() + " " + controlPanelHint + " " + regionLearnMore,
	}
}

func projectIDFrameworkResourceSchema() schema.StringAttribute {
	return schema.StringAttribute{
		Required:      true,
		Description:   projectIDDescription + " " + projectIDFromResource + " " + projectIDLearnMore,
		PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
	}
}

func projectIDFrameworkIdentitySchema() identityschema.StringAttribute {
	return identityschema.StringAttribute{
		RequiredForImport: true,
		Description:       projectIDDescription + " " + projectIDFromControlPanel + " " + projectIDLearnMore,
	}
}

// withFrameworkDocsHints is the framework version of withDocsHints.
func (r resourceDocs) withFrameworkDocsHints(attrs map[string]schema.Attribute) map[string]schema.Attribute {
	ctx := context.Background()

	for name, attr := range attrs {
		switch a := attr.(type) {
		case schema.BoolAttribute:
			appendHint(&a.Description, &a.MarkdownDescription, replaceHint(ctx, r.Name, a.PlanModifiers)+frameworkDefaultHint(ctx, a.Default))
			attrs[name] = a
		case schema.DynamicAttribute:
			appendHint(&a.Description, &a.MarkdownDescription, replaceHint(ctx, r.Name, a.PlanModifiers)+frameworkDefaultHint(ctx, a.Default))
			attrs[name] = a
		case schema.Float32Attribute:
			appendHint(&a.Description, &a.MarkdownDescription, replaceHint(ctx, r.Name, a.PlanModifiers)+frameworkDefaultHint(ctx, a.Default))
			attrs[name] = a
		case schema.Float64Attribute:
			appendHint(&a.Description, &a.MarkdownDescription, replaceHint(ctx, r.Name, a.PlanModifiers)+frameworkDefaultHint(ctx, a.Default))
			attrs[name] = a
		case schema.Int32Attribute:
			appendHint(&a.Description, &a.MarkdownDescription, replaceHint(ctx, r.Name, a.PlanModifiers)+frameworkDefaultHint(ctx, a.Default))
			attrs[name] = a
		case schema.Int64Attribute:
			appendHint(&a.Description, &a.MarkdownDescription, replaceHint(ctx, r.Name, a.PlanModifiers)+frameworkDefaultHint(ctx, a.Default))
			attrs[name] = a
		case schema.ListAttribute:
			appendHint(&a.Description, &a.MarkdownDescription, replaceHint(ctx, r.Name, a.PlanModifiers)+frameworkDefaultHint(ctx, a.Default))
			attrs[name] = a
		case schema.MapAttribute:
			appendHint(&a.Description, &a.MarkdownDescription, replaceHint(ctx, r.Name, a.PlanModifiers)+frameworkDefaultHint(ctx, a.Default))
			attrs[name] = a
		case schema.NumberAttribute:
			appendHint(&a.Description, &a.MarkdownDescription, replaceHint(ctx, r.Name, a.PlanModifiers)+frameworkDefaultHint(ctx, a.Default))
			attrs[name] = a
		case schema.ObjectAttribute:
			appendHint(&a.Description, &a.MarkdownDescription, replaceHint(ctx, r.Name, a.PlanModifiers)+frameworkDefaultHint(ctx, a.Default))
			attrs[name] = a
		case schema.SetAttribute:
			appendHint(&a.Description, &a.MarkdownDescription, replaceHint(ctx, r.Name, a.PlanModifiers)+frameworkDefaultHint(ctx, a.Default))
			attrs[name] = a
		case schema.StringAttribute:
			appendHint(&a.Description, &a.MarkdownDescription, replaceHint(ctx, r.Name, a.PlanModifiers)+frameworkDefaultHint(ctx, a.Default))
			attrs[name] = a
		case schema.ListNestedAttribute:
			appendHint(&a.Description, &a.MarkdownDescription, replaceHint(ctx, r.Name, a.PlanModifiers)+frameworkDefaultHint(ctx, a.Default))
			r.withFrameworkDocsHints(a.NestedObject.Attributes)
			attrs[name] = a
		case schema.MapNestedAttribute:
			appendHint(&a.Description, &a.MarkdownDescription, replaceHint(ctx, r.Name, a.PlanModifiers)+frameworkDefaultHint(ctx, a.Default))
			r.withFrameworkDocsHints(a.NestedObject.Attributes)
			attrs[name] = a
		case schema.SetNestedAttribute:
			appendHint(&a.Description, &a.MarkdownDescription, replaceHint(ctx, r.Name, a.PlanModifiers)+frameworkDefaultHint(ctx, a.Default))
			r.withFrameworkDocsHints(a.NestedObject.Attributes)
			attrs[name] = a
		case schema.SingleNestedAttribute:
			appendHint(&a.Description, &a.MarkdownDescription, replaceHint(ctx, r.Name, a.PlanModifiers)+frameworkDefaultHint(ctx, a.Default))
			r.withFrameworkDocsHints(a.Attributes)
			attrs[name] = a
		}
	}

	return attrs
}

// withFrameworkBlockDocsHints covers nested blocks, which SDKv2 declares as
// attributes with an Elem resource.
func (r resourceDocs) withFrameworkBlockDocsHints(blocks map[string]schema.Block) map[string]schema.Block {
	ctx := context.Background()

	for name, block := range blocks {
		switch b := block.(type) {
		case schema.ListNestedBlock:
			appendHint(&b.Description, &b.MarkdownDescription, replaceHint(ctx, r.Name, b.PlanModifiers))
			r.withFrameworkDocsHints(b.NestedObject.Attributes)
			r.withFrameworkBlockDocsHints(b.NestedObject.Blocks)
			blocks[name] = b
		case schema.SetNestedBlock:
			appendHint(&b.Description, &b.MarkdownDescription, replaceHint(ctx, r.Name, b.PlanModifiers))
			r.withFrameworkDocsHints(b.NestedObject.Attributes)
			r.withFrameworkBlockDocsHints(b.NestedObject.Blocks)
			blocks[name] = b
		case schema.SingleNestedBlock:
			appendHint(&b.Description, &b.MarkdownDescription, replaceHint(ctx, r.Name, b.PlanModifiers))
			r.withFrameworkDocsHints(b.Attributes)
			r.withFrameworkBlockDocsHints(b.Blocks)
			blocks[name] = b
		}
	}

	return blocks
}

// frameworkDefaultHint is defaultHint for framework defaults. Collections, which
// SDKv2 cannot default, are rendered the way the framework prints them.
func frameworkDefaultHint(ctx context.Context, d any) string {
	switch d := d.(type) {
	case defaults.Bool:
		resp := &defaults.BoolResponse{}
		d.DefaultBool(ctx, defaults.BoolRequest{}, resp)

		return defaultHint(resp.PlanValue.ValueBool())
	case defaults.Float32:
		resp := &defaults.Float32Response{}
		d.DefaultFloat32(ctx, defaults.Float32Request{}, resp)

		return defaultHint(resp.PlanValue.ValueFloat32())
	case defaults.Float64:
		resp := &defaults.Float64Response{}
		d.DefaultFloat64(ctx, defaults.Float64Request{}, resp)

		return defaultHint(resp.PlanValue.ValueFloat64())
	case defaults.Int32:
		resp := &defaults.Int32Response{}
		d.DefaultInt32(ctx, defaults.Int32Request{}, resp)

		return defaultHint(resp.PlanValue.ValueInt32())
	case defaults.Int64:
		resp := &defaults.Int64Response{}
		d.DefaultInt64(ctx, defaults.Int64Request{}, resp)

		return defaultHint(resp.PlanValue.ValueInt64())
	case defaults.Number:
		resp := &defaults.NumberResponse{}
		d.DefaultNumber(ctx, defaults.NumberRequest{}, resp)

		return defaultHint(resp.PlanValue.ValueBigFloat())
	case defaults.String:
		resp := &defaults.StringResponse{}
		d.DefaultString(ctx, defaults.StringRequest{}, resp)

		return defaultHint(resp.PlanValue.ValueString())
	case defaults.Dynamic:
		resp := &defaults.DynamicResponse{}
		d.DefaultDynamic(ctx, defaults.DynamicRequest{}, resp)

		return defaultHint(resp.PlanValue.String())
	case defaults.List:
		resp := &defaults.ListResponse{}
		d.DefaultList(ctx, defaults.ListRequest{}, resp)

		return defaultHint(resp.PlanValue.String())
	case defaults.Map:
		resp := &defaults.MapResponse{}
		d.DefaultMap(ctx, defaults.MapRequest{}, resp)

		return defaultHint(resp.PlanValue.String())
	case defaults.Object:
		resp := &defaults.ObjectResponse{}
		d.DefaultObject(ctx, defaults.ObjectRequest{}, resp)

		return defaultHint(resp.PlanValue.String())
	case defaults.Set:
		resp := &defaults.SetResponse{}
		d.DefaultSet(ctx, defaults.SetRequest{}, resp)

		return defaultHint(resp.PlanValue.String())
	}

	return ""
}

// appendHint adds the hint to both descriptions: the framework serves
// MarkdownDescription instead of Description when it is set.
func appendHint(description, markdownDescription *string, hint string) {
	*description += hint
	if *markdownDescription != "" {
		*markdownDescription += hint
	}
}

type describer interface {
	Description(ctx context.Context) string
}

func replaceHint[M describer](ctx context.Context, name string, modifiers []M) string {
	for _, m := range modifiers {
		// The framework has no ForceNew flag; RequiresReplace is recognised by
		// its stock description, so a custom RequiresReplaceIf text is missed.
		if strings.Contains(m.Description(ctx), "Terraform will destroy and recreate the resource") {
			return fmt.Sprintf(" Changing this creates a new %s.", name)
		}
	}

	return ""
}
