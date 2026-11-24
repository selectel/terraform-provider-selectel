// Package mks implements the Managed Kubernetes API v2 Terraform Framework resources.
package mks

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	mksv2 "github.com/selectel/mks-go/v2/pkg"
	"github.com/terraform-providers/terraform-provider-selectel/selectel/internal/schemadoc"
)

// ProviderConfig is the authentication boundary shared by the resources.
// API operations use the concrete service SDK so tests exercise the actual wire contract.
type ProviderConfig interface {
	MKSV2Client(ctx context.Context, projectID, pool string) (*mksv2.ServiceClient, error)
	MKSV2Defaults() (projectID, pool string)
}

type connection struct{ config ProviderConfig }

// configure stores the configuration that the provider passes to Configure.
// Data is nil while Terraform validates the configuration before Configure.
func (c *connection) configure(data any) diag.Diagnostics {
	if data == nil {
		return nil
	}
	config, ok := data.(ProviderConfig)
	if !ok {
		return diag.Diagnostics{diag.NewErrorDiagnostic(
			"Invalid provider configuration", "Expected Managed Kubernetes provider configuration.",
		)}
	}
	c.config = config

	return nil
}

// importScope is the project and pool of an imported object, taken from the
// provider configuration, because the cluster ID carries neither.
func (c connection) importScope() (string, string, diag.Diagnostics) {
	var diags diag.Diagnostics
	var projectID, pool string
	if c.config != nil {
		projectID, pool = c.config.MKSV2Defaults()
	}
	if projectID == "" {
		diags.AddError("Missing project ID", "INFRA_PROJECT_ID must be set for the resource import")
	}
	if pool == "" {
		diags.AddError("Missing pool", "INFRA_REGION must be set for the resource import")
	}

	return keystoneProjectID(projectID), pool, diags
}

// projectID returns the project_id attribute as set, or the provider one in
// the form keystoneProjectID keeps, when the attribute is not set.
func (c connection) projectID(attr types.String) (string, diag.Diagnostics) {
	if !attr.IsNull() && !attr.IsUnknown() && attr.ValueString() != "" {
		return attr.ValueString(), nil
	}
	if c.config != nil {
		projectID, _ := c.config.MKSV2Defaults()
		if projectID != "" {
			return keystoneProjectID(projectID), nil
		}
	}

	return "", diag.Diagnostics{diag.NewAttributeErrorDiagnostic(path.Root("project_id"), "Missing project ID",
		"Set project_id in the resource or data source, or in the provider configuration.")}
}

// client returns the MKS API v2 client for the pool and for the project of the
// project_id attribute, and that project ID as projectID resolves it.
func (c connection) client(ctx context.Context, attr types.String, pool string) (*mksv2.ServiceClient, string, diag.Diagnostics) {
	projectID, diags := c.projectID(attr)
	if diags.HasError() {
		return nil, "", diags
	}
	err := ctx.Err()
	if err != nil {
		diags.AddError("Error initializing MKS client", err.Error())

		return nil, "", diags
	}
	if c.config == nil {
		diags.AddError("Error initializing MKS client", "managed kubernetes provider is not configured")

		return nil, "", diags
	}
	client, err := c.config.MKSV2Client(ctx, keystoneProjectID(projectID), pool)
	if err != nil {
		diags.AddError("Error initializing MKS client",
			fmt.Sprintf("connect to Managed Kubernetes (project_id=%s, pool=%s): %s", projectID, pool, err))

		return nil, "", diags
	}

	return client, projectID, diags
}

// keystoneProjectID returns the project ID as lower-case hex without dashes, the form Keystone scopes a token by.
// The API returns a dashed UUID.
func keystoneProjectID(projectID string) string {
	return strings.ToLower(strings.ReplaceAll(projectID, "-", ""))
}

// sameProject reports whether two project IDs name the same project.
func sameProject(a, b string) bool {
	return keystoneProjectID(a) == keystoneProjectID(b)
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
