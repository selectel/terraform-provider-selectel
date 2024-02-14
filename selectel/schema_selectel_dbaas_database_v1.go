package selectel

import "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

func resourceDBaaSDatabaseV1Schema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"name": {
			Type:     schema.TypeString,
			Required: true,
			ForceNew: true,
		},
		"project_id": {
			Type:     schema.TypeString,
			Required: true,
			ForceNew: true,
		},
		"region": {
			Type:     schema.TypeString,
			Required: true,
			ForceNew: true,
		},
		"datastore_id": {
			Type:     schema.TypeString,
			Required: true,
			ForceNew: true,
		},
		"status": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"owner_id": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"lc_collate": {
			Type:             schema.TypeString,
			Optional:         true,
			ForceNew:         true,
			DiffSuppressFunc: dbaasDatabaseV1LocaleDiffSuppressFunc,
		},
		"lc_ctype": {
			Type:             schema.TypeString,
			Optional:         true,
			ForceNew:         true,
			DiffSuppressFunc: dbaasDatabaseV1LocaleDiffSuppressFunc,
		},
	}
}

func dbaasDatabaseV1LocaleDiffSuppressFunc(k, old, new string, d *schema.ResourceData) bool {
	// The default locale value - C is the same as null value, so we need to suppress
	if old == "C" && new == "" {
		return true
	}

	return false
}
