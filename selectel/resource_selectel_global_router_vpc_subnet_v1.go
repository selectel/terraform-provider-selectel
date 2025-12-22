package selectel

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	globalrouter "github.com/selectel/globalrouter-go/pkg/v1"
	waiters "github.com/terraform-providers/terraform-provider-selectel/selectel/waiters/globalrouter"
)

func resourceGlobalRouterVPCSubnetV1() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceGlobalRouterVPCSubnetV1Create,
		ReadContext:   resourceGlobalRouterVPCSubnetV1Read,
		UpdateContext: resourceGlobalRouterVPCSubnetV1Update,
		DeleteContext: resourceGlobalRouterVPCSubnetV1Delete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(5 * time.Minute),
			Update: schema.DefaultTimeout(5 * time.Minute),
			Delete: schema.DefaultTimeout(5 * time.Minute),
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Human-readable name of the subnet",
			},
			"network_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "UUID of the parent vpc network",
				ForceNew:    true,
			},
			"cidr": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Cidr of the subnet in vpc",
				ForceNew:    true,
			},
			"os_subnet_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "UUID of the subnet in the vpc",
				ForceNew:    true,
			},
			"gateway": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Subnet gateway address from specified cidr",
				ForceNew:    true,
			},
			"service_addresses": {
				Type:        schema.TypeSet,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Set:         schema.HashString,
				MaxItems:    2,
				MinItems:    2,
				Description: "List of two ip addresses which will be reserved for internal use",
				ForceNew:    true,
			},
			"tags": {
				Type:        schema.TypeSet,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Set:         schema.HashString,
				Description: "List of the resource tags",
			},
			"project_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "UUID of a project in vpc (taken from the parent network)",
			},
			"created_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Resource creation time",
			},
			"updated_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Resource last update time",
			},
			"status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Resource status",
			},
			"account_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Resource owner account UUID",
			},
			"netops_subnet_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Field for internal usage",
			},
			"sv_subnet_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Field for internal usage",
			},
		},
	}
}

func resourceGlobalRouterVPCSubnetV1Create(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, diagErr := getGlobalRouterClient(meta)
	if diagErr != nil {
		return diagErr
	}

	var createOpts globalrouter.VPCSubnetCreateRequest

	networkID := d.Get("network_id").(string)
	createOpts.NetworkID = networkID
	cidr := d.Get("cidr").(string)
	createOpts.Cidr = cidr
	name := d.Get("name").(string)
	createOpts.Name = name
	osSubnetID := d.Get("os_subnet_id").(string)
	createOpts.OsSubnetID = osSubnetID

	// optional args
	if v, ok := d.GetOk("gateway"); ok {
		gateway := v.(string)
		createOpts.Gateway = gateway
	}

	if v, ok := d.GetOk("service_addresses"); ok {
		serviceAddresses := v.(*schema.Set).List()
		createOpts.ServiceAddresses = expandToStringSlice(serviceAddresses)
	}

	if v, ok := d.GetOk("tags"); ok {
		tags := v.(*schema.Set).List()
		createOpts.Tags = expandToStringSlice(tags)
	}

	createdRouter, _, err := client.VPCSubnetCreate(ctx, &createOpts)
	if err != nil {
		return diag.FromErr(errCreatingObject(objectGlobalRouterVPCSubnet, err))
	}

	d.SetId(createdRouter.ID)

	diagErr = waiters.WaitForSubnetV1ActiveState(ctx, client, d.Id(), d.Timeout(schema.TimeoutCreate))
	if diagErr != nil {
		return diagErr
	}

	return resourceGlobalRouterVPCSubnetV1Read(ctx, d, meta)
}

func resourceGlobalRouterVPCSubnetV1Read(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, diagErr := getGlobalRouterClient(meta)
	if diagErr != nil {
		return diagErr
	}

	res, _, err := client.Subnet(ctx, d.Id())
	if err != nil {
		return diag.FromErr(errGettingObject(objectGlobalRouterVPCSubnet, d.Id(), err))
	}

	if res == nil {
		return diag.FromErr(fmt.Errorf("can't find subnet %q", d.Id()))
	}

	d.Set("name", res.Name)
	d.Set("network_id", res.NetworkID)
	d.Set("cidr", res.Cidr)
	d.Set("os_subnet_id", res.OsSubnetID)
	d.Set("project_id", res.ProjectID)
	d.Set("gateway", res.Gateway)
	d.Set("service_addresses", res.ServiceAddresses)
	d.Set("tags", res.Tags)
	d.Set("created_at", res.CreatedAt)
	d.Set("updated_at", res.UpdatedAt)
	d.Set("status", res.Status)
	d.Set("account_id", res.AccountID)
	d.Set("netops_subnet_id", res.NetopsSubnetID)
	d.Set("sv_subnet_id", res.SvSubnetID)

	return nil
}

func resourceGlobalRouterVPCSubnetV1Update(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, diagErr := getGlobalRouterClient(meta)
	if diagErr != nil {
		return diagErr
	}

	var (
		updateOpts globalrouter.SubnetUpdateRequest
		hasChange  bool
	)

	if d.HasChange("name") {
		hasChange = true
		name := d.Get("name").(string)
		updateOpts.Name = &name
	}

	if d.HasChange("tags") {
		hasChange = true
		if v, ok := d.GetOk("tags"); ok {
			tags := v.(*schema.Set).List()
			tagsToUpdate := expandToStringSlice(tags)
			updateOpts.Tags = &tagsToUpdate
		} else {
			updateOpts.Tags = &[]string{}
		}
	}

	if hasChange {
		_, _, err := client.SubnetUpdate(ctx, d.Id(), &updateOpts)
		if err != nil {
			return diag.FromErr(errUpdatingObject(objectGlobalRouterVPCSubnet, d.Id(), err))
		}

		diagErr = waiters.WaitForSubnetV1ActiveState(ctx, client, d.Id(), d.Timeout(schema.TimeoutUpdate))
		if diagErr != nil {
			return diagErr
		}

		return resourceGlobalRouterVPCSubnetV1Read(ctx, d, meta)
	}

	return nil
}

func resourceGlobalRouterVPCSubnetV1Delete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, diagErr := getGlobalRouterClient(meta)
	if diagErr != nil {
		return diagErr
	}

	resourceID := d.Id()

	_, err := client.SubnetDisconnect(ctx, resourceID)
	if err != nil {
		return diag.FromErr(errDeletingObject(objectGlobalRouterVPCSubnet, d.Id(), err))
	}

	diagErr = waiters.WaitForSubnetV1Deleted(ctx, client, d.Id(), d.Timeout(schema.TimeoutDelete))
	if diagErr != nil {
		return diagErr
	}

	d.SetId("")

	return nil
}
