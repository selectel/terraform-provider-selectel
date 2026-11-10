package filestorage

import (
	"context"
	"fmt"
	"maps"

	"github.com/hashicorp/go-uuid"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	sdk "github.com/selectel/filestorage-go/pkg/v1"
	"github.com/selectel/filestorage-go/pkg/v1/share"
	"github.com/selectel/filestorage-go/pkg/v1/sharenetwork"
)

func (r *shareResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan shareModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(reservedMetadataErrors(plan.Metadata)...)
	duration, d := plan.Timeouts.Create(ctx, shareTimeouts.Create)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}
	ctx, cancel := context.WithTimeout(ctx, duration)
	defer cancel()
	region := plan.Region.ValueString()
	client, err := r.client(ctx, plan.ProjectID.ValueString(), region)
	if err != nil {
		resp.Diagnostics.AddError("Connect to File Storage", err.Error())
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
	created, _, err := share.Create(ctx, client, share.CreateOpts{
		ShareProto:  plan.ShareProto.ValueString(),
		Size:        int(plan.Size.ValueInt64()),
		ShareType:   plan.ShareType.ValueString(),
		Name:        plan.Name.ValueString(),
		Description: plan.Description.ValueString(),
		Metadata:    metadata,
		NetworkID:   plan.NetworkID.ValueString(),
		SubnetID:    plan.SubnetID.ValueString(),
		IP:          plan.IP.ValueString(),
	})
	if err != nil && createRejected(err) {
		resp.Diagnostics.Append(apiError("Create file storage", region, err))
		return
	}
	id := ""
	if err == nil {
		id = created.ID
	}
	if id == "" {
		// The request must never be repeated: the share may exist. The token
		// finds it now or, through the pending ID, at a later refresh or destroy.
		cause := err
		if cause == nil {
			cause = fmt.Errorf("the response has no file storage ID")
		}
		id, err = awaitShareByToken(ctx, client, token, r.pollInterval, r.recoveryWindow)
		if err != nil || id == "" {
			resp.Diagnostics.Append(resp.State.Set(ctx, plan.pending(&share.View{ID: pendingPrefix + token}))...)
			if err != nil {
				cause = fmt.Errorf("%w; %w", cause, err)
			}
			resp.Diagnostics.AddError("File storage creation outcome is uncertain", uncertainCreateDetail("file storage", token, cause))

			return
		}
		created = &share.View{ID: id, Status: statusCreating}
	}
	// Publish identity before waiting: a timeout or an error status must not
	// lose a share that already exists. Terraform taints it, so the next apply
	// replaces it instead of creating a second one.
	resp.Diagnostics.Append(resp.State.Set(ctx, plan.pending(created))...)
	ready, err := waitShareAvailable(ctx, client, id, r.pollInterval)
	if err != nil {
		resp.Diagnostics.AddError("Wait for created file storage", err.Error())
		return
	}
	observed, err := observeShare(ctx, client, ready)
	if err != nil {
		resp.Diagnostics.Append(apiError("Read created file storage", region, err))
		return
	}
	resp.Diagnostics.Append(plan.observe(ctx, observed, plan.Metadata)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// pending is the state of a share that the API accepted but that is not ready:
// the planned arguments with the identity and status the API returned.
func (m shareModel) pending(created *share.View) *shareModel {
	m.ID = types.StringValue(created.ID)
	m.Status = types.StringValue(created.Status)
	m.AvailabilityZone = types.StringValue(created.AvailabilityZone)
	if m.IP.IsUnknown() {
		m.IP = types.StringValue("")
	}
	m.AllMetadata = types.MapNull(types.StringType)
	m.ExportLocations = types.ListNull(types.ObjectType{AttrTypes: exportLocationTypes})

	return &m
}

func (r *shareResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state shareModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	ctx, cancel := context.WithTimeout(ctx, readTimeout)
	defer cancel()
	region := state.Region.ValueString()
	client, err := r.client(ctx, state.ProjectID.ValueString(), region)
	if err != nil {
		resp.Diagnostics.AddError("Connect to File Storage", err.Error())
		return
	}
	id, err := recoverShareID(ctx, client, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.Append(apiError("Recover file storage", region, err))
		return
	}
	current, _, err := share.Get(ctx, client, id)
	if sdk.IsKind(err, sdk.KindNotFound) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.Append(apiError("Read file storage", region, err))
		return
	}
	observed, err := observeShare(ctx, client, current)
	if err != nil {
		resp.Diagnostics.Append(apiError("Read file storage", region, err))
		return
	}
	resp.Diagnostics.Append(state.observe(ctx, observed, state.Metadata)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *shareResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, prior shareModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &prior)...)
	duration, d := plan.Timeouts.Update(ctx, shareTimeouts.Update)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}
	if reserved := reservedMetadataErrors(plan.Metadata); reserved.HasError() {
		resp.Diagnostics.Append(reserved...)
		resp.Diagnostics.Append(resp.State.Set(ctx, &prior)...)

		return
	}
	ctx, cancel := context.WithTimeout(ctx, duration)
	defer cancel()
	region := plan.Region.ValueString()
	client, err := r.client(ctx, plan.ProjectID.ValueString(), region)
	if err != nil {
		resp.Diagnostics.AddError("Connect to File Storage", err.Error())
		return
	}
	id, err := recoverShareID(ctx, client, prior.ID.ValueString())
	if err != nil {
		resp.Diagnostics.Append(apiError("Recover file storage", region, err))
		resp.Diagnostics.Append(resp.State.Set(ctx, &prior)...)

		return
	}
	planned, priorKeys := map[string]string{}, map[string]string{}
	if !plan.Metadata.IsNull() {
		resp.Diagnostics.Append(plan.Metadata.ElementsAs(ctx, &planned, false)...)
	}
	if !prior.Metadata.IsNull() {
		resp.Diagnostics.Append(prior.Metadata.ElementsAs(ctx, &priorKeys, false)...)
	}
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.updateShare(ctx, client, id, plan, prior, planned, priorKeys); err != nil {
		resp.Diagnostics.Append(apiError("Update file storage", region, err))
	}
	// Read after both success and partial failure, so state holds the size and
	// metadata the API reports rather than the unfulfilled plan.
	current, _, err := share.Get(ctx, client, id)
	var observed *shareObservation
	if err == nil {
		observed, err = observeShare(ctx, client, current)
	}
	if err != nil {
		resp.Diagnostics.Append(apiError("Refresh updated file storage", region, err))
		// Keep the last known state rather than the unfulfilled plan.
		resp.Diagnostics.Append(resp.State.Set(ctx, &prior)...)

		return
	}
	managed := plan.Metadata
	if resp.Diagnostics.HasError() {
		managed = ownedAfterFailedUpdate(prior.Metadata, plan.Metadata)
	}
	plan.IP = prior.IP
	resp.Diagnostics.Append(plan.observe(ctx, observed, managed)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// updateShare changes the name and description, then the metadata, then the
// size, and stops at the first failure.
func (r *shareResource) updateShare(
	ctx context.Context,
	client *sdk.Client,
	id string,
	plan, prior shareModel,
	planned, priorKeys map[string]string,
) error {
	opts := share.UpdateOpts{}
	if !plan.Name.Equal(prior.Name) {
		v := plan.Name.ValueString()
		opts.Name = &v
	}
	if !plan.Description.Equal(prior.Description) {
		v := plan.Description.ValueString()
		opts.Description = &v
	}
	if opts.Name != nil || opts.Description != nil {
		if _, _, err := share.Update(ctx, client, id, opts); err != nil {
			return fmt.Errorf("update name and description: %w", err)
		}
	}
	// POST merges the keys, so the service keys of the fork stay; PUT would
	// replace the whole map and drop them.
	if changed := changedMetadata(prior.AllMetadata, planned); len(changed) > 0 {
		if _, _, err := share.SetMetadata(ctx, client, id, changed); err != nil {
			return fmt.Errorf("set metadata: %w", err)
		}
	}
	for key := range priorKeys {
		if _, kept := planned[key]; kept {
			continue
		}
		if _, err := share.DeleteMetadatum(ctx, client, id, key); err != nil && !sdk.IsKind(err, sdk.KindNotFound) {
			return fmt.Errorf("delete metadata key %s: %w", key, err)
		}
	}
	if plan.Size.ValueInt64() > prior.Size.ValueInt64() {
		if _, err := share.Extend(ctx, client, id, int(plan.Size.ValueInt64())); err != nil {
			return fmt.Errorf("extend: %w", err)
		}
		if _, err := waitShareExtended(ctx, client, id, int(plan.Size.ValueInt64()), r.pollInterval); err != nil {
			return fmt.Errorf("wait for extension: %w", err)
		}
	}

	return nil
}

// changedMetadata is the planned keys whose remote value differs.
func changedMetadata(remote types.Map, planned map[string]string) map[string]string {
	changed := map[string]string{}
	elements := remote.Elements()
	for key, value := range planned {
		if current, ok := elements[key]; !ok || !current.Equal(types.StringValue(value)) {
			changed[key] = value
		}
	}

	return changed
}

// ownedAfterFailedUpdate keeps every prior and planned key managed after a
// partial update; the refresh drops the keys the share does not have, and the
// next plan shows what is left to apply.
func ownedAfterFailedUpdate(prior, planned types.Map) types.Map {
	if prior.IsNull() && planned.IsNull() {
		return prior
	}
	owned := maps.Clone(prior.Elements())
	if owned == nil {
		owned = map[string]attr.Value{}
	}
	maps.Copy(owned, planned.Elements())

	return types.MapValueMust(types.StringType, owned)
}

func (r *shareResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state shareModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	duration, d := state.Timeouts.Delete(ctx, shareTimeouts.Delete)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}
	ctx, cancel := context.WithTimeout(ctx, duration)
	defer cancel()
	region := state.Region.ValueString()
	client, err := r.client(ctx, state.ProjectID.ValueString(), region)
	if err != nil {
		resp.Diagnostics.AddError("Connect to File Storage", err.Error())
		return
	}
	id, err := recoverShareID(ctx, client, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.Append(apiError("Recover file storage", region, err))
		return
	}
	current, _, err := share.Get(ctx, client, id)
	if sdk.IsKind(err, sdk.KindNotFound) {
		return
	}
	if err != nil {
		resp.Diagnostics.Append(apiError("Read file storage before deletion", region, err))
		return
	}
	// The network is read first: once the share is gone, the fork may remove it.
	var network *sharenetwork.View
	if _, flagged := current.Metadata[deleteNetworkKey]; flagged && current.ShareNetworkID != "" {
		network, _, err = sharenetwork.Get(ctx, client, current.ShareNetworkID)
		if err != nil && !sdk.IsKind(err, sdk.KindNotFound) {
			resp.Diagnostics.Append(apiError("Read share network before deletion", region, err))
			return
		}
	}
	if current.Status != statusDeleting {
		if _, err := share.Delete(ctx, client, id); err != nil && !sdk.IsKind(err, sdk.KindNotFound) {
			resp.Diagnostics.Append(apiError("Delete file storage", region, err))
			return
		}
	}
	if err := waitShareDeleted(ctx, client, id, r.pollInterval); err != nil {
		resp.Diagnostics.AddError("Wait for deleted file storage", err.Error())
		return
	}
	resp.Diagnostics.Append(r.waitNetworkDeleted(ctx, client, network)...)
}

// waitNetworkDeleted waits for the service network that the fork deletes with
// the share. The fork keeps a network that still has other subnets.
func (r *shareResource) waitNetworkDeleted(
	ctx context.Context, client *sdk.Client, network *sharenetwork.View,
) diag.Diagnostics {
	var diags diag.Diagnostics
	switch {
	case network == nil:
	case len(network.Subnets) > 1:
		// The service keeps it, and the user cannot manage it: nothing to wait for or report.
	default:
		if err := waitShareNetworkDeleted(ctx, client, network.ID, r.pollInterval); err != nil {
			diags.AddError("Wait for deleted share network", err.Error())
		}
	}

	return diags
}
