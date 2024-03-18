package waiters

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/selectel/dbaas-go"
)

func WaitForDBaaSDatastoreV1ActiveState(
	ctx context.Context, client *dbaas.API, datastoreID string, timeout time.Duration,
) error {
	pending := []string{
		string(dbaas.StatusPendingCreate),
		string(dbaas.StatusPendingUpdate),
		string(dbaas.StatusResizing),
	}
	target := []string{
		string(dbaas.StatusActive),
	}

	stateConf := &resource.StateChangeConf{
		Pending:    pending,
		Target:     target,
		Refresh:    dbaasDatastoreV1StateRefreshFunc(ctx, client, datastoreID),
		Timeout:    timeout,
		Delay:      10 * time.Second,
		MinTimeout: 20 * time.Second,
	}

	_, err := stateConf.WaitForState()
	if err != nil {
		return fmt.Errorf(
			"error waiting for the datastore %s to become 'ACTIVE': %s",
			datastoreID, err)
	}

	return nil
}

func DBaaSDatastoreV1DeleteStateRefreshFunc(ctx context.Context, client *dbaas.API, datastoreID string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		d, err := client.Datastore(ctx, datastoreID)
		if err != nil {
			var dbaasError *dbaas.DBaaSAPIError
			if errors.As(err, &dbaasError) {
				return d, strconv.Itoa(dbaasError.StatusCode()), nil
			}

			return nil, "", err
		}

		return d, strconv.Itoa(http.StatusOK), err
	}
}

func dbaasDatastoreV1StateRefreshFunc(ctx context.Context, client *dbaas.API, datastoreID string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		d, err := client.Datastore(ctx, datastoreID)
		if err != nil {
			return nil, "", err
		}

		return d, string(d.Status), nil
	}
}
