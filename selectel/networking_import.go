package selectel

import (
	"errors"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// resourceNetworkingImportProjectRegion fills project_id and region from the
// provider-level configuration during import. Neutron import IDs carry only the
// remote identifier, following the pattern already established by the other
// resources of this provider (see resourceMKSClusterV1ImportState).
func resourceNetworkingImportProjectRegion(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	if config.ProjectID == "" {
		return errors.New("INFRA_PROJECT_ID must be set for the resource import")
	}
	if config.Region == "" {
		return errors.New("INFRA_REGION must be set for the resource import")
	}

	if err := d.Set("project_id", config.ProjectID); err != nil {
		return err
	}

	return d.Set("region", config.Region)
}
