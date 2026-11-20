package mks

import (
	"cmp"
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	mksv2 "github.com/selectel/mks-go/v2/pkg"
	"github.com/selectel/mks-go/v2/pkg/cluster"
	"github.com/selectel/mks-go/v2/pkg/mksclient"
	"github.com/terraform-providers/terraform-provider-selectel/selectel/internal/schemadoc"
)

const defaultTimeout = 60 * time.Minute

var (
	_ resource.ResourceWithConfigure      = &clusterResource{}
	_ resource.ResourceWithImportState    = &clusterResource{}
	_ resource.ResourceWithModifyPlan     = &clusterResource{}
	_ resource.ResourceWithValidateConfig = &clusterResource{}
)

// clusterLocation is the managed kubernetes section of the product
// availability page.
const clusterLocation = "managed-kubernetes"

var clusterDocs = schemadoc.Resource{Name: "cluster", Location: clusterLocation}

// workersType is the user-facing name of the API network_type.
var workersTypes = map[string]string{
	"CLOUD":     "STANDARD",
	"DEDICATED": "L3VPN",
}

var (
	ciliumAttrTypes = map[string]attr.Type{
		"envoy_daemonset": types.BoolType,
		"hubble_relay":    types.BoolType,
	}
	auditLogsAttrTypes = map[string]attr.Type{
		"enabled":     types.BoolType,
		"secret_name": types.StringType,
	}
	oidcAttrTypes = map[string]attr.Type{
		"enabled":        types.BoolType,
		"provider_name":  types.StringType,
		"issuer_url":     types.StringType,
		"client_id":      types.StringType,
		"username_claim": types.StringType,
		"groups_claim":   types.StringType,
		"ca_certs":       trimmedStringType{},
	}
	kubeOptionsAttrTypes = map[string]attr.Type{
		"feature_gates":         types.SetType{ElemType: types.StringType},
		"admission_controllers": types.SetType{ElemType: types.StringType},
		"audit_logs":            types.ObjectType{AttrTypes: auditLogsAttrTypes},
		"oidc":                  types.ObjectType{AttrTypes: oidcAttrTypes},
		"x509_ca_certificates":  types.StringType,
	}
)

type clusterModel struct {
	ID                            types.String   `tfsdk:"id"`
	Name                          types.String   `tfsdk:"name"`
	Pool                          types.String   `tfsdk:"pool"`
	ProjectID                     types.String   `tfsdk:"project_id"`
	KubeVersion                   types.String   `tfsdk:"kube_version"`
	ClusterType                   types.String   `tfsdk:"cluster_type"`
	WorkersType                   types.String   `tfsdk:"workers_type"`
	NetworkID                     types.String   `tfsdk:"network_id"`
	SubnetID                      types.String   `tfsdk:"subnet_id"`
	CloudSubnetCIDR               types.String   `tfsdk:"cloud_subnet_cidr"`
	PrivateKubeAPI                types.Bool     `tfsdk:"private_kube_api"`
	MaintenanceWindowStart        types.String   `tfsdk:"maintenance_window_start"`
	MaintenanceWindowEnd          types.String   `tfsdk:"maintenance_window_end"`
	EnableAutorepair              types.Bool     `tfsdk:"enable_autorepair"`
	EnablePatchVersionAutoUpgrade types.Bool     `tfsdk:"enable_patch_version_auto_upgrade"`
	CNIType                       types.String   `tfsdk:"cni_type"`
	CNICiliumSettings             types.Object   `tfsdk:"cni_cilium_settings"`
	KubernetesOptions             types.Object   `tfsdk:"kubernetes_options"`
	Status                        types.String   `tfsdk:"status"`
	KubeAPIIP                     types.String   `tfsdk:"kube_api_ip"`
	Timeouts                      timeouts.Value `tfsdk:"timeouts"`
}

type clusterCiliumModel struct {
	EnvoyDaemonset types.Bool `tfsdk:"envoy_daemonset"`
	HubbleRelay    types.Bool `tfsdk:"hubble_relay"`
}

type clusterKubeOptionsModel struct {
	FeatureGates         types.Set    `tfsdk:"feature_gates"`
	AdmissionControllers types.Set    `tfsdk:"admission_controllers"`
	AuditLogs            types.Object `tfsdk:"audit_logs"`
	OIDC                 types.Object `tfsdk:"oidc"`
	X509CACertificates   types.String `tfsdk:"x509_ca_certificates"`
}

type clusterAuditLogsModel struct {
	Enabled    types.Bool   `tfsdk:"enabled"`
	SecretName types.String `tfsdk:"secret_name"`
}

type clusterOIDCModel struct {
	Enabled       types.Bool         `tfsdk:"enabled"`
	ProviderName  types.String       `tfsdk:"provider_name"`
	IssuerURL     types.String       `tfsdk:"issuer_url"`
	ClientID      types.String       `tfsdk:"client_id"`
	UsernameClaim types.String       `tfsdk:"username_claim"`
	GroupsClaim   types.String       `tfsdk:"groups_claim"`
	CACerts       trimmedStringValue `tfsdk:"ca_certs"`
}

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

