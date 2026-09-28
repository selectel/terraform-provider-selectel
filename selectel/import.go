package selectel

import (
	"errors"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func importProjectRegion(d *schema.ResourceData, meta any) error {
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
