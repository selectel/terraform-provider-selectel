// Package blockstorage implements the Block Storage Terraform Framework resources.
package blockstorage

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	sdk "github.com/selectel/blockstorage-go/pkg/v1"
	"github.com/selectel/blockstorage-go/pkg/v1/volume"
)

// ProviderConfig is the authentication boundary shared by resources and data sources.
// API operations use the concrete service SDK so tests exercise the actual wire contract.
type ProviderConfig interface {
	BlockStorageClient(ctx context.Context, projectID, region string) (*sdk.Client, error)
	BlockStorageDefaults() (projectID, region string)
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
			"Invalid provider configuration", "Expected Block Storage provider configuration.",
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
		return nil, fmt.Errorf("block storage provider is not configured")
	}
	client, err := c.config.BlockStorageClient(ctx, projectID, region)
	if err != nil {
		return nil, fmt.Errorf("connect to Block Storage (project_id=%s, region=%s): %w", projectID, region, err)
	}

	return client, nil
}

// waitVolumeReady waits until the volume is available or attached.
func waitVolumeReady(ctx context.Context, client *sdk.Client, id string, interval time.Duration) (*volume.View, error) {
	ready, err := waitVolume(ctx, client, id, interval,
		[]string{"creating", "downloading", "restoring-backup", "extending"}, []string{"available", "in-use"})
	if err != nil {
		return nil, err
	}

	return ready.(*volume.View), nil
}

// waitVolumeDeleted waits until the volume is gone and a complete list confirms it.
func waitVolumeDeleted(ctx context.Context, client *sdk.Client, id string, interval time.Duration) error {
	if _, err := waitVolume(ctx, client, id, interval, []string{"available", "deleting"}, nil); err != nil {
		return err
	}

	return confirmVolumeAbsent(ctx, client, id)
}

// waitVolume polls the volume status until the operation deadline, which every
// caller sets. An empty target waits for the volume to disappear.
func waitVolume(
	ctx context.Context,
	client *sdk.Client,
	id string,
	interval time.Duration,
	pending, target []string,
) (any, error) {
	deadline, _ := ctx.Deadline()
	wait := retry.StateChangeConf{
		Pending:      pending,
		Target:       target,
		Timeout:      time.Until(deadline),
		PollInterval: interval,
		Refresh: func() (any, string, error) {
			current, _, err := volume.Get(ctx, client, id)
			if sdk.IsKind(err, sdk.KindNotFound) {
				return nil, "", nil
			}
			if err != nil {
				return nil, "", fmt.Errorf("observe volume %s: %w", id, err)
			}

			return current, current.Status, nil
		},
	}
	current, err := wait.WaitForStateContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("wait for volume %s: %w", id, err)
	}

	return current, nil
}

// A scoped GET may return 404 for a volume hidden from the token. Requiring a
// complete, successful list avoids dropping state on permission/pagination errors.
func confirmVolumeAbsent(ctx context.Context, client *sdk.Client, id string) error {
	all, err := volume.ListDetail(ctx, client, volume.ListOpts{})
	if err != nil {
		return fmt.Errorf("confirm absence of volume %s: %w", id, err)
	}
	for _, item := range all {
		if item.ID == id {
			return fmt.Errorf("volume %s is listed but cannot be read; check project permissions", id)
		}
	}

	return nil
}
