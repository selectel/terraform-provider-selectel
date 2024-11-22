package selectel

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	v1 "github.com/selectel/craas-go/pkg"
	"github.com/selectel/craas-go/pkg/v1/registry"
	"github.com/selectel/go-selvpcclient/v4/selvpcclient"
)

const (
	craasV1TokenUsername = "token"
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

func getCRaaSClient(d *schema.ResourceData, meta interface{}) (*v1.ServiceClient, diag.Diagnostics) {
	config := meta.(*Config)
	selvpcClient, err := config.GetSelVPCClientWithProjectScope(d.Get("project_id").(string))
	if err != nil {
		return nil, diag.FromErr(fmt.Errorf("can't get project-scope selvpc client for craas: %w", err))
	}

	endpoint, diagErr := getEndpointForCRaaS(selvpcClient)
	if diagErr != nil {
		return nil, diag.FromErr(fmt.Errorf("can't get endpoint to init craas client: %w", err))
	}

	craasClient := v1.NewCRaaSClientV1(selvpcClient.GetXAuthToken(), endpoint)

	return craasClient, nil
}

// https://cr.selcloud.ru/api/v1 -> https://cr.selcloud.ru
func getHostNameForCRaaS(endpoint string) (string, error) {
	parsedEndpoint, err := url.Parse(endpoint)
	if err != nil {
		return "", fmt.Errorf("cant parse url for craas endpoint: %w", err)
	}

	return fmt.Sprintf("%s://%s", parsedEndpoint.Scheme, parsedEndpoint.Host), nil
}

func getEndpointForCRaaS(selvpcClient *selvpcclient.Client) (string, error) {
	endpoints, err := selvpcClient.Catalog.GetEndpoints(CRaaS)
	if err != nil {
		return "", fmt.Errorf("can't get endpoint to for craas: %w", err)
	}

	// There is no actual regionality for CRaaS, but we need to support any environments where the region is
	// called whatever
	if len(endpoints) > 1 {
		return "", fmt.Errorf("unexpectedly received more than one endpoint for craas")
	}

	return endpoints[0].URL, nil
}
