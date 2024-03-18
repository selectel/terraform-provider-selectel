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

func WaitForDBaaSDatabaseV1ActiveState(
	ctx context.Context, client *dbaas.API, databaseID string, timeout time.Duration,
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
		Refresh:    dbaasDatabaseV1StateRefreshFunc(ctx, client, databaseID),
		Timeout:    timeout,
		Delay:      10 * time.Second,
		MinTimeout: 10 * time.Second,
	}

	_, err := stateConf.WaitForStateContext(ctx)
	if err != nil {
		return fmt.Errorf(
			"error waiting for the database %s to become 'ACTIVE': %s",
			databaseID, err)
	}

	return nil
}

func DBaaSDatabaseV1DeleteStateRefreshFunc(ctx context.Context, client *dbaas.API, datastoreID string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		d, err := client.Database(ctx, datastoreID)
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

func dbaasDatabaseV1StateRefreshFunc(ctx context.Context, client *dbaas.API, databaseID string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		d, err := client.Database(ctx, databaseID)
		if err != nil {
			return nil, "", err
		}

		return d, string(d.Status), nil
	}
}
