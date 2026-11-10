package images

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"maps"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/terraform-providers/terraform-provider-selectel/selectel/internal/schemadoc"
)

type imageIDsDataSource struct{ dataConnection }

func NewImageIDsDataSource() datasource.DataSource { return &imageIDsDataSource{} }

func (d *imageIDsDataSource) Metadata(
	_ context.Context,
	req datasource.MetadataRequest,
	resp *datasource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_compute_image_ids_v2"
}

type imageIDsDataModel struct {
	ID        types.String `tfsdk:"id"`
	ProjectID types.String `tfsdk:"project_id"`
	Region    types.String `tfsdk:"region"`
	imageFilter
	Sort types.String `tfsdk:"sort"`
	IDs  types.List   `tfsdk:"ids"`
}

// defaultSort applies when sort is omitted. A data source attribute has no
// Default, so Read applies it.
const defaultSort = "name:asc"

func (d *imageIDsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	docs := schemadoc.DataSource{Name: "image"}
	attributes := map[string]schema.Attribute{
		"id": schema.StringAttribute{
			Computed:    true,
			Description: "Identifier of the search result derived from the list of image IDs.",
		},
		"project_id": docs.ProjectID(),
		"region":     docs.Region(),
		"sort": docs.String(schema.StringAttribute{
			Optional:   true,
			Validators: []validator.String{sortOrder()},
		}, schemadoc.Field{
			Definition: "Sort order of the found images.",
			Behavior:   "If omitted, the sort order is `" + defaultSort + "`.",
		}),
		"ids": docs.List(schema.ListAttribute{
			Computed:    true,
			ElementType: types.StringType,
		}, schemadoc.Field{
			Definition: "List of unique identifiers of the found images in the `sort` order.",
		}),
	}
	maps.Copy(attributes, filterAttributes(docs, false))
	resp.Schema = schema.Schema{
		MarkdownDescription: "Provides a list of image IDs in Cloud Servers. The data source searches active images visible to the project. A search that returns no images gives an empty list. For more information about images, see the [official Selectel documentation](" + aboutImages + ").",
		Attributes:          attributes,
	}
}

func (d *imageIDsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config imageIDsDataModel
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
	sort := defaultSort
	if !config.Sort.IsNull() {
		sort = config.Sort.ValueString()
	}
	found, diags := findImages(ctx, client, config.imageFilter, sort)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	ids := make([]string, len(found))
	for i, candidate := range found {
		ids[i] = candidate.ID
	}
	digest := sha256.Sum256([]byte(strings.Join(ids, ",")))
	config.ID = types.StringValue(hex.EncodeToString(digest[:]))
	config.IDs, diags = types.ListValueFrom(ctx, types.StringType, ids)
	resp.Diagnostics.Append(diags...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}
