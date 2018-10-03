package selvpc

import (
	"context"
	"errors"
	"log"
	"strconv"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/selectel/go-selvpcclient/selvpcclient/resell/v2/licenses"
)

func resourceResellLicenseV2() *schema.Resource {
	return &schema.Resource{
		Create: resourceResellLicenseV2Create,
		Read:   resourceResellLicenseV2Read,
		Delete: resourceResellLicenseV2Delete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
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
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"type": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"servers": {
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
						"status": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func resourceResellLicenseV2Create(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	resellV2Client := config.resellV2Client()
	ctx := context.Background()

	projectID := d.Get("project_id").(string)
	region := d.Get("region").(string)
	licenseType := d.Get("type").(string)
	opts := licenses.LicenseOpts{
		Licenses: []licenses.LicenseOpt{
			{
				Region:   region,
				Type:     licenseType,
				Quantity: 1,
			},
		},
	}

	log.Printf("[DEBUG] Creating license with options: %v\n", opts)
	newLicenses, _, err := licenses.Create(ctx, resellV2Client, projectID, opts)
	if err != nil {
		return err
	}

	if len(newLicenses) != 1 {
		return errors.New("can't get license from the response")
	}

	d.SetId(strconv.Itoa(newLicenses[0].ID))

	return resourceResellLicenseV2Read(d, meta)
}

func resourceResellLicenseV2Read(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	resellV2Client := config.resellV2Client()
	ctx := context.Background()

	log.Printf("[DEBUG] Getting license %s", d.Id())
	license, _, err := licenses.Get(ctx, resellV2Client, d.Id())
	if err != nil {
		return err
	}

	d.Set("project_id", license.ProjectID)
	d.Set("region", license.Region)
	d.Set("status", license.Status)
	d.Set("type", license.Type)
	associatedServers := serversMapsFromStructs(license.Servers)
	d.Set("servers", associatedServers)

	return nil
}

func resourceResellLicenseV2Delete(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	resellV2Client := config.resellV2Client()
	ctx := context.Background()

	log.Printf("[DEBUG] Deleting license %s\n", d.Id())
	_, err := licenses.Delete(ctx, resellV2Client, d.Id())
	if err != nil {
		return err
	}

	return nil
}
