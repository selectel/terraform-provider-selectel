package selectel

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	vpc "github.com/selectel/vpc-go/pkg/v2"
	"github.com/selectel/vpc-go/pkg/v2/network"
)

const (
	networkingNetworkV2Object = "networking network"

	// Neutron network statuses.
	networkingNetworkV2StatusActive = "ACTIVE"
	networkingNetworkV2StatusDown   = "DOWN"
	networkingNetworkV2StatusBuild  = "BUILD"
	networkingNetworkV2StatusError  = "ERROR"

	// networkingNetworkV2StatusDeleted is the synthetic status the delete
	// waiter reports once the network is gone.
	networkingNetworkV2StatusDeleted = "DELETED"
)

// networkingNetworkV2DNSDomainRegexp accepts either an empty string or a
// fully-qualified domain name, which must end with a dot.
var networkingNetworkV2DNSDomainRegexp = regexp.MustCompile(`^$|\.$`)

func resourceVPCNetworkV2() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceVPCNetworkV2Create,
		ReadContext:   resourceVPCNetworkV2Read,
		UpdateContext: resourceVPCNetworkV2Update,
		DeleteContext: resourceVPCNetworkV2Delete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceVPCNetworkV2ImportState,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
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
			"name": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"admin_state_up": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
			},
			"dns_domain": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ValidateFunc: validation.StringMatch(
					networkingNetworkV2DNSDomainRegexp,
					"fully-qualified (unambiguous) DNS domain names must have a dot at the end",
				),
			},
			"availability_zone_hints": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				ForceNew: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"tags": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"all_tags": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			// Every attribute below is read-only: the cloud's Neutron policy
			// grants none of the granular Selectel VPC roles permission to
			// change it, so the resource only reflects what the API reports.
			"shared": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"external": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"mtu": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"port_security_enabled": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"qos_policy_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"provider_network_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"provider_physical_network": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"provider_segmentation_id": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"blocked": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"is_public": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"is_dns_enabled": {
				Type:     schema.TypeBool,
				Computed: true,
			},
		},
	}
}

