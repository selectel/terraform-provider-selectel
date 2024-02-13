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

func WaitForDBaaSLogicalReplicationSlotV1ActiveState(
	ctx context.Context, client *dbaas.API, slotID string, timeout time.Duration,
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
		Refresh:    dbaasLogicalReplicationSlotV1StateRefreshFunc(ctx, client, slotID),
		Timeout:    timeout,
		Delay:      10 * time.Second,
		MinTimeout: 15 * time.Second,
	}

	_, err := stateConf.WaitForStateContext(ctx)
	if err != nil {
		return fmt.Errorf(
			"error waiting for the slot %s to become 'ACTIVE': %s",
			slotID, err)
	}

	return nil
}

func DBaaSLogicalReplicationSlotV1DeleteStateRefreshFunc(ctx context.Context, client *dbaas.API, slotID string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		d, err := client.LogicalReplicationSlot(ctx, slotID)
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

func dbaasLogicalReplicationSlotV1StateRefreshFunc(ctx context.Context, client *dbaas.API, slotID string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		d, err := client.LogicalReplicationSlot(ctx, slotID)
		if err != nil {
			return nil, "", err
		}

		return d, string(d.Status), nil
	}
}
