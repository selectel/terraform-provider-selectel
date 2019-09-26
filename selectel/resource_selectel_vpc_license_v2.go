package selectel

import (
	"context"
	"log"
	"net/http"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/selectel/go-selvpcclient/selvpcclient/resell/v2/licenses"
)

func resourceVPCLicenseV2() *schema.Resource {
	return &schema.Resource{
		Create: resourceVPCLicenseV2Create,
		Read:   resourceVPCLicenseV2Read,
		Delete: resourceVPCLicenseV2Delete,
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
			"type": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"servers": {
				Type:     schema.TypeSet,
				Computed: true,
				Set:      hashServers,
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

func resourceVPCLicenseV2Create(d *schema.ResourceData, meta interface{}) error {
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

	log.Print(msgCreate(objectLicense, opts))
	newLicenses, _, err := licenses.Create(ctx, resellV2Client, projectID, opts)
	if err != nil {
		return errCreatingObject(objectLicense, err)
	}

	if len(newLicenses) != 1 {
		return errReadFromResponse(objectLicense)
	}

	d.SetId(strconv.Itoa(newLicenses[0].ID))

	return resourceVPCLicenseV2Read(d, meta)
}

func resourceVPCLicenseV2Read(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	resellV2Client := config.resellV2Client()
	ctx := context.Background()

	log.Print(msgGet(objectLicense, d.Id()))
	license, response, err := licenses.Get(ctx, resellV2Client, d.Id())
	if err != nil {
		if response.StatusCode == http.StatusNotFound {
			d.SetId("")
			return nil
		}

		return errGettingObject(objectLicense, d.Id(), err)
	}

	d.Set("project_id", license.ProjectID)
	d.Set("region", license.Region)
	d.Set("status", license.Status)
	d.Set("type", license.Type)
	associatedServers := serversMapsFromStructs(license.Servers)
	if err := d.Set("servers", associatedServers); err != nil {
		log.Print(errSettingComplexAttr("servers", err))
	}

	return nil
}

func resourceVPCLicenseV2Delete(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	resellV2Client := config.resellV2Client()
	ctx := context.Background()

	log.Print(msgDelete(objectLicense, d.Id()))
	response, err := licenses.Delete(ctx, resellV2Client, d.Id())
	if err != nil {
		if response.StatusCode == http.StatusNotFound {
			d.SetId("")
			return nil
		}

		return errDeletingObject(objectLicense, d.Id(), err)
	}

	return nil
}
