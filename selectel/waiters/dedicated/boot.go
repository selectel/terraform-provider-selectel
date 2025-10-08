package dedicated

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/terraform-providers/terraform-provider-selectel/selectel/internal/api/dedicated"
)

func WaitForServersServerInstallNewOSV1ActiveState(
	ctx context.Context, client *dedicated.ServiceClient, resourceID string, timeout time.Duration,
) error {
	timer := time.NewTimer(30 * time.Minute)

	stateConf := &resource.StateChangeConf{
		Pending: []string{
			"1",
		},
		Target: []string{
			"0",
		},
		Refresh:    serversServerInstallNewOSV1StateRefreshFunc(ctx, client, resourceID, timer),
		Timeout:    timeout,
		MinTimeout: 15 * time.Second,
	}

	_, err := stateConf.WaitForStateContext(ctx)
	if err != nil {
		return fmt.Errorf("error waiting for the server %s to become 'active': %s", resourceID, err)
	}

	return nil
}

func serversServerInstallNewOSV1StateRefreshFunc(
	ctx context.Context, client *dedicated.ServiceClient, resourceID string, timer *time.Timer,
) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		select {
		case <-timer.C:
			log.Printf("[WARN] reinstalling the OS is taking more than 30 minutes, contact support")
		default:
		}

		d, _, err := client.OperatingSystemByResource(ctx, resourceID)
		if err != nil {
			return nil, "", err
		}

		return d, strconv.Itoa(d.Reinstall), nil
	}
}
