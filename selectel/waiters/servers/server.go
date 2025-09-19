package servers

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/terraform-providers/terraform-provider-selectel/selectel/internal/api/dedicatedservers"
)

func WaitForServersServerV1ActiveState(
	ctx context.Context, client *dedicatedservers.ServiceClient, resourceID string, timeout time.Duration,
) error {
	stateConf := &resource.StateChangeConf{
		Pending: []string{
			dedicatedservers.ResourceDetailsStatePending,
			dedicatedservers.ResourceDetailsStateProcessing,
			dedicatedservers.ResourceDetailsStatePaid,
			dedicatedservers.ResourceDetailsStateDeploy,
		},
		Target: []string{
			dedicatedservers.ResourceDetailsStateActive,
		},
		Timeout:    timeout,
		Refresh:    serversServerV1StateRefreshFunc(ctx, client, resourceID),
		MinTimeout: 10 * time.Second,
	}

	_, err := stateConf.WaitForStateContext(ctx)
	if err != nil {
		return fmt.Errorf("error waiting for the server %s to become 'active': %w", resourceID, err)
	}

	return nil
}

func serversServerV1StateRefreshFunc(ctx context.Context, client *dedicatedservers.ServiceClient, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		d, _, err := client.ResourceDetails(ctx, id)
		if err != nil {
			return nil, "", err
		}

		return d, d.State, nil
	}
}

func WaitForServersServerV1RefusedToRenewState(
	ctx context.Context, client *dedicatedservers.ServiceClient, resourceID string, timeout time.Duration,
) error {
	stateConf := &resource.StateChangeConf{
		Pending: []string{
			dedicatedservers.ResourceDetailsStatePending,
			dedicatedservers.ResourceDetailsStateProcessing,
			dedicatedservers.ResourceDetailsStatePaid,
			dedicatedservers.ResourceDetailsStateDeploy,
			dedicatedservers.ResourceDetailsStateActive,
		},
		Target: []string{
			dedicatedservers.ResourceDetailsStateExpiring,
			dedicatedservers.ResourceDetailsStateEnding,
		},
		Timeout:    timeout,
		Refresh:    serversServerV1StateRefreshFunc(ctx, client, resourceID),
		MinTimeout: 10 * time.Second,
	}

	_, err := stateConf.WaitForStateContext(ctx)
	if err != nil {
		return fmt.Errorf("error waiting for the server %s to become 'active': %w", resourceID, err)
	}

	return nil
}
