package selectel

import (
	"errors"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func importProjectRegion(d *schema.ResourceData, meta any) error {
	if used, err := importByIdentity(d); used || err != nil {
		return err
	}

	return importProjectRegionFromConfig(d, meta)
}

func importProjectRegionFromConfig(d *schema.ResourceData, meta any) error {
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

func importIdentity(d *schema.ResourceData) (*schema.IdentityData, bool) {
	if d.Id() != "" {
		return nil, false
	}

	identity, err := d.Identity()
	if err != nil {
		return nil, false
	}

	projectID, _ := identity.Get("project_id").(string)

	return identity, projectID != ""
}

func importByIdentity(d *schema.ResourceData) (bool, error) {
	identity, ok := importIdentity(d)
	if !ok {
		return false, nil
	}

	d.SetId(identity.Get("id").(string))

	return true, setProjectRegionFromIdentity(d, identity)
}

func setProjectRegionFromIdentity(d *schema.ResourceData, identity *schema.IdentityData) error {
	if err := d.Set("project_id", identity.Get("project_id")); err != nil {
		return err
	}

	return d.Set("region", identity.Get("region"))
}

func setProjectRegionIdentity(d *schema.ResourceData) {
	setIdentity(d, "id")
}

func setIdentity(d *schema.ResourceData, attrs ...string) {
	identity, err := d.Identity()
	if err != nil {
		return
	}

	_ = identity.Set("project_id", d.Get("project_id"))
	_ = identity.Set("region", d.Get("region"))

	for _, attr := range attrs {
		if attr == "id" {
			_ = identity.Set("id", d.Id())

			continue
		}

		_ = identity.Set(attr, d.Get(attr))
	}
}
