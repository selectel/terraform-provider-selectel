// Package mks implements the Managed Kubernetes API v2 Terraform Framework resources.
package mks

import (
	"context"
	"fmt"
	"strings"

	"github.com/terraform-providers/terraform-provider-selectel/selectel/internal/schemadoc"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	mksv2 "github.com/selectel/mks-go/v2/pkg"
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
	if c.config == nil {
		diags.AddError("Missing project ID", "INFRA_PROJECT_ID must be set for the resource import")
		diags.AddError("Missing pool", "INFRA_REGION must be set for the resource import")

		return "", "", diags
	}
	projectID, pool := c.config.MKSV2Defaults()
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

// client builds the MKS API v2 client for the Keystone form of projectID.
func (c connection) client(ctx context.Context, projectID, pool string) (*mksv2.ServiceClient, error) {
	err := ctx.Err()
	if err != nil {
		return nil, err
	}
	if c.config == nil {
		return nil, fmt.Errorf("managed kubernetes provider is not configured")
	}
	client, err := c.config.MKSV2Client(ctx, keystoneProjectID(projectID), pool)
	if err != nil {
		return nil, fmt.Errorf("connect to Managed Kubernetes (project_id=%s, pool=%s): %w", projectID, pool, err)
	}

	return client, nil
}

// keystoneProjectID returns the project ID in the form Keystone scopes a
// token by: lower-case hex without dashes. The API v2 returns project_id as
// a dashed UUID, and Keystone answers 401 for that form.
func keystoneProjectID(projectID string) string {
	return strings.ToLower(strings.ReplaceAll(projectID, "-", ""))
}

// sameProject reports whether two project IDs name the same project.
func sameProject(a, b string) bool {
	return keystoneProjectID(a) == keystoneProjectID(b)
}

// sameProjectID plans the state project_id for a configured one that names
// the same project in another form, see keystoneProjectID.
type sameProjectID struct {
	schemadoc.Facts
}

func (m sameProjectID) Description(_ context.Context) string {
	return "A project ID that differs only in dashes or letter case keeps the state value."
}

func (m sameProjectID) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}

func (m sameProjectID) PlanModifyString(_ context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	if req.StateValue.IsNull() || req.PlanValue.IsUnknown() || req.PlanValue.IsNull() {
		return
	}
	if sameProject(req.PlanValue.ValueString(), req.StateValue.ValueString()) {
		resp.PlanValue = req.StateValue
	}
}
