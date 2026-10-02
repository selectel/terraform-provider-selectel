package selectel

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/selectel/dbaas-go"
)

type rolesSearchFilter struct {
	datastoreTypeID string
	name            string
}

func dataSourceDBaaSRolesV1() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceDBaaSRolesV1Read,
		Schema: map[string]*schema.Schema{
			"project_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"region": {
				Type:     schema.TypeString,
				Required: true,
			},
			"filter": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"datastore_type_id": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"name": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
			"roles": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"datastore_type_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"name": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func dataSourceDBaaSRolesV1Read(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	dbaasClient, diagErr := getDBaaSClient(d, meta)
	if diagErr != nil {
		return diagErr
	}

	roles, err := dbaasClient.Roles(ctx)
	if err != nil {
		return diag.FromErr(errGettingObjects(objectRoles, err))
	}

	rolesIDs := make([]string, 0, len(roles))
	for _, param := range roles {
		rolesIDs = append(rolesIDs, param.ID)
	}

	filter, err := expandRolesSearchFilter(d.Get("filter").(*schema.Set))
	if err != nil {
		return diag.FromErr(err)
	}

	roles = filterRolesByDatastoreTypeID(roles, filter.datastoreTypeID)
	roles = filterRolesByName(roles, filter.name)

	rolesFlatter := flattenDBaaSRoles(roles)
	if err := d.Set("roles", rolesFlatter); err != nil {
		return diag.FromErr(err)
	}
	checksum, err := stringListChecksum(rolesIDs)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(checksum)

	return nil
}

func expandRolesSearchFilter(filterSet *schema.Set) (rolesSearchFilter, error) {
	filter := rolesSearchFilter{}
	if filterSet.Len() == 0 {
		return filter, nil
	}

	resourceFilterMap := filterSet.List()[0].(map[string]any)

	datastoreTypeID, ok := resourceFilterMap["datastore_type_id"]
	if ok {
		filter.datastoreTypeID = datastoreTypeID.(string)
	}

	name, ok := resourceFilterMap["name"]
	if ok {
		filter.name = name.(string)
	}

	return filter, nil
}

func filterRolesByDatastoreTypeID(roles []dbaas.Roles, datastoreTypeID string) []dbaas.Roles {
	if datastoreTypeID == "" {
		return roles
	}

	var filteredRoles []dbaas.Roles
	for _, param := range roles {
		if param.DatastoreTypeID == datastoreTypeID {
			filteredRoles = append(filteredRoles, param)
		}
	}

	return filteredRoles
}

func filterRolesByName(roles []dbaas.Roles, name string) []dbaas.Roles {
	if name == "" {
		return roles
	}

	var filteredRoles []dbaas.Roles
	for _, param := range roles {
		if param.Name == name {
			filteredRoles = append(filteredRoles, param)
		}
	}

	return filteredRoles
}

func flattenDBaaSRoles(roles []dbaas.Roles) []any {
	rolesList := make([]any, len(roles))
	for i, param := range roles {
		rolesMap := make(map[string]any)
		rolesMap["id"] = param.ID
		rolesMap["datastore_type_id"] = param.DatastoreTypeID
		rolesMap["name"] = param.Name

		rolesList[i] = rolesMap
	}

	return rolesList
}
