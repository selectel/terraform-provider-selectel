package blockstorage

import (
	"context"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVolumeConfigValidation(t *testing.T) {
	for _, tc := range []struct {
		name      string
		change    func(*volumeModel)
		errorText string
	}{
		{"empty volume", func(_ *volumeModel) {}, ""},
		{"missing size", func(m *volumeModel) { m.Size = types.Int64Null() }, "Missing volume size"},
		{"unknown source supplies size", func(m *volumeModel) { m.Size = types.Int64Null(); m.SnapshotID = types.StringUnknown() }, ""},
		{"snapshot supplies size", func(m *volumeModel) { m.Size = types.Int64Null(); m.SnapshotID = types.StringValue(volumeID) }, ""},
		{"clone supplies size", func(m *volumeModel) { m.Size = types.Int64Null(); m.SourceVolumeID = types.StringValue(volumeID) }, ""},
		{"backup supplies size", func(m *volumeModel) { m.Size = types.Int64Null(); m.BackupID = types.StringValue(volumeID) }, ""},
		{"backup with empty name", func(m *volumeModel) { m.BackupID = types.StringValue(volumeID); m.Name = types.StringValue("") }, "Empty volume name"},
		{"image requires size", func(m *volumeModel) { m.Size = types.Int64Null(); m.ImageID = types.StringValue(volumeID) }, "Missing volume size"},
		{"unknown metadata", func(m *volumeModel) { m.Metadata = types.MapUnknown(types.StringType) }, ""},
		{"valid metadata", func(m *volumeModel) { m.Metadata = metadata(map[string]string{"key": "", iopsKey: "2000"}) }, ""},
		{"reserved token", func(m *volumeModel) { m.Metadata = metadata(map[string]string{createTokenKey: "abc"}) }, "Reserved metadata key"},
		{"server bandwidth", func(m *volumeModel) { m.Metadata = metadata(map[string]string{bandwidthKey: "100"}) }, "Reserved metadata key"},
		{"empty key", func(m *volumeModel) { m.Metadata = metadata(map[string]string{"": "value"}) }, "Invalid metadata key"},
		{"invalid key", func(m *volumeModel) { m.Metadata = metadata(map[string]string{"!": "value"}) }, "Invalid metadata key"},
		{"long key", func(m *volumeModel) { m.Metadata = metadata(map[string]string{strings.Repeat("a", 256): "value"}) }, "Invalid metadata key"},
		{"boundary key", func(m *volumeModel) {
			m.Metadata = metadata(map[string]string{strings.Repeat("a", 255): strings.Repeat("я", 255)})
		}, ""},
		{"long value", func(m *volumeModel) { m.Metadata = metadata(map[string]string{"key": strings.Repeat("я", 256)}) }, "Metadata value too long"},
		{"nondecimal iops", func(m *volumeModel) { m.Metadata = metadata(map[string]string{iopsKey: "2e3"}) }, "Invalid IOPS"},
		{"null value", func(m *volumeModel) {
			m.Metadata = types.MapValueMust(types.StringType, map[string]attr.Value{"key": types.StringNull()})
		}, "Null metadata value"},
		{"unknown value", func(m *volumeModel) {
			m.Metadata = types.MapValueMust(types.StringType, map[string]attr.Value{iopsKey: types.StringUnknown()})
		}, ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r := testResource(nil)
			config := model()
			config.SnapshotID = types.StringNull()
			config.SourceVolumeID = types.StringNull()
			tc.change(&config)
			state := stateOf(t, r, config)
			response := resource.ValidateConfigResponse{}
			r.ValidateConfig(
				context.Background(),
				resource.ValidateConfigRequest{Config: tfsdk.Config(state)},
				&response,
			)
			if tc.errorText == "" {
				require.False(t, response.Diagnostics.HasError(), "%v", response.Diagnostics)
			} else {
				require.True(t, response.Diagnostics.HasError())
				assert.Contains(t, response.Diagnostics.Errors()[0].Summary(), tc.errorText)
			}
		})
	}
}

func TestVolumeSourcesConflict(t *testing.T) {
	for _, name := range volumeSources {
		for _, other := range []string{"", otherSources(name)[0]} {
			t.Run(name+" "+other, func(t *testing.T) {
				config := model()
				config.SnapshotID, config.SourceVolumeID = types.StringNull(), types.StringNull()
				sources := map[string]*types.String{
					"snapshot_id": &config.SnapshotID, "source_volume_id": &config.SourceVolumeID,
					"image_id": &config.ImageID, "backup_id": &config.BackupID,
				}
				*sources[name] = types.StringValue(volumeID)
				if other != "" {
					*sources[other] = types.StringValue(volumeID)
				}
				diags := volumeAttributeErrors(t, config)
				assert.Equal(t, other != "", diags.HasError(), "%v", diags)
			})
		}
	}
}

