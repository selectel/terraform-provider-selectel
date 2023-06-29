package selectel

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	v1 "github.com/selectel/craas-go/pkg"
	"github.com/selectel/craas-go/pkg/v1/registry"
)

const (
	craasV1Endpoint         = "https://cr.selcloud.ru/api/v1"
	craasV1RegistryHostName = "cr.selcloud.ru"
	craasV1TokenUsername    = "token"
)

func waitForCRaaSRegistryV1StableState(
	ctx context.Context, client *v1.ServiceClient, registryID string, timeout time.Duration,
) error {
	pending := []string{
		string(registry.StatusCreating),
		string(registry.StatusDeleting),
		string(registry.StatusGC),
	}
	target := []string{
		string(registry.StatusActive),
	}

	stateConf := &resource.StateChangeConf{
		Pending:      pending,
		Target:       target,
		Timeout:      timeout,
		Refresh:      craasRegistryV1StateRefreshFunc(ctx, client, registryID),
		Delay:        1 * time.Second,
		PollInterval: 1 * time.Second,
	}

	_, err := stateConf.WaitForStateContext(ctx)
	if err != nil {
		return fmt.Errorf(
			"error waiting for registry %s to achieve a stable state: %s",
			registryID, err)
	}

	return nil
}

func craasRegistryV1StateRefreshFunc(
	ctx context.Context, client *v1.ServiceClient, registryID string,
) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		r, _, err := registry.Get(ctx, client, registryID)
		if err != nil {
			return nil, "", err
		}

		return r, string(r.Status), nil
	}
}
