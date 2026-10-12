package blockstorage

import (
	"context"
	"fmt"
	"maps"
	"slices"
	"strings"
	"time"

	"github.com/hashicorp/go-uuid"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	sdk "github.com/selectel/blockstorage-go/pkg/v1"
	"github.com/selectel/blockstorage-go/pkg/v1/volume"
)

func (r *volumeResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan volumeModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	duration, d := plan.Timeouts.Create(ctx, volumeTimeouts.Create)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}
	ctx, cancel := context.WithTimeout(ctx, duration)
	defer cancel()
	client, err := r.client(ctx, plan.ProjectID.ValueString(), plan.Region.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Connect to Block Storage", err.Error())
		return
	}
	metadata := map[string]string{}
	if !plan.Metadata.IsNull() {
		resp.Diagnostics.Append(plan.Metadata.ElementsAs(ctx, &metadata, false)...)
	}
	if resp.Diagnostics.HasError() {
		return
	}
	token, err := uuid.GenerateUUID()
	if err != nil {
		resp.Diagnostics.AddError("Generate creation token", err.Error())
		return
	}
	metadata[createTokenKey] = token
	opts := volume.CreateOpts{
		Size:             int(plan.Size.ValueInt64()),
		Name:             plan.Name.ValueString(),
		Description:      plan.Description.ValueString(),
		AvailabilityZone: plan.AvailabilityZone.ValueString(),
		VolumeType:       plan.VolumeType.ValueString(),
		SnapshotID:       plan.SnapshotID.ValueString(),
		SourceVolID:      plan.SourceVolumeID.ValueString(),
		ImageID:          plan.ImageID.ValueString(),
		BackupID:         plan.BackupID.ValueString(),
		Metadata:         metadata,
	}
	created, _, err := volume.Create(ctx, client, opts)
	if err != nil && !createOutcomeUncertain(err) {
		resp.Diagnostics.AddError("Create volume", err.Error())
		return
	}
	if err != nil || created.ID == "" {
		// An uncertain POST must never be retried: retain a searchable token so
		// a later Read or Delete can recover its identity.
		pending := &volume.View{
			ID:               "pending:" + token,
			Size:             opts.Size,
			Name:             opts.Name,
			Description:      opts.Description,
			AvailabilityZone: opts.AvailabilityZone,
			VolumeType:       opts.VolumeType,
			SnapshotID:       opts.SnapshotID,
			SourceVolID:      opts.SourceVolID,
			Metadata:         metadata,
		}
		resp.Diagnostics.Append(plan.observe(ctx, pending, plan.Metadata)...)
		resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
		resp.Diagnostics.AddError(
			"Volume creation outcome is uncertain",
			fmt.Sprintf(
				"Creation token %s was retained in state. Refresh to recover the volume before retrying. API error: %v",
				token,
				err,
			),
		)

		return
	}
	// Publish identity before waiting: a timeout or terminal API failure must not
	// lose an object that has already been created. The configured type is kept
	// until Read can verify it against the type name reported by the API.
	partial := plan
	resp.Diagnostics.Append(partial.observe(ctx, created, plan.Metadata)...)
	if !plan.VolumeType.IsUnknown() {
		partial.VolumeType = plan.VolumeType
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &partial)...)
	current, err := waitVolumeReady(ctx, client, created.ID, r.pollInterval)
	if err != nil {
		resp.Diagnostics.AddError("Wait for created volume", err.Error())
		return
	}
	volumeType, err := stateVolumeType(ctx, client, plan.VolumeType.ValueString(), plan.Region.ValueString(), current)
	if err != nil {
		resp.Diagnostics.AddError("Read volume type", err.Error())
		return
	}
	resp.Diagnostics.Append(plan.observe(ctx, current, plan.Metadata)...)
	plan.VolumeType = types.StringValue(volumeType)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// createOutcomeUncertain reports whether a failed POST may still have created
// a volume: the API may have accepted it, but no definite answer came back.
// A rejection, such as a missing source or a rate limit, created nothing.
func createOutcomeUncertain(err error) bool {
	for _, kind := range []sdk.Kind{
		sdk.KindServerError, sdk.KindTransport, sdk.KindTimeout, sdk.KindCanceled, sdk.KindUnexpected,
	} {
		if sdk.IsKind(err, kind) {
			return true
		}
	}

	return false
}