func TestVolumeSourceRequirements(t *testing.T) {
	for _, tc := range []struct {
		name   string
		change func(*volumeModel)
	}{
		{"snapshot without type", func(m *volumeModel) {
			m.SnapshotID, m.VolumeType = types.StringValue(volumeID), types.StringNull()
		}},
		{"backup without name", func(m *volumeModel) {
			m.SnapshotID, m.BackupID, m.Name = types.StringNull(), types.StringValue(volumeID), types.StringNull()
		}},
		// The API returns lowercase IDs, so an uppercase one would never match state.
		{"uppercase snapshot", func(m *volumeModel) { m.SnapshotID = types.StringValue("ABCDEF12-1111-4111-8111-111111111111") }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			config := model()
			config.SnapshotID, config.SourceVolumeID = types.StringValue(volumeID), types.StringNull()
			diags := volumeAttributeErrors(t, config)
			require.False(t, diags.HasError(), "%v", diags)
			tc.change(&config)
			assert.True(t, volumeAttributeErrors(t, config).HasError())
		})
	}
}

func TestVolumeWhitespace(t *testing.T) {
	for name, emptyAllowed := range map[string]bool{
		"name": true, "description": true, "availability_zone": false, "volume_type": false, "image_id": false,
		"project_id": false, "region": false,
	} {
		for value, valid := range map[string]bool{
			"disk": true, "a b": true, "": emptyAllowed, " disk": false, "disk\n": false, " ": false,
			"\u00a0disk": false, "disk\v": false,
		} {
			t.Run(name+"/"+value, func(t *testing.T) {
				config := model()
				config.SnapshotID, config.SourceVolumeID = types.StringNull(), types.StringNull()
				*map[string]*types.String{
					"name": &config.Name, "description": &config.Description, "availability_zone": &config.AvailabilityZone,
					"volume_type": &config.VolumeType, "image_id": &config.ImageID, "project_id": &config.ProjectID, "region": &config.Region,
				}[name] = types.StringValue(value)
				diags := volumeAttributeErrors(t, config)
				assert.Equal(t, !valid, diags.HasError(), "%v", diags)
			})
		}
	}
}

func TestVolumePlanValidation(t *testing.T) {
	smaller := func(m *volumeModel) { m.Size = types.Int64Value(9) }
	for name, tc := range map[string]struct {
		change          func(*volumeModel)
		create, destroy bool
		err             string
	}{
		"unchanged":       {change: func(*volumeModel) {}},
		"shrink":          {change: smaller, err: "Volume shrinking is unsupported"},
		"replace smaller": {change: func(m *volumeModel) { smaller(m); m.VolumeType = types.StringValue("other") }},
		// An unknown volume type replaces the volume, as RequiresReplace does.
		"replace smaller, type not known yet": {change: func(m *volumeModel) { smaller(m); m.VolumeType = types.StringUnknown() }},
		"grow":                                {change: func(m *volumeModel) { m.Size = types.Int64Value(11) }},
		"unknown size":                        {change: func(m *volumeModel) { m.Size = types.Int64Unknown() }},
		"remove iops":                         {change: func(m *volumeModel) { m.Metadata = metadata(map[string]string{}) }, err: "Cannot remove IOPS"},
		"null metadata":                       {change: func(m *volumeModel) { m.Metadata = types.MapNull(types.StringType) }, err: "Cannot remove IOPS"},
		"unknown metadata":                    {change: func(m *volumeModel) { m.Metadata = types.MapUnknown(types.StringType) }},
		"create":                              {change: smaller, create: true},
		"destroy":                             {change: smaller, destroy: true},
	} {
		t.Run(name, func(t *testing.T) {
			r := testResource(nil)
			prior := model()
			prior.Metadata = metadata(map[string]string{iopsKey: "2000"})
			plan := prior
			tc.change(&plan)
			req := resource.ModifyPlanRequest{State: stateOf(t, r, prior), Plan: planOf(t, r, plan)}
			if tc.create {
				req.State.Raw = tftypes.NewValue(req.State.Raw.Type(), nil)
			}
			if tc.destroy {
				req.Plan.Raw = tftypes.NewValue(req.Plan.Raw.Type(), nil)
			}
			resp := resource.ModifyPlanResponse{Plan: req.Plan}
			r.ModifyPlan(context.Background(), req, &resp)
			if tc.err == "" {
				assert.False(t, resp.Diagnostics.HasError(), "%v", resp.Diagnostics)
				return
			}
			require.Len(t, resp.Diagnostics.Errors(), 1, "%v", resp.Diagnostics)
			assert.Equal(t, tc.err, resp.Diagnostics.Errors()[0].Summary())
		})
	}
}

