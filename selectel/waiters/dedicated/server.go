package dedicated

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	dedicated "github.com/selectel/dedicated-go/pkg/v2"
)

func WaitForServersServerV1ActiveState(
	ctx context.Context, client *dedicated.ServiceClient, resourceID string, timeout time.Duration,
) error {
	timer := time.NewTimer(30 * time.Minute)

	stateConf := &resource.StateChangeConf{
		Pending: []string{
			dedicated.ResourceDetailsStatePending,
			dedicated.ResourceDetailsStateProcessing,
			dedicated.ResourceDetailsStatePaid,
			dedicated.ResourceDetailsStateDeploy,
		},
		Target: []string{
			dedicated.ResourceDetailsStateActive,
		},
		Timeout:    timeout,
		Refresh:    serversServerV1StateRefreshFunc(ctx, client, resourceID, timer),
		MinTimeout: 10 * time.Second,
	}

	_, err := stateConf.WaitForStateContext(ctx)
	if err != nil {
		return fmt.Errorf("error waiting for the server %s to become 'active': %w", resourceID, err)
	}

	return nil
}

func serversServerV1StateRefreshFunc(ctx context.Context, client *dedicated.ServiceClient, id string, timer *time.Timer) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		select {
		case <-timer.C:
			log.Printf("[WARN] operation is taking more than 30 minutes, contact support")
		default:
		}

		d, _, err := client.ResourceDetails(ctx, id)
		if err != nil {
			return nil, "", err
		}

		return d, d.State, nil
	}
}

func WaitForServersServerV1RefusedToRenewState(
	ctx context.Context, client *dedicated.ServiceClient, resourceID string, timeout time.Duration,
) error {
	timer := time.NewTimer(30 * time.Minute)

	stateConf := &resource.StateChangeConf{
		Pending: []string{
			dedicated.ResourceDetailsStatePending,
			dedicated.ResourceDetailsStateProcessing,
			dedicated.ResourceDetailsStatePaid,
			dedicated.ResourceDetailsStateDeploy,
			dedicated.ResourceDetailsStateActive,
		},
		Target: []string{
			dedicated.ResourceDetailsStateExpiring,
			dedicated.ResourceDetailsStateEnding,
		},
		Timeout:    timeout,
		Refresh:    serversServerV1StateRefreshFunc(ctx, client, resourceID, timer),
		MinTimeout: 10 * time.Second,
	}

	_, err := stateConf.WaitForStateContext(ctx)
	if err != nil {
		return fmt.Errorf("error waiting for the server %s to become 'active': %w", resourceID, err)
	}

	return nil
}