func (r *clusterResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	optionalString := func() schema.StringAttribute {
		return schema.StringAttribute{
			Optional: true, Computed: true,
			PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
		}
	}
	optionalBool := func() schema.BoolAttribute {
		return schema.BoolAttribute{
			Optional: true, Computed: true,
			PlanModifiers: []planmodifier.Bool{boolplanmodifier.UseStateForUnknown()},
		}
	}
	oidcParam := func(enabledDefault string) schema.StringAttribute {
		return schema.StringAttribute{
			Optional:      true,
			Computed:      true,
			PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown(), oidcPlanModifier{enabledDefault: enabledDefault}},
		}
	}

	resp.Schema = schema.Schema{
		MarkdownDescription: "Creates and manages a cluster in Managed Kubernetes using the v2 public API. " +
			"For more information about Managed Kubernetes, see the [official Selectel documentation](https://docs.selectel.ru/en/cloud/managed-kubernetes/).\n\n" +
			"When a `kubernetes_options` or `cni_cilium_settings` task fails, the API has already recorded the new values, " +
			"so the next plan shows no change. Check `status`.",
		Attributes: map[string]schema.Attribute{
			"id": clusterDocs.ID(),
			"name": clusterDocs.String(schema.StringAttribute{
				Required:      true,
				PlanModifiers: []planmodifier.String{nameCase{}, schemadoc.RequiresReplace()},
			}, schemadoc.Field{
				Definition: "Cluster name. It is included in the names of the cluster entities: " +
					"node groups, nodes, load balancers, networks, and volumes. " +
					"The API stores the name in lower case, so a change of only the letter case is not a change.",
			}),
			"pool": clusterDocs.Region(),
			"project_id": clusterDocs.String(schema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					sameProjectID{},
					stringplanmodifier.UseStateForUnknown(),
					schemadoc.RequiresReplace(),
				},
			}, schemadoc.Field{
				Definition: "Unique identifier of the associated project, for example, `a07abc12310546f1b9291ab301a7d75b`.",
				Behavior:   "If omitted, the `project_id` of the provider configuration is used.",
				Source:     schemadoc.FromResource("selectel_vpc_project_v2", "project ID"),
				Concept: schemadoc.Link{
					Title: "Projects",
					URL:   "https://docs.selectel.ru/en/control-panel-actions/projects/about-projects/",
				},
			}),
			"kube_version": clusterDocs.String(schema.StringAttribute{
				Required: true,
			}, schemadoc.Field{
				Definition: "Kubernetes version of the cluster. The format is `x.y.z`. Raising it upgrades the cluster. " +
					"A patch upgrade is possible only to the latest patch version of the current minor version. " +
					"Another patch version fails the apply. " +
					"A minor upgrade takes the next minor version. Skipping a minor version fails the plan. " +
					"The cluster may run a newer version than configured, for example after patch auto-upgrade: " +
					"a configured version lower than the one in the state is accepted with no change and never downgrades the cluster.",
			}),
			"cluster_type": clusterDocs.String(schema.StringAttribute{
				Optional: true,
				Computed: true,
				Validators: []validator.String{schemadoc.OneOf(
					string(mksclient.ClusterCreateStructClusterTypeBASIC),
					string(mksclient.ClusterCreateStructClusterTypeHIGHAVAILABILITY),
					string(mksclient.ClusterCreateStructClusterTypeHIGHAVAILABILITYMULTIAZ),
				)},
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown(), schemadoc.RequiresReplace()},
			}, schemadoc.Field{
				Definition: "Type of the cluster: `BASIC` has one master node, `HIGH_AVAILABILITY` has three master nodes " +
					"in one availability zone, and `HIGH_AVAILABILITY_MULTI_AZ` has three master nodes in three availability zones.",
				Behavior: "If omitted, the API creates a `HIGH_AVAILABILITY` cluster. If `cluster_type` is `BASIC`, " +
					"set `enable_patch_version_auto_upgrade` to `false` or omit it.",
			}),
			"workers_type": clusterDocs.String(schema.StringAttribute{
				Required:      true,
				Validators:    []validator.String{schemadoc.OneOf("CLOUD", "DEDICATED")},
				PlanModifiers: []planmodifier.String{schemadoc.RequiresReplace()},
			}, schemadoc.Field{
				Definition: "Type of the worker nodes the cluster accepts: `CLOUD` allows only cloud node groups, " +
					"and `DEDICATED` allows only dedicated servers.",
				Behavior: "Node groups of the other type cannot be added to the cluster, " +
					"and switching the type recreates the cluster with all its node groups.",
			}),
			"network_id": clusterDocs.String(schema.StringAttribute{
				Optional:      true,
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown(), schemadoc.RequiresReplace()},
			}, schemadoc.Field{
				Definition: "Unique identifier of the network of the cluster.",
				Behavior:   "If omitted, the API creates a network.",
				Source:     schemadoc.FromResource("openstack_networking_network_v2", "network ID"),
			}),
			"subnet_id": clusterDocs.String(schema.StringAttribute{
				Optional:      true,
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown(), schemadoc.RequiresReplace()},
			}, schemadoc.Field{
				Definition: "Unique identifier of the subnet of the cluster.",
				Behavior:   "If omitted, the API creates a subnet.",
				Source:     schemadoc.FromResource("openstack_networking_subnet_v2", "subnet ID"),
			}),
			"cloud_subnet_cidr": clusterDocs.String(schema.StringAttribute{
				Optional: true,
				PlanModifiers: []planmodifier.String{schemadoc.RequiresReplaceIf(
					func(ctx context.Context, req planmodifier.StringRequest, resp *stringplanmodifier.RequiresReplaceIfFuncResponse) {
						imported, diags := req.Private.GetKey(ctx, importedKey)
						resp.Diagnostics.Append(diags...)
						resp.RequiresReplace = !req.StateValue.IsNull() || len(imported) == 0
					},
					"A change from an unset value after import does not replace the cluster.")},
			}, schemadoc.Field{
				Definition: "CIDR of the subnet the API creates for the cluster. The API does not return it, " +
					"so an imported cluster has no value.",
			}),
			"private_kube_api": clusterDocs.Bool(schema.BoolAttribute{
				Optional:      true,
				Computed:      true,
				Default:       schemadoc.StaticBool(false),
				PlanModifiers: []planmodifier.Bool{schemadoc.BoolRequiresReplace()},
			}, schemadoc.Field{
				Definition: "Whether the Kube API is available only from the cluster network instead of the Internet.",
			}),
			"maintenance_window_start": clusterDocs.String(optionalString(), schemadoc.Field{
				Definition: "Time in UTC when maintenance in the cluster starts. The format is `hh:mm:ss`.",
			}),
			"enable_autorepair": clusterDocs.Bool(schema.BoolAttribute{
				Optional:      true,
				Computed:      true,
				Default:       schemadoc.StaticBool(true),
				PlanModifiers: []planmodifier.Bool{boolplanmodifier.UseStateForUnknown()},
			}, schemadoc.Field{
				Definition: "Enables or disables automatic reinstallation of worker nodes that are unavailable or unhealthy. " +
					"Auto-recovery is unavailable with a single worker node and on dedicated servers, " +
					"and it deletes all data on the boot disk of the node. " +
					"For more information, see the [official Selectel documentation](https://docs.selectel.ru/en/cloud/managed-kubernetes/node-groups/reinstall-nodes/).",
			}),
			"enable_patch_version_auto_upgrade": clusterDocs.Bool(optionalBool(), schemadoc.Field{
				Definition: "Enables or disables automatic upgrade of the cluster to the latest patch version during the maintenance window. " +
					"If omitted, the API enables it for all cluster types except `BASIC`. " +
					"If `cluster_type` is `BASIC`, set this argument to `false` or omit it.",
			}),
			"cni_type": clusterDocs.String(schema.StringAttribute{
				Optional: true,
				Computed: true,
				Validators: []validator.String{schemadoc.OneOf(
					string(mksclient.ClusterCniTypeCALICO), string(mksclient.ClusterCniTypeCILIUM),
				)},
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown(), schemadoc.RequiresReplace()},
			}, schemadoc.Field{
				Definition: "Type of the CNI of the cluster.",
				Behavior:   "If omitted, the API uses `CALICO`. To set `cni_cilium_settings`, set this argument to `CILIUM`.",
			}),
			"cni_cilium_settings": schema.SingleNestedAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Settings of the Cilium CNI. If you set this argument, set `cni_type` to `CILIUM`.",
				PlanModifiers:       []planmodifier.Object{objectplanmodifier.UseStateForUnknown()},
				Attributes: map[string]schema.Attribute{
					"envoy_daemonset": clusterDocs.Bool(optionalBool(), schemadoc.Field{
						Definition: "Enables or disables the Envoy DaemonSet for Cilium. If omitted, the API uses `true`.",
					}),
					"hubble_relay": clusterDocs.Bool(optionalBool(), schemadoc.Field{
						Definition: "Enables or disables Hubble Relay for Cilium. If omitted, the API uses `true`.",
					}),
				},
			},
			"kubernetes_options": schema.SingleNestedAttribute{
				Optional: true,
				Computed: true,
				MarkdownDescription: "Kubernetes options of the cluster. The provider always sends the whole object, " +
					"because the API replaces the options entirely.",
				PlanModifiers: []planmodifier.Object{objectplanmodifier.UseStateForUnknown()},
				Attributes: map[string]schema.Attribute{
					"feature_gates": clusterDocs.Set(schema.SetAttribute{
						Optional:      true,
						Computed:      true,
						ElementType:   types.StringType,
						PlanModifiers: []planmodifier.Set{setplanmodifier.UseStateForUnknown()},
					}, schemadoc.Field{
						Definition: "Enabled feature gates.",
						Concept: schemadoc.Link{
							Title: "Feature gates",
							URL:   "https://docs.selectel.ru/en/cloud/managed-kubernetes/clusters/feature-gates/",
						},
					}),
					"admission_controllers": clusterDocs.Set(schema.SetAttribute{
						Optional:      true,
						Computed:      true,
						ElementType:   types.StringType,
						PlanModifiers: []planmodifier.Set{setplanmodifier.UseStateForUnknown()},
					}, schemadoc.Field{
						Definition: "Enabled admission controllers.",
						Concept: schemadoc.Link{
							Title: "Admission controllers",
							URL:   "https://docs.selectel.ru/en/cloud/managed-kubernetes/clusters/admission-controllers/",
						},
					}),
					"audit_logs": schema.SingleNestedAttribute{
						Optional:            true,
						Computed:            true,
						MarkdownDescription: "Collection of Kubernetes audit logs. Learn how to [configure export of audit logs to a logging system](https://docs.selectel.ru/en/cloud/managed-kubernetes/clusters/logs/#configure-export-of-audit-logs).",
						PlanModifiers:       []planmodifier.Object{objectplanmodifier.UseStateForUnknown()},
						Attributes: map[string]schema.Attribute{
							"enabled": clusterDocs.Bool(optionalBool(), schemadoc.Field{
								Definition: "Enables or disables collection of audit logs.",
							}),
							"secret_name": clusterDocs.String(optionalString(), schemadoc.Field{
								Definition: "Name of the secret in the `kube-system` namespace " +
									"with the credentials of the logging system. The API does not apply a change of only this field.",
							}),
						},
					},
					"oidc": schema.SingleNestedAttribute{
						Optional: true,
						Computed: true,
						MarkdownDescription: "Connection of an OpenID Connect (OIDC) provider to the cluster. " +
							"Learn how to [configure the OIDC provider in the cluster](https://docs.selectel.ru/en/cloud/managed-kubernetes/clusters/access-to-cluster-with-oidc-provider/#configure-oidc-connection). " +
							"Disabling OIDC clears its settings, so with `enabled = false` omit the other arguments. A non-empty one fails the plan.",
						PlanModifiers: []planmodifier.Object{objectplanmodifier.UseStateForUnknown()},
						Attributes: map[string]schema.Attribute{
							"enabled": clusterDocs.Bool(optionalBool(), schemadoc.Field{
								Definition: "Enables or disables authentication with OIDC.",
								Behavior:   "If `false`, omit the other arguments of `oidc`.",
							}),
							"provider_name": clusterDocs.String(oidcParam(""), schemadoc.Field{
								Definition: "Name of the connection, for identification only. " +
									"The API does not apply a change of only this field.",
							}),
							"issuer_url": clusterDocs.String(oidcParam(""), schemadoc.Field{
								Definition: "URL of the OIDC provider. It must start with `https://`.",
							}),
							"client_id": clusterDocs.String(oidcParam(""), schemadoc.Field{
								Definition: "Client ID that all tokens must be issued for.",
							}),
							"username_claim": clusterDocs.String(oidcParam(oidcUsernameClaim), schemadoc.Field{
								Definition: "JWT claim to use as the username. " +
									"When OIDC is enabled, omit it to get `sub`. An empty value fails the plan.",
							}),
							"groups_claim": clusterDocs.String(oidcParam(oidcGroupsClaim), schemadoc.Field{
								Definition: "JWT claim to use as the user's group. " +
									"When OIDC is enabled, omit it to get `groups`. An empty value fails the plan.",
							}),
							"ca_certs": clusterDocs.String(schema.StringAttribute{
								Optional:      true,
								Computed:      true,
								CustomType:    trimmedStringType{},
								PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown(), trimmedStringKeep{}, oidcPlanModifier{}},
							}, schemadoc.Field{
								Definition: "CA certificates of the OIDC provider. The format is PEM. " +
									"Leading and trailing whitespace is ignored.",
							}),
						},
					},
					"x509_ca_certificates": clusterDocs.String(schema.StringAttribute{
						Optional: true,
					}, schemadoc.Field{
						Definition: "Custom X509 CA certificates for the cluster components, base64-encoded. " +
							"The API does not return them, so an imported cluster has no value. If setting them fails " +
							"after the cluster is created, the apply ends with a warning and the next apply sets them. " +
							"The API keeps the certificates of the cluster: removing them from the configuration " +
							"does not remove them from the cluster. New certificates can be set instead.",
					}),
				},
			},
			"status": clusterDocs.String(schema.StringAttribute{
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			}, schemadoc.Field{
				Definition: "Cluster status.",
			}),
			"kube_api_ip": clusterDocs.String(schema.StringAttribute{
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			}, schemadoc.Field{
				Definition: "IP address of the Kube API.",
			}),
			"maintenance_window_end": clusterDocs.String(schema.StringAttribute{
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			}, schemadoc.Field{
				Definition: "Time in UTC when maintenance in the cluster ends. The format is `hh:mm:ss`.",
			}),
		},
		Blocks: map[string]schema.Block{
			"timeouts": clusterDocs.Timeouts(ctx, schemadoc.Timeouts{Create: defaultTimeout, Update: defaultTimeout, Delete: defaultTimeout}),
		},
	}
}

