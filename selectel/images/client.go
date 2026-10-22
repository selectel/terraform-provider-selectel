// Package images implements the image Terraform Framework resources and data sources.
package images

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	sdk "github.com/selectel/images-go/pkg/v1"
	"github.com/selectel/images-go/pkg/v1/image"
)

// ProviderConfig is the authentication boundary shared by resources and data sources.
// API operations use the concrete service SDK so tests exercise the actual wire contract.
type ProviderConfig interface {
	ImagesClient(ctx context.Context, projectID, region string) (*sdk.Client, error)
	ImagesDefaults() (projectID, region string)
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
			"Invalid provider configuration", "Expected image provider configuration.",
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
		return nil, fmt.Errorf("image provider is not configured")
	}
	client, err := c.config.ImagesClient(ctx, projectID, region)
	if err != nil {
		return nil, fmt.Errorf("connect to the Image service (project_id=%s, region=%s): %w", projectID, region, err)
	}

	return client, nil
}

// importScope is the project and pool of an imported object. They come from
// the provider configuration, except a project that the import ID names.
func (c connection) importScope(projectFromID string) (string, string, diag.Diagnostics) {
	var diags diag.Diagnostics
	if c.config == nil {
		diags.AddError("Missing provider configuration", "Configure the provider before importing.")
		return "", "", diags
	}
	project, region := c.config.ImagesDefaults()
	if projectFromID != "" {
		project = projectFromID
	}
	if project == "" || region == "" {
		diags.AddError(
			"Missing import scope",
			"Set provider project_id and region, or INFRA_PROJECT_ID and INFRA_REGION.",
		)
	}

	return project, region, diags
}

// requestTimeout bounds the operations that have no configurable timeout.
const requestTimeout = 10 * time.Minute

// imageGone reports whether a read of the image shows that it no longer
// exists. The Image service keeps a deleted image readable for some time with
// a deleted status, and then returns 404.
func imageGone(current *image.View, err error) bool {
	if err != nil {
		return sdk.IsKind(err, sdk.KindNotFound)
	}

	return current.Status == image.StatusDeleted || current.Status == image.StatusPendingDelete
}

// isTransient reports whether a failed read may succeed when repeated. Such an
// error does not end a wait: the image keeps changing on the service side.
func isTransient(err error) bool {
	for _, kind := range []sdk.Kind{sdk.KindRateLimited, sdk.KindServerError, sdk.KindTransport, sdk.KindTimeout} {
		if sdk.IsKind(err, kind) {
			return true
		}
	}

	return false
}

// createOutcomeUncertain reports whether a failed create request may still
// have made the object: the API may have accepted it, but no definite answer
// came back. Unlike isTransient, a rate limit is excluded: the API rejected
// the request. A canceled request is excluded too: any read that resolves the
// outcome uses the same context and fails as well.
func createOutcomeUncertain(err error) bool {
	for _, kind := range []sdk.Kind{sdk.KindServerError, sdk.KindTransport, sdk.KindTimeout, sdk.KindUnexpected} {
		if sdk.IsKind(err, kind) {
			return true
		}
	}

	return false
}

// waitImageActive waits until the image data is in place. With checkImport,
// every poll also reads the import tasks, so a failed web-download ends the
// wait with its reason instead of a timeout.
func waitImageActive(
	ctx context.Context,
	client *sdk.Client,
	id string,
	interval time.Duration,
	checkImport bool,
) (*image.View, error) {
	// Every caller sets a deadline on ctx, and the wait ends at it.
	deadline, _ := ctx.Deadline()
	// The last observed image answers a poll that failed with a transient
	// error: a nil result would count as a missing image.
	last := &image.View{ID: id, Status: image.StatusQueued}
	wait := retry.StateChangeConf{
		Pending:      []string{image.StatusQueued, image.StatusSaving, image.StatusImporting},
		Target:       []string{image.StatusActive},
		Timeout:      time.Until(deadline),
		PollInterval: interval,
		Refresh: func() (any, string, error) {
			current, _, err := image.Get(ctx, client, id)
			if isTransient(err) {
				return last, last.Status, nil
			}
			if err != nil {
				return nil, "", fmt.Errorf("observe image %s: %w", id, err)
			}
			last = current
			if checkImport && current.Status != image.StatusActive {
				if reason := failedImportReason(ctx, client, id); reason != "" {
					return nil, "", fmt.Errorf("the API could not download the image data: %s", reason)
				}
			}

			return current, current.Status, nil
		},
	}
	current, err := wait.WaitForStateContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("wait for image %s: %w", id, err)
	}

	return current.(*image.View), nil
}

// timedOut reports whether a wait ended at the operation deadline.
func timedOut(err error) bool {
	var timeout *retry.TimeoutError

	return errors.As(err, &timeout) || errors.Is(err, context.DeadlineExceeded)
}

// failedImportReason is the reason key of the newest import task, such as
// url_not_accessible, when that task failed, and empty otherwise. Reading the
// tasks is best effort: without them the wait relies on the image status.
func failedImportReason(ctx context.Context, client *sdk.Client, id string) string {
	tasks, _, err := image.Tasks(ctx, client, id)
	if err != nil {
		return ""
	}
	var newest *image.Task
	for i := range tasks {
		task := &tasks[i]
		if task.Type == image.TaskTypeImport && (newest == nil || task.CreatedAt.After(newest.CreatedAt.Time)) {
			newest = task
		}
	}
	if newest == nil || newest.Status != image.TaskStatusFailure {
		return ""
	}
	if newest.Message == "" {
		return "unknown"
	}

	return newest.Message
}

// waitImageDeleted waits until the API reports the image as deleted or no
// longer returns it.
func waitImageDeleted(ctx context.Context, client *sdk.Client, id string, interval time.Duration) error {
	const gone = "gone"
	// Every caller sets a deadline on ctx, and the wait ends at it.
	deadline, _ := ctx.Deadline()
	last := image.StatusActive
	wait := retry.StateChangeConf{
		Pending: []string{
			image.StatusQueued, image.StatusSaving, image.StatusImporting,
			image.StatusActive, image.StatusKilled, image.StatusDeactivated,
		},
		Target:       []string{gone},
		Timeout:      time.Until(deadline),
		PollInterval: interval,
		Refresh: func() (any, string, error) {
			current, _, err := image.Get(ctx, client, id)
			switch {
			case imageGone(current, err):
				return id, gone, nil
			case isTransient(err):
				return id, last, nil
			case err != nil:
				return nil, "", fmt.Errorf("observe image %s: %w", id, err)
			}
			last = current.Status

			return id, last, nil
		},
	}
	if _, err := wait.WaitForStateContext(ctx); err != nil {
		return fmt.Errorf("wait for deleted image %s: %w", id, err)
	}

	return nil
}
