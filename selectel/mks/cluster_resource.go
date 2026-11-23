package mks

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	mksv2 "github.com/selectel/mks-go/v2/pkg"
	"github.com/selectel/mks-go/v2/pkg/cluster"
	mksclient "github.com/selectel/mks-go/v2/pkg/mksclient"
)

const defaultTimeout = 60 * time.Minute

var (
	_ resource.ResourceWithConfigure      = &clusterResource{}
	_ resource.ResourceWithImportState    = &clusterResource{}
	_ resource.ResourceWithModifyPlan     = &clusterResource{}
	_ resource.ResourceWithValidateConfig = &clusterResource{}
)

type clusterResource struct {
	connection
}

// NewClusterResource creates the cluster resource.
func NewClusterResource() resource.Resource {
	return &clusterResource{}
}

func (r *clusterResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_mks_cluster_v2"
}

func (r *clusterResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	resp.Diagnostics.Append(r.configure(req.ProviderData)...)
}

// x509PendingKey marks a created cluster whose x509_ca_certificates failed to
// be set.
const x509PendingKey = "x509_pending"

// importedKey marks an imported cluster: fields the API never returns, such
// as cloud_subnet_cidr, are null in its state although the cluster has them.
const importedKey = "imported"

func (r *clusterResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan clusterModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	timeout, diags := plan.Timeouts.Create(ctx, defaultTimeout)
	resp.Diagnostics.Append(diags...)
	projectID, diags := r.projectID(plan.ProjectID)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	pool := plan.Pool.ValueString()
	client, err := r.client(ctx, projectID, pool)
	if err != nil {
		resp.Diagnostics.AddError("Error initializing MKS client", err.Error())

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

	waitErr := newCreatedTaskWaiter(client, created.Id).Wait(ctx)

	// Save the cluster even when the wait failed, so Terraform taints it instead of losing it.
	readCtx, readCancel := readAfterWaitContext(ctx)
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
	if waitErr != nil {
		resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
		resp.Diagnostics.AddError("Error waiting for the cluster to become ready", waitErr.Error())

		return
	}

	x509Err := applyX509(ctx, client, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
	if x509Err != nil {
		// An error would taint the healthy cluster just created: the state keeps the planned value and the pending mark sends it on the next apply.
		if resp.Private != nil {
			resp.Diagnostics.Append(resp.Private.SetKey(ctx, x509PendingKey, []byte("true"))...)
		}
		resp.Diagnostics.AddWarning("X509 CA certificates of the cluster are not set",
			"The cluster is created, but setting x509_ca_certificates failed, so the next plan shows them "+
				"and the next apply sets them: "+errUpdatingObject(objectCluster, created.Id, x509Err).Error())
	}
}

// applyX509 sends the planned x509_ca_certificates of a created cluster: the
// API ignores them on create and stores them only on a PATCH that carries the
// whole kubernetes_options.
func applyX509(ctx context.Context, client *mksv2.ServiceClient, state *clusterModel) error {
	var options clusterKubeOptionsModel
	diags := state.KubernetesOptions.As(ctx, &options, basetypes.ObjectAsOptions{})
	if diags.HasError() {
		return fmt.Errorf("error reading kubernetes_options: %v", diags)
	}
	if options.X509CACertificates.ValueString() == "" {
		return nil
	}

	kubeOptions, diags := expandKubeOptions(ctx, state.KubernetesOptions)
	if diags.HasError() {
		return fmt.Errorf("error expanding kubernetes_options: %v", diags)
	}
	clusterID := state.ID.ValueString()
	err := callAndWait(ctx, client, clusterID, func() error {
		_, err := cluster.Patch(ctx, client, clusterID, &mksclient.ClusterUpdateStruct{KubernetesOptions: kubeOptions})

		return err
	})
	var got *mksclient.ClusterDetailed
	if err == nil {
		got, err = cluster.Get(ctx, client, clusterID)
	}
	if err != nil {
		return err
	}

	prior := *state
	diags = state.fromAPI(ctx, got, prior, true)
	if diags.HasError() {
		return fmt.Errorf("error reading the cluster: %v", diags)
	}

	return nil
}

func (r *clusterResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state clusterModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID, diags := r.projectID(state.ProjectID)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	pool := state.Pool.ValueString()
	client, err := r.client(ctx, projectID, pool)
	if err != nil {
		resp.Diagnostics.AddError("Error initializing MKS client", err.Error())

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
	pending, diags := req.Private.GetKey(ctx, x509PendingKey)
	resp.Diagnostics.Append(diags...)
	if len(pending) > 0 {
		// The cluster has none, see Create.
		state.KubernetesOptions, diags = withoutX509(ctx, state.KubernetesOptions)
		resp.Diagnostics.Append(diags...)
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

// withoutX509 returns the kubernetes_options with null
// x509_ca_certificates.
func withoutX509(ctx context.Context, kubeOptions types.Object) (types.Object, diag.Diagnostics) {
	if kubeOptions.IsNull() || kubeOptions.IsUnknown() {
		return kubeOptions, nil
	}
	var options clusterKubeOptionsModel
	diags := kubeOptions.As(ctx, &options, basetypes.ObjectAsOptions{})
	if diags.HasError() {
		return kubeOptions, diags
	}
	options.X509CACertificates = types.StringNull()
	withoutX509Object, d := types.ObjectValueFrom(ctx, kubeOptionsAttrTypes, options)
	diags.Append(d...)

	return withoutX509Object, diags
}

func (r *clusterResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state clusterModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	timeout, diags := plan.Timeouts.Update(ctx, defaultTimeout)
	resp.Diagnostics.Append(diags...)
	projectID, diags := r.projectID(state.ProjectID)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	pool := state.Pool.ValueString()
	client, err := r.client(ctx, projectID, pool)
	if err != nil {
		resp.Diagnostics.AddError("Error initializing MKS client", err.Error())

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
	if changed {
		err = waitStatus(ctx, client, clusterID, clusterUpdate)
		if err == nil {
			err = callAndWait(ctx, client, clusterID, func() error {
				_, err := cluster.Patch(ctx, client, clusterID, patch)

				return err
			})
		}
		if err != nil {
			resp.Diagnostics.AddError("Error updating cluster", errUpdatingObject(objectCluster, clusterID, err).Error())

			return
		}
	}

	if !plan.KubeVersion.Equal(state.KubeVersion) {
		err = upgradeKubeVersion(ctx, client, clusterID, plan.KubeVersion.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("Error upgrading cluster", errUpdatingObject(objectCluster, clusterID, err).Error())

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
	// The plan carried any pending x509_ca_certificates, see Create.
	if resp.Private != nil {
		resp.Diagnostics.Append(resp.Private.SetKey(ctx, x509PendingKey, nil)...)
	}
}

func (r *clusterResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state clusterModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	timeout, diags := state.Timeouts.Delete(ctx, defaultTimeout)
	resp.Diagnostics.Append(diags...)
	projectID, diags := r.projectID(state.ProjectID)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	pool := state.Pool.ValueString()
	client, err := r.client(ctx, projectID, pool)
	if err != nil {
		resp.Diagnostics.AddError("Error initializing MKS client", err.Error())

		return
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	clusterID := state.ID.ValueString()
	err = waitStatus(ctx, client, clusterID, clusterDelete)
	var waiter *taskWaiter
	if err == nil {
		waiter, err = newTaskWaiter(ctx, client, clusterID)
	}
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
	if resp.Private != nil {
		resp.Diagnostics.Append(resp.Private.SetKey(ctx, importedKey, []byte("true"))...)
	}
}

// upgradeKubeVersion starts from the version the cluster runs, which patch
// auto-upgrade can move past the state.
func upgradeKubeVersion(ctx context.Context, client *mksv2.ServiceClient, clusterID, desired string) error {
	err := waitStatus(ctx, client, clusterID, clusterUpgrade)
	if err != nil {
		return err
	}
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

		err = callAndWait(ctx, client, clusterID, func() error {
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

	// The minor upgrade goes to the latest patch of the next minor, so the desired version has to be that patch.
	latestPatchVersions, err := latestKubePatchVersions(ctx, client)
	if err != nil {
		return err
	}
	latest, ok := latestPatchVersions[desiredMinor]
	if !ok {
		return fmt.Errorf("unable to find the latest patch version for the minor version %s", desiredMinor)
	}
	if desired != latest {
		return fmt.Errorf("current version %s can't be upgraded to version %s, the latest available patch version is: %s",
			current, desired, latest)
	}

	err = callAndWait(ctx, client, clusterID, func() error {
		_, err := cluster.UpgradeMinorVersion(ctx, client, clusterID)

		return err
	})
	if err != nil {
		return fmt.Errorf("error upgrading minor version: %w", err)
	}

	return nil
}
