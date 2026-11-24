package mks

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	mksv2 "github.com/selectel/mks-go/v2/pkg"
	"github.com/selectel/mks-go/v2/pkg/cluster"
	"github.com/terraform-providers/terraform-provider-selectel/selectel/internal/schemadoc"
)

var clusterTimeouts = schemadoc.Timeouts{
	Create: 60 * time.Minute,
	Update: 60 * time.Minute,
	Delete: 60 * time.Minute,
}

var (
	_ resource.ResourceWithConfigure      = &clusterResource{}
	_ resource.ResourceWithImportState    = &clusterResource{}
	_ resource.ResourceWithModifyPlan     = &clusterResource{}
	_ resource.ResourceWithValidateConfig = &clusterResource{}
)

type clusterResource struct {
	connection
	pollInterval time.Duration
}

func NewClusterResource() resource.Resource {
	return &clusterResource{pollInterval: 10 * time.Second}
}

func (r *clusterResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_mks_cluster_v2"
}

func (r *clusterResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	resp.Diagnostics.Append(r.configure(req.ProviderData)...)
}

func (r *clusterResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan clusterModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	timeout, diags := plan.Timeouts.Create(ctx, clusterTimeouts.Create)
	resp.Diagnostics.Append(diags...)
	client, projectID, diags := r.client(ctx, plan.ProjectID, plan.Pool.ValueString())
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	opts, diags := expandClusterCreateOpts(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	// The state keeps the project as configured, see fromAPI, or as projectID resolved it.
	plan.ProjectID = types.StringValue(projectID)

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	created, err := cluster.Create(ctx, client, opts)
	if err != nil {
		resp.Diagnostics.AddError("Error creating cluster", errCreatingObject(objectCluster, err).Error())

		return
	}

	waitErr := newCreatedTaskWaiter(client, created.Id, r.pollInterval).Wait(ctx)
	if waitErr == nil {
		waitErr = waitActive(ctx, client, created.Id, r.pollInterval)
	}

	// Save the cluster even when the wait failed, which can be ctx ending, so Terraform taints it instead of losing it.
	readCtx, readCancel := context.WithTimeout(context.WithoutCancel(ctx), readAfterWaitTimeout)
	defer readCancel()
	got, err := cluster.Get(readCtx, client, created.Id)
	if err != nil {
		resp.Diagnostics.Append(plannedState(req.Plan, &resp.State)...)
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), created.Id)...)
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("project_id"), projectID)...)
		resp.Diagnostics.AddError("Error reading cluster", errGettingObject(objectCluster, created.Id, errors.Join(waitErr, err)).Error())

		return
	}
	state := plan
	resp.Diagnostics.Append(state.fromAPI(ctx, got, plan, true)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
	if waitErr != nil {
		resp.Diagnostics.AddError("Error waiting for the cluster to become ready", waitErr.Error())
	}
}

