package selectel

import (
	"context"
	"errors"
	"fmt"
	"log"
	"slices"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	vpc "github.com/selectel/vpc-go/pkg/v2"
	"github.com/selectel/vpc-go/pkg/v2/floatingip"
)

const (
	vpcV2StatusActive        = "ACTIVE"
	vpcV2StatusDown          = "DOWN"
	vpcV2StatusError         = "ERROR"
	vpcV2StatusBuild         = "BUILD"
	vpcV2StatusAllocating    = "ALLOCATING"
	vpcV2StatusPendingCreate = "PENDING_CREATE"
	vpcV2StatusPendingUpdate = "PENDING_UPDATE"
)

const (
	vpcV2StatusPollInterval = 2 * time.Second
	vpcV2StatusTimeout      = time.Minute
)

const vpcFloatingIPV2Object = "VPC floating ip"

var errVPCV2StatusTimeout = errors.New("timeout waiting for status")

type vpcV2StatusRefreshFunc func(ctx context.Context) (string, error)

func waitForVPCV2Status(
	ctx context.Context, object, id string, pending []string,
	interval, timeout time.Duration, refresh vpcV2StatusRefreshFunc,
) (string, error) {
	deadline := time.Now().Add(timeout)
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		status, err := refresh(ctx)
		if err != nil {
			return "", fmt.Errorf("error getting %s %s status: %w", object, id, err)
		}
		if status == vpcV2StatusError {
			return status, fmt.Errorf("%s %s is in %s status", object, id, status)
		}
		if !slices.Contains(pending, status) {
			return status, nil
		}
		if time.Now().After(deadline) {
			return status, fmt.Errorf("%w: %s %s is still %s after %s",
				errVPCV2StatusTimeout, object, id, status, timeout)
		}

		log.Printf("[DEBUG] %s %s is %s, waiting", object, id, status)

		select {
		case <-ctx.Done():
			return status, ctx.Err()
		case <-ticker.C:
		}
	}
}

func waitForVPCFloatingIPV2Active(ctx context.Context, client *vpc.Client, id string) diag.Diagnostics {
	status, err := waitForVPCV2Status(
		ctx, vpcFloatingIPV2Object, id, []string{vpcV2StatusDown},
		vpcV2StatusPollInterval, vpcV2StatusTimeout,
		func(ctx context.Context) (string, error) {
			found, err := floatingip.Get(ctx, client, id)
			if err != nil {
				return "", err
			}

			return found.Status, nil
		},
	)
	if errors.Is(err, errVPCV2StatusTimeout) {
		return diag.Diagnostics{{
			Severity: diag.Warning,
			Summary:  fmt.Sprintf("%s %s is still %s", vpcFloatingIPV2Object, id, status),
			Detail: "The API accepted the change, but the network backend has not " +
				"reported the floating ip as ACTIVE yet. Check its status before " +
				"relying on the address.",
		}}
	}
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}
