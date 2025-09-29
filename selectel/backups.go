package selectel

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/terraform-providers/terraform-provider-selectel/selectel/internal/api/cloudbackup"
)

func getScheduledBackupClient(d *schema.ResourceData, meta interface{}) (*cloudbackup.ServiceClient, diag.Diagnostics) {
	config := meta.(*Config)
	projectID := d.Get("project_id").(string)
	region := d.Get("region").(string)

	selvpcClient, err := config.GetSelVPCClientWithProjectScope(projectID)
	if err != nil {
		return nil, diag.FromErr(fmt.Errorf("can't get project-scope selvpc client for scheduled backup api: %w", err))
	}

	err = validateRegion(selvpcClient, DataProtectV2, region)
	if err != nil {
		return nil, diag.FromErr(fmt.Errorf("can't validate region: %w", err))
	}

	endpoint, err := selvpcClient.Catalog.GetEndpoint(DataProtectV2, region)
	if err != nil {
		return nil, diag.FromErr(fmt.Errorf("can't get endpoint to init client: %w", err))
	}

	return cloudbackup.NewClientV2(selvpcClient.GetXAuthToken(), endpoint.URL), nil
}
