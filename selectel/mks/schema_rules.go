package mks

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/defaults"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setdefault"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/terraform-providers/terraform-provider-selectel/selectel/internal/schemadoc"
)

// kubeVersionPattern is the x.y.z form of a Kubernetes version.
var kubeVersionPattern = regexp.MustCompile(`^\d+\.\d+\.\d+$`)

var maintenanceWindowPattern = regexp.MustCompile(`^([01]\d|2[0-3]):[0-5]\d:[0-5]\d$`)

var namePattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9-]*$`)

var maintenanceWindowLink = schemadoc.Link{
	Title: "Maintenance window",
	URL:   "https://docs.selectel.ru/en/managed-kubernetes/clusters/set-up-maintenance-window/",
}

// keepState plans the state value when the planned one means the same.
type keepState struct {
	description string
	same        func(a, b string) bool
	schemadoc.Facts
}

var (
	// nameCase exists because the API lower-cases the name, and v1 ignored the case too.
	nameCase = keepState{
		description: "A change of only the letter case keeps the state value.",
		same:        strings.EqualFold,
	}
	// sameProjectID exists because a project ID in another form names the same project, see keystoneProjectID.
	sameProjectID = keepState{
		description: "A project ID that differs only in dashes or letter case keeps the state value.",
		same:        sameProject,
	}
	trimmedStringKeep = keepState{
		description: "A change of only the surrounding whitespace keeps the state value.",
		same:        func(a, b string) bool { return strings.TrimSpace(a) == strings.TrimSpace(b) },
	}
)

func (m keepState) Description(_ context.Context) string {
	return m.description
}

func (m keepState) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}

func (m keepState) PlanModifyString(_ context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	if req.StateValue.IsNull() || req.PlanValue.IsUnknown() || req.PlanValue.IsNull() {
		return
	}
	if m.same(req.PlanValue.ValueString(), req.StateValue.ValueString()) {
		resp.PlanValue = req.StateValue
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
