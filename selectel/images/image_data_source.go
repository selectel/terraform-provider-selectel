package images

import (
	"context"
	"maps"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/selectel/images-go/pkg/v1/image"
	"github.com/terraform-providers/terraform-provider-selectel/selectel/internal/schemadoc"
)

type imageDataSource struct{ dataConnection }

func NewImageDataSource() datasource.DataSource { return &imageDataSource{} }

func (d *imageDataSource) Metadata(
	_ context.Context,
	req datasource.MetadataRequest,
	resp *datasource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_compute_image_v2"
}

type imageDataModel struct {
	ID        types.String `tfsdk:"id"`
	ProjectID types.String `tfsdk:"project_id"`
	Region    types.String `tfsdk:"region"`
	imageFilter
	MostRecent types.Bool   `tfsdk:"most_recent"`
	Protected  types.Bool   `tfsdk:"protected"`
	MinDiskGB  types.Int64  `tfsdk:"min_disk_gb"`
	MinRAMMB   types.Int64  `tfsdk:"min_ram_mb"`
	SizeBytes  types.Int64  `tfsdk:"size_bytes"`
	Checksum   types.String `tfsdk:"checksum"`
	CreatedAt  types.String `tfsdk:"created_at"`
	UpdatedAt  types.String `tfsdk:"updated_at"`
}

func (d *imageDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	docs := schemadoc.DataSource{Name: "image"}
	attributes := map[string]schema.Attribute{
		"id":         docs.ID(),
		"project_id": docs.ProjectID(),
		"region":     docs.Region(),
		"most_recent": schema.BoolAttribute{
			Optional:            true,
			MarkdownDescription: "Specifies if the data source selects the image with the latest `created_at` timestamp when several images match. Equal timestamps are resolved by the lexicographically smallest `id`. If omitted or `false`, the search must return exactly one image.",
		},
		"protected":   schema.BoolAttribute{Computed: true, Description: "Whether the image is protected from deletion."},
		"min_disk_gb": schema.Int64Attribute{Computed: true, Description: "Minimum disk size in GB required to boot the image."},
		"min_ram_mb":  schema.Int64Attribute{Computed: true, Description: "Minimum RAM in MB required to boot the image."},
		"size_bytes":  schema.Int64Attribute{Computed: true, Description: "Size of the image data in bytes."},
		"checksum":    schema.StringAttribute{Computed: true, Description: "MD5 checksum of the image data."},
		"created_at":  schema.StringAttribute{Computed: true, Description: "Time when the image was created in the RFC 3339 format."},
		"updated_at":  schema.StringAttribute{Computed: true, Description: "Time when the image was last updated in the RFC 3339 format."},
	}
	maps.Copy(attributes, filterAttributes(docs, true))
	resp.Schema = schema.Schema{
		MarkdownDescription: "Provides an image in Cloud Servers. The data source searches active images visible to the project and returns exactly one. A search that returns no images fails. A search that returns several images fails unless `most_recent` is `true`. For more information about images, see the [official Selectel documentation](" + aboutImages + ").",
		Attributes:          attributes,
	}
}

func (d *imageDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config imageDataModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}
	ctx, cancel := context.WithTimeout(ctx, requestTimeout)
	defer cancel()
	client, err := d.client(ctx, config.ProjectID.ValueString(), config.Region.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Connect to the Image service", err.Error())
		return
	}
	// The selection below does not depend on the order, so the request
	// keeps the default sort of the API.
	found, diags := findImages(ctx, client, config.imageFilter, "")
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	if len(found) == 0 {
		resp.Diagnostics.AddError("Image not found", "No active images matched the configured filters.")
		return
	}
	if len(found) > 1 && !config.MostRecent.ValueBool() {
		resp.Diagnostics.AddError("Ambiguous image lookup", "Refine filters or set most_recent = true.")
		return
	}
	selected := found[0]
	for _, candidate := range found[1:] {
		if candidate.CreatedAt.After(selected.CreatedAt) ||
			(candidate.CreatedAt.Equal(selected.CreatedAt) && candidate.ID < selected.ID) {
			selected = candidate
		}
	}
	resp.Diagnostics.Append(config.observe(ctx, &selected)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}

// observe writes the selected image into the model. The filters that also
// describe the image take its values.
func (m *imageDataModel) observe(ctx context.Context, selected *image.View) diag.Diagnostics {
	var diags, d diag.Diagnostics
	m.ID = types.StringValue(selected.ID)
	m.Name = types.StringValue(selected.Name)
	m.Visibility = types.StringValue(selected.Visibility)
	m.Owner = types.StringValue(selected.Owner)
	m.Hidden = types.BoolValue(selected.Hidden)
	m.ContainerFormat = types.StringValue(selected.ContainerFormat)
	m.DiskFormat = types.StringValue(selected.DiskFormat)
	m.Tags, d = types.SetValueFrom(ctx, types.StringType, selected.Tags)
	diags.Append(d...)
	m.Properties, d = types.MapValueFrom(ctx, types.StringType, selected.Properties)
	diags.Append(d...)
	m.Protected = types.BoolValue(selected.Protected)
	m.MinDiskGB = types.Int64Value(int64(selected.MinDisk))
	m.MinRAMMB = types.Int64Value(int64(selected.MinRAM))
	m.SizeBytes = types.Int64Value(selected.Size)
	m.Checksum = types.StringValue(selected.Checksum)
	m.CreatedAt = types.StringValue(selected.CreatedAt.Format(time.RFC3339))
	m.UpdatedAt = types.StringValue(selected.UpdatedAt.Format(time.RFC3339))

	return diags
}
