package selectel

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceDBaaSUserV1Schema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"datastore_id": {
			Type:     schema.TypeString,
			Required: true,
			ForceNew: true,
		},
		"region": {
			Type:     schema.TypeString,
			Required: true,
			ForceNew: true,
		},
		"name": {
			Type:     schema.TypeString,
			Required: true,
			ForceNew: true,
		},
		"password": {
			Type:      schema.TypeString,
			Required:  true,
			Sensitive: true,
		},
		"status": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"project_id": {
			Type:     schema.TypeString,
			Required: true,
			ForceNew: true,
		},
		"settings": {
			Type:     schema.TypeList,
			Optional: true,
			MaxItems: 1,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"conn_limit": {
						Type:     schema.TypeInt,
						Optional: true,
					},
					"login": {
						Type:     schema.TypeBool,
						Optional: true,
					},
					"statement_timeout": {
						Type:     schema.TypeInt,
						Optional: true,
					},
					"lock_timeout": {
						Type:     schema.TypeInt,
						Optional: true,
					},
					"idle_in_transaction_session_timeout": {
						Type:     schema.TypeInt,
						Optional: true,
					},
					"idle_session_timeout": {
						Type:     schema.TypeInt,
						Optional: true,
					},
					"temp_file_limit": {
						Type:     schema.TypeInt,
						Optional: true,
					},
					"work_mem": {
						Type:     schema.TypeInt,
						Optional: true,
					},
					"log_min_duration_statement": {
						Type:     schema.TypeInt,
						Optional: true,
					},
					"search_path": {
						Type:     schema.TypeString,
						Optional: true,
					},
					"default_transaction_isolation": {
						Type:     schema.TypeString,
						Optional: true,
						ValidateFunc: validation.StringInSlice([]string{
							"serializable",
							"repeatable read",
							"read committed",
							"read uncommitted",
						}, false),
					},
					"log_statement": {
						Type:     schema.TypeString,
						Optional: true,
						ValidateFunc: validation.StringInSlice([]string{
							"none",
							"ddl",
							"mod",
							"all",
						}, false),
					},
					"synchronous_commit": {
						Type:     schema.TypeString,
						Optional: true,
						ValidateFunc: validation.StringInSlice([]string{
							"on",
							"off",
							"local",
							"remote_write",
							"remote_apply",
						}, false),
					},
					"default_transaction_read_only": {
						Type:     schema.TypeBool,
						Optional: true,
					},
				},
			},
		},
	}
}

var dbaasUserSettingsKeys = []string{
	"conn_limit",
	"login",
	"statement_timeout",
	"lock_timeout",
	"idle_in_transaction_session_timeout",
	"idle_session_timeout",
	"temp_file_limit",
	"work_mem",
	"log_min_duration_statement",
	"search_path",
	"default_transaction_isolation",
	"log_statement",
	"synchronous_commit",
	"default_transaction_read_only",
}