func (r *clusterResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var clusterType, cniType types.String
	var autoUpgrade types.Bool
	var cilium types.Object
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("cluster_type"), &clusterType)...)
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("enable_patch_version_auto_upgrade"), &autoUpgrade)...)
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("cni_type"), &cniType)...)
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("cni_cilium_settings"), &cilium)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// The API stores Cilium settings only for a Cilium cluster and returns
	// null otherwise.
	if !cilium.IsNull() && !cilium.IsUnknown() && !cniType.IsUnknown() &&
		cniType.ValueString() != string(mksclient.ClusterCniTypeCILIUM) {
		resp.Diagnostics.AddAttributeError(path.Root("cni_cilium_settings"),
			"Cilium settings need the Cilium CNI",
			"Set cni_type to CILIUM or remove cni_cilium_settings.")
	}

	// The API rejects it.
	if clusterType.ValueString() == string(mksclient.ClusterCreateStructClusterTypeBASIC) && autoUpgrade.ValueBool() {
		resp.Diagnostics.AddAttributeError(path.Root("enable_patch_version_auto_upgrade"),
			"Patch version auto-upgrade is not available for BASIC clusters",
			"Set enable_patch_version_auto_upgrade to false or omit it when cluster_type is BASIC.")
	}
}

func (r *clusterResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	if req.State.Raw.IsNull() || req.Plan.Raw.IsNull() {
		return
	}

	var state, plan clusterModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !plan.KubeVersion.IsUnknown() && !plan.KubeVersion.Equal(state.KubeVersion) {
		order, err := compareKubeVersions(plan.KubeVersion.ValueString(), state.KubeVersion.ValueString())
		if err == nil && order < 0 {
			// Like the v1 diff suppress: a lower configured version, which
			// patch auto-upgrade leaves behind, changes nothing.
			plan.KubeVersion = state.KubeVersion
			resp.Diagnostics.Append(resp.Plan.SetAttribute(ctx, path.Root("kube_version"), state.KubeVersion)...)
		} else {
			err = validateKubeVersionChange(state.KubeVersion.ValueString(), plan.KubeVersion.ValueString())
			if err != nil {
				resp.Diagnostics.AddAttributeError(path.Root("kube_version"), "Invalid Kubernetes version change", err.Error())

				return
			}
		}
	}

	// Any update can change the status: a task-creating PATCH moves the
	// cluster to PENDING_UPGRADE_CLUSTER_CONFIG, and a finished task leaves it
	// ACTIVE or MAINTENANCE, whatever it was at refresh.
	if !resp.Plan.Raw.Equal(req.State.Raw) {
		resp.Diagnostics.Append(resp.Plan.SetAttribute(ctx, path.Root("status"), types.StringUnknown())...)
	}
	// The window end follows these changes and cannot be kept from the state.
	if !plan.KubeVersion.Equal(state.KubeVersion) || !plan.MaintenanceWindowStart.Equal(state.MaintenanceWindowStart) {
		resp.Diagnostics.Append(resp.Plan.SetAttribute(ctx, path.Root("maintenance_window_end"), types.StringUnknown())...)
	}

	x509Plan := ""
	planKubeOptions := plan.KubernetesOptions
	if !planKubeOptions.IsNull() && !planKubeOptions.IsUnknown() {
		var options clusterKubeOptionsModel
		if !planKubeOptions.As(ctx, &options, basetypes.ObjectAsOptions{}).HasError() {
			x509Plan = options.X509CACertificates.ValueString()
		}
	}
	if x509Plan == "" {
		x509State := ""
		stateKubeOptions := state.KubernetesOptions
		if !stateKubeOptions.IsNull() && !stateKubeOptions.IsUnknown() {
			var options clusterKubeOptionsModel
			if !stateKubeOptions.As(ctx, &options, basetypes.ObjectAsOptions{}).HasError() {
				x509State = options.X509CACertificates.ValueString()
			}
		}
		if x509State != "" {
			resp.Diagnostics.AddAttributeError(path.Root("kubernetes_options").AtName("x509_ca_certificates"),
				"X509 CA certificates cannot be removed",
				"The API keeps the X509 CA certificates of the cluster, an empty value does not remove them. "+
					"New certificates can be set instead.")
		}
	}
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
	// The state keeps the project as configured, see fromAPI, or as
	// projectID resolved it.
	plan.ProjectID = types.StringValue(projectID)

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	created, err := cluster.Create(ctx, client, opts)
	if err != nil {
		resp.Diagnostics.AddError("Error creating cluster", errCreatingObject(objectCluster, err).Error())

		return
	}

	waitErr := newCreatedTaskWaiter(client, created.Id, "").Wait(ctx)

	// Save the cluster even when the wait failed, so Terraform taints it
	// instead of losing it. The wait may have used up ctx.
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
		// An error would taint the healthy cluster just created. The state
		// keeps the planned value, which Terraform requires, and the mark
		// makes the next refresh drop it, so the next apply sends it.
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
		err = callAndWait(ctx, client, clusterID, func() error {
			_, err := cluster.Patch(ctx, client, clusterID, patch)

			return err
		})
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
	err = waitDeletable(ctx, func() ([]statusValue, error) {
		c, err := cluster.Get(ctx, client, clusterID)
		if err != nil {
			return nil, err
		}

		return []statusValue{{object: objectCluster, status: string(c.Status), accepted: clusterDeleteStatuses}}, nil
	})
	var waiter *taskWaiter
	if err == nil {
		waiter, err = newTaskWaiter(ctx, client, clusterID, "")
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

// validateKubeVersionChange rejects at plan the changes no upgrade action
// performs: a major change, a downgrade, a skipped minor version.
func validateKubeVersionChange(current, desired string) error {
	currentMajor, err := kubeVersionToMajor(current)
	if err != nil {
		return err
	}
	desiredMajor, err := kubeVersionToMajor(desired)
	if err != nil {
		return err
	}
	if currentMajor != desiredMajor {
		return fmt.Errorf("current version %s can't be upgraded to version %s", current, desired)
	}

	currentMinor, err := kubeVersionToMinor(current)
	if err != nil {
		return err
	}
	desiredMinor, err := kubeVersionToMinor(desired)
	if err != nil {
		return err
	}
	currentPatch, err := kubeVersionToPatch(current)
	if err != nil {
		return err
	}
	desiredPatch, err := kubeVersionToPatch(desired)
	if err != nil {
		return err
	}

	switch {
	case desiredMinor > currentMinor+1:
		return fmt.Errorf("current version %s can't be upgraded to version %s, kubernetes versions must be upgraded one minor version at a time",
			current, desired)
	case desiredMinor < currentMinor, desiredMinor == currentMinor && desiredPatch < currentPatch:
		return fmt.Errorf("current version %s can't be downgraded to version %s", current, desired)
	}

	return nil
}

// compareKubeVersions compares two x.y.z versions like cmp.Compare.
func compareKubeVersions(a, b string) (int, error) {
	for _, part := range []func(string) (int, error){kubeVersionToMajor, kubeVersionToMinor, kubeVersionToPatch} {
		partA, err := part(a)
		if err != nil {
			return 0, err
		}
		partB, err := part(b)
		if err != nil {
			return 0, err
		}
		order := cmp.Compare(partA, partB)
		if order != 0 {
			return order, nil
		}
	}

	return 0, nil
}

// upgradeKubeVersion starts from the version the cluster runs, which patch
// auto-upgrade can move past the state.
func upgradeKubeVersion(ctx context.Context, client *mksv2.ServiceClient, clusterID, desired string) error {
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
		// Patch auto-upgrade, or an upgrade outside Terraform, got there
		// first.
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

	// The minor upgrade goes to the latest patch of the next minor, so the
	// desired version has to be that patch.
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

// kubeVersionKeepsState keeps the version from the state while the cluster
// runs a newer patch of the same minor version: patch auto-upgrade moves the
// cluster past the configuration. Any other difference, such as a minor
// upgrade outside Terraform, is read as is; ModifyPlan then accepts a lower
// configured version with no change.
func kubeVersionKeepsState(prior, actual string) string {
	if prior == "" {
		return actual
	}
	priorMinor, err := kubeVersionTrimToMinor(prior)
	if err != nil {
		return actual
	}
	actualMinor, err := kubeVersionTrimToMinor(actual)
	if err != nil || priorMinor != actualMinor {
		return actual
	}
	order, err := compareKubeVersions(prior, actual)
	if err != nil || order > 0 {
		return actual
	}

	return prior
}

func expandClusterCreateOpts(ctx context.Context, plan clusterModel) (*mksclient.ClusterCreateStruct, diag.Diagnostics) {
	opts := &mksclient.ClusterCreateStruct{
		Name:        plan.Name.ValueString(),
		Pool:        plan.Pool.ValueString(),
		KubeVersion: plan.KubeVersion.ValueString(),
		// Deprecated in the API, cluster_type wins.
		Basic:                         false,
		NetworkType:                   workersTypes[plan.WorkersType.ValueString()],
		NetworkId:                     plan.NetworkID.ValueString(),
		SubnetId:                      plan.SubnetID.ValueString(),
		CloudSubnetCidr:               plan.CloudSubnetCIDR.ValueString(),
		PrivateKubeApi:                plan.PrivateKubeAPI.ValueBool(),
		MaintenanceWindowStart:        plan.MaintenanceWindowStart.ValueString(),
		EnableAutorepair:              knownBoolPointer(plan.EnableAutorepair),
		EnablePatchVersionAutoUpgrade: knownBoolPointer(plan.EnablePatchVersionAutoUpgrade),
		CniType:                       knownStringPointer(plan.CNIType),
	}
	if !plan.ClusterType.IsNull() && !plan.ClusterType.IsUnknown() {
		clusterType := mksclient.ClusterCreateStructClusterType(plan.ClusterType.ValueString())
		opts.ClusterType = &clusterType
	}

	var diags diag.Diagnostics
	opts.CniCiliumSettings, diags = expandCilium(ctx, plan.CNICiliumSettings)
	kubeOptions, kubeOptionsDiags := expandKubeOptions(ctx, plan.KubernetesOptions)
	diags.Append(kubeOptionsDiags...)
	opts.KubernetesOptions = kubeOptions

	return opts, diags
}

// expandClusterPatch collects the changed in-place fields. The API replaces
// kubernetes_options entirely, so the whole plan object goes.
func expandClusterPatch(ctx context.Context, plan, state clusterModel) (*mksclient.ClusterUpdateStruct, bool, diag.Diagnostics) {
	var diags diag.Diagnostics
	patch := &mksclient.ClusterUpdateStruct{}
	changed := false

	if !plan.MaintenanceWindowStart.IsUnknown() && !plan.MaintenanceWindowStart.Equal(state.MaintenanceWindowStart) {
		patch.MaintenanceWindowStart = plan.MaintenanceWindowStart.ValueStringPointer()
		changed = true
	}
	if !plan.EnableAutorepair.IsUnknown() && !plan.EnableAutorepair.Equal(state.EnableAutorepair) {
		patch.EnableAutorepair = plan.EnableAutorepair.ValueBoolPointer()
		changed = true
	}
	if !plan.EnablePatchVersionAutoUpgrade.IsUnknown() && !plan.EnablePatchVersionAutoUpgrade.Equal(state.EnablePatchVersionAutoUpgrade) {
		patch.EnablePatchVersionAutoUpgrade = plan.EnablePatchVersionAutoUpgrade.ValueBoolPointer()
		changed = true
	}
	if !plan.CNICiliumSettings.Equal(state.CNICiliumSettings) {
		cilium, ciliumDiags := expandCilium(ctx, plan.CNICiliumSettings)
		diags.Append(ciliumDiags...)
		patch.CniCiliumSettings = cilium
		changed = changed || cilium != nil
	}
	if !plan.KubernetesOptions.Equal(state.KubernetesOptions) {
		kubeOptions, kubeOptionsDiags := expandKubeOptions(ctx, plan.KubernetesOptions)
		diags.Append(kubeOptionsDiags...)
		patch.KubernetesOptions = kubeOptions
		changed = changed || kubeOptions != nil
	}

	return patch, changed, diags
}

func expandCilium(ctx context.Context, obj types.Object) (*mksclient.CNICiliumSettings, diag.Diagnostics) {
	if obj.IsNull() || obj.IsUnknown() {
		return nil, nil
	}

	var m clusterCiliumModel
	diags := obj.As(ctx, &m, basetypes.ObjectAsOptions{})

	return &mksclient.CNICiliumSettings{
		EnvoyDaemonset: knownBoolPointer(m.EnvoyDaemonset),
		HubbleRelay:    knownBoolPointer(m.HubbleRelay),
	}, diags
}

// expandKubeOptions converts the object; unknown parts, possible only on
// create, are left to the API defaults.
func expandKubeOptions(ctx context.Context, obj types.Object) (*mksclient.KubernetesOptions, diag.Diagnostics) {
	if obj.IsNull() || obj.IsUnknown() {
		return nil, nil
	}

	var m clusterKubeOptionsModel
	diags := obj.As(ctx, &m, basetypes.ObjectAsOptions{})
	if diags.HasError() {
		return nil, diags
	}

	opts := &mksclient.KubernetesOptions{
		FeatureGates:         []string{},
		AdmissionControllers: []string{},
		X509CaCertificates:   m.X509CACertificates.ValueString(),
	}
	if !m.FeatureGates.IsNull() && !m.FeatureGates.IsUnknown() {
		diags.Append(m.FeatureGates.ElementsAs(ctx, &opts.FeatureGates, false)...)
	}
	if !m.AdmissionControllers.IsNull() && !m.AdmissionControllers.IsUnknown() {
		diags.Append(m.AdmissionControllers.ElementsAs(ctx, &opts.AdmissionControllers, false)...)
	}
	if !m.AuditLogs.IsNull() && !m.AuditLogs.IsUnknown() {
		var audit clusterAuditLogsModel
		diags.Append(m.AuditLogs.As(ctx, &audit, basetypes.ObjectAsOptions{})...)
		opts.AuditLogs = mksclient.AuditLogs{Enabled: audit.Enabled.ValueBool(), SecretName: audit.SecretName.ValueString()}
	}
	if !m.OIDC.IsNull() && !m.OIDC.IsUnknown() {
		var oidc clusterOIDCModel
		diags.Append(m.OIDC.As(ctx, &oidc, basetypes.ObjectAsOptions{})...)
		opts.Oidc = mksclient.OIDC{
			Enabled:       oidc.Enabled.ValueBool(),
			ProviderName:  oidc.ProviderName.ValueString(),
			IssuerUrl:     oidc.IssuerURL.ValueString(),
			ClientId:      oidc.ClientID.ValueString(),
			UsernameClaim: oidc.UsernameClaim.ValueString(),
			GroupsClaim:   oidc.GroupsClaim.ValueString(),
			CaCerts:       oidc.CACerts.ValueString(),
		}
		if !opts.Oidc.Enabled {
			// The API rejects any parameter of a disabled OIDC and clears
			// them itself.
			opts.Oidc = mksclient.OIDC{}
		}
	}

	return opts, diags
}

// fromAPI maps the cluster onto the model, taking cloud_subnet_cidr and
// x509_ca_certificates from prior, which the API never returns, and keeping
// kube_version via kubeVersionKeepsState after a refresh.
func (m *clusterModel) fromAPI(ctx context.Context, c *mksclient.ClusterDetailed, prior clusterModel, applied bool) diag.Diagnostics {
	var diags diag.Diagnostics

	m.ID = types.StringValue(c.Id)
	m.Name = types.StringValue(c.Name)
	if strings.EqualFold(prior.Name.ValueString(), c.Name) {
		// The API lower-cases the name, see nameCase.
		m.Name = prior.Name
	}
	m.Pool = types.StringValue(c.Pool)
	if c.ProjectId != "" && !sameProject(prior.ProjectID.ValueString(), c.ProjectId) {
		// The API returns the project dashed: the state takes the form the
		// provider stores, and keeps a configured one, see
		// keystoneProjectID.
		m.ProjectID = types.StringValue(keystoneProjectID(c.ProjectId))
	}
	if !applied {
		m.KubeVersion = types.StringValue(kubeVersionKeepsState(prior.KubeVersion.ValueString(), c.KubeVersion))
	}
	m.ClusterType = types.StringValue(string(c.ClusterType))
	m.WorkersType = types.StringValue(string(c.NetworkType))
	for workersType, networkType := range workersTypes {
		if networkType == string(c.NetworkType) {
			m.WorkersType = types.StringValue(workersType)
		}
	}
	m.NetworkID = types.StringValue(c.NetworkId)
	m.SubnetID = types.StringValue(c.SubnetId)
	m.CloudSubnetCIDR = prior.CloudSubnetCIDR
	m.PrivateKubeAPI = types.BoolValue(c.PrivateKubeApi)
	m.MaintenanceWindowStart = types.StringValue(c.MaintenanceWindowStart)
	m.MaintenanceWindowEnd = types.StringValue(c.MaintenanceWindowEnd)
	m.EnableAutorepair = types.BoolValue(c.EnableAutorepair)
	m.EnablePatchVersionAutoUpgrade = types.BoolValue(c.EnablePatchVersionAutoUpgrade)
	m.CNIType = types.StringValue(string(c.CniType))
	m.Status = types.StringValue(string(c.Status))
	m.KubeAPIIP = types.StringValue(c.KubeApiIp)

	m.CNICiliumSettings = types.ObjectNull(ciliumAttrTypes)
	if c.CniCiliumSettings != nil {
		var d diag.Diagnostics
		m.CNICiliumSettings, d = types.ObjectValueFrom(ctx, ciliumAttrTypes, clusterCiliumModel{
			EnvoyDaemonset: types.BoolPointerValue(c.CniCiliumSettings.EnvoyDaemonset),
			HubbleRelay:    types.BoolPointerValue(c.CniCiliumSettings.HubbleRelay),
		})
		diags.Append(d...)
	}

	x509 := types.StringNull()
	if !prior.KubernetesOptions.IsNull() && !prior.KubernetesOptions.IsUnknown() {
		var priorOptions clusterKubeOptionsModel
		diags.Append(prior.KubernetesOptions.As(ctx, &priorOptions, basetypes.ObjectAsOptions{})...)
		x509 = priorOptions.X509CACertificates
	}
	kubeOptions, d := flattenKubeOptions(ctx, c.KubernetesOptions, x509)
	diags.Append(d...)
	m.KubernetesOptions = kubeOptions

	return diags
}

func flattenKubeOptions(ctx context.Context, o mksclient.KubernetesOptions, x509 types.String) (types.Object, diag.Diagnostics) {
	var diags diag.Diagnostics

	stringSet := func(values []string) types.Set {
		// A framework set rejects duplicates.
		values = slices.Compact(slices.Sorted(slices.Values(values)))
		set, d := types.SetValueFrom(ctx, types.StringType, append([]string{}, values...))
		diags.Append(d...)

		return set
	}
	audit, d := types.ObjectValueFrom(ctx, auditLogsAttrTypes, clusterAuditLogsModel{
		Enabled:    types.BoolValue(o.AuditLogs.Enabled),
		SecretName: types.StringValue(o.AuditLogs.SecretName),
	})
	diags.Append(d...)
	oidcObject, d := types.ObjectValueFrom(ctx, oidcAttrTypes, clusterOIDCModel{
		Enabled:       types.BoolValue(o.Oidc.Enabled),
		ProviderName:  types.StringValue(o.Oidc.ProviderName),
		IssuerURL:     types.StringValue(o.Oidc.IssuerUrl),
		ClientID:      types.StringValue(o.Oidc.ClientId),
		UsernameClaim: types.StringValue(o.Oidc.UsernameClaim),
		GroupsClaim:   types.StringValue(o.Oidc.GroupsClaim),
		CACerts:       trimmedString(o.Oidc.CaCerts),
	})
	diags.Append(d...)

	obj, d := types.ObjectValueFrom(ctx, kubeOptionsAttrTypes, clusterKubeOptionsModel{
		FeatureGates:         stringSet(o.FeatureGates),
		AdmissionControllers: stringSet(o.AdmissionControllers),
		AuditLogs:            audit,
		OIDC:                 oidcObject,
		X509CACertificates:   x509,
	})
	diags.Append(d...)

	return obj, diags
}

// knownBoolPointer is nil for a null or unknown value: ValueBoolPointer turns
// unknown into false.
func knownBoolPointer(v types.Bool) *bool {
	if v.IsUnknown() {
		return nil
	}

	return v.ValueBoolPointer()
}

func knownStringPointer(v types.String) *string {
	if v.IsUnknown() {
		return nil
	}

	return v.ValueStringPointer()
}

// plannedState sets state to the plan with every unknown value null. It saves
// a created object that could not be read, so Terraform keeps it as tainted
// instead of losing it.
func plannedState(plan tfsdk.Plan, state *tfsdk.State) diag.Diagnostics {
	var diags diag.Diagnostics

	raw, err := tftypes.Transform(plan.Raw, func(_ *tftypes.AttributePath, v tftypes.Value) (tftypes.Value, error) {
		if !v.IsKnown() {
			return tftypes.NewValue(v.Type(), nil), nil
		}

		return v, nil
	})
	if err != nil {
		diags.AddError("Error saving the planned state", err.Error())

		return diags
	}
	state.Raw = raw

	return diags
}

// The claims the API sets for an enabled OIDC that omits them.
const (
	oidcUsernameClaim = "sub"
	oidcGroupsClaim   = "groups"
)

// oidcPlanModifier plans an OIDC parameter the way the API stores it: empty
// when OIDC is disabled, the API default for an omitted claim when enabled.
type oidcPlanModifier struct {
	enabledDefault string
	schemadoc.Facts
}

func (m oidcPlanModifier) Description(_ context.Context) string {
	return "Plans the value the API stores for the OIDC parameter."
}

func (m oidcPlanModifier) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}

func (m oidcPlanModifier) PlanModifyString(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	var oidc types.Object
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, req.Path.ParentPath(), &oidc)...)
	if resp.Diagnostics.HasError() || oidc.IsNull() || oidc.IsUnknown() {
		// An omitted oidc object keeps the state, see UseStateForUnknown.
		return
	}

	enabled, known, diags := oidcEnabled(ctx, req, oidc)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	if !known {
		if req.ConfigValue.IsNull() {
			resp.PlanValue = types.StringUnknown()
		}

		return
	}

	if req.ConfigValue.IsUnknown() {
		// The plan must keep an unknown configured value; the apply-time plan
		// sees the known one.
		return
	}

	if enabled {
		if m.enabledDefault == "" {
			return
		}
		if req.ConfigValue.IsNull() {
			resp.PlanValue = types.StringValue(m.enabledDefault)

			return
		}
		if req.ConfigValue.ValueString() == "" {
			// The API would store the default instead, and the plan cannot
			// differ from the configured "".
			name, _ := req.Path.Steps().LastStep()
			resp.Diagnostics.AddAttributeError(req.Path, "Empty OIDC claim while OIDC is enabled",
				fmt.Sprintf("%s cannot be empty while OIDC is enabled: omit it to get `%s`.", name, m.enabledDefault))
		}

		return
	}

	if !req.ConfigValue.IsNull() && req.ConfigValue.ValueString() != "" {
		resp.Diagnostics.AddAttributeError(req.Path, "OIDC parameter set while OIDC is disabled",
			fmt.Sprintf("The API rejects it: oidc parameters cannot be configured when it is disabled. "+
				"Disabling OIDC clears its settings, so remove %s.", req.Path))

		return
	}
	resp.PlanValue = types.StringValue("")
}

