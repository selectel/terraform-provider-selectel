package selectel

import (
	"fmt"
	"time"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	publicnetapi "github.com/selectel/public-net-api-go/pkg/v1"
)

const (
	defaultTimeout      = 150 * time.Second
	defaultRetryWaitMin = time.Second //nolint: revive
	defaultRetryWaitMax = 5 * time.Second
	defaultRetryMax     = 2
)

func getPublicNetAPIClient(
	d *schema.ResourceData,
	meta interface{},
) (*publicnetapi.PublicNetAPIClient, diag.Diagnostics) {
	config := meta.(*Config)
	projectID := d.Get("project_id").(string)
	region := d.Get("region").(string)

	selectelVPCClient, err := config.GetSelVPCClientWithProjectScope(projectID)
	if err != nil {
		return nil, diag.FromErr(
			fmt.Errorf("failed to get project-scope Selectel VPC client (project_id=%s): %w", projectID, err),
		)
	}

	endpoint, err := selectelVPCClient.Catalog.GetEndpoint(PublicNetAPI, region)
	if err != nil {
		return nil, diag.FromErr(
			fmt.Errorf(
				"failed to get endpoint for %s (region=%s, project_id=%s): %w",
				PublicNetAPI,
				region,
				projectID,
				err,
			),
		)
	}

	if endpoint.URL == "" {
		return nil, diag.FromErr(
			fmt.Errorf(
				"empty endpoint URL returned for %s (region=%s, project_id=%s)",
				PublicNetAPI,
				region,
				projectID,
			),
		)
	}

	retryClient := retryablehttp.NewClient()
	retryClient.Logger = nil
	retryClient.RetryWaitMin = defaultRetryWaitMin
	retryClient.RetryWaitMax = defaultRetryWaitMax
	retryClient.RetryMax = defaultRetryMax
	retryClient.HTTPClient.Timeout = defaultTimeout

	cfg := &publicnetapi.Config{
		URL:        endpoint.URL,
		AuthToken:  selectelVPCClient.GetXAuthToken(),
		HTTPClient: retryClient.StandardClient(),
		UserAgent:  config.UserAgent,
	}

	client, err := publicnetapi.NewPublicNetAPIClient(cfg)
	if err != nil {
		return nil, diag.FromErr(fmt.Errorf("failed to initialize %s client: %w", PublicNetAPI, err))
	}

	return client, nil
}
