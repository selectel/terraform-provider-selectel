package blockstorage

import (
	"context"
	"fmt"
	"maps"
	"strings"

	"github.com/hashicorp/go-uuid"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	sdk "github.com/selectel/blockstorage-go/pkg/v1"
	"github.com/selectel/blockstorage-go/pkg/v1/volume"
	"github.com/selectel/blockstorage-go/pkg/v1/volumetype"
)

// observe writes the API representation of the volume into the model. managed
// is the user-owned metadata; its keys select what metadata keeps, and a null
// map leaves metadata unmanaged.
func (m *volumeModel) observe(ctx context.Context, current *volume.View, managed types.Map) diag.Diagnostics {
	var diags, d diag.Diagnostics
	m.ID = types.StringValue(current.ID)
	m.Size = types.Int64Value(int64(current.Size))
	m.Name = types.StringValue(current.Name)
	m.Description = types.StringValue(current.Description)
	m.AvailabilityZone = types.StringValue(current.AvailabilityZone)
	m.VolumeType = types.StringValue(current.VolumeType)
	m.SnapshotID = types.StringValue(current.SnapshotID)
	m.SourceVolumeID = types.StringValue(current.SourceVolID)
	remote := serviceMetadata(current)
	m.ServiceMetadata, d = types.MapValueFrom(ctx, types.StringType, remote)
	diags.Append(d...)
	m.Metadata, d = managedMetadata(ctx, managed, remote)
	diags.Append(d...)
	m.Attachments, d = attachmentsValue(current.Attachments)
	diags.Append(d...)

	return diags
}

// serviceMetadata is all metadata returned by the API except the creation token.
func serviceMetadata(current *volume.View) map[string]string {
	metadata := maps.Clone(current.Metadata)
	delete(metadata, createTokenKey)
	if metadata == nil {
		metadata = map[string]string{}
	}

	return metadata
}

// managedMetadata keeps ownership separate from API-added keys. It refreshes
// every managed key, dropping missing ones, so an out-of-band change still
// produces a diff.
func managedMetadata(ctx context.Context, managed types.Map, remote map[string]string) (types.Map, diag.Diagnostics) {
	if managed.IsNull() {
		return managed, nil
	}
	values := map[string]string{}
	for key := range managed.Elements() {
		if value, exists := remote[key]; exists {
			values[key] = value
		}
	}

	return types.MapValueFrom(ctx, types.StringType, values)
}

// ownedAfterFailedUpdate is the managed metadata after an update that failed
// part way: it may have applied none, some or all planned keys. The prior keys
// stay managed, so leftovers still produce a diff. A planned key becomes managed
// only once the API reports its planned value; otherwise a value the update did
// not set, such as the server default of total_iops_sec, would enter state as
// if the user had configured it.
func ownedAfterFailedUpdate(prior, planned types.Map, remote map[string]string) (types.Map, diag.Diagnostics) {
	owned := prior.Elements()
	for key, value := range planned.Elements() {
		applied, exists := remote[key]
		if exists && value.Equal(types.StringValue(applied)) {
			owned[key] = value
		}
	}
	if prior.IsNull() && len(owned) == 0 {
		return prior, nil
	}

	return types.MapValue(types.StringType, owned)
}

func attachmentsValue(attachments []volume.Attachment) (types.Set, diag.Diagnostics) {
	var diags diag.Diagnostics
	values := make([]attr.Value, 0, len(attachments))
	for _, attachment := range attachments {
		item, d := types.ObjectValue(attachmentTypes, map[string]attr.Value{
			"id":          types.StringValue(attachment.ID),
			"instance_id": types.StringValue(attachment.ServerID),
			"device":      types.StringValue(attachment.Device),
		})
		diags.Append(d...)
		values = append(values, item)
	}
	set, d := types.SetValue(types.ObjectType{AttrTypes: attachmentTypes}, values)
	diags.Append(d...)

	return set, diags
}

func isUUID(value string) bool {
	_, err := uuid.ParseUUID(value)

	return err == nil
}

func isAttached(current *volume.View) bool {
	return current.Status == "in-use" || len(current.Attachments) > 0
}

// stateVolumeType is the volume type to store. The API reports a type name, but
// Terraform requires a configured UUID or regional alias to stay unchanged, so
// a configured value that verifiably refers to the reported type is kept.
func stateVolumeType(ctx context.Context, client *sdk.Client, configured, region string, current *volume.View) (string, error) {
	if configured == current.VolumeType {
		return configured, nil
	}
	if isUUID(configured) {
		typ, _, err := volumetype.Get(ctx, client, configured)
		if err != nil {
			return "", fmt.Errorf("resolve volume type %s: %w", configured, err)
		}
		if typ.Name == current.VolumeType {
			return configured, nil
		}

		return current.VolumeType, nil
	}
	base, suffix, _ := strings.Cut(configured, ".")
	if suffix == region && current.VolumeType == base+"."+current.AvailabilityZone {
		return configured, nil
	}

	return current.VolumeType, nil
}

func (r *volumeResource) ImportState(
	ctx context.Context,
	req resource.ImportStateRequest,
	resp *resource.ImportStateResponse,
) {
	if !isUUID(req.ID) {
		resp.Diagnostics.AddError("Invalid volume ID", "Import requires a volume UUID.")
		return
	}
	if r.config == nil {
		resp.Diagnostics.AddError("Missing provider configuration", "Configure the provider before importing.")
		return
	}
	project, region := r.config.BlockStorageDefaults()
	if project == "" || region == "" {
		resp.Diagnostics.AddError(
			"Missing import scope",
			"Set provider project_id and region, or INFRA_PROJECT_ID and INFRA_REGION.",
		)

		return
	}
	// A refresh fills all API-owned fields. Image and backup provenance cannot be
	// recovered from the volume API, so they intentionally remain null.
	state := volumeModel{
		Timeouts: timeouts.Value{
			Object: types.ObjectNull(
				map[string]attr.Type{
					"create": types.StringType,
					"update": types.StringType,
					"delete": types.StringType,
				},
			),
		},
		ID:              types.StringValue(req.ID),
		ProjectID:       types.StringValue(project),
		Region:          types.StringValue(region),
		OnlineResize:    types.BoolValue(false),
		Metadata:        types.MapNull(types.StringType),
		ServiceMetadata: types.MapNull(types.StringType),
		Attachments:     types.SetNull(types.ObjectType{AttrTypes: attachmentTypes}),
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