// oidcEnabled is the planned oidc.enabled: the configured value, otherwise
// the state one, otherwise false, which create sends.
func oidcEnabled(ctx context.Context, req planmodifier.StringRequest, oidc types.Object) (bool, bool, diag.Diagnostics) {
	enabled, _ := oidc.Attributes()["enabled"].(types.Bool)
	if enabled.IsUnknown() {
		return false, false, nil
	}
	if !enabled.IsNull() {
		return enabled.ValueBool(), true, nil
	}
	if req.State.Raw.IsNull() {
		return false, true, nil
	}

	var stateEnabled types.Bool
	diags := req.State.GetAttribute(ctx, req.Path.ParentPath().AtName("enabled"), &stateEnabled)
	if stateEnabled.IsNull() || stateEnabled.IsUnknown() {
		return false, false, diags
	}

	return stateEnabled.ValueBool(), true, diags
}

// nameCase plans the state name for a configured name that differs only in
// letter case: the API lower-cases the name, and v1 ignored the case too.
type nameCase struct {
	schemadoc.Facts
}

func (m nameCase) Description(_ context.Context) string {
	return "A change of only the letter case keeps the state value."
}

func (m nameCase) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}

func (m nameCase) PlanModifyString(_ context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	if req.StateValue.IsNull() || req.PlanValue.IsUnknown() || req.PlanValue.IsNull() {
		return
	}
	if strings.EqualFold(req.PlanValue.ValueString(), req.StateValue.ValueString()) {
		resp.PlanValue = req.StateValue
	}
}

