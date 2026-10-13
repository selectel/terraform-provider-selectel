package blockstorage

import (
	"context"
	"fmt"
	"regexp"
	"unicode/utf8"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const (
	createTokenKey = "selectel_tf_create_token" //nolint:gosec // Metadata key name, not a credential.
	iopsKey        = "total_iops_sec"
	bandwidthKey   = "total_bytes_sec"
)

// recreateHint is the way around the in-place restrictions of ModifyPlan.
const recreateHint = "To recreate the volume instead, run terraform taint for this resource and apply."

var (
	metadataKeyPattern = regexp.MustCompile(`^[a-zA-Z0-9_:. /-]+$`)
	decimalPattern     = regexp.MustCompile(`^[0-9]+$`)
)

func (r *volumeResource) ValidateConfig(
	ctx context.Context,
	req resource.ValidateConfigRequest,
	resp *resource.ValidateConfigResponse,
) {
	var config volumeModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if config.Size.IsNull() && config.SnapshotID.IsNull() && config.SourceVolumeID.IsNull() &&
		config.BackupID.IsNull() {
		resp.Diagnostics.AddAttributeError(
			path.Root("size"),
			"Missing volume size",
			"Set size unless creating from a snapshot, volume, or backup.",
		)
	}
	// backup_id requires name; an explicitly empty name lets the restore copy
	// the name of the backed-up volume as well.
	if !config.BackupID.IsNull() && !config.Name.IsUnknown() && !config.Name.IsNull() && config.Name.ValueString() == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("name"),
			"Empty volume name",
			"Set a nonempty name when restoring from a backup.",
		)
	}
	for key, element := range config.Metadata.Elements() {
		at := path.Root("metadata").AtMapKey(key)
		if key == createTokenKey || key == bandwidthKey {
			resp.Diagnostics.AddAttributeError(
				at,
				"Reserved metadata key",
				"This key is managed by the provider or the Block Storage API.",
			)
		}
		if !metadataKeyPattern.MatchString(key) || utf8.RuneCountInString(key) > 255 {
			resp.Diagnostics.AddAttributeError(
				at,
				"Invalid metadata key",
				"Use 1 to 255 ASCII letters, digits, underscores, colons, periods, spaces, slashes, or hyphens.",
			)
		}
		value := element.(types.String)
		if value.IsNull() {
			resp.Diagnostics.AddAttributeError(at, "Null metadata value", "Metadata values must be strings.")
			continue
		}
		if value.IsUnknown() {
			continue
		}
		if utf8.RuneCountInString(value.ValueString()) > 255 {
			resp.Diagnostics.AddAttributeError(at, "Metadata value too long", "Use at most 255 characters.")
		}
		if key == iopsKey && !decimalPattern.MatchString(value.ValueString()) {
			resp.Diagnostics.AddAttributeError(
				at,
				"Invalid IOPS",
				"IOPS must be a nonempty decimal string. The API validates the supported range for the chosen volume type.",
			)
		}
	}
}

func (r *volumeResource) ModifyPlan(
	ctx context.Context,
	req resource.ModifyPlanRequest,
	resp *resource.ModifyPlanResponse,
) {
	if req.State.Raw.IsNull() || req.Plan.Raw.IsNull() {
		return
	}
	var prior, plan volumeModel
	resp.Diagnostics.Append(req.State.Get(ctx, &prior)...)
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	// A replacement may use a smaller volume or drop the IOPS, but only one caused
	// by the configuration is visible here: Terraform decides on -replace after
	// planning, so the errors below point to taint instead.
	if hasCreationChange(prior, plan) {
		return
	}
	if !plan.Size.IsUnknown() && plan.Size.ValueInt64() < prior.Size.ValueInt64() {
		resp.Diagnostics.AddAttributeError(
			path.Root("size"),
			"Volume shrinking is unsupported",
			fmt.Sprintf("Current size is %d GiB. Set size to this value or greater. %s", prior.Size.ValueInt64(), recreateHint),
		)
	}
	if !plan.Metadata.IsUnknown() {
		oldIOPS, hadIOPS := prior.Metadata.Elements()[iopsKey]
		if hadIOPS && !oldIOPS.IsNull() {
			if _, kept := plan.Metadata.Elements()[iopsKey]; !kept {
				resp.Diagnostics.AddAttributeError(
					path.Root("metadata"),
					"Cannot remove IOPS",
					"The Block Storage API cannot remove existing total_iops_sec. Keep it or set another supported value. "+recreateHint,
				)
			}
		}
	}
}

// Framework invokes resource plan modification separately from attribute plan
// modifiers, so their RequiresReplace paths are not available in this response.
func hasCreationChange(prior, plan volumeModel) bool {
	for _, change := range []struct{ before, after types.String }{
		{prior.ProjectID, plan.ProjectID},
		{prior.Region, plan.Region},
		{prior.AvailabilityZone, plan.AvailabilityZone},
		{prior.VolumeType, plan.VolumeType},
		{prior.SnapshotID, plan.SnapshotID},
		{prior.SourceVolumeID, plan.SourceVolumeID},
		{prior.ImageID, plan.ImageID},
		{prior.BackupID, plan.BackupID},
	} {
		if !change.after.IsUnknown() && !change.after.Equal(change.before) {
			return true
		}
	}

	return false
}
