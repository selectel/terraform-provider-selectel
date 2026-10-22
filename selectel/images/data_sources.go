package images

import (
	"context"
	"regexp"
	"slices"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	sdk "github.com/selectel/images-go/pkg/v1"
	"github.com/selectel/images-go/pkg/v1/image"
	"github.com/terraform-providers/terraform-provider-selectel/selectel/internal/schemadoc"
)

// dataConnection provides the configuration boundary to both image lookups.
type dataConnection struct{ connection }

func (d *dataConnection) Configure(
	_ context.Context,
	req datasource.ConfigureRequest,
	resp *datasource.ConfigureResponse,
) {
	resp.Diagnostics.Append(d.configure(req.ProviderData)...)
}

// imageFilter is the search criteria shared by the image data sources.
type imageFilter struct {
	Name            types.String `tfsdk:"name"`
	NameRegex       types.String `tfsdk:"name_regex"`
	Visibility      types.String `tfsdk:"visibility"`
	MemberStatus    types.String `tfsdk:"member_status"`
	Owner           types.String `tfsdk:"owner"`
	SizeMin         types.Int64  `tfsdk:"size_min"`
	SizeMax         types.Int64  `tfsdk:"size_max"`
	Tag             types.String `tfsdk:"tag"`
	Tags            types.Set    `tfsdk:"tags"`
	Properties      types.Map    `tfsdk:"properties"`
	Hidden          types.Bool   `tfsdk:"hidden"`
	ContainerFormat types.String `tfsdk:"container_format"`
	DiskFormat      types.String `tfsdk:"disk_format"`
}

// filterAttributes are the arguments of imageFilter. When returnsSelected is
// set, the data source selects one image, and a filter that describes the
// image also returns its value.
func filterAttributes(docs schemadoc.DataSource, returnsSelected bool) map[string]schema.Attribute {
	// selected completes the behavior of a filter that returns a value.
	selected := func(behavior, returned string) string {
		if !returnsSelected {
			return behavior
		}
		if behavior == "" {
			return returned
		}

		return behavior + " " + returned
	}
	omitted := func(what string) string {
		return "If omitted, the data source returns the " + what + " of the selected image."
	}
	appliedByProvider := "The provider applies this filter to the images found by the other filters."

	return map[string]schema.Attribute{
		"name": docs.String(schema.StringAttribute{
			Optional:   true,
			Computed:   returnsSelected,
			Validators: []validator.String{schemadoc.LengthAtLeast(1), schemadoc.ConflictsWith("name_regex")},
		}, schemadoc.Field{
			Definition: "Exact image name to search for.",
			Behavior:   selected("", omitted("name")),
			Source:     schemadoc.FromOpenStackCLI("image name", "openstack image list", "Name"),
		}),
		"name_regex": docs.String(schema.StringAttribute{
			Optional:   true,
			Validators: []validator.String{validRegexp(), schemadoc.ConflictsWith("name")},
		}, schemadoc.Field{
			Definition: "Regular expression that the image name must match.",
			Behavior:   appliedByProvider,
		}),
		"visibility": docs.String(schema.StringAttribute{
			Optional:   true,
			Computed:   returnsSelected,
			Validators: []validator.String{schemadoc.OneOf("public", "private", "shared", "community")},
		}, schemadoc.Field{
			Definition: "Visibility of the image to search for.",
			Behavior: selected(
				"If omitted, the data source searches public images, images of the project, and shared images that the project accepted. Community images of other projects are found only with `visibility = \"community\"`.",
				"The data source returns the visibility of the selected image.",
			),
			Concept: sharingLink,
		}),
		"member_status": docs.String(schema.StringAttribute{
			Optional:   true,
			Validators: []validator.String{schemadoc.OneOf("accepted", "pending", "rejected", "all")},
		}, schemadoc.Field{
			Definition: "Status of the project access to shared images to search for.",
			Behavior:   "If omitted, only shared images with accepted access are found.",
		}),
		"owner": docs.String(schema.StringAttribute{
			Optional:   true,
			Computed:   returnsSelected,
			Validators: []validator.String{schemadoc.LengthAtLeast(1)},
		}, schemadoc.Field{
			Definition: "Unique identifier of the project that owns the image.",
			Behavior:   selected("", omitted("owner")),
			Source:     schemadoc.FromOpenStackCLI("owner ID", "openstack image list --long", "Project"),
		}),
		"size_min": docs.Int64(schema.Int64Attribute{
			Optional:   true,
			Validators: []validator.Int64{schemadoc.AtLeast(0)},
		}, schemadoc.Field{
			Definition: "Minimum size of the image data in bytes.",
		}),
		"size_max": docs.Int64(schema.Int64Attribute{
			Optional:   true,
			Validators: []validator.Int64{schemadoc.AtLeast(0)},
		}, schemadoc.Field{
			Definition: "Maximum size of the image data in bytes.",
		}),
		"tag": docs.String(schema.StringAttribute{
			Optional:   true,
			Validators: []validator.String{schemadoc.LengthAtLeast(1)},
		}, schemadoc.Field{
			Definition: "Tag that the image must have.",
			Behavior:   "Use `tags` to search by several tags. Both arguments can be set.",
		}),
		"tags": docs.Set(schema.SetAttribute{
			Optional:    true,
			Computed:    returnsSelected,
			ElementType: types.StringType,
		}, schemadoc.Field{
			Definition: "Set of tags that the image must have.",
			Behavior:   selected("", "The data source returns all tags of the selected image."),
		}),
		"properties": docs.Map(schema.MapAttribute{
			Optional:    true,
			Computed:    returnsSelected,
			ElementType: types.StringType,
		}, schemadoc.Field{
			Definition: "Key-value pairs that the image properties must contain.",
			Behavior:   selected(appliedByProvider, "The data source returns all properties of the selected image."),
		}),
		"hidden": docs.Bool(schema.BoolAttribute{
			Optional: true,
			Computed: returnsSelected,
		}, schemadoc.Field{
			Definition: "Specifies if the data source searches hidden images.",
			Behavior: selected(
				"If omitted or `false`, only images that are not hidden are found. If `true`, only hidden images are found.",
				"The data source returns whether the selected image is hidden.",
			),
		}),
		"container_format": docs.String(schema.StringAttribute{
			Optional:   true,
			Computed:   returnsSelected,
			Validators: []validator.String{schemadoc.OneOf(containerFormats...)},
		}, schemadoc.Field{
			Definition: "Container format of the image to search for.",
			Behavior:   selected("", omitted("container format")),
			Concept:    formatsLink,
		}),
		"disk_format": docs.String(schema.StringAttribute{
			Optional:   true,
			Computed:   returnsSelected,
			Validators: []validator.String{schemadoc.OneOf(diskFormats...)},
		}, schemadoc.Field{
			Definition: "Disk format of the image to search for.",
			Behavior:   selected("", omitted("disk format")),
			Concept:    formatsLink,
		}),
	}
}

