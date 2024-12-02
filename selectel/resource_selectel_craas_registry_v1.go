package selectel

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/selectel/craas-go/pkg/v1/registry"
)

func resourceCRaaSRegistryV1() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceCRaaSRegistryV1Create,
		ReadContext:   resourceCRaaSRegistryV1Read,
		DeleteContext: resourceCRaaSRegistryV1Delete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceCRaaSRegistryV1ImportState,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"project_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"endpoint": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceCRaaSRegistryV1Create(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	craasClient, diagErr := getCRaaSClient(d, meta)
	if diagErr != nil {
		return diagErr
	}

	name := d.Get("name").(string)

	log.Print(msgCreate(objectRegistry, name))
	newRegistry, _, err := registry.Create(ctx, craasClient, name)
	if err != nil {
		return diag.FromErr(errCreatingObject(objectRegistry, err))
	}

	log.Printf("[DEBUG] Waiting for registry %s to achieve a stable state", newRegistry.ID)
	timeout := d.Timeout(schema.TimeoutCreate)
	err = waitForCRaaSRegistryV1StableState(ctx, craasClient, newRegistry.ID, timeout)
	if err != nil {
		return diag.FromErr(errCreatingObject(objectRegistry, err))
	}

	d.SetId(newRegistry.ID)

	return resourceCRaaSRegistryV1Read(ctx, d, meta)
}

func resourceCRaaSRegistryV1Read(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	craasClient, diagErr := getCRaaSClient(d, meta)
	if diagErr != nil {
		return diagErr
	}

	craasHostName, err := getHostNameForCRaaS(craasClient.Endpoint)
	if err != nil {
		return diag.FromErr(err)
	}

	log.Print(msgGet(objectRegistry, d.Id()))
	craasRegistry, response, err := registry.Get(ctx, craasClient, d.Id())
	if err != nil {
		if response != nil {
			if response.StatusCode == http.StatusNotFound {
				d.SetId("")
				return nil
			}
		}

		return diag.FromErr(errGettingObject(objectRegistry, d.Id(), err))
	}

	d.Set("name", craasRegistry.Name)
	d.Set("status", craasRegistry.Status)
	d.Set("endpoint", fmt.Sprintf("%s/%s", craasHostName, craasRegistry.Name))

	return nil
}

func resourceCRaaSRegistryV1Delete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	craasClient, diagErr := getCRaaSClient(d, meta)
	if diagErr != nil {
		return diagErr
	}

	log.Print(msgDelete(objectRegistry, d.Id()))
	_, err := registry.Delete(ctx, craasClient, d.Id())
	if err != nil {
		return diag.FromErr(errDeletingObject(objectRegistry, d.Id(), err))
	}

	stateConf := &resource.StateChangeConf{
		Pending: []string{strconv.Itoa(http.StatusOK)},
		Target:  []string{strconv.Itoa(http.StatusNotFound)},
		Timeout: d.Timeout(schema.TimeoutDelete),
		Refresh: func() (result interface{}, state string, err error) {
			result, response, err := registry.Get(ctx, craasClient, d.Id())
			if err != nil {
				if response != nil {
					return result, strconv.Itoa(response.StatusCode), nil
				}

				return nil, "", err
			}

			return result, strconv.Itoa(response.StatusCode), err
		},
		Delay:        1 * time.Second,
		PollInterval: 1 * time.Second,
	}

	log.Printf("[DEBUG] Waiting for registry %s to become deleted", d.Id())
	_, err = stateConf.WaitForStateContext(ctx)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error waiting for registry %s to become deleted: %s", d.Id(), err))
	}

	return nil
}

func resourceCRaaSRegistryV1ImportState(_ context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	config := meta.(*Config)
	if config.ProjectID == "" {
		return nil, errors.New("INFRA_PROJECT_ID must be set for the CRaaS registry resource import")
	}
	d.Set("project_id", config.ProjectID)

	return []*schema.ResourceData{d}, nil
}
