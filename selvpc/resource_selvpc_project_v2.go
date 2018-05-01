package selvpc

import (
	"context"
	"fmt"
	"log"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/selectel/go-selvpcclient/selvpcclient/resell/v2/projects"
	"github.com/selectel/go-selvpcclient/selvpcclient/resell/v2/quotas"
)

func resourceResellProjectV2() *schema.Resource {
	return &schema.Resource{
		Create: resourceResellProjectV2Create,
		Read:   resourceResellProjectV2Read,
		Update: resourceResellProjectV2Update,
		Delete: resourceResellProjectV2Delete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: false,
			},
			"url": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"enabled": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"custom_url": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: false,
			},
			"theme": {
				Type:     schema.TypeMap,
				Optional: true,
				ForceNew: false,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"color": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: false,
						},
						"logo": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: false,
						},
					},
				},
			},
			"quotas": {
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: false,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"resource_name": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: false,
						},
						"resource_quotas": {
							Type:     schema.TypeList,
							Required: true,
							ForceNew: false,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"region": {
										Type:     schema.TypeString,
										Optional: true,
										ForceNew: false,
									},
									"zone": {
										Type:     schema.TypeString,
										Optional: true,
										ForceNew: false,
									},
									"value": {
										Type:     schema.TypeInt,
										Required: true,
										ForceNew: false,
									},
								},
							},
						},
					},
				},
			},
			"all_quotas": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"resource_name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"resource_quotas": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"region": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"zone": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"value": {
										Type:     schema.TypeInt,
										Computed: true,
									},
									"used": {
										Type:     schema.TypeInt,
										Computed: true,
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func resourceResellProjectV2Create(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	resellV2Client := config.resellV2Client()
	ctx := context.Background()

	var opts projects.CreateOpts
	quotasList := d.Get("quotas").([]interface{})
	if len(quotasList) != 0 {
		quotasOpts, err := resourceResellProjectV2QuotasOptsFromList(quotasList)
		if err != nil {
			return fmt.Errorf(errParseProjectV2QuotasFmt, err)
		}
		opts.Quotas = quotasOpts
	}
	opts.Name = d.Get("name").(string)

	log.Printf("[DEBUG] Creating project with options: %v\n", opts)
	project, _, err := projects.Create(ctx, resellV2Client, opts)
	if err != nil {
		return err
	}

	d.SetId(project.ID)

	return resourceResellProjectV2Read(d, meta)
}

func resourceResellProjectV2Read(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	resellV2Client := config.resellV2Client()
	ctx := context.Background()

	log.Printf("[DEBUG] Getting project %s\n", d.Id())
	project, _, err := projects.Get(ctx, resellV2Client, d.Id())
	if err != nil {
		return err
	}

	d.Set("name", project.Name)
	d.Set("url", project.URL)
	d.Set("enabled", project.Enabled)
	d.Set("custom_url", project.CustomURL)
	d.Set("theme", project.Theme)

	// Set all quotas.
	// This can be different from what the user specified since the project
	// will have all available resource quotas automatically applied.
	projectQuotas := resourceResellProjectV2QuotasToList(project.Quotas)
	d.Set("all_quotas", projectQuotas)

	return nil
}

func resourceResellProjectV2Update(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	resellV2Client := config.resellV2Client()
	ctx := context.Background()

	var hasChange, projectChange, quotaChange bool
	var projectOpts projects.UpdateOpts
	var projectQuotasOpts quotas.UpdateProjectQuotasOpts

	if d.HasChange("name") {
		hasChange, projectChange = true, true
		projectOpts.Name = d.Get("name").(string)
	}
	if d.HasChange("custom_url") {
		hasChange, projectChange = true, true
		customURL := d.Get("custom_url").(string)
		projectOpts.CustomURL = &customURL
	}
	if d.HasChange("theme") {
		hasChange, projectChange = true, true
		themeMap := d.Get("theme").(map[string]interface{})
		updateThemeOpts := resourceProjectV2UpdateThemeOptsFromMap(themeMap)
		projectOpts.Theme = updateThemeOpts
	}
	if d.HasChange("quotas") {
		hasChange, quotaChange = true, true
		quotasList := d.Get("quotas").([]interface{})
		quotasOpts, err := resourceResellProjectV2QuotasOptsFromList(quotasList)
		if err != nil {
			return fmt.Errorf(errParseProjectV2QuotasFmt, err)
		}
		projectQuotasOpts.QuotasOpts = quotasOpts
	}

	if hasChange {
		// Update project options if needed.
		if projectChange {
			log.Printf("[DEBUG] Updating project %s with options: %v\n", d.Id(), projectOpts)
			_, _, err := projects.Update(ctx, resellV2Client, d.Id(), projectOpts)
			if err != nil {
				return err
			}
		}
		// Update project quotas if needed.
		if quotaChange {
			log.Printf("[DEBUG] Updating project %s quotas with options: %v\n", d.Id(), projectQuotasOpts)
			_, _, err := quotas.UpdateProjectQuotas(ctx, resellV2Client, d.Id(), projectQuotasOpts)
			if err != nil {
				return err
			}
		}
	}

	return resourceResellProjectV2Read(d, meta)
}

func resourceResellProjectV2Delete(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	resellV2Client := config.resellV2Client()
	ctx := context.Background()

	log.Printf("[DEBUG] Deleting project %s\n", d.Id())
	_, err := projects.Delete(ctx, resellV2Client, d.Id())
	if err != nil {
		return err
	}

	return nil
}

// resourceResellProjectV2QuotasOptsFromList converts the provided quotasList to
// the slice of quotas.QuotaOpts. It then can be used to make requests with quotas data.
func resourceResellProjectV2QuotasOptsFromList(quotasList []interface{}) ([]quotas.QuotaOpts, error) {
	if len(quotasList) == 0 {
		return nil, fmt.Errorf("got empty quotas list")
	}

	// Pre-populate quotasOpts slice in memory as we already know it's length.
	quotasOpts := make([]quotas.QuotaOpts, len(quotasList))

	// Iterate over each billing resource quotas map.
	for i, resourceQuotasData := range quotasList {
		var resourceNameRaw, resourceQuotasRaw interface{}
		var ok bool

		// Cast type of the current resource quotas map and check provided values.
		resourceQuotasMap := resourceQuotasData.(map[string]interface{})
		if resourceNameRaw, ok = resourceQuotasMap["resource_name"]; !ok {
			return nil, fmt.Errorf("resource_name value isn't provided")
		}
		if resourceQuotasRaw, ok = resourceQuotasMap["resource_quotas"]; !ok {
			return nil, fmt.Errorf("resource_quotas value isn't provided")
		}

		// Cast types of provided values and pre-populate slice of []quotas.ResourceQuotaOpts
		// in memory as we already know it's length.
		resourceName := resourceNameRaw.(string)
		resourceQuotasEntities := resourceQuotasRaw.([]interface{})
		resourceQuotasOpts := make([]quotas.ResourceQuotaOpts, len(resourceQuotasEntities))

		// Populate every quotas.ResourceQuotaOpts with data from a single
		// resourceQuotasMap's region zone and value.
		for j, resourceQuotasEntityRaw := range resourceQuotasEntities {
			var resourceQuotasEntityRegion, resourceQuotasEntityZone string
			var resourceQuotasEntityValue int
			resourceQuotasEntity := resourceQuotasEntityRaw.(map[string]interface{})
			if region, ok := resourceQuotasEntity["region"]; ok {
				resourceQuotasEntityRegion = region.(string)
			}
			if zone, ok := resourceQuotasEntity["zone"]; ok {
				resourceQuotasEntityZone = zone.(string)
			}
			if value, ok := resourceQuotasEntity["value"]; ok {
				resourceQuotasEntityValue = value.(int)
			}
			// Populate single entity of billing resource data with the region,
			// zone and value information.
			resourceQuotasOpts[j] = quotas.ResourceQuotaOpts{
				Region: resourceQuotasEntityRegion,
				Zone:   resourceQuotasEntityZone,
				Value:  &resourceQuotasEntityValue,
			}
		}

		// Populate single quota options element.
		quotasOpts[i] = quotas.QuotaOpts{
			Name:               resourceName,
			ResourceQuotasOpts: resourceQuotasOpts,
		}
	}

	return quotasOpts, nil
}

// resourceResellProjectV2QuotasToMap converts the provided quotas.Quota slice
// to a nested complex map structure correspondingly to the resource's schema.
func resourceResellProjectV2QuotasToMap(quotasStructures []quotas.Quota) map[string]interface{} {
	quotasMap := make(map[string]interface{}, len(quotasStructures))

	if quotasStructures != nil {
		// Iterate over each billing resource quota.
		for _, quota := range quotasStructures {
			// For each billing resource populate corresponding slice of maps that
			// contain quota data (region, zone and value).
			resourceQuotasEntities := make([]map[string]interface{}, len(quota.ResourceQuotasEntities))
			for i, resourceQuotasEntity := range quota.ResourceQuotasEntities {
				resourceQuotasEntities[i] = map[string]interface{}{
					"region": resourceQuotasEntity.Region,
					"zone":   resourceQuotasEntity.Zone,
					"value":  resourceQuotasEntity.Value,
					"used":   resourceQuotasEntity.Used,
				}
			}

			quotasMap[quota.Name] = resourceQuotasEntities
		}
	}

	return quotasMap
}

// resourceResellProjectV2QuotasToMap converts the provided quotas.Quota slice
// to a nested complex list structure correspondingly to the resource's schema.
func resourceResellProjectV2QuotasToList(quotasStructures []quotas.Quota) []interface{} {
	quotasList := make([]interface{}, len(quotasStructures))

	if quotasStructures != nil {
		// Iterate over each billing resource quota.
		for i, quota := range quotasStructures {
			// For each billing resource populate corresponding slice of maps that
			// contain quota data (region, zone and value).
			resourceQuotasEntities := make([]map[string]interface{}, len(quota.ResourceQuotasEntities))
			for i, resourceQuotasEntity := range quota.ResourceQuotasEntities {
				resourceQuotasEntities[i] = map[string]interface{}{
					"region": resourceQuotasEntity.Region,
					"zone":   resourceQuotasEntity.Zone,
					"value":  resourceQuotasEntity.Value,
					"used":   resourceQuotasEntity.Used,
				}
			}

			// Populate single quota element.
			quotasList[i] = map[string]interface{}{
				"resource_name":   quota.Name,
				"resource_quotas": resourceQuotasEntities,
			}
		}
	}

	return quotasList
}

// resourceProjectV2UpdateThemeOptsFromMap converts the provided themeOptsMap to
// the *project.ThemeUpdateOpts. It then can be used to make requests with theme data.
func resourceProjectV2UpdateThemeOptsFromMap(themeOptsMap map[string]interface{}) *projects.ThemeUpdateOpts {
	themeUpdateOpts := &projects.ThemeUpdateOpts{}

	var themeColor, themeLogo string
	if color, ok := themeOptsMap["color"]; ok {
		themeColor = color.(string)
	}
	if logo, ok := themeOptsMap["logo"]; ok {
		themeLogo = logo.(string)
	}
	themeUpdateOpts.Color = &themeColor
	themeUpdateOpts.Logo = &themeLogo

	return themeUpdateOpts
}
