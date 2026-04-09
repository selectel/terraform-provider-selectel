package selectel

import (
	"context"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	dedicated "github.com/selectel/dedicated-go/v2/pkg/v2"
)

func resourceDedicatedPrivateSubnetV1() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceDedicatedPrivateSubnetV1Create,
		ReadContext:   resourceDedicatedPrivateSubnetV1Read,
		DeleteContext: resourceDedicatedPrivateSubnetV1Delete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(5 * time.Minute),
			Delete: schema.DefaultTimeout(5 * time.Minute),
		},
		Schema: map[string]*schema.Schema{
			"id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "ID of the created private subnet",
			},
			"location_id": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				Description:  "Location ID where the private subnet will be created",
				ValidateFunc: validation.IsUUID,
			},
			"vlan": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "VLAN TAG for the private subnet",
			},
			"subnet": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				Description:  "CIDR notation for the private subnet (e.g., 192.168.1.0/24)",
				ValidateFunc: validation.IsCIDR,
			},
		},
	}
}

func resourceDedicatedPrivateSubnetV1Create(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	client, diagErr := getDedicatedClient(d, meta, false)
	if diagErr != nil {
		return diagErr
	}

	locationID := d.Get("location_id").(string)
	vlan := d.Get("vlan").(string)
	subnetCIDR := d.Get("subnet").(string)

	// Validate private subnet
	err := validatePrivateSubnet(subnetCIDR)
	if err != nil {
		return diag.FromErr(err)
	}

	networks, _, err := client.Networks(ctx, locationID, dedicated.NetworkTypeLocal, vlan)
	if err != nil {
		return diag.Errorf("failed to get networks for location %s: %s", locationID, err)
	}
	if len(networks) == 0 {
		return diag.Errorf("vlan %s not found in location %s", vlan, locationID)
	}

	localSubnet, _, err := client.CreateNetworkLocalSubnet(ctx, networks[0].UUID, subnetCIDR)
	if err != nil {
		return diag.FromErr(err)
	}

	_ = d.Set("vlan", strconv.Itoa(localSubnet.Network))
	_ = d.Set("location_id", localSubnet.LocationUUID)
	_ = d.Set("subnet", localSubnet.Subnet)
	d.SetId(localSubnet.UUID)

	return nil
}

func resourceDedicatedPrivateSubnetV1Read(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	client, diagErr := getDedicatedClient(d, meta, false)
	if diagErr != nil {
		return diagErr
	}

	localSubnet, _, err := client.GetNetworkLocalSubnet(ctx, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	_ = d.Set("vlan", strconv.Itoa(localSubnet.Network))
	_ = d.Set("location_id", localSubnet.LocationUUID)
	_ = d.Set("subnet", localSubnet.Subnet)
	d.SetId(localSubnet.UUID)

	return nil
}

func resourceDedicatedPrivateSubnetV1Delete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	client, diagErr := getDedicatedClient(d, meta, false)
	if diagErr != nil {
		return diagErr
	}

	_, err := client.DeleteNetworkLocalSubnet(ctx, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")

	return nil
}
