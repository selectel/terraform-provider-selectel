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

	filter, err := expandDBaaSRolesSearchFilter(d.Get("filter").(*schema.Set))
	if err != nil {
		return diag.FromErr(err)
	}

	roles = filterDBaaSRolesByDatastoreTypeID(roles, filter.datastoreTypeID)
	roles = filterDBaaSRolesByName(roles, filter.name)

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

func expandDBaaSRolesSearchFilter(filterSet *schema.Set) (rolesSearchFilter, error) {
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

func filterDBaaSRolesByDatastoreTypeID(roles []dbaas.Role, datastoreTypeID string) []dbaas.Role {
	if datastoreTypeID == "" {
		return roles
	}

	var filteredRoles []dbaas.Role
	for _, param := range roles {
		if param.DatastoreTypeID == datastoreTypeID {
			filteredRoles = append(filteredRoles, param)
		}
	}

	return filteredRoles
}

func filterDBaaSRolesByName(roles []dbaas.Role, name string) []dbaas.Role {
	if name == "" {
		return roles
	}

	var filteredRoles []dbaas.Role
	for _, role := range roles {
		if role.Name == name {
			filteredRoles = append(filteredRoles, role)
		}
	}

	return filteredRoles
}

func flattenDBaaSRoles(roles []dbaas.Role) []any {
	rolesList := make([]any, len(roles))
	for i, role := range roles {
		rolesMap := make(map[string]any)
		rolesMap["id"] = role.ID
		rolesMap["datastore_type_id"] = role.DatastoreTypeID
		rolesMap["name"] = role.Name

		rolesList[i] = rolesMap
	}

	return rolesList
}
