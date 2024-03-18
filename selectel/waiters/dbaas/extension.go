package waiters

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/selectel/dbaas-go"
)

func WaitForDBaaSExtensionV1ActiveState(
	ctx context.Context, client *dbaas.API, extensionID string, timeout time.Duration,
) error {
	pending := []string{
		string(dbaas.StatusPendingCreate),
		string(dbaas.StatusPendingUpdate),
	}
	target := []string{
		string(dbaas.StatusActive),
	}

	stateConf := &resource.StateChangeConf{
		Pending:    pending,
		Target:     target,
		Refresh:    dbaasExtensionV1StateRefreshFunc(ctx, client, extensionID),
		Timeout:    timeout,
		Delay:      10 * time.Second,
		MinTimeout: 15 * time.Second,
	}

	_, err := stateConf.WaitForStateContext(ctx)
	if err != nil {
		return fmt.Errorf(
			"error waiting for the extension %s to become 'ACTIVE': %s",
			extensionID, err)
	}

	return nil
}

func DBaaSExtensionV1DeleteStateRefreshFunc(ctx context.Context, client *dbaas.API, extensionID string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		d, err := client.Extension(ctx, extensionID)
		if err != nil {
			var dbaasError *dbaas.DBaaSAPIError
			if errors.As(err, &dbaasError) {
				return d, strconv.Itoa(dbaasError.StatusCode()), nil
			}

			return nil, "", err
		}

		return d, strconv.Itoa(200), err
	}
}

func dbaasExtensionV1StateRefreshFunc(ctx context.Context, client *dbaas.API, extensionID string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		d, err := client.Extension(ctx, extensionID)
		if err != nil {
			return nil, "", err
		}

		return d, string(d.Status), nil
	}
}
