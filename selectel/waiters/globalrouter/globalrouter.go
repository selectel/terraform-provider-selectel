package globalrouter

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	globalrouter "github.com/selectel/globalrouter-go/pkg/v1"
)

func WaitForRouterV1ActiveState(
	ctx context.Context, client *globalrouter.ServiceClient, id string, timeout time.Duration,
) diag.Diagnostics {
	stateConf := &resource.StateChangeConf{
		Pending: []string{
			globalrouter.GlobalRouterResourceStatusUpdating,
			globalrouter.GlobalRouterResourceStatusCreating,
		},
		Target: []string{
			globalrouter.GlobalRouterResourceStatusActive,
		},
		Timeout:    timeout,
		Refresh:    routerV1RefreshFunc(ctx, client, id),
		MinTimeout: 10 * time.Second,
	}

	_, err := stateConf.WaitForStateContext(ctx)
	if err != nil {
		return diag.Errorf(
			"error waiting for the router %s to become '%s': %v",
			id, globalrouter.GlobalRouterResourceStatusActive, err,
		)
	}

	return nil
}

func WaitForRouterV1Deleted(
	ctx context.Context, client *globalrouter.ServiceClient, id string, timeout time.Duration,
) diag.Diagnostics {
	stateConf := &resource.StateChangeConf{
		Pending: []string{
			globalrouter.GlobalRouterResourceStatusDeleting,
		},
		Target:  []string{},
		Timeout: timeout,
		Refresh: routerV1DeleteRefreshFunc(ctx, client, id),
	}

	_, err := stateConf.WaitForStateContext(ctx)
	if err != nil {
		return diag.Errorf("error waiting for the router %s to be deleted: %v", id, err)
	}

	return nil
}

func WaitForNetworkV1ActiveState(
	ctx context.Context, client *globalrouter.ServiceClient, id string, timeout time.Duration,
) diag.Diagnostics {
	stateConf := &resource.StateChangeConf{
		Pending: []string{
			globalrouter.GlobalRouterResourceStatusUpdating,
			globalrouter.GlobalRouterResourceStatusCreating,
		},
		Target: []string{
			globalrouter.GlobalRouterResourceStatusActive,
		},
		Timeout:    timeout,
		Refresh:    networkV1RefreshFunc(ctx, client, id),
		MinTimeout: 10 * time.Second,
	}

	_, err := stateConf.WaitForStateContext(ctx)
	if err != nil {
		return diag.Errorf(
			"error waiting for the network %s to become '%s': %v",
			id, globalrouter.GlobalRouterResourceStatusActive, err,
		)
	}

	return nil
}

func WaitForNetworkV1Deleted(
	ctx context.Context, client *globalrouter.ServiceClient, id string, timeout time.Duration,
) diag.Diagnostics {
	stateConf := &resource.StateChangeConf{
		Pending: []string{
			globalrouter.GlobalRouterResourceStatusDeleting,
		},
		Target:  []string{},
		Timeout: timeout,
		Refresh: networkV1DeleteRefreshFunc(ctx, client, id),
	}

	_, err := stateConf.WaitForStateContext(ctx)
	if err != nil {
		return diag.Errorf("error waiting for the network %s to be deleted: %v", id, err)
	}

	return nil
}

func WaitForSubnetV1ActiveState(
	ctx context.Context, client *globalrouter.ServiceClient, id string, timeout time.Duration,
) diag.Diagnostics {
	stateConf := &resource.StateChangeConf{
		Pending: []string{
			globalrouter.GlobalRouterResourceStatusUpdating,
			globalrouter.GlobalRouterResourceStatusCreating,
		},
		Target: []string{
			globalrouter.GlobalRouterResourceStatusActive,
		},
		Timeout:    timeout,
		Refresh:    subnetV1RefreshFunc(ctx, client, id),
		MinTimeout: 10 * time.Second,
	}

	_, err := stateConf.WaitForStateContext(ctx)
	if err != nil {
		return diag.Errorf(
			"error waiting for the subnet %s to become '%s': %v",
			id, globalrouter.GlobalRouterResourceStatusActive, err,
		)
	}

	return nil
}

func WaitForSubnetV1Deleted(
	ctx context.Context, client *globalrouter.ServiceClient, id string, timeout time.Duration,
) diag.Diagnostics {
	stateConf := &resource.StateChangeConf{
		Pending: []string{
			globalrouter.GlobalRouterResourceStatusDeleting,
		},
		Target:  []string{},
		Timeout: timeout,
		Refresh: subnetV1DeleteRefreshFunc(ctx, client, id),
	}

	_, err := stateConf.WaitForStateContext(ctx)
	if err != nil {
		return diag.Errorf("error waiting for the subnet %s to be deleted: %v", id, err)
	}

	return nil
}