// findImages returns the active images that match every filter, in the
// order of sort. The API filters only narrow the list: they compare strings
// case-insensitively, and the API keeps only one tag filter on the following
// pages. So the provider checks the filters exactly.
func findImages(ctx context.Context, client *sdk.Client, filter imageFilter, sort string) ([]image.View, diag.Diagnostics) {
	tags, diags := setStrings(ctx, filter.Tags)
	if !filter.Tag.IsNull() {
		tags = append(tags, filter.Tag.ValueString())
	}
	properties, d := mapStrings(ctx, filter.Properties)
	diags.Append(d...)
	if diags.HasError() {
		return nil, diags
	}
	// validRegexp has already checked the expression. A null name_regex is an
	// empty expression, which matches every name.
	nameRegex := regexp.MustCompile(filter.NameRegex.ValueString())
	opts := image.ListOpts{
		Name:            filter.Name.ValueString(),
		Visibility:      filter.Visibility.ValueString(),
		MemberStatus:    filter.MemberStatus.ValueString(),
		Owner:           filter.Owner.ValueString(),
		Status:          image.StatusActive,
		SizeMin:         filter.SizeMin.ValueInt64(),
		SizeMax:         filter.SizeMax.ValueInt64(),
		Sort:            sort,
		ContainerFormat: filter.ContainerFormat.ValueString(),
		DiskFormat:      filter.DiskFormat.ValueString(),
		Tags:            tags,
	}
	if !filter.Hidden.IsNull() {
		hidden := filter.Hidden.ValueBool()
		opts.Hidden = &hidden
	}
	all, err := image.List(ctx, client, opts)
	if err != nil {
		diags.AddError("Find images", err.Error())
		return nil, diags
	}
	found := slices.DeleteFunc(all, func(candidate image.View) bool {
		return !matchesFilter(filter.Name, candidate.Name) ||
			!nameRegex.MatchString(candidate.Name) ||
			!matchesFilter(filter.Owner, candidate.Owner) ||
			!matchesFilter(filter.ContainerFormat, candidate.ContainerFormat) ||
			!matchesFilter(filter.DiskFormat, candidate.DiskFormat) ||
			!hasAll(candidate.Tags, tags) ||
			!containsProperties(candidate.Properties, properties)
	})

	return found, diags
}

// matchesFilter applies a string filter to a found image; an omitted filter
// matches any value.
func matchesFilter(filter types.String, value string) bool {
	return filter.IsNull() || filter.ValueString() == value
}

func hasAll(tags, wanted []string) bool {
	for _, tag := range wanted {
		if !slices.Contains(tags, tag) {
			return false
		}
	}

	return true
}

func containsProperties(properties, wanted map[string]string) bool {
	for key, value := range wanted {
		if actual, exists := properties[key]; !exists || actual != value {
			return false
		}
	}

	return true
}
