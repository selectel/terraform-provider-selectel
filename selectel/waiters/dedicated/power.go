package dedicated

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	dedicated "github.com/selectel/dedicated-go/v2/pkg/v2"
)

const (
	powerStateOn      = "on"
	powerStateOff     = "off"
	powerActionReboot = "reboot"
	powerUnknown      = "unknown"
)

func WaitForServersV1PowerRunning(
	ctx context.Context, client *dedicated.ServiceClient, resourceID string, timeout time.Duration,
) error {
	timer := time.NewTimer(30 * time.Minute)

	stateConf := &resource.StateChangeConf{
		Pending: []string{
			powerStateOff,
			powerUnknown,
		},
		Target: []string{
			powerStateOn,
		},
		Timeout:    timeout,
		Refresh:    serversServerV1PowerRefreshFunc(ctx, client, resourceID, timer),
		MinTimeout: 10 * time.Second,
	}

	_, err := stateConf.WaitForStateContext(ctx)
	if err != nil {
		return fmt.Errorf("error waiting for the server %s to become 'running': %w", resourceID, err)
	}

	return nil
}

func WaitForServersV1PowerRunningAfterReboot(
	ctx context.Context, client *dedicated.ServiceClient, resourceID string, timeout time.Duration,
) error {
	timer := time.NewTimer(30 * time.Minute)

	stateConf := &resource.StateChangeConf{
		Pending: []string{
			powerActionReboot,
			powerUnknown,
		},
		Target: []string{
			powerStateOn,
		},
		Timeout:    timeout,
		Refresh:    serversServerV1PowerRefreshFunc(ctx, client, resourceID, timer),
		MinTimeout: 10 * time.Second,
	}

	_, err := stateConf.WaitForStateContext(ctx)
	if err != nil {
		return fmt.Errorf("error waiting for the server %s to become 'running': %w", resourceID, err)
	}

	return nil
}

func WaitForServersV1PowerShutdown(
	ctx context.Context, client *dedicated.ServiceClient, resourceID string, timeout time.Duration,
) error {
	timer := time.NewTimer(30 * time.Minute)

	stateConf := &resource.StateChangeConf{
		Pending: []string{
			powerStateOn,
			powerUnknown,
		},
		Target: []string{
			powerStateOff,
		},
		Timeout:    timeout,
		Refresh:    serversServerV1PowerRefreshFunc(ctx, client, resourceID, timer),
		MinTimeout: 10 * time.Second,
	}

	_, err := stateConf.WaitForStateContext(ctx)
	if err != nil {
		return fmt.Errorf("error waiting for the server %s to become 'shutdown': %w", resourceID, err)
	}

	return nil
}

func serversServerV1PowerRefreshFunc(ctx context.Context, client *dedicated.ServiceClient, id string, timer *time.Timer) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		select {
		case <-timer.C:
			log.Printf("[WARN] operation is taking more than 30 minutes, contact support")
		default:
		}

		d, _, err := client.ShowPowerState(ctx, id)
		if err != nil {
			return nil, "", err
		}

		if d.IsOn() {
			return d, powerStateOn, nil
		}

		if d.IsOff() {
			return d, powerStateOff, nil
		}

		if d.IsReboot() {
			return d, powerActionReboot, nil
		}

		return d, powerUnknown, nil
	}
}