func (r *volumeResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state volumeModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	ctx, cancel := context.WithTimeout(ctx, 10*time.Minute)
	defer cancel()
	client, err := r.client(ctx, state.ProjectID.ValueString(), state.Region.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Connect to Block Storage", err.Error())
		return
	}
	id, err := recoverVolumeID(ctx, client, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Recover volume", err.Error())
		return
	}
	current, _, err := volume.Get(ctx, client, id)
	if sdk.IsKind(err, sdk.KindNotFound) {
		if err := confirmVolumeAbsent(ctx, client, id); err != nil {
			resp.Diagnostics.AddError("Confirm volume absence", err.Error())
			return
		}
		resp.State.RemoveResource(ctx)

		return
	}
	if err != nil {
		resp.Diagnostics.AddError("Read volume", err.Error())
		return
	}
	volumeType, err := stateVolumeType(ctx, client, state.VolumeType.ValueString(), state.Region.ValueString(), current)
	if err != nil {
		resp.Diagnostics.AddError("Read volume type", err.Error())
		return
	}
	resp.Diagnostics.Append(state.observe(ctx, current, state.Metadata)...)
	state.VolumeType = types.StringValue(volumeType)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *volumeResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, prior volumeModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &prior)...)
	duration, d := plan.Timeouts.Update(ctx, volumeTimeouts.Update)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}
	ctx, cancel := context.WithTimeout(ctx, duration)
	defer cancel()
	client, err := r.client(ctx, plan.ProjectID.ValueString(), plan.Region.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Connect to Block Storage", err.Error())
		return
	}
	id, err := recoverVolumeID(ctx, client, prior.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Recover volume", err.Error())
		return
	}
	current, _, err := volume.Get(ctx, client, id)
	if err != nil {
		resp.Diagnostics.AddError("Read volume before update", err.Error())
		return
	}
	target := int(plan.Size.ValueInt64())
	if target < current.Size {
		resp.Diagnostics.AddError(
			"Volume shrinking is unsupported",
			fmt.Sprintf("Remote size is %d GiB; requested %d GiB.", current.Size, target),
		)

		return
	}
	attached := isAttached(current)
	if target > current.Size && attached && !plan.OnlineResize.ValueBool() {
		resp.Diagnostics.AddError("Online resize is disabled", "Detach the volume or set enable_online_resize = true.")
		return
	}
	opts := volume.UpdateOpts{}
	if !plan.Name.Equal(prior.Name) {
		v := plan.Name.ValueString()
		opts.Name = &v
	}
	if !plan.Description.Equal(prior.Description) {
		v := plan.Description.ValueString()
		opts.Description = &v
	}
	if !plan.Metadata.Equal(prior.Metadata) {
		planned := map[string]string{}
		if !plan.Metadata.IsNull() {
			resp.Diagnostics.Append(plan.Metadata.ElementsAs(ctx, &planned, false)...)
		}
		if resp.Diagnostics.HasError() {
			return
		}
		opts.Metadata = desiredMetadata(current.Metadata, slices.Collect(maps.Keys(prior.Metadata.Elements())), planned)
	}
	current, updateErr := r.updateVolume(ctx, client, current, opts, target, attached)
	if updateErr != nil {
		resp.Diagnostics.AddError("Update volume", updateErr.Error())
	}
	// Read after both success and partial failure. If refresh fails, retain the
	// last successful response and identity instead of writing the unfulfilled plan.
	fresh, _, readErr := volume.Get(ctx, client, id)
	if readErr == nil {
		current = fresh
	} else {
		resp.Diagnostics.AddError("Refresh updated volume", readErr.Error())
	}
	managed := plan.Metadata
	if resp.Diagnostics.HasError() {
		var d diag.Diagnostics
		managed, d = ownedAfterFailedUpdate(prior.Metadata, plan.Metadata, current.Metadata)
		resp.Diagnostics.Append(d...)
	}
	resp.Diagnostics.Append(plan.observe(ctx, current, managed)...)
	// volume_type requires replacement, so Update cannot change it. Keep the
	// stored value and leave drift detection to Read.
	plan.VolumeType = prior.VolumeType
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// updateVolume applies descriptive changes before resizing and returns the last
// confirmed API representation even when a later operation fails.
func (r *volumeResource) updateVolume(
	ctx context.Context,
	client *sdk.Client,
	current *volume.View,
	opts volume.UpdateOpts,
	target int,
	attached bool,
) (*volume.View, error) {
	if opts.Name != nil || opts.Description != nil || opts.Metadata != nil {
		updated, _, err := volume.Update(ctx, client, current.ID, opts)
		if err != nil {
			return current, fmt.Errorf("update name, description or metadata: %w", err)
		}
		current = updated
	}
	if target > current.Size && current.Status != "extending" {
		_, err := volume.Extend(ctx, client, current.ID, volume.ExtendOpts{NewSize: target, Attached: attached})
		if err != nil {
			return current, fmt.Errorf("request resize: %w", err)
		}
	}
	if target > current.Size || current.Status == "extending" {
		resized, err := waitVolumeReady(ctx, client, current.ID, r.pollInterval)
		if err != nil {
			return current, fmt.Errorf("wait for resize: %w", err)
		}
		if resized.Size != target {
			return resized, fmt.Errorf("volume %s finished at %d GiB; expected %d GiB", resized.ID, resized.Size, target)
		}
		current = resized
	}

	return current, nil
}

