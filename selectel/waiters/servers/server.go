package servers

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/terraform-providers/terraform-provider-selectel/selectel/internal/api/servers"
)

func WaitForServersServerV1ActiveState(
	ctx context.Context, client *servers.ServiceClient, resourceID string, timeout time.Duration,
) error {
	stateConf := &resource.StateChangeConf{
		Pending: []string{
			servers.ResourceDetailsStatePending,
			servers.ResourceDetailsStateProcessing,
			servers.ResourceDetailsStatePaid,
			servers.ResourceDetailsStateDeploy,
		},
		Target: []string{
			servers.ResourceDetailsStateActive,
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

func serversServerV1StateRefreshFunc(ctx context.Context, client *servers.ServiceClient, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		d, _, err := client.ResourceDetails(ctx, id)
		if err != nil {
			return nil, "", err
		}

		return d, d.State, nil
	}
}

func WaitForServersServerV1RefusedToRenewState(
	ctx context.Context, client *servers.ServiceClient, resourceID string, timeout time.Duration,
) error {
	stateConf := &resource.StateChangeConf{
		Pending: []string{
			servers.ResourceDetailsStatePending,
			servers.ResourceDetailsStateProcessing,
			servers.ResourceDetailsStatePaid,
			servers.ResourceDetailsStateDeploy,
			servers.ResourceDetailsStateActive,
		},
		Target: []string{
			servers.ResourceDetailsStateExpiring,
			servers.ResourceDetailsStateEnding,
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
