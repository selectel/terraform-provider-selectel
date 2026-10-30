// Package filestorage implements the file storage Terraform Framework resources and data sources.
package filestorage

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	sdk "github.com/selectel/filestorage-go/pkg/v1"
)

// ProviderConfig is the authentication boundary shared by resources and data sources.
// API operations use the concrete service SDK so tests exercise the actual wire contract.
type ProviderConfig interface {
	FileStorageClient(ctx context.Context, projectID, region string) (*sdk.Client, error)
	FileStorageDefaults() (projectID, region string)
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
			"Invalid provider configuration", "Expected file storage provider configuration.",
		)}
	}
	c.config = config

	return nil
}

func (c connection) client(ctx context.Context, projectID, region string) (*sdk.Client, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if c.config == nil {
		return nil, fmt.Errorf("file storage provider is not configured")
	}
	client, err := c.config.FileStorageClient(ctx, projectID, region)
	if err != nil {
		return nil, fmt.Errorf("connect to File Storage (project_id=%s, region=%s): %w", projectID, region, err)
	}

	return client, nil
}

// importScope is the project and pool of an imported object, taken from the
// provider configuration.
func (c connection) importScope() (string, string, diag.Diagnostics) {
	var diags diag.Diagnostics
	if c.config == nil {
		diags.AddError("Missing provider configuration", "Configure the provider before importing.")
		return "", "", diags
	}
	project, region := c.config.FileStorageDefaults()
	if project == "" {
		diags.AddError("Missing import scope", "Set provider project_id or INFRA_PROJECT_ID.")
	}
	if region == "" {
		diags.AddError("Missing import scope", "Set provider region or INFRA_REGION.")
	}

	return project, region, diags
}

// isTransient reports whether a failed read may succeed when repeated. Such an
// error does not end a wait: the share keeps changing on the service side.
func isTransient(err error) bool {
	for _, kind := range []sdk.Kind{sdk.KindRateLimited, sdk.KindServerError, sdk.KindTransport, sdk.KindTimeout} {
		if sdk.IsKind(err, kind) {
			return true
		}
	}

	return false
}

// apiError is the diagnostic of a failed API operation. A pool whose API
// does not serve the SDK microversion rejects every request the same way, so
// that error names the pool and the version instead of the raw API answer.
func apiError(summary, region string, err error) diag.Diagnostic {
	if sdk.IsKind(err, sdk.KindMicroversion) {
		return diag.NewErrorDiagnostic(summary, fmt.Sprintf(
			"File storage in pool %s does not support API microversion %s, which the provider requires: %s",
			region, sdk.Microversion, err,
		))
	}

	return diag.NewErrorDiagnostic(summary, err.Error())
}

// readTimeout bounds a refresh, which has no configurable timeout.
const readTimeout = 10 * time.Minute
