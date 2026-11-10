package filestorage

import (
	"context"
	"fmt"
	"maps"
	"slices"
	"sort"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	sdk "github.com/selectel/filestorage-go/pkg/v1"
	"github.com/selectel/filestorage-go/pkg/v1/availabilityzone"
	"github.com/selectel/filestorage-go/pkg/v1/share"
	"github.com/terraform-providers/terraform-provider-selectel/selectel/internal/schemadoc"
)

// dataConnection provides the identical configuration boundary to the three
// lookups; each lookup keeps its own schema, model, selection and state write.
type dataConnection struct{ connection }

func (d *dataConnection) Configure(
	_ context.Context,
	req datasource.ConfigureRequest,
	resp *datasource.ConfigureResponse,
) {
	resp.Diagnostics.Append(d.configure(req.ProviderData)...)
}

// filterString is a lookup selector: an exact match that, when omitted,
// returns the value of the found object.
func filterString(validators ...validator.String) schema.StringAttribute {
	return schema.StringAttribute{
		Optional: true, Computed: true,
		Validators: append([]validator.String{schemadoc.LengthAtLeast(1)}, validators...),
	}
}

type shareDataSource struct{ dataConnection }

func NewShareDataSource() datasource.DataSource { return &shareDataSource{} }

type shareDataModel struct {
	ID                 types.String `tfsdk:"id"`
	ProjectID          types.String `tfsdk:"project_id"`
	Region             types.String `tfsdk:"region"`
	ShareID            types.String `tfsdk:"share_id"`
	Name               types.String `tfsdk:"name"`
	Description        types.String `tfsdk:"description"`
	Status             types.String `tfsdk:"status"`
	Metadata           types.Map    `tfsdk:"metadata"`
	ExportLocationPath types.String `tfsdk:"export_location_path"`
	ShareProto         types.String `tfsdk:"share_proto"`
	ShareType          types.String `tfsdk:"share_type"`
	Size               types.Int64  `tfsdk:"size"`
	NetworkID          types.String `tfsdk:"network_id"`
	SubnetID           types.String `tfsdk:"subnet_id"`
	IP                 types.String `tfsdk:"ip"`
	AllMetadata        types.Map    `tfsdk:"all_metadata"`
	AvailabilityZone   types.String `tfsdk:"availability_zone"`
	ExportLocations    types.List   `tfsdk:"export_locations"`
}

func (d *shareDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_filestorage_share_v2"
}

func (d *shareDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	docs := schemadoc.DataSource{Name: "file storage", Location: shareLocation}
	byID := schemadoc.ConflictsWith("share_id")
	computed := func(description string) schema.StringAttribute {
		return schema.StringAttribute{Computed: true, Description: description}
	}
	resp.Schema = schema.Schema{
		MarkdownDescription: "Provides a file storage in File Storage. The data source finds the file storage by `share_id` or exactly one file storage of the project that matches all set selectors. Without `share_id` and selectors, the project must have exactly one file storage in the pool. For more information about file storage, see the [official Selectel documentation](" + aboutFileStorage + ").",
		Attributes: map[string]schema.Attribute{
			"id":         docs.ID(),
			"project_id": docs.ProjectID(),
			"region":     docs.Region(),
			"share_id": docs.String(schema.StringAttribute{
				Optional: true,
				Validators: []validator.String{
					schemadoc.UUID(),
					schemadoc.ConflictsWith("name", "description", "status", "export_location_path", "metadata"),
				},
			}, schemadoc.Field{
				Definition: "Unique identifier of the file storage.",
				Source:     schemadoc.FromOpenStackCLI("file storage ID", "openstack share list", "ID"),
			}),
			"name": docs.String(filterString(byID), schemadoc.Field{
				Definition: "Exact name of the file storage.",
				Behavior:   "When omitted, it returns the name of the found file storage.",
				Source:     schemadoc.FromOpenStackCLI("file storage name", "openstack share list", "Name"),
			}),
			"description": docs.String(filterString(byID), schemadoc.Field{
				Definition: "Exact description of the file storage.",
				Behavior:   "When omitted, it returns the description of the found file storage.",
			}),
			"status": docs.String(filterString(byID), schemadoc.Field{
				Definition: "Status of the file storage, for example, `available`.",
				Behavior:   "When omitted, it returns the status of the found file storage.",
			}),
			"export_location_path": docs.String(schema.StringAttribute{
				Optional:   true,
				Validators: []validator.String{schemadoc.LengthAtLeast(1), byID},
			}, schemadoc.Field{
				Definition: "Mount path of the file storage, for example, `192.168.199.5:/shares/share-<id>`.",
			}),
			"metadata": docs.Map(schema.MapAttribute{
				Optional:    true,
				Computed:    true,
				ElementType: types.StringType,
				Validators:  []validator.Map{byID},
			}, schemadoc.Field{
				Definition: "Key-value pairs of the file storage.",
				Behavior: "As a selector, it matches a file storage that has all the pairs among all its keys, including the keys of the service. " +
					"When omitted, it returns the keys of the found file storage without the reserved keys, so you can copy it to `selectel_filestorage_share_v2`.",
			}),
			"share_proto":       computed("File storage protocol."),
			"share_type":        computed("Name of the file storage type, in the form that the `share_type` argument of `selectel_filestorage_share_v2` takes."),
			"size":              schema.Int64Attribute{Computed: true, Description: "Size of the file storage in GB."},
			"network_id":        computed("Unique identifier of the private network of the file storage."),
			"subnet_id":         computed("Unique identifier of the private subnet of the file storage."),
			"ip":                computed("IP address of the file storage in the subnet."),
			"availability_zone": computed("Pool segment of the file storage."),
			"all_metadata": schema.MapAttribute{
				Computed:    true,
				ElementType: types.StringType,
				Description: "All metadata of the file storage, including the keys added by the service.",
			},
			"export_locations": schema.ListNestedAttribute{
				Computed:    true,
				Description: "Paths to mount the file storage.",
				NestedObject: schema.NestedAttributeObject{Attributes: map[string]schema.Attribute{
					"path":      computed("Mount path."),
					"preferred": schema.BoolAttribute{Computed: true, Description: "Whether the API recommends this path."},
				}},
			},
		},
	}
}

