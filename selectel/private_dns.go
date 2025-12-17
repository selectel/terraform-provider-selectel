package selectel

import (
	"fmt"
	"time"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	privatedns "github.com/selectel/private-dns-go/pkg/v1"
)

const (
	privateDNSDefaultRetryWaitMin = time.Second
	privateDNSDefaultRetryWaitMax = 5 * time.Second
	privateDNSDefaultRetry        = 5
)

func getPrivateDNSClient(d *schema.ResourceData, meta interface{}) (*privatedns.PrivateDNSClient, diag.Diagnostics) {
	config := meta.(*Config)
	projectID := d.Get("project_id").(string)
	region := d.Get("region").(string)

	selvpcClient, err := config.GetSelVPCClientWithProjectScope(projectID)
	if err != nil {
		return nil, diag.FromErr(fmt.Errorf("can't get project-scope selvpc client for private dns: %w", err))
	}

	err = validateRegion(selvpcClient, PrivateDNS, region)
	if err != nil {
		return nil, diag.FromErr(fmt.Errorf("can't validate region: %w", err))
	}

	endpoint, err := selvpcClient.Catalog.GetEndpoint(PrivateDNS, region)
	if err != nil {
		return nil, diag.FromErr(fmt.Errorf("can't get endpoint to init private dns client: %w", err))
	}

	retryClient := retryablehttp.NewClient()
	retryClient.Logger = nil
	retryClient.RetryWaitMin = privateDNSDefaultRetryWaitMin
	retryClient.RetryWaitMax = privateDNSDefaultRetryWaitMax
	retryClient.RetryMax = privateDNSDefaultRetry

	cfg := &privatedns.Config{
		URL:        endpoint.URL,
		AuthToken:  selvpcClient.GetXAuthToken(),
		HTTPClient: retryClient.StandardClient(),
		UserAgent:  config.UserAgent,
	}
	client := privatedns.NewPrivateDNSClient(cfg)

	return client, nil
}