// trimmedStringKeep plans the state value when the configured one differs
// from it only in surrounding whitespace: semantic equality covers apply and
// read results, but not the plan, e.g. after an import.
type trimmedStringKeep struct {
	schemadoc.Facts
}

func (m trimmedStringKeep) Description(_ context.Context) string {
	return "A change of only the surrounding whitespace keeps the state value."
}

func (m trimmedStringKeep) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}

func (m trimmedStringKeep) PlanModifyString(_ context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	if req.StateValue.IsNull() || req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}
	if strings.TrimSpace(req.ConfigValue.ValueString()) == strings.TrimSpace(req.StateValue.ValueString()) {
		resp.PlanValue = req.StateValue
	}
}

// trimmedStringType is a string whose surrounding whitespace does not
// matter: the API trims oidc.ca_certs, while file("ca.pem") keeps the
// trailing newline.
type trimmedStringType struct {
	basetypes.StringType
}

var _ basetypes.StringTypable = trimmedStringType{}

func (t trimmedStringType) Equal(o attr.Type) bool {
	_, ok := o.(trimmedStringType)

	return ok
}

func (t trimmedStringType) String() string {
	return "trimmedStringType"
}

func (t trimmedStringType) ValueFromString(_ context.Context, in basetypes.StringValue) (basetypes.StringValuable, diag.Diagnostics) {
	return trimmedStringValue{StringValue: in}, nil
}