func (d *shareDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config shareDataModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	filters := map[string]string{}
	if !config.Metadata.IsNull() {
		resp.Diagnostics.Append(config.Metadata.ElementsAs(ctx, &filters, false)...)
	}
	if resp.Diagnostics.HasError() {
		return
	}
	ctx, cancel := context.WithTimeout(ctx, readTimeout)
	defer cancel()
	region := config.Region.ValueString()
	client, err := d.client(ctx, config.ProjectID.ValueString(), region)
	if err != nil {
		resp.Diagnostics.AddError("Connect to File Storage", err.Error())
		return
	}
	found, ok := d.find(ctx, client, config, filters, &resp.Diagnostics)
	if !ok {
		return
	}
	observed, err := observeShare(ctx, client, found)
	if err != nil {
		resp.Diagnostics.Append(apiError("Read file storage", region, err))
		return
	}
	resp.Diagnostics.Append(config.observe(ctx, observed)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}

// find returns the file storage with share_id, or the only one that matches
// the selectors.
func (d *shareDataSource) find(
	ctx context.Context, client *sdk.Client, config shareDataModel, filters map[string]string, diags *diag.Diagnostics,
) (*share.View, bool) {
	region := config.Region.ValueString()
	if !config.ShareID.IsNull() {
		id := config.ShareID.ValueString()
		found, _, err := share.Get(ctx, client, id)
		switch {
		case sdk.IsKind(err, sdk.KindNotFound):
			diags.AddError("File storage not found", fmt.Sprintf("No file storage %s is in the project in pool %s.", id, region))
			return nil, false
		case err != nil:
			diags.Append(apiError("Read file storage", region, err))
			return nil, false
		}

		return found, true
	}
	// The API filters the list; the exact comparison below guards against a
	// filter that the API matches more loosely.
	all, err := share.ListDetail(ctx, client, share.ListOpts{
		Name:               config.Name.ValueString(),
		Description:        config.Description.ValueString(),
		Status:             config.Status.ValueString(),
		ExportLocationPath: config.ExportLocationPath.ValueString(),
		Metadata:           filters,
	})
	if err != nil {
		diags.Append(apiError("Find file storage", region, err))
		return nil, false
	}
	var matches []*share.View
	for i := range all {
		candidate := &all[i]
		if matchesFilter(config.Name, candidate.Name) && matchesFilter(config.Description, candidate.Description) &&
			matchesFilter(config.Status, candidate.Status) && hasPairs(candidate.Metadata, filters) {
			matches = append(matches, candidate)
		}
	}
	switch len(matches) {
	case 0:
		diags.AddError("File storage not found", "No file storage of the project matches the selectors.")
		return nil, false
	case 1:
		return matches[0], true
	default:
		diags.AddError("Multiple file storages found",
			fmt.Sprintf("%d file storages match the selectors. Set more selectors or share_id to find exactly one.", len(matches)))

		return nil, false
	}
}

// observe writes the found file storage into the model. Omitted metadata
// returns the keys of the file storage without the reserved ones.
func (m *shareDataModel) observe(ctx context.Context, observed *shareObservation) diag.Diagnostics {
	var found shareModel
	diags := found.observe(ctx, observed, types.MapNull(types.StringType))
	m.ID = found.ID
	m.Name = found.Name
	m.Description = found.Description
	m.Status = found.Status
	m.ShareProto = found.ShareProto
	m.ShareType = found.ShareType
	m.Size = found.Size
	m.NetworkID = found.NetworkID
	m.SubnetID = found.SubnetID
	m.IP = found.IP
	m.AllMetadata = found.AllMetadata
	m.AvailabilityZone = found.AvailabilityZone
	m.ExportLocations = found.ExportLocations
	if m.Metadata.IsNull() {
		clientKeys := maps.Clone(observed.share.Metadata)
		maps.DeleteFunc(clientKeys, func(key, _ string) bool { return slices.Contains(reservedMetadataKeys, key) })
		var d diag.Diagnostics
		m.Metadata, d = types.MapValueFrom(ctx, types.StringType, nonNil(clientKeys))
		diags.Append(d...)
	}

	return diags
}

// matchesFilter applies a filterString to a value of a found object; an omitted
// filter matches any value.
func matchesFilter(filter types.String, value string) bool {
	return filter.IsNull() || filter.ValueString() == value
}

func hasPairs(metadata, pairs map[string]string) bool {
	for key, value := range pairs {
		if current, ok := metadata[key]; !ok || current != value {
			return false
		}
	}

	return true
}

func nonNil(values map[string]string) map[string]string {
	if values == nil {
		return map[string]string{}
	}

	return values
}

type availabilityZonesDataSource struct{ dataConnection }

func NewAvailabilityZonesDataSource() datasource.DataSource { return &availabilityZonesDataSource{} }

type availabilityZonesDataModel struct {
	ID        types.String `tfsdk:"id"`
	ProjectID types.String `tfsdk:"project_id"`
	Region    types.String `tfsdk:"region"`
	Names     types.List   `tfsdk:"names"`
}

func (d *availabilityZonesDataSource) Metadata(
	_ context.Context,
	req datasource.MetadataRequest,
	resp *datasource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_filestorage_availability_zones_v2"
}

func (d *availabilityZonesDataSource) Schema(
	_ context.Context,
	_ datasource.SchemaRequest,
	resp *datasource.SchemaResponse,
) {
	docs := schemadoc.DataSource{Name: "file storage availability zone list", Location: shareLocation}
	resp.Schema = schema.Schema{
		MarkdownDescription: "Provides a list of file storage availability zones of a pool in File Storage. For more information about file storage availability, see the [official Selectel documentation](" + schemadoc.LocationLink("", shareLocation).URL + ").",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Pool of the availability zones.",
			},
			"project_id": docs.ProjectID(),
			"region":     docs.Region(),
			"names": schema.ListAttribute{
				Computed:    true,
				ElementType: types.StringType,
				Description: "Sorted names of the file storage availability zones, for example, `ru-6a`.",
			},
		},
	}
}

func (d *availabilityZonesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config availabilityZonesDataModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}
	ctx, cancel := context.WithTimeout(ctx, readTimeout)
	defer cancel()
	region := config.Region.ValueString()
	client, err := d.client(ctx, config.ProjectID.ValueString(), region)
	if err != nil {
		resp.Diagnostics.AddError("Connect to File Storage", err.Error())
		return
	}
	zones, _, err := availabilityzone.List(ctx, client)
	if err != nil {
		resp.Diagnostics.Append(apiError("List file storage availability zones", region, err))
		return
	}
	names := make([]string, 0, len(zones))
	for _, zone := range zones {
		names = append(names, zone.Name)
	}
	sort.Strings(names)
	config.ID = config.Region
	var diags diag.Diagnostics
	config.Names, diags = types.ListValueFrom(ctx, types.StringType, names)
	resp.Diagnostics.Append(diags...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}
