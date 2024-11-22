package selectel

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/selectel/go-selvpcclient/v4/selvpcclient/clients"
	"github.com/selectel/go-selvpcclient/v4/selvpcclient/resell/v2/licenses"
)

func resourceVPCLicenseV2() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceVPCLicenseV2Create,
		ReadContext:   resourceVPCLicenseV2Read,
		DeleteContext: resourceVPCLicenseV2Delete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceVPCLicenseV2ImportState,
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
			"network_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"subnet_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"port_id": {
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

func resourceVPCLicenseV2Create(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*Config)
	projectID := d.Get("project_id").(string)
	selvpcClient, err := config.GetSelVPCClientWithProjectScope(projectID)
	if err != nil {
		return diag.FromErr(fmt.Errorf("can't get selvpc client for license: %w", err))
	}

	region := d.Get("region").(string)
	err = validateRegion(selvpcClient, clients.ResellServiceType, region)
	if err != nil {
		return diag.FromErr(fmt.Errorf("can't validate region: %w", err))
	}

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
	newLicenses, _, err := licenses.Create(selvpcClient, projectID, opts)
	if err != nil {
		return diag.FromErr(errCreatingObject(objectLicense, err))
	}

	if len(newLicenses) != 1 {
		return diag.FromErr(errReadFromResponse(objectLicense))
	}

	d.SetId(strconv.Itoa(newLicenses[0].ID))

	return resourceVPCLicenseV2Read(ctx, d, meta)
}

func resourceVPCLicenseV2Read(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*Config)
	projectID := d.Get("project_id").(string)
	selvpcClient, err := config.GetSelVPCClientWithProjectScope(projectID)
	if err != nil {
		return diag.FromErr(fmt.Errorf("can't get selvpc client for license: %w", err))
	}

	log.Print(msgGet(objectLicense, d.Id()))
	license, response, err := licenses.Get(selvpcClient, d.Id())
	if err != nil {
		if response != nil {
			if response.StatusCode == http.StatusNotFound {
				d.SetId("")
				return nil
			}
		}

		return diag.FromErr(errGettingObject(objectLicense, d.Id(), err))
	}

	d.Set("project_id", license.ProjectID)
	d.Set("region", license.Region)
	d.Set("status", license.Status)
	d.Set("network_id", license.NetworkID)
	d.Set("subnet_id", license.SubnetID)
	d.Set("port_id", license.PortID)
	d.Set("type", license.Type)
	associatedServers := serversMapsFromStructs(license.Servers)
	if err := d.Set("servers", associatedServers); err != nil {
		log.Print(errSettingComplexAttr("servers", err))
	}

	return nil
}

func resourceVPCLicenseV2Delete(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*Config)
	projectID := d.Get("project_id").(string)
	selvpcClient, err := config.GetSelVPCClientWithProjectScope(projectID)
	if err != nil {
		return diag.FromErr(fmt.Errorf("can't get selvpc client for license: %w", err))
	}

	log.Print(msgDelete(objectLicense, d.Id()))
	response, err := licenses.Delete(selvpcClient, d.Id())
	if err != nil {
		if response != nil {
			if response.StatusCode == http.StatusNotFound {
				d.SetId("")
				return nil
			}
		}

		return diag.FromErr(errDeletingObject(objectLicense, d.Id(), err))
	}

	return nil
}

func resourceVPCLicenseV2ImportState(_ context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	config := meta.(*Config)
	if config.ProjectID == "" {
		return nil, fmt.Errorf("INFRA_PROJECT_ID must be set for the resource import")
	}

	d.Set("project_id", config.ProjectID)

	return []*schema.ResourceData{d}, nil
}
