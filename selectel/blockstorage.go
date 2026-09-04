package selectel

import (
	"errors"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	blockstorage "github.com/selectel/blockstorage-go/pkg/v1"
)

func getBlockStorageClient(d *schema.ResourceData, meta any) (*blockstorage.Client, diag.Diagnostics) {
	config := meta.(*Config)
	projectID := d.Get("project_id").(string)
	region := d.Get("region").(string)

	selvpcClient, err := config.GetSelVPCClientWithProjectScope(projectID)
	if err != nil {
		return nil, diag.FromErr(fmt.Errorf(
			"failed to get project-scoped Selectel VPC client for Block Storage (project_id=%s): %w",
			projectID,
			err,
		))
	}

	if err := validateRegion(selvpcClient, BlockStorage, region); err != nil {
		return nil, diag.FromErr(fmt.Errorf(
			"failed to validate Block Storage region (service=%s, region=%s, project_id=%s): %w",
			BlockStorage,
			region,
			projectID,
			err,
		))
	}

	endpoint, err := selvpcClient.Catalog.GetEndpoint(BlockStorage, region)
	if err != nil {
		return nil, diag.FromErr(fmt.Errorf(
			"failed to get Block Storage endpoint (service=%s, region=%s, project_id=%s): %w",
			BlockStorage,
			region,
			projectID,
			err,
		))
	}

	if endpoint.URL == "" {
		return nil, diag.Errorf(
			"Block Storage endpoint is empty (service=%s, region=%s, project_id=%s)",
			BlockStorage,
			region,
			projectID,
		)
	}

	client, err := blockstorage.NewClient(blockstorage.Config{
		Endpoint:  endpoint.URL,
		Token:     selvpcClient.GetXAuthToken(),
		UserAgent: config.UserAgent,
	})
	if err != nil {
		return nil, diag.FromErr(fmt.Errorf(
			"failed to initialize Block Storage client (region=%s, project_id=%s): %w",
			region,
			projectID,
			err,
		))
	}

	return client, nil
}

func blockStorageOperationDiagnostics(operation string, err error) diag.Diagnostics {
	operationErr := fmt.Errorf("failed to %s: %w", operation, err)

	var sdkErr *blockstorage.Error
	if errors.As(err, &sdkErr) &&
		sdkErr.Kind == blockstorage.KindIncompleteList &&
		sdkErr.Err != nil {
		operationErr = fmt.Errorf("%w; cause: %v", operationErr, sdkErr.Err)
	}

	return diag.FromErr(operationErr)
}
