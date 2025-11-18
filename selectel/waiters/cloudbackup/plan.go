package cloudbackup

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	cloudbackup "github.com/selectel/cloudbackup-go/pkg/v2"
)

func WaitForPlanV2StartedState(
	ctx context.Context, client *cloudbackup.ServiceClient, id string, timeout time.Duration,
) diag.Diagnostics {
	stateConf := &resource.StateChangeConf{
		Pending: []string{
			cloudbackup.PlanStatusSuspended,
		},
		Target: []string{
			cloudbackup.PlanStatusStarted,
		},
		Timeout:    timeout,
		Refresh:    planV2RefreshFunc(ctx, client, id),
		MinTimeout: 10 * time.Second,
	}

	_, err := stateConf.WaitForStateContext(ctx)
	if err != nil {
		return diag.Errorf(
			"error waiting for the plan %s to become '%s': %v",
			id, cloudbackup.PlanStatusStarted, err,
		)
	}

	return nil
}

func planV2RefreshFunc(ctx context.Context, client *cloudbackup.ServiceClient, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		p, _, err := client.Plan(ctx, id)
		if err != nil {
			return nil, "", err
		}

		if p == nil {
			return nil, "", fmt.Errorf("can't find created plan %s", id)
		}

		return p, p.Status, nil
	}
}

func WaitForPlanV2Deleted(
	ctx context.Context, client *cloudbackup.ServiceClient, id string, timeout time.Duration,
) diag.Diagnostics {
	stateConf := &resource.StateChangeConf{
		Pending: []string{
			cloudbackup.PlanStatusStarted,
			cloudbackup.PlanStatusSuspended,
		},
		Target:  []string{},
		Timeout: timeout,
		Refresh: planV2DeleteRefreshFunc(ctx, client, id),
	}

	_, err := stateConf.WaitForStateContext(ctx)
	if err != nil {
		return diag.Errorf("error waiting for the plan %s to be deleted: %v", id, err)
	}

	return nil
}

func planV2DeleteRefreshFunc(ctx context.Context, client *cloudbackup.ServiceClient, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		p, resp, err := client.Plan(ctx, id)
		switch {
		case resp != nil && resp.StatusCode == http.StatusNotFound:
			return nil, "", nil
		case err != nil:
			return nil, "", err
		}

		return p, p.Status, nil
	}
}