func resourceVPCNetworkV2Create(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, err := networkingV2Client(meta, d.Get("project_id").(string), d.Get("region").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	request := network.CreateRequest{}
	if v, ok := d.GetOk("name"); ok {
		request.Name = vpcString(v.(string))
	}
	if v, ok := d.GetOk("description"); ok {
		request.Description = vpcString(v.(string))
	}
	if v, ok := d.GetOkExists("admin_state_up"); ok {
		request.AdminStateUp = vpcBool(v.(bool))
	}
	if v, ok := d.GetOk("dns_domain"); ok {
		request.DNSDomain = vpcString(v.(string))
	}
	if v, ok := d.GetOk("availability_zone_hints"); ok {
		hints := expandToStringSlice(v.([]interface{}))
		request.AvailabilityZoneHints = &hints
	}

	log.Print("[DEBUG] creating networking network")
	created, err := network.Create(ctx, client, request)
	if err != nil {
		return diag.FromErr(errCreatingObject(networkingNetworkV2Object, err))
	}

	d.SetId(created.ID)

	if err := waitForNetworkingNetworkV2ActiveState(
		ctx, client, created.ID, d.Timeout(schema.TimeoutCreate),
	); err != nil {
		return diag.FromErr(err)
	}

	// Tags are a separate API call; the identifier is already in the state, so
	// a failure here leaves a partially configured but tracked resource.
	if tags, ok := d.GetOk("tags"); ok {
		values := expandToStringSlice(tags.(*schema.Set).List())
		if _, err := network.TagOperations(client, created.ID).Replace(ctx, values); err != nil {
			return diag.FromErr(errUpdatingObject(networkingNetworkV2Object, created.ID, err))
		}
	}

	return resourceVPCNetworkV2Read(ctx, d, meta)
}

func resourceVPCNetworkV2Read(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, err := networkingV2Client(meta, d.Get("project_id").(string), d.Get("region").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[DEBUG] getting networking network %s", d.Id())
	found, err := network.Get(ctx, client, d.Id())
	if err != nil {
		// Only a not-found means the network is gone; a denied access, a
		// conflict, a server error or a transport failure keeps the resource.
		return networkingCheckDeleted(d, err, networkingNetworkV2Object)
	}

	flattenNetworkingNetworkV2(d, found)

	return nil
}

func resourceVPCNetworkV2Update(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, err := networkingV2Client(meta, d.Get("project_id").(string), d.Get("region").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	var (
		request     network.UpdateRequest
		needsUpdate bool
	)

	if d.HasChange("name") {
		request.Name = vpcString(d.Get("name").(string))
		needsUpdate = true
	}
	if d.HasChange("description") {
		request.Description = vpcString(d.Get("description").(string))
		needsUpdate = true
	}
	if d.HasChange("admin_state_up") {
		request.AdminStateUp = vpcBool(d.Get("admin_state_up").(bool))
		needsUpdate = true
	}
	if d.HasChange("dns_domain") {
		request.DNSDomain = vpcString(d.Get("dns_domain").(string))
		needsUpdate = true
	}

	if needsUpdate {
		log.Printf("[DEBUG] updating networking network %s", d.Id())
		if _, err := network.Update(ctx, client, d.Id(), request); err != nil {
			return diag.FromErr(errUpdatingObject(networkingNetworkV2Object, d.Id(), err))
		}
	}

	if d.HasChange("tags") {
		values := expandToStringSlice(d.Get("tags").(*schema.Set).List())
		if _, err := network.TagOperations(client, d.Id()).Replace(ctx, values); err != nil {
			return diag.FromErr(errUpdatingObject(networkingNetworkV2Object, d.Id(), err))
		}
	}

	return resourceVPCNetworkV2Read(ctx, d, meta)
}

func resourceVPCNetworkV2Delete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, err := networkingV2Client(meta, d.Get("project_id").(string), d.Get("region").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[DEBUG] deleting networking network %s", d.Id())
	if err := network.Delete(ctx, client, d.Id()); err != nil {
		// A not-found from Delete does not prove absence on its own, so it is
		// confirmed by a read before Delete is treated as idempotent success.
		return networkingCheckDeletedOnDelete(d, err, networkingNetworkV2Object, func() error {
			_, confirmErr := network.Get(ctx, client, d.Id())

			return confirmErr
		})
	}

	if err := waitForNetworkingNetworkV2DeletedState(
		ctx, client, d.Id(), d.Timeout(schema.TimeoutDelete),
	); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceVPCNetworkV2ImportState(_ context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	if err := resourceNetworkingImportProjectRegion(d, meta); err != nil {
		return nil, err
	}

	return []*schema.ResourceData{d}, nil
}

// flattenNetworkingNetworkV2 writes an API network into the resource state.
func flattenNetworkingNetworkV2(d *schema.ResourceData, found *network.Network) {
	d.Set("name", found.Name)
	d.Set("description", found.Description)
	d.Set("admin_state_up", found.AdminStateUp)
	d.Set("dns_domain", found.DNSDomain)
	d.Set("status", found.Status)
	d.Set("shared", found.Shared)
	d.Set("external", found.RouterExternal)
	d.Set("mtu", found.MTU)
	d.Set("port_security_enabled", found.PortSecurityEnabled)
	d.Set("qos_policy_id", derefString(found.QoSPolicyID))
	d.Set("provider_network_type", found.ProviderNetworkType)
	d.Set("provider_physical_network", derefString(found.ProviderPhysicalNet))
	d.Set("provider_segmentation_id", derefInt(found.ProviderSegmentation))
	d.Set("blocked", found.Blocked)
	d.Set("is_public", found.IsPublic)
	d.Set("is_dns_enabled", found.IsDNSEnabled)

	if err := d.Set("availability_zone_hints", found.AvailabilityZoneHints); err != nil {
		log.Print(errSettingComplexAttr("availability_zone_hints", err))
	}
	if err := d.Set("all_tags", found.Tags); err != nil {
		log.Print(errSettingComplexAttr("all_tags", err))
	}
}

// waitForNetworkingNetworkV2ActiveState waits until the network leaves the
// transient BUILD status for a stable one: ACTIVE, or DOWN when the network is
// administratively down.
func waitForNetworkingNetworkV2ActiveState(
	ctx context.Context, client *vpc.Client, networkID string, timeout time.Duration,
) error {
	stateConf := &resource.StateChangeConf{
		Pending:    []string{networkingNetworkV2StatusBuild},
		Target:     []string{networkingNetworkV2StatusActive, networkingNetworkV2StatusDown},
		Refresh:    networkingNetworkV2StateRefreshFunc(ctx, client, networkID),
		Timeout:    timeout,
		Delay:      2 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	log.Printf("[DEBUG] waiting for networking network %s to become active", networkID)
	if _, err := stateConf.WaitForStateContext(ctx); err != nil {
		return fmt.Errorf("error waiting for the networking network %s to become active: %w", networkID, err)
	}

	return nil
}

// waitForNetworkingNetworkV2DeletedState waits until the network is gone.
func waitForNetworkingNetworkV2DeletedState(
	ctx context.Context, client *vpc.Client, networkID string, timeout time.Duration,
) error {
	stateConf := &resource.StateChangeConf{
		Pending: []string{
			networkingNetworkV2StatusActive,
			networkingNetworkV2StatusDown,
			networkingNetworkV2StatusBuild,
		},
		Target:     []string{networkingNetworkV2StatusDeleted},
		Refresh:    networkingNetworkV2StateRefreshFunc(ctx, client, networkID),
		Timeout:    timeout,
		Delay:      2 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	log.Printf("[DEBUG] waiting for networking network %s to become deleted", networkID)
	if _, err := stateConf.WaitForStateContext(ctx); err != nil {
		return fmt.Errorf("error waiting for the networking network %s to become deleted: %w", networkID, err)
	}

	return nil
}

// networkingNetworkV2StateRefreshFunc reports the current status of a network,
// reporting DELETED once it no longer exists and failing on ERROR.
func networkingNetworkV2StateRefreshFunc(
	ctx context.Context, client *vpc.Client, networkID string,
) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		found, err := network.Get(ctx, client, networkID)
		if err != nil {
			if networkingIsNotFound(err) {
				return network.Network{}, networkingNetworkV2StatusDeleted, nil
			}

			return nil, "", err
		}

		if found.Status == networkingNetworkV2StatusError {
			return found, found.Status, fmt.Errorf(
				"the networking network %s is in the %s status", networkID, found.Status)
		}

		return found, found.Status, nil
	}
}