func TestLookupConfiguration(t *testing.T) {
	volume := NewVolumeDataSource().(*volumeDataSource)
	var volumeSchema datasource.SchemaResponse
	volume.Schema(context.Background(), datasource.SchemaRequest{}, &volumeSchema)
	for _, hasID := range []bool{false, true} {
		for _, filter := range []string{"none", "name", "status", "metadata"} {
			t.Run(filter+map[bool]string{true: " id", false: " no id"}[hasID], func(t *testing.T) {
				config := volumeDataModel{
					ProjectID:   types.StringValue("project"),
					Region:      types.StringValue("ru-1"),
					Metadata:    types.MapNull(types.StringType),
					Attachments: types.SetNull(types.ObjectType{AttrTypes: attachmentTypes}),
				}
				if hasID {
					config.VolumeID = types.StringValue(volumeID)
				}
				switch filter {
				case "name":
					config.Name = types.StringValue("disk")
				case "status":
					config.Status = types.StringValue("available")
				case "metadata":
					config.Metadata = metadata(map[string]string{})
				}
				diags := attributeErrors(t, volumeSchema.Schema.Attributes, dataConfig(t, volume, config))
				assert.Equal(t, hasID && filter != "none", diags.HasError(), "%v", diags)
			})
		}
	}
	typ := NewVolumeTypeDataSource().(*volumeTypeDataSource)
	var typeSchema datasource.SchemaResponse
	typ.Schema(context.Background(), datasource.SchemaRequest{}, &typeSchema)
	for _, hasID := range []bool{false, true} {
		for _, hasName := range []bool{false, true} {
			config := volumeTypeDataModel{
				ProjectID:  types.StringValue("project"),
				Region:     types.StringValue("ru-1"),
				ExtraSpecs: types.MapNull(types.StringType),
			}
			if hasID {
				config.TypeID = types.StringValue(typeID)
			}
			if hasName {
				config.Name = types.StringValue("fast")
			}
			diags := attributeErrors(t, typeSchema.Schema.Attributes, dataConfig(t, typ, config))
			assert.Equal(t, hasID == hasName, diags.HasError(), "id=%v name=%v: %v", hasID, hasName, diags)
		}
	}
}

// TestVolumePlanInPlace plans through the framework server, which runs the
// attribute plan modifiers the way Terraform does: any change marks omitted
// computed values unknown before the modifiers see them.
func TestVolumePlanInPlace(t *testing.T) {
	ctx := context.Background()
	server := providerserver.NewProtocol6(planTestProvider{})()
	schemas, err := server.GetProviderSchema(ctx, &tfprotov6.GetProviderSchemaRequest{})
	require.NoError(t, err)
	valueType := schemas.ResourceSchemas["selectel_compute_volume_v3"].ValueType()
	r := testResource(nil)
	encode := func(m volumeModel) *tfprotov6.DynamicValue {
		value, err := tfprotov6.NewDynamicValue(valueType, stateOf(t, r, m).Raw)
		require.NoError(t, err)

		return &value
	}
	prior := model()
	for _, tc := range []struct {
		name    string
		change  func(*volumeModel)
		replace []*tftypes.AttributePath
	}{
		{"rename", func(m *volumeModel) { m.Name = types.StringValue("renamed") }, nil},
		{"grow", func(m *volumeModel) { m.Size = types.Int64Value(20) }, nil},
		{"zone", func(m *volumeModel) { m.AvailabilityZone = types.StringValue("ru-1b") }, []*tftypes.AttributePath{tftypes.NewAttributePath().WithAttributeName("availability_zone")}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			// Terraform proposes the configured values and keeps the prior ones
			// of computed attributes that the configuration omits.
			proposed := prior
			tc.change(&proposed)
			config := proposed
			config.ID = types.StringNull()
			config.ServiceMetadata = types.MapNull(types.StringType)
			config.Attachments = types.SetNull(types.ObjectType{AttrTypes: attachmentTypes})
			config.SnapshotID, config.SourceVolumeID = types.StringNull(), types.StringNull()
			if config.AvailabilityZone.Equal(prior.AvailabilityZone) {
				config.AvailabilityZone = types.StringNull()
			}
			config.VolumeType = types.StringNull()

			resp, err := server.PlanResourceChange(ctx, &tfprotov6.PlanResourceChangeRequest{
				TypeName:         "selectel_compute_volume_v3",
				PriorState:       encode(prior),
				ProposedNewState: encode(proposed),
				Config:           encode(config),
			})
			require.NoError(t, err)
			require.Empty(t, resp.Diagnostics)
			assert.ElementsMatch(t, tc.replace, resp.RequiresReplace)
		})
	}
}

type planTestProvider struct{}

func (planTestProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "selectel"
}

func (planTestProvider) Schema(context.Context, provider.SchemaRequest, *provider.SchemaResponse) {}

func (planTestProvider) Configure(context.Context, provider.ConfigureRequest, *provider.ConfigureResponse) {
}

func (planTestProvider) Resources(context.Context) []func() resource.Resource {
	return []func() resource.Resource{NewVolumeResource}
}

func (planTestProvider) DataSources(context.Context) []func() datasource.DataSource { return nil }
