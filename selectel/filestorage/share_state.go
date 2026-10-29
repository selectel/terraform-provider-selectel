package filestorage

import (
	"context"
	"fmt"
	"maps"
	"net/netip"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	sdk "github.com/selectel/filestorage-go/pkg/v1"
	"github.com/selectel/filestorage-go/pkg/v1/share"
	"github.com/selectel/filestorage-go/pkg/v1/sharenetwork"
)

// deleteNetworkKey is the flag of the fork: while the share has it, deleting
// the share also deletes its service share network.
const deleteNetworkKey = "delete_empty_share_network"

// shareObservation is everything a refresh reads: the share, its service
// share network and its mount paths.
type shareObservation struct {
	share     *share.View
	network   *sharenetwork.View
	locations []share.ExportLocation
}

// observeShare completes a read share with its network and mount paths. A
// missing network or path list is an error, not a missing share.
func observeShare(ctx context.Context, client *sdk.Client, current *share.View) (*shareObservation, error) {
	observed := &shareObservation{share: current}
	var err error
	if current.ShareNetworkID != "" {
		observed.network, _, err = sharenetwork.Get(ctx, client, current.ShareNetworkID)
		if err != nil {
			return nil, fmt.Errorf("read share network %s: %w", current.ShareNetworkID, err)
		}
	}
	observed.locations, _, err = share.ListExportLocations(ctx, client, current.ID)
	if err != nil {
		return nil, fmt.Errorf("read mount paths of file storage %s: %w", current.ID, err)
	}

	return observed, nil
}

// observe writes the API representation into the model. managed is the
// configured metadata; its keys select what metadata keeps.
func (m *shareModel) observe(ctx context.Context, observed *shareObservation, managed types.Map) diag.Diagnostics {
	var diags, d diag.Diagnostics
	current := observed.share
	m.ID = types.StringValue(current.ID)
	m.ShareProto = types.StringValue(current.ShareProto)
	m.Size = types.Int64Value(int64(current.Size))
	m.ShareType = types.StringValue(current.ShareTypeName)
	m.Name = types.StringValue(current.Name)
	m.Description = types.StringValue(current.Description)
	m.Status = types.StringValue(current.Status)
	m.AvailabilityZone = types.StringValue(current.AvailabilityZone)
	// The creation token stays on the share for debugging, but the user does
	// not set it, so all_metadata hides it as the volume resource does.
	all := maps.Clone(current.Metadata)
	delete(all, createTokenKey)
	if all == nil {
		all = map[string]string{}
	}
	m.AllMetadata, d = types.MapValueFrom(ctx, types.StringType, all)
	diags.Append(d...)
	m.Metadata, d = managedMetadata(ctx, managed, all)
	diags.Append(d...)
	if observed.network != nil && len(observed.network.Subnets) > 0 {
		subnet := observed.network.Subnets[0]
		m.NetworkID = types.StringValue(subnet.NetworkID)
		m.SubnetID = types.StringValue(subnet.SubnetID)
	}
	m.ExportLocations, d = exportLocationsValue(observed.locations)
	diags.Append(d...)
	if ip, ok := shareIP(current.ShareProto, observed.locations); ok {
		m.IP = types.StringValue(ip)
	} else {
		// A path without an IPv4 host keeps the known address, so the plan does
		// not replace the share; an imported share has none to keep.
		if m.IP.IsNull() || m.IP.IsUnknown() {
			m.IP = types.StringValue("")
		}
		diags.AddAttributeWarning(path.Root("ip"), "File storage address is unknown",
			fmt.Sprintf("No mount path of file storage %s contains an IPv4 address; ip keeps %q.", current.ID, m.IP.ValueString()))
	}

	return diags
}

// shareIP is the host of the preferred mount path, or of the first one when
// none is preferred: `10.0.0.5:/shares/x` for NFS and `\\10.0.0.5\x` for CIFS.
func shareIP(proto string, locations []share.ExportLocation) (string, bool) {
	if len(locations) == 0 {
		return "", false
	}
	location := locations[0]
	for _, candidate := range locations {
		if candidate.Preferred {
			location = candidate
			break
		}
	}
	var host string
	if strings.EqualFold(proto, protoCIFS) {
		host, _, _ = strings.Cut(strings.TrimPrefix(location.Path, `\\`), `\`)
	} else {
		host, _, _ = strings.Cut(location.Path, ":")
	}
	address, err := netip.ParseAddr(host)
	if err != nil || !address.Is4() {
		return "", false
	}

	return address.String(), true
}

// managedMetadata refreshes every configured key, dropping missing ones, so an
// out-of-band change still produces a diff. Other keys stay in all_metadata.
func managedMetadata(ctx context.Context, managed types.Map, remote map[string]string) (types.Map, diag.Diagnostics) {
	if managed.IsNull() || managed.IsUnknown() {
		return types.MapNull(types.StringType), nil
	}
	values := map[string]string{}
	for key := range managed.Elements() {
		if value, exists := remote[key]; exists {
			values[key] = value
		}
	}

	return types.MapValueFrom(ctx, types.StringType, values)
}

func exportLocationsValue(locations []share.ExportLocation) (types.List, diag.Diagnostics) {
	var diags diag.Diagnostics
	values := make([]attr.Value, 0, len(locations))
	for _, location := range locations {
		item, d := types.ObjectValue(exportLocationTypes, map[string]attr.Value{
			"path":      types.StringValue(location.Path),
			"preferred": types.BoolValue(location.Preferred),
		})
		diags.Append(d...)
		values = append(values, item)
	}
	list, d := types.ListValue(types.ObjectType{AttrTypes: exportLocationTypes}, values)
	diags.Append(d...)

	return list, diags
}

func (r *shareResource) ImportState(
	ctx context.Context,
	req resource.ImportStateRequest,
	resp *resource.ImportStateResponse,
) {
	if !uuidPattern.MatchString(req.ID) {
		resp.Diagnostics.AddError("Invalid file storage ID", "Import requires a file storage UUID.")
		return
	}
	project, region, diags := r.importScope()
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("project_id"), project)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("region"), region)...)
}
