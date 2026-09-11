package selectel

import (
	"context"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	dbaas_v2 "github.com/selectel/dbaas-go/v2/common"
)

type flavorSearchFilterV2 struct {
	vcpus           int
	ram             int
	disk            int
	flSize          string
	datastoreTypeID string
	allowedRole     string
}

func dataSourceDBaaSV2Flavor() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceDBaaSV2FlavorRead,
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
						"allowed_roles": {
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
						"allowed_role": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
		},
	}
}

func dataSourceDBaaSV2FlavorRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	dbaasClient, diagErr := getDBaaSV2Client(d, meta)
	if diagErr != nil {
		return diagErr
	}

	response, err := dbaasClient.Flavor.GetFlavorList(ctx)
	if err != nil {
		return diag.FromErr(errGettingObjects(objectFlavors, err))
	}

	if response.Errors != "" {
		log.Printf("[WARN] flavors got with error: %s", response.Errors)

	}

	flavors := response.Flavors

	flavorIDs := []string{}
	for _, flavor := range flavors {
		flavorIDs = append(flavorIDs, flavor.ID)
	}

	filter, err := expandFlavorSearchFilterV2(d.Get("filter").(*schema.Set))
	if err != nil {
		return diag.FromErr(err)
	}

	flavors = filterDBaaSV2FlavorByVcpus(flavors, filter.vcpus)
	flavors = filterDBaaSV2FlavorByRAM(flavors, filter.ram)
	flavors = filterDBaaSV2FlavorByDisk(flavors, filter.disk)
	flavors = filterDBaaSV2FlavorByFlSize(flavors, filter.flSize)
	flavors = filterDBaaSV2FlavorByDatastoreTypeID(flavors, filter.datastoreTypeID)
	flavors = filterDBaaSV2FlavorByAllowedRole(flavors, filter.allowedRole)

	flavorsFlatten := flattenDBaaSV2Flavors(flavors)
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

func expandFlavorSearchFilterV2(filterSet *schema.Set) (flavorSearchFilterV2, error) {
	filter := flavorSearchFilterV2{}
	if filterSet.Len() == 0 {
		return filter, nil
	}

	resourceFilterMap := filterSet.List()[0].(map[string]any)

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

	allowedRole, ok := resourceFilterMap["allowed_role"]
	if ok {
		filter.allowedRole = allowedRole.(string)
	}

	return filter, nil
}

func filterDBaaSV2FlavorByVcpus(flavors []dbaas_v2.FlavorResponse, vcpus int) []dbaas_v2.FlavorResponse {
	if vcpus == 0 {
		return flavors
	}

	var filteredFlavors []dbaas_v2.FlavorResponse
	for _, f := range flavors {
		if f.VCPUs == vcpus {
			filteredFlavors = append(filteredFlavors, f)
		}
	}

	return filteredFlavors
}

func filterDBaaSV2FlavorByRAM(flavors []dbaas_v2.FlavorResponse, ram int) []dbaas_v2.FlavorResponse {
	if ram == 0 {
		return flavors
	}

	var filteredFlavors []dbaas_v2.FlavorResponse
	for _, f := range flavors {
		if f.RAM == ram {
			filteredFlavors = append(filteredFlavors, f)
		}
	}

	return filteredFlavors
}

func filterDBaaSV2FlavorByDisk(flavors []dbaas_v2.FlavorResponse, disk int) []dbaas_v2.FlavorResponse {
	if disk == 0 {
		return flavors
	}

	var filteredFlavors []dbaas_v2.FlavorResponse
	for _, f := range flavors {
		if f.Disk == disk {
			filteredFlavors = append(filteredFlavors, f)
		}
	}

	return filteredFlavors
}

func filterDBaaSV2FlavorByFlSize(flavors []dbaas_v2.FlavorResponse, flSize string) []dbaas_v2.FlavorResponse {
	if flSize == "" {
		return flavors
	}

	var filteredFlavors []dbaas_v2.FlavorResponse
	for _, f := range flavors {
		if f.FlSize == flSize {
			filteredFlavors = append(filteredFlavors, f)
		}
	}

	return filteredFlavors
}

func filterDBaaSV2FlavorByDatastoreTypeID(flavors []dbaas_v2.FlavorResponse, datastoreTypeID string) []dbaas_v2.FlavorResponse {
	if datastoreTypeID == "" {
		return flavors
	}

	var filteredFlavors []dbaas_v2.FlavorResponse
	for _, f := range flavors {
		for _, flavorDatastoreTypeID := range f.DatastoreTypeIDs {
			if flavorDatastoreTypeID == datastoreTypeID {
				filteredFlavors = append(filteredFlavors, f)
			}
		}
	}

	return filteredFlavors
}

func filterDBaaSV2FlavorByAllowedRole(flavors []dbaas_v2.FlavorResponse, allowedRole string) []dbaas_v2.FlavorResponse {
	if allowedRole == "" {
		return flavors
	}

	var filteredFlavors []dbaas_v2.FlavorResponse
	for _, f := range flavors {
		for _, flavorAllowedRole := range f.AllowedRoles {
			if flavorAllowedRole == allowedRole {
				filteredFlavors = append(filteredFlavors, f)
			}
		}
	}

	return filteredFlavors
}

func flattenDBaaSV2Flavors(flavors []dbaas_v2.FlavorResponse) []any {
	flavorsList := make([]any, len(flavors))
	for i, flavor := range flavors {
		flavorMap := make(map[string]any)
		flavorMap["id"] = flavor.ID
		flavorMap["vcpus"] = flavor.VCPUs
		flavorMap["ram"] = flavor.RAM
		flavorMap["disk"] = flavor.Disk
		flavorMap["fl_size"] = flavor.FlSize
		flavorMap["datastore_type_ids"] = flavor.DatastoreTypeIDs
		flavorMap["allowed_roles"] = flavor.AllowedRoles

		flavorsList[i] = flavorMap
	}

	return flavorsList
}