// desiredMetadata is the full metadata map to send, because the API replaces
// the whole map. It keeps API-added keys and the creation token, drops the
// previously managed keys and adds the planned ones. total_iops_sec is never
// dropped because the API cannot remove it: ModifyPlan rejects such a plan,
// but it skips the check while the planned metadata is unknown.
func desiredMetadata(remote map[string]string, previouslyManaged []string, planned map[string]string) map[string]string {
	desired := maps.Clone(remote)
	if desired == nil {
		desired = map[string]string{}
	}
	for _, key := range previouslyManaged {
		if key != iopsKey {
			delete(desired, key)
		}
	}
	maps.Copy(desired, planned)

	return desired
}

func (r *volumeResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state volumeModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	duration, d := state.Timeouts.Delete(ctx, volumeTimeouts.Delete)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}
	ctx, cancel := context.WithTimeout(ctx, duration)
	defer cancel()
	client, err := r.client(ctx, state.ProjectID.ValueString(), state.Region.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Connect to Block Storage", err.Error())
		return
	}
	id, err := recoverVolumeID(ctx, client, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Recover volume", err.Error())
		return
	}
	current, _, err := volume.Get(ctx, client, id)
	if sdk.IsKind(err, sdk.KindNotFound) {
		err = confirmVolumeAbsent(ctx, client, id)
		if err != nil {
			resp.Diagnostics.AddError("Confirm volume absence", err.Error())
		}

		return
	}
	if err != nil {
		resp.Diagnostics.AddError("Read volume before deletion", err.Error())
		return
	}
	if isAttached(current) {
		resp.Diagnostics.AddError(
			"Volume is attached",
			"Detach the volume from all servers before deletion. Its identity is retained in state.",
		)

		return
	}
	if current.Status != "deleting" {
		_, err = volume.Delete(ctx, client, id)
		if err != nil && !sdk.IsKind(err, sdk.KindNotFound) {
			resp.Diagnostics.AddError("Delete volume", err.Error())
			return
		}
	}
	if err := waitVolumeDeleted(ctx, client, id, r.pollInterval); err != nil {
		resp.Diagnostics.AddError("Wait for deleted volume", err.Error())
	}
}

func recoverVolumeID(ctx context.Context, client *sdk.Client, id string) (string, error) {
	token, pending := strings.CutPrefix(id, "pending:")
	if !pending {
		return id, nil
	}
	all, err := volume.ListDetail(ctx, client, volume.ListOpts{})
	if err != nil {
		return "", err
	}
	var found string
	for _, candidate := range all {
		if candidate.Metadata[createTokenKey] != token {
			continue
		}
		if found != "" {
			return "", fmt.Errorf(
				"multiple volumes have creation token %s; import the intended volume UUID manually",
				token,
			)
		}
		found = candidate.ID
	}
	if found == "" {
		return "", fmt.Errorf(
			"creation token %s is not visible yet; check the project and refresh again. Do not remove this state entry until the creation outcome is known; if no volume with this token appears, remove the entry with terraform state rm",
			token,
		)
	}

	return found, nil
}
