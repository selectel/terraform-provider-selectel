package selectel

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/selectel/dbaas-go"
)

type flavorSearchFilter struct {
	vcpus           int
	ram             int
	disk            int
	flSize          string
	datastoreTypeID string
}

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
						"fl_size": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"datastore_type_ids": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
					},
				},
			},
			"filter": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"vcpus": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"ram": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"disk": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"fl_size": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"datastore_type_id": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
		},
	}
}

func dataSourceDBaaSFlavorV1Read(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	dbaasClient, diagErr := getDBaaSClient(d, meta)
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

	filter, err := expandFlavorSearchFilter(d.Get("filter").(*schema.Set))
	if err != nil {
		return diag.FromErr(err)
	}

	flavors = filterFlavorByVcpus(flavors, filter.vcpus)
	flavors = filterFlavorByRAM(flavors, filter.ram)
	flavors = filterFlavorByDisk(flavors, filter.disk)
	flavors = filterFlavorByFlSize(flavors, filter.flSize)
	flavors = filterFlavorByDatastoreTypeID(flavors, filter.datastoreTypeID)

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

func expandFlavorSearchFilter(filterSet *schema.Set) (flavorSearchFilter, error) {
	filter := flavorSearchFilter{}
	if filterSet.Len() == 0 {
		return filter, nil
	}

	resourceFilterMap := filterSet.List()[0].(map[string]interface{})

	vcpus, ok := resourceFilterMap["vcpus"]
	if ok {
		filter.vcpus = vcpus.(int)
	}

	ram, ok := resourceFilterMap["ram"]
	if ok {
		filter.ram = ram.(int)
	}

	disk, ok := resourceFilterMap["disk"]
	if ok {
		filter.disk = disk.(int)
	}

	flSize, ok := resourceFilterMap["fl_size"]
	if ok {
		filter.flSize = flSize.(string)
	}

	datastoreTypeID, ok := resourceFilterMap["datastore_type_id"]
	if ok {
		filter.datastoreTypeID = datastoreTypeID.(string)
	}

	return filter, nil
}

func filterFlavorByVcpus(flavors []dbaas.FlavorResponse, vcpus int) []dbaas.FlavorResponse {
	if vcpus == 0 {
		return flavors
	}

	var filteredFlavors []dbaas.FlavorResponse
	for _, f := range flavors {
		if f.Vcpus == vcpus {
			filteredFlavors = append(filteredFlavors, f)
		}
	}

	return filteredFlavors
}

func filterFlavorByRAM(flavors []dbaas.FlavorResponse, ram int) []dbaas.FlavorResponse {
	if ram == 0 {
		return flavors
	}

	var filteredFlavors []dbaas.FlavorResponse
	for _, f := range flavors {
		if f.RAM == ram {
			filteredFlavors = append(filteredFlavors, f)
		}
	}

	return filteredFlavors
}

func filterFlavorByDisk(flavors []dbaas.FlavorResponse, disk int) []dbaas.FlavorResponse {
	if disk == 0 {
		return flavors
	}

	var filteredFlavors []dbaas.FlavorResponse
	for _, f := range flavors {
		if f.Disk == disk {
			filteredFlavors = append(filteredFlavors, f)
		}
	}

	return filteredFlavors
}

func filterFlavorByFlSize(flavors []dbaas.FlavorResponse, flSize string) []dbaas.FlavorResponse {
	if flSize == "" {
		return flavors
	}

	var filteredFlavors []dbaas.FlavorResponse
	for _, f := range flavors {
		if f.FlSize == flSize {
			filteredFlavors = append(filteredFlavors, f)
		}
	}

	return filteredFlavors
}

func filterFlavorByDatastoreTypeID(flavors []dbaas.FlavorResponse, datastoreTypeID string) []dbaas.FlavorResponse {
	if datastoreTypeID == "" {
		return flavors
	}

	var filteredFlavors []dbaas.FlavorResponse
	for _, f := range flavors {
		for _, flavorDatastoreTypeID := range f.DatastoreTypeIDs {
			if flavorDatastoreTypeID == datastoreTypeID {
				filteredFlavors = append(filteredFlavors, f)
			}
		}
	}

	return filteredFlavors
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
		flavorMap["fl_size"] = flavor.FlSize
		flavorMap["datastore_type_ids"] = flavor.DatastoreTypeIDs

		flavorsList[i] = flavorMap
	}

	return flavorsList
}