func (t trimmedStringType) ValueFromTerraform(ctx context.Context, in tftypes.Value) (attr.Value, error) {
	value, err := t.StringType.ValueFromTerraform(ctx, in)
	if err != nil {
		return nil, err
	}
	stringValue, ok := value.(basetypes.StringValue)
	if !ok {
		return nil, fmt.Errorf("unexpected value type %T", value)
	}

	return trimmedStringValue{StringValue: stringValue}, nil
}

func (t trimmedStringType) ValueType(_ context.Context) attr.Value {
	return trimmedStringValue{}
}

type trimmedStringValue struct {
	basetypes.StringValue
}

var _ basetypes.StringValuableWithSemanticEquals = trimmedStringValue{}

func (v trimmedStringValue) Equal(o attr.Value) bool {
	other, ok := o.(trimmedStringValue)

	return ok && v.StringValue.Equal(other.StringValue)
}

func (v trimmedStringValue) Type(_ context.Context) attr.Type {
	return trimmedStringType{}
}

func (v trimmedStringValue) StringSemanticEquals(_ context.Context, newValuable basetypes.StringValuable) (bool, diag.Diagnostics) {
	var diags diag.Diagnostics

	newValue, ok := newValuable.(trimmedStringValue)
	if !ok {
		diags.AddError("Semantic Equality Check Error", fmt.Sprintf("Expected trimmedStringValue, got %T.", newValuable))

		return false, diags
	}

	return strings.TrimSpace(v.ValueString()) == strings.TrimSpace(newValue.ValueString()), diags
}

func trimmedString(value string) trimmedStringValue {
	return trimmedStringValue{StringValue: types.StringValue(value)}
}
