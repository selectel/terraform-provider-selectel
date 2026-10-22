package images

import (
	"context"
	"strings"
	"time"

	"github.com/hashicorp/go-uuid"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/selectel/images-go/pkg/v1/image"
)

// observe writes the API representation of the image into the model. The
// image sources are not reported by the API and stay as they are. Tags and
// properties keep the form that the model has: the keys of properties select
// the user-owned properties, and a null map leaves them unmanaged.
func (m *imageModel) observe(ctx context.Context, current *image.View) diag.Diagnostics {
	var diags, d diag.Diagnostics
	m.ID = types.StringValue(current.ID)
	m.Name = types.StringValue(current.Name)
	m.ContainerFormat = types.StringValue(current.ContainerFormat)
	m.DiskFormat = types.StringValue(current.DiskFormat)
	m.MinDiskGB = types.Int64Value(int64(current.MinDisk))
	m.MinRAMMB = types.Int64Value(int64(current.MinRAM))
	m.Protected = types.BoolValue(current.Protected)
	m.Hidden = types.BoolValue(current.Hidden)
	m.Visibility = types.StringValue(current.Visibility)
	m.Tags, d = tagsValue(m.Tags, current.Tags)
	diags.Append(d...)
	m.Properties, d = managedProperties(m.Properties, current.Properties)
	diags.Append(d...)
	m.ServiceProperties, d = types.MapValueFrom(ctx, types.StringType, current.Properties)
	diags.Append(d...)
	m.Status = types.StringValue(current.Status)
	m.Owner = types.StringValue(current.Owner)
	m.SizeBytes = types.Int64Value(current.Size)
	m.Checksum = types.StringValue(current.Checksum)
	m.CreatedAt = types.StringValue(current.CreatedAt.Format(time.RFC3339))
	m.UpdatedAt = types.StringValue(current.UpdatedAt.Format(time.RFC3339))

	return diags
}

// tagsValue is the set of remote tags. An image without tags keeps the form
// of the model value, null or empty, so both spellings in the configuration
// stay without a diff.
func tagsValue(modeled types.Set, remote []string) (types.Set, diag.Diagnostics) {
	if len(remote) == 0 && modeled.IsNull() {
		return modeled, nil
	}
	values := make([]attr.Value, len(remote))
	for i, tag := range remote {
		values[i] = types.StringValue(tag)
	}

	return types.SetValue(types.StringType, values)
}

// managedProperties keeps ownership separate from the properties that the
// Image service adds. It refreshes every managed key, dropping missing ones,
// so an out-of-band change still produces a diff.
func managedProperties(managed types.Map, remote map[string]string) (types.Map, diag.Diagnostics) {
	if managed.IsNull() {
		return managed, nil
	}
	values := map[string]attr.Value{}
	for key := range managed.Elements() {
		if value, exists := remote[key]; exists {
			values[key] = types.StringValue(value)
		}
	}

	return types.MapValue(types.StringType, values)
}

// isImageID accepts the lowercase UUID in which the API reports image IDs.
func isImageID(id string) bool {
	_, err := uuid.ParseUUID(id)

	return err == nil && id == strings.ToLower(id)
}

func (r *imageResource) ImportState(
	ctx context.Context,
	req resource.ImportStateRequest,
	resp *resource.ImportStateResponse,
) {
	if !isImageID(req.ID) {
		resp.Diagnostics.AddError("Invalid image ID", "Import requires an image ID in the lowercase UUID format.")
		return
	}
	project, region, diags := r.importScope("")
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	// A refresh fills every attribute that the API reports. The sources and
	// the ownership of properties cannot be recovered, so they stay null.
	state := imageModel{
		ID:                types.StringValue(req.ID),
		ProjectID:         types.StringValue(project),
		Region:            types.StringValue(region),
		Tags:              types.SetNull(types.StringType),
		Properties:        types.MapNull(types.StringType),
		ServiceProperties: types.MapNull(types.StringType),
		Timeouts: timeouts.Value{Object: types.ObjectNull(map[string]attr.Type{
			"create": types.StringType,
			"delete": types.StringType,
		})},
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// setStrings is the elements of a set of strings; null converts to none.
func setStrings(ctx context.Context, set types.Set) ([]string, diag.Diagnostics) {
	values := []string{}
	if set.IsNull() {
		return values, nil
	}
	diags := set.ElementsAs(ctx, &values, false)

	return values, diags
}

// mapStrings is the elements of a map of strings; null converts to none.
func mapStrings(ctx context.Context, m types.Map) (map[string]string, diag.Diagnostics) {
	values := map[string]string{}
	if m.IsNull() {
		return values, nil
	}
	diags := m.ElementsAs(ctx, &values, false)

	return values, diags
}
