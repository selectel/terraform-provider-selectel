package mks

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/defaults"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	mksclient "github.com/selectel/mks-go/v2/pkg/mksclient"
	"github.com/terraform-providers/terraform-provider-selectel/selectel/internal/schemadoc"
)

var clusterDocs = schemadoc.Resource{Name: "cluster", Location: "managed-kubernetes"}

// kubeVersionPattern pins the form the API accepts on create: it takes a
// plain minor version too, and every later plan then fails.
var kubeVersionPattern = regexp.MustCompile(`^\d+\.\d+\.\d+$`)

var maintenanceWindowPattern = regexp.MustCompile(`^([01]\d|2[0-3]):[0-5]\d:[0-5]\d$`)

var namePattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9-]*$`)

// The claims the API sets for an enabled OIDC that omits them.
const (
	oidcUsernameClaim = "sub"
	oidcGroupsClaim   = "groups"
)

// workersTypes maps each user-facing workers_type to the API network_type.
var workersTypes = map[string]string{
	"CLOUD":     string(mksclient.ClusterNetworkTypeSTANDARD),
	"DEDICATED": string(mksclient.ClusterNetworkTypeL3VPN),
}

var (
	ciliumAttrTypes = map[string]attr.Type{
		"envoy_daemonset": types.BoolType,
		"hubble_relay":    types.BoolType,
	}
	auditLogsAttrTypes = map[string]attr.Type{
		"enabled": types.BoolType,
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
	}
)

func (r *clusterResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
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
			"When a `kube_version`, `kubernetes_options`, or `cni_cilium_settings` task fails, the API has already recorded the new values, " +
			"so the next plan shows no change. Check `status`.",
		Attributes: map[string]schema.Attribute{
			"id": clusterDocs.ID(),
			"name": clusterDocs.String(schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					schemadoc.LengthAtMost(32),
					schemadoc.DescribeString(stringvalidator.RegexMatches(namePattern, "must contain Latin letters, digits, and hyphens and not start with a hyphen"),
						"The name can contain Latin letters, digits, and hyphens and cannot start with a hyphen. "+
							"If `workers_type` is `DEDICATED`, the name must start with a letter."),
				},
				PlanModifiers: []planmodifier.String{nameCase, schemadoc.RequiresReplace()},
			}, schemadoc.Field{
				Definition: "Cluster name. It is included in the names of the cluster entities: " +
					"node groups, nodes, load balancers, networks, and volumes. " +
					"The API stores the name in lower case, so a change of only the letter case is not a change.",
			}),
			"pool": clusterDocs.Region(),
			"project_id": clusterDocs.String(schema.StringAttribute{
				Optional:   true,
				Computed:   true,
				Validators: []validator.String{schemadoc.LengthAtLeast(1)},
				PlanModifiers: []planmodifier.String{
					sameProjectID,
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
				Validators: []validator.String{
					schemadoc.DescribeString(stringvalidator.RegexMatches(kubeVersionPattern, "must be in the x.y.z form"),
						"The format is `x.y.z`."),
				},
			}, schemadoc.Field{
				Definition: "Kubernetes version of the cluster. The format is `x.y.z`. Raising it upgrades the cluster. " +
					"A patch upgrade is possible only to the latest patch version of the current minor version. " +
					"Another patch version fails the apply. " +
					"A minor upgrade is possible only to the next minor version, and the API picks its patch version. " +
					"Skipping a minor version fails the apply. " +
					"The cluster may run a newer version than configured, for example after patch auto-upgrade: " +
					"a configured version lower than the one in the state is accepted with no change and never downgrades the cluster.",
				Source: "You can retrieve the available versions with the [selectel_mks_kube_versions_v1](https://registry.terraform.io/providers/selectel/selectel/latest/docs/data-sources/mks_kube_versions_v1) data source.",
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
				Validators:    []validator.String{schemadoc.LengthAtLeast(1)},
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown(), schemadoc.RequiresReplace()},
			}, schemadoc.Field{
				Definition: "Unique identifier of the network of the cluster.",
				Behavior:   "If omitted, the API creates a network.",
				Source:     schemadoc.FromResource("selectel_vpc_network_v2", "network ID"),
			}),
			"subnet_id": clusterDocs.String(schema.StringAttribute{
				Optional:      true,
				Computed:      true,
				Validators:    []validator.String{schemadoc.LengthAtLeast(1)},
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown(), schemadoc.RequiresReplace()},
			}, schemadoc.Field{
				Definition: "Unique identifier of the subnet of the cluster.",
				Behavior:   "If omitted, the API creates a subnet.",
				Source:     schemadoc.FromResource("selectel_vpc_network_subnet_v2", "subnet ID"),
			}),
			"cloud_subnet_cidr": clusterDocs.String(schema.StringAttribute{
				Optional:      true,
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown(), schemadoc.RequiresReplace()},
			}, schemadoc.Field{
				Definition: "CIDR of the cloud subnet to which the master nodes are connected. " +
					"Applicable only if `workers_type` is `DEDICATED`. If omitted, the API uses `10.20.0.0/24`.",
			}),
			"private_kube_api": clusterDocs.Bool(schema.BoolAttribute{
				Optional:      true,
				Computed:      true,
				Default:       schemadoc.StaticBool(false),
				PlanModifiers: []planmodifier.Bool{schemadoc.BoolRequiresReplace()},
			}, schemadoc.Field{
				Definition: "Whether the Kube API is available only from the cluster network instead of the Internet.",
			}),
			"maintenance_window_start": clusterDocs.String(schema.StringAttribute{
				Optional: true,
				Computed: true,
				Validators: []validator.String{
					schemadoc.DescribeString(stringvalidator.RegexMatches(maintenanceWindowPattern, "must be in the hh:mm:ss form"),
						"The format is `hh:mm:ss`."),
				},
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			}, schemadoc.Field{
				Definition: "Time in UTC when maintenance in the cluster starts.",
				Behavior:   "If omitted when the cluster is created, the API uses `01:00:00`.",
			}),
			"enable_autorepair": clusterDocs.Bool(optionalBool(), schemadoc.Field{
				Definition: "Enables or disables automatic reinstallation of worker nodes that are unavailable or unhealthy. " +
					"Auto-recovery is unavailable with a single worker node and on dedicated servers, " +
					"and it deletes all data on the boot disk of the node.",
				Behavior: "If omitted, the API enables it, except for `DEDICATED` clusters, where it is disabled.",
				Concept: schemadoc.Link{
					Title: "node auto-recovery",
					URL:   "https://docs.selectel.ru/en/cloud/managed-kubernetes/node-groups/reinstall-nodes/",
				},
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
				Optional: true,
				Computed: true,
				MarkdownDescription: "Settings of the Cilium CNI. If you set this argument, set `cni_type` to `CILIUM`. " +
					"If omitted when `cni_type` is `CILIUM`, the API uses `true` for `envoy_daemonset` and `false` for `hubble_relay`.",
				PlanModifiers: []planmodifier.Object{objectplanmodifier.UseStateForUnknown()},
				Attributes: map[string]schema.Attribute{
					"envoy_daemonset": clusterDocs.Bool(optionalBool(), schemadoc.Field{
						Definition: "Enables or disables the Envoy DaemonSet for Cilium. If omitted, the API uses `true`.",
					}),
					"hubble_relay": clusterDocs.Bool(schema.BoolAttribute{
						Optional: true,
						Computed: true,
						Default:  schemadoc.StaticBool(true),
					}, schemadoc.Field{
						Definition: "Enables or disables Hubble Relay for Cilium.",
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
						Optional:    true,
						Computed:    true,
						ElementType: types.StringType,
						Default:     emptySet(),
					}, schemadoc.Field{
						Definition: "Enabled feature gates.",
						Concept: schemadoc.Link{
							Title: "Feature gates",
							URL:   "https://docs.selectel.ru/en/cloud/managed-kubernetes/clusters/feature-gates/",
						},
					}),
					"admission_controllers": clusterDocs.Set(schema.SetAttribute{
						Optional:    true,
						Computed:    true,
						ElementType: types.StringType,
						Default:     emptySet(),
					}, schemadoc.Field{
						Definition: "Enabled admission controllers.",
						Concept: schemadoc.Link{
							Title: "Admission controllers",
							URL:   "https://docs.selectel.ru/en/cloud/managed-kubernetes/clusters/admission-controllers/",
						},
					}),
					"audit_logs": schema.SingleNestedAttribute{
						Optional: true,
						Computed: true,
						MarkdownDescription: "Collection of Kubernetes audit logs. " +
							"Learn more about [exporting audit logs to a logging system](https://docs.selectel.ru/en/cloud/managed-kubernetes/clusters/logs/#configure-export-of-audit-logs).",
						PlanModifiers: []planmodifier.Object{objectplanmodifier.UseStateForUnknown()},
						Attributes: map[string]schema.Attribute{
							"enabled": clusterDocs.Bool(optionalBool(), schemadoc.Field{
								Definition: "Enables or disables collection of audit logs.",
								Behavior:   "If omitted when the cluster is created, the value is `false`.",
							}),
						},
					},
					"oidc": schema.SingleNestedAttribute{
						Optional: true,
						Computed: true,
						MarkdownDescription: "Connection of an OpenID Connect (OIDC) provider to the cluster. " +
							"Disabling OIDC clears its settings, so with `enabled = false` omit the other arguments. A non-empty one fails the plan. " +
							"Learn more about [connecting an OIDC provider to the cluster](https://docs.selectel.ru/en/cloud/managed-kubernetes/clusters/access-to-cluster-with-oidc-provider/#configure-oidc-connection).",
						PlanModifiers: []planmodifier.Object{objectplanmodifier.UseStateForUnknown()},
						Attributes: map[string]schema.Attribute{
							"enabled": clusterDocs.Bool(optionalBool(), schemadoc.Field{
								Definition: "Enables or disables authentication with OIDC.",
								Behavior:   "If `false`, omit the other arguments of `oidc`. If omitted when the cluster is created, the value is `false`.",
							}),
							"provider_name": clusterDocs.String(oidcParam(""), schemadoc.Field{
								Definition: "Name of the connection, for identification only.",
								Behavior:   "If `enabled` is `false`, omit this argument.",
							}),
							"issuer_url": clusterDocs.String(oidcParam(""), schemadoc.Field{
								Definition: "URL of the OIDC provider. It must start with `https://`.",
								Behavior:   "If `enabled` is `false`, omit this argument.",
							}),
							"client_id": clusterDocs.String(oidcParam(""), schemadoc.Field{
								Definition: "Client ID that all tokens must be issued for.",
								Behavior:   "If `enabled` is `false`, omit this argument.",
							}),
							"username_claim": clusterDocs.String(oidcParam(oidcUsernameClaim), schemadoc.Field{
								Definition: "JWT claim to use as the username. " +
									"When OIDC is enabled, omit it to get `sub`. An empty value fails the plan.",
								Behavior: "If `enabled` is `false`, omit this argument.",
							}),
							"groups_claim": clusterDocs.String(oidcParam(oidcGroupsClaim), schemadoc.Field{
								Definition: "JWT claim to use as the user's group. " +
									"When OIDC is enabled, omit it to get `groups`. An empty value fails the plan.",
								Behavior: "If `enabled` is `false`, omit this argument.",
							}),
							"ca_certs": clusterDocs.String(schema.StringAttribute{
								Optional:      true,
								Computed:      true,
								CustomType:    trimmedStringType{},
								PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown(), trimmedStringKeep, oidcPlanModifier{}},
							}, schemadoc.Field{
								Definition: "CA certificates of the OIDC provider. The format is PEM. " +
									"Leading and trailing whitespace is ignored.",
							}),
							// ca_certs carries no enabled=false sentence: any PEM content is a value the plan rejects.
						},
					},
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
			"timeouts": clusterDocs.Timeouts(ctx, clusterTimeouts),
		},
	}
}

type emptySetDefault struct {
	defaults.Set
	schemadoc.Facts
}

// emptySet plans an omitted set as the empty one, which clears it.
func emptySet() defaults.Set {
	return emptySetDefault{
		setdefault.StaticValue(types.SetValueMust(types.StringType, nil)),
		schemadoc.Facts{Default: "The default value is an empty set."},
	}
}

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
