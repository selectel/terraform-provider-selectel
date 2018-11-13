package selvpc

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log"
	"net/url"

	"github.com/hashicorp/terraform/helper/hashcode"
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
			"auto_quotas": {
				Type:     schema.TypeBool,
				Optional: true,
				ForceNew: false,
			},
			"quotas": {
				Type:     schema.TypeSet,
				Optional: true,
				ForceNew: false,
				Set:      hashQuotas,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"resource_name": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: false,
						},
						"resource_quotas": {
							Type:     schema.TypeSet,
							Required: true,
							ForceNew: false,
							Set:      hashResourceQuotas,
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
				Type:     schema.TypeSet,
				Computed: true,
				Set:      hashQuotas,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"resource_name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"resource_quotas": {
							Type:     schema.TypeSet,
							Computed: true,
							Set:      hashResourceQuotas,
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
	quotaSet := d.Get("quotas").(*schema.Set)
	if quotaSet.Len() != 0 {
		quotasOpts, err := resourceResellProjectV2QuotasOptsFromSet(quotaSet)
		if err != nil {
			return fmt.Errorf(errParseProjectV2QuotasFmt, err)
		}
		opts.Quotas = quotasOpts
	}
	opts.Name = d.Get("name").(string)
	opts.AutoQuotas = d.Get("auto_quotas").(bool)

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

	projectCustomURL, err := resourceResellProjectV2URLWithoutSchema(project.CustomURL)
	if err != nil {
		return err
	}
	d.Set("custom_url", projectCustomURL)
	d.Set("name", project.Name)
	d.Set("url", project.URL)
	d.Set("enabled", project.Enabled)
	if err := d.Set("theme", project.Theme); err != nil {
		log.Printf("[DEBUG] theme: %s", err)
	}

	// Set all quotas. This can be different from what the user specified since
	// the project will have all available resource quotas automatically applied.
	allQuotas := resourceResellProjectV2QuotasToSet(project.Quotas)
	if err := d.Set("all_quotas", allQuotas); err != nil {
		log.Printf("[DEBUG] all_quotas: %s", err)
	}

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
		quotaSet := d.Get("quotas").(*schema.Set)
		quotasOpts, err := resourceResellProjectV2QuotasOptsFromSet(quotaSet)
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

// resourceResellProjectV2QuotasOptsFromSet converts the provided quotaSet to
// the slice of quotas.QuotaOpts. It then can be used to make requests with
// quotas data.
func resourceResellProjectV2QuotasOptsFromSet(quotaSet *schema.Set) ([]quotas.QuotaOpts, error) {
	quotaSetLen := quotaSet.Len()
	if quotaSetLen == 0 {
		return nil, errors.New("got empty quotas")
	}

	// Pre-allocate memory for quotasOpts slice since we already know it's length.
	quotasOpts := make([]quotas.QuotaOpts, quotaSetLen)

	// Iterate over each billing resource quotas map.
	for i, resourceQuotasData := range quotaSet.List() {
		var resourceNameRaw, resourceQuotasRaw interface{}
		var ok bool

		// Cast type of the current resource quotas map and check provided values.
		resourceQuotasMap := resourceQuotasData.(map[string]interface{})
		if resourceNameRaw, ok = resourceQuotasMap["resource_name"]; !ok {
			return nil, errors.New("resource_name value isn't provided")
		}
		if resourceQuotasRaw, ok = resourceQuotasMap["resource_quotas"]; !ok {
			return nil, errors.New("resource_quotas value isn't provided")
		}

		// Cast types of provided values and pre-populate slice of []quotas.ResourceQuotaOpts
		// in memory as we already know it's length.
		resourceName := resourceNameRaw.(string)
		resourceQuotasEntities := resourceQuotasRaw.(*schema.Set)
		resourceQuotasOpts := make([]quotas.ResourceQuotaOpts, resourceQuotasEntities.Len())

		// Populate every quotas.ResourceQuotaOpts with data from a single
		// resourceQuotasMap's region zone and value.
		for j, resourceQuotasEntityRaw := range resourceQuotasEntities.List() {
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

// resourceResellProjectV2QuotasToSet converts the provided quotas.Quota slice
// to a nested complex set structure correspondingly to the resource's schema.
func resourceResellProjectV2QuotasToSet(quotasStructures []quotas.Quota) *schema.Set {
	quotaSet := &schema.Set{
		F: quotasHashSetFunc(),
	}

	// Iterate over each billing resource quota.
	for _, quota := range quotasStructures {
		// For each billing resource populate corresponding resourceQuotasSet that
		// contain quota data (region, zone and value).
		resourceQuotasSet := &schema.Set{
			F: resourceQuotasHashSetFunc(),
		}
		for _, resourceQuotasEntity := range quota.ResourceQuotasEntities {
			resourceQuotasSet.Add(map[string]interface{}{
				"region": resourceQuotasEntity.Region,
				"zone":   resourceQuotasEntity.Zone,
				"value":  resourceQuotasEntity.Value,
				"used":   resourceQuotasEntity.Used,
			})
		}

		// Populate single quota element.
		quotaSet.Add(map[string]interface{}{
			"resource_name":   quota.Name,
			"resource_quotas": resourceQuotasSet,
		})
	}

	return quotaSet
}

// resourceProjectV2UpdateThemeOptsFromMap converts the provided themeOptsMap to
// the *project.ThemeUpdateOpts.
// It can be used to make requests with project theme parameters.
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

// resourceResellProjectV2URLWithoutSchema strips the scheme part from project URL.
func resourceResellProjectV2URLWithoutSchema(customURL string) (string, error) {
	var customURLWithoutSchema string

	if customURL != "" {
		u, err := url.Parse(customURL)
		if err != nil {
			return "", err
		}
		customURLWithoutSchema = u.Hostname()
	}

	return customURLWithoutSchema, nil
}

// quotasSchema returns *schema.Resource from the "quotas" attribute.
func quotasSchema() *schema.Resource {
	return resourceResellProjectV2().Schema["quotas"].Elem.(*schema.Resource)
}

// quotasSchema returns *schema.Resource from the "resource_quotas" attribute.
func resourceQuotasSchema() *schema.Resource {
	return quotasSchema().Schema["resource_quotas"].Elem.(*schema.Resource)
}

// quotasHashSetFunc returns schema.SchemaSetFunc that can be used to
// create a new schema.Set for the "quotas" or "all_quotas" attributes.
func quotasHashSetFunc() schema.SchemaSetFunc {
	return schema.HashResource(quotasSchema())
}

// resourceQuotasHashSetFunc returns schema.SchemaSetFunc that can be used to
// create a new schema.Set for the "resource_quotas" attribute.
func resourceQuotasHashSetFunc() schema.SchemaSetFunc {
	return schema.HashResource(resourceQuotasSchema())
}

// hashQuotas is a hash function to use with the "quotas" and "all_quotas" sets.
func hashQuotas(v interface{}) int {
	m := v.(map[string]interface{})
	return hashcode.String(fmt.Sprintf("%s-", m["resource_name"].(string)))
}

// hashResourceQuotas is a hash function to use with the "resource_quotas" set.
func hashResourceQuotas(v interface{}) int {
	var buf bytes.Buffer
	m := v.(map[string]interface{})
	if m["region"] != "" {
		buf.WriteString(fmt.Sprintf("%s-", m["region"].(string)))
	}
	if m["zone"] != "" {
		buf.WriteString(fmt.Sprintf("%s-", m["zone"].(string)))
	}
	return hashcode.String(buf.String())
}
