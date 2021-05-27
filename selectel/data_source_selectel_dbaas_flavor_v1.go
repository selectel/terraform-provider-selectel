package selectel

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/selectel/dbaas-go"
)

func dataSourceDBaaSFlavorV1() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceDBaaSFlavorV1Read,
		Schema: map[string]*schema.Schema{
			"project_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"region": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.StringInSlice([]string{
					ru1Region,
					ru2Region,
					ru3Region,
					ru7Region,
					ru8Region,
					ru9Region,
				}, false),
			},
			"flavors": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"description": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"vcpus": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"ram": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"disk": {
							Type:     schema.TypeInt,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func dataSourceDBaaSFlavorV1Read(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	dbaasClient, diagErr := getDBaaSClient(ctx, d, meta)
	if diagErr != nil {
		return diagErr
	}

	flavors, err := dbaasClient.Flavors(ctx)
	if err != nil {
		return diag.FromErr(errGettingObjects(objectFlavors, err))
	}

	flavorIDs := []string{}
	for _, flavor := range flavors {
		flavorIDs = append(flavorIDs, flavor.ID)
	}

	flavorsFlatten := flattenDBaaSFlavors(flavors)
	if err := d.Set("flavors", flavorsFlatten); err != nil {
		return diag.FromErr(err)
	}
	checksum, err := stringListChecksum(flavorIDs)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(checksum)

	return nil
}

func flattenDBaaSFlavors(flavors []dbaas.FlavorResponse) []interface{} {
	flavorsList := make([]interface{}, len(flavors))
	for i, flavor := range flavors {
		flavorMap := make(map[string]interface{})
		flavorMap["id"] = flavor.ID
		flavorMap["name"] = flavor.Name
		flavorMap["description"] = flavor.Description
		flavorMap["vcpus"] = flavor.Vcpus
		flavorMap["ram"] = flavor.RAM
		flavorMap["disk"] = flavor.Disk

		flavorsList[i] = flavorMap
	}

	return flavorsList
}
