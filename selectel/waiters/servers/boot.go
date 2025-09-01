package servers

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/terraform-providers/terraform-provider-selectel/selectel/internal/api/servers"
)

func WaitForServersServerInstallNewOSV1ActiveState(
	ctx context.Context, client *servers.ServiceClient, resourceID string, timeout time.Duration,
) error {
	stateConf := &resource.StateChangeConf{
		Pending: []string{
			"1",
		},
		Target: []string{
			"0",
		},
		Refresh:    serversServerInstallNewOSV1StateRefreshFunc(ctx, client, resourceID),
		Timeout:    timeout,
		MinTimeout: 15 * time.Second,
	}

	_, err := stateConf.WaitForStateContext(ctx)
	if err != nil {
		return fmt.Errorf("error waiting for the server %s to become 'active': %s", resourceID, err)
	}

	return nil
}

func serversServerInstallNewOSV1StateRefreshFunc(ctx context.Context, client *servers.ServiceClient, resourceID string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		d, _, err := client.OperatingSystemByResource(ctx, resourceID)
		if err != nil {
			return nil, "", err
		}

		return d, strconv.Itoa(d.Reinstall), nil
	}
}