func WaitForStaticRouteV1ActiveState(
	ctx context.Context, client *globalrouter.ServiceClient, id string, timeout time.Duration,
) diag.Diagnostics {
	stateConf := &resource.StateChangeConf{
		Pending: []string{
			globalrouter.GlobalRouterResourceStatusUpdating,
			globalrouter.GlobalRouterResourceStatusCreating,
		},
		Target: []string{
			globalrouter.GlobalRouterResourceStatusActive,
		},
		Timeout:    timeout,
		Refresh:    staticRouteV1RefreshFunc(ctx, client, id),
		MinTimeout: 10 * time.Second,
	}

	_, err := stateConf.WaitForStateContext(ctx)
	if err != nil {
		return diag.Errorf(
			"error waiting for the static route %s to become '%s': %v",
			id, globalrouter.GlobalRouterResourceStatusActive, err,
		)
	}

	return nil
}

func WaitForStaticRouteV1Deleted(
	ctx context.Context, client *globalrouter.ServiceClient, id string, timeout time.Duration,
) diag.Diagnostics {
	stateConf := &resource.StateChangeConf{
		Pending: []string{
			globalrouter.GlobalRouterResourceStatusDeleting,
		},
		Target:  []string{},
		Timeout: timeout,
		Refresh: staticRouteV1DeleteRefreshFunc(ctx, client, id),
	}

	_, err := stateConf.WaitForStateContext(ctx)
	if err != nil {
		return diag.Errorf("error waiting for the static route %s to be deleted: %v", id, err)
	}

	return nil
}

func routerV1RefreshFunc(ctx context.Context, client *globalrouter.ServiceClient, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		p, _, err := client.Router(ctx, id)
		if err != nil {
			return nil, "", err
		}

		if p == nil {
			return nil, "", fmt.Errorf("can't find created router %s", id)
		}

		return p, p.Status, nil
	}
}

func routerV1DeleteRefreshFunc(ctx context.Context, client *globalrouter.ServiceClient, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		p, resp, err := client.Router(ctx, id)
		switch {
		case resp != nil && resp.StatusCode == http.StatusNotFound:
			return nil, "", nil
		case err != nil:
			return nil, "", err
		}

		return p, p.Status, nil
	}
}

func networkV1RefreshFunc(ctx context.Context, client *globalrouter.ServiceClient, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		p, _, err := client.Network(ctx, id)
		if err != nil {
			return nil, "", err
		}

		if p == nil {
			return nil, "", fmt.Errorf("can't find created network %s", id)
		}

		return p, p.Status, nil
	}
}

func networkV1DeleteRefreshFunc(ctx context.Context, client *globalrouter.ServiceClient, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		p, resp, err := client.Network(ctx, id)
		switch {
		case resp != nil && resp.StatusCode == http.StatusNotFound:
			return nil, "", nil
		case err != nil:
			return nil, "", err
		}

		return p, p.Status, nil
	}
}

func subnetV1RefreshFunc(ctx context.Context, client *globalrouter.ServiceClient, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		p, _, err := client.Subnet(ctx, id)
		if err != nil {
			return nil, "", err
		}

		if p == nil {
			return nil, "", fmt.Errorf("can't find created subnet %s", id)
		}

		return p, p.Status, nil
	}
}

func subnetV1DeleteRefreshFunc(ctx context.Context, client *globalrouter.ServiceClient, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		p, resp, err := client.Subnet(ctx, id)
		switch {
		case resp != nil && resp.StatusCode == http.StatusNotFound:
			return nil, "", nil
		case err != nil:
			return nil, "", err
		}

		return p, p.Status, nil
	}
}

func staticRouteV1RefreshFunc(ctx context.Context, client *globalrouter.ServiceClient, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		p, _, err := client.StaticRoute(ctx, id)
		if err != nil {
			return nil, "", err
		}

		if p == nil {
			return nil, "", fmt.Errorf("can't find created static route %s", id)
		}

		return p, p.Status, nil
	}
}

func staticRouteV1DeleteRefreshFunc(ctx context.Context, client *globalrouter.ServiceClient, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		p, resp, err := client.StaticRoute(ctx, id)
		switch {
		case resp != nil && resp.StatusCode == http.StatusNotFound:
			return nil, "", nil
		case err != nil:
			return nil, "", err
		}

		return p, p.Status, nil
	}
}