func (r *clusterResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state clusterModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	client, _, diags := r.client(ctx, state.ProjectID, state.Pool.ValueString())
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	got, err := cluster.Get(ctx, client, state.ID.ValueString())
	if isNotFound(err) {
		resp.State.RemoveResource(ctx)

		return
	}
	if err != nil {
		resp.Diagnostics.AddError("Error reading cluster", errGettingObject(objectCluster, state.ID.ValueString(), err).Error())

		return
	}

	prior := state
	resp.Diagnostics.Append(state.fromAPI(ctx, got, prior, false)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *clusterResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state clusterModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	timeout, diags := plan.Timeouts.Update(ctx, clusterTimeouts.Update)
	resp.Diagnostics.Append(diags...)
	client, _, diags := r.client(ctx, state.ProjectID, state.Pool.ValueString())
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	patch, changed, diags := expandClusterPatch(ctx, plan, state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	clusterID := state.ID.ValueString()
	if !plan.KubeVersion.Equal(state.KubeVersion) {
		err := upgradeKubeVersion(ctx, client, clusterID, plan.KubeVersion.ValueString(), r.pollInterval)
		if err != nil {
			resp.Diagnostics.AddError("Error upgrading cluster", errUpdatingObject(objectCluster, clusterID, err).Error())

			return
		}
	}

	if changed {
		err := callAndWait(ctx, client, clusterID, r.pollInterval, func() error {
			_, err := cluster.Patch(ctx, client, clusterID, patch)

			return err
		})
		if err != nil {
			resp.Diagnostics.AddError("Error updating cluster", errUpdatingObject(objectCluster, clusterID, err).Error())

			return
		}
	}

	got, err := cluster.Get(ctx, client, clusterID)
	if err != nil {
		resp.Diagnostics.AddError("Error reading cluster", errGettingObject(objectCluster, clusterID, err).Error())

		return
	}
	resp.Diagnostics.Append(plan.fromAPI(ctx, got, plan, true)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *clusterResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state clusterModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	timeout, diags := state.Timeouts.Delete(ctx, clusterTimeouts.Delete)
	resp.Diagnostics.Append(diags...)
	client, _, diags := r.client(ctx, state.ProjectID, state.Pool.ValueString())
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	clusterID := state.ID.ValueString()
	waiter, err := newTaskWaiter(ctx, client, clusterID, r.pollInterval)
	if err == nil {
		err = cluster.Delete(ctx, client, clusterID)
	}
	if isNotFound(err) {
		return
	}
	if err == nil {
		err = waiter.WaitClusterDeleted(ctx)
	}
	if err != nil {
		resp.Diagnostics.AddError("Error deleting cluster", errDeletingObject(objectCluster, clusterID, err).Error())
	}
}

// ImportState takes the project and the pool from the provider configuration,
// because the cluster ID carries neither, like selectel_mks_cluster_v1.
func (r *clusterResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	projectID, pool, diags := r.importScope()
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("project_id"), projectID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("pool"), pool)...)
}

// upgradeKubeVersion starts from the version the cluster runs, which patch
// auto-upgrade can move past the state.
func upgradeKubeVersion(ctx context.Context, client *mksv2.ServiceClient, clusterID, desired string, interval time.Duration) error {
	got, err := cluster.Get(ctx, client, clusterID)
	if err != nil {
		return err
	}
	current := got.KubeVersion
	order, err := compareKubeVersions(desired, current)
	if err != nil {
		return err
	}
	if order <= 0 {
		// Patch auto-upgrade, or an upgrade outside Terraform, got there first.
		return nil
	}

	currentMinor, err := kubeVersionTrimToMinor(current)
	if err != nil {
		return err
	}
	desiredMinor, err := kubeVersionTrimToMinor(desired)
	if err != nil {
		return err
	}

	if desiredMinor == currentMinor {
		latestPatchVersions, err := latestKubePatchVersions(ctx, client)
		if err != nil {
			return err
		}
		latest, ok := latestPatchVersions[currentMinor]
		if !ok {
			return fmt.Errorf("unable to find the latest patch version for the current minor version %s", currentMinor)
		}
		if desired != latest {
			return fmt.Errorf("current version %s can't be upgraded to version %s, the latest available patch version is: %s",
				current, desired, latest)
		}

		err = callAndWait(ctx, client, clusterID, interval, func() error {
			_, err := cluster.UpgradePatchVersion(ctx, client, clusterID)

			return err
		})
		if err != nil {
			return fmt.Errorf("error upgrading patch version: %w", err)
		}

		return nil
	}

	nextMinor, err := kubeVersionTrimToMinorIncremented(current)
	if err != nil {
		return err
	}
	if desiredMinor != nextMinor {
		return fmt.Errorf("current version %s can't be upgraded to version %s, kubernetes versions must be upgraded one minor version at a time",
			current, desired)
	}

	err = callAndWait(ctx, client, clusterID, interval, func() error {
		_, err := cluster.UpgradeMinorVersion(ctx, client, clusterID)

		return err
	})
	if err != nil {
		return fmt.Errorf("error upgrading minor version: %w", err)
	}

	return nil
}
