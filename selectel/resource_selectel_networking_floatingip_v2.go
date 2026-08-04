package selectel

import (
	"context"
	"fmt"
	"log"
	"regexp"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	vpc "github.com/selectel/vpc-go/pkg/v2"
	"github.com/selectel/vpc-go/pkg/v2/floatingip"
	"github.com/selectel/vpc-go/pkg/v2/network"
)

const networkingFloatingIPV2Object = "networking floating ip"

// networkingFloatingIPV2DNSDomainRegexp accepts either an empty string or a
// fully-qualified domain name, which must end with a dot.
var networkingFloatingIPV2DNSDomainRegexp = regexp.MustCompile(`^$|\.$`)

func resourceNetworkingFloatingIPV2() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceNetworkingFloatingIPV2Create,
		ReadContext:   resourceNetworkingFloatingIPV2Read,
		UpdateContext: resourceNetworkingFloatingIPV2Update,
		DeleteContext: resourceNetworkingFloatingIPV2Delete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceNetworkingFloatingIPV2ImportState,
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
			// pool is the name of the external network the address comes from;
			// the provider resolves it to the network identifier.
			"pool": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			// subnet_id is write-only on the API: the create request carries it and
			// the floating IP model never returns it, so Read cannot fill it in and
			// it must not be Computed — there would be nothing to populate from.
			// After an import the stored value is therefore empty, and without the
			// suppression below a configured subnet_id would plan a replacement
			// that releases an already allocated address over a value the provider
			// merely could not read.
			"subnet_id": {
				Type:             schema.TypeString,
				Optional:         true,
				ForceNew:         true,
				DiffSuppressFunc: suppressNetworkingFloatingIPV2SubnetIDDiff,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			// The association may be changed in place, so neither attribute
			// replaces the floating IP.
			"port_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"fixed_ip": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			// dns_name and dns_domain are write-only in the same way subnet_id is: the
			// API accepts both on create and returns neither, so Read has nothing to
			// fill them from. Computed would therefore populate nothing, and without
			// the suppression a configured value against the empty stored one plans a
			// replacement on every run — releasing an allocated address over values
			// the provider merely could not read.
			"dns_name": {
				Type:             schema.TypeString,
				Optional:         true,
				ForceNew:         true,
				DiffSuppressFunc: suppressNetworkingFloatingIPV2WriteOnlyDiff,
			},
			"dns_domain": {
				Type:             schema.TypeString,
				Optional:         true,
				ForceNew:         true,
				DiffSuppressFunc: suppressNetworkingFloatingIPV2WriteOnlyDiff,
				ValidateFunc: validation.StringMatch(
					networkingFloatingIPV2DNSDomainRegexp,
					"fully-qualified (unambiguous) DNS domain names must have a dot at the end",
				),
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
			// The API always assigns the address itself: a specific one cannot
			// be requested.
			"address": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"floating_network_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"router_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			// Read-only: writable only with admin privileges.
			"qos_policy_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"blocked": {
				Type:     schema.TypeBool,
				Computed: true,
			},
		},
	}
}

func resourceNetworkingFloatingIPV2Create(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, err := networkingV2Client(meta, d.Get("project_id").(string), d.Get("region").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	// pool is a name, while the API takes an identifier, so it is resolved
	// before the floating IP is created: an unknown or ambiguous pool fails here
	// rather than leaving a half-created address behind.
	networkID, err := resolveNetworkingFloatingIPV2Pool(ctx, client, d.Get("pool").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	request := floatingip.CreateRequest{FloatingNetworkID: networkID}
	if v, ok := d.GetOk("subnet_id"); ok {
		request.SubnetID = vpcString(v.(string))
	}
	if v, ok := d.GetOk("description"); ok {
		request.Description = vpcString(v.(string))
	}
	if v, ok := d.GetOk("port_id"); ok {
		request.PortID = vpcString(v.(string))
	}
	if v, ok := d.GetOk("fixed_ip"); ok {
		request.FixedIPAddress = vpcString(v.(string))
	}
	if v, ok := d.GetOk("dns_name"); ok {
		request.DNSName = vpcString(v.(string))
	}
	if v, ok := d.GetOk("dns_domain"); ok {
		request.DNSDomain = vpcString(v.(string))
	}

	log.Print("[DEBUG] creating networking floating ip")
	created, err := floatingip.Create(ctx, client, request)
	if err != nil {
		return diag.FromErr(errCreatingObject(networkingFloatingIPV2Object, err))
	}

	d.SetId(created.ID)

	if tags, ok := d.GetOk("tags"); ok {
		values := expandToStringSlice(tags.(*schema.Set).List())
		if _, err := floatingip.TagOperations(client, created.ID).Replace(ctx, values); err != nil {
			return diag.FromErr(errUpdatingObject(networkingFloatingIPV2Object, created.ID, err))
		}
	}

	return resourceNetworkingFloatingIPV2Read(ctx, d, meta)
}

func resourceNetworkingFloatingIPV2Read(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, err := networkingV2Client(meta, d.Get("project_id").(string), d.Get("region").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[DEBUG] getting networking floating ip %s", d.Id())
	found, err := floatingip.Get(ctx, client, d.Id())
	if err != nil {
		return networkingCheckDeleted(d, err, networkingFloatingIPV2Object)
	}

	flattenNetworkingFloatingIPV2Resource(d, found)

	return nil
}

func resourceNetworkingFloatingIPV2Update(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, err := networkingV2Client(meta, d.Get("project_id").(string), d.Get("region").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	var (
		request     floatingip.UpdateRequest
		needsUpdate bool
	)

	if d.HasChange("description") {
		request.Description = vpcString(d.Get("description").(string))
		needsUpdate = true
	}
	if d.HasChange("port_id") || d.HasChange("fixed_ip") {
		// Clearing port_id detaches the address, which the API expresses as an
		// explicit JSON null rather than an omitted field. Neutron also expects
		// the current port when fixed_ip alone changes.
		if portID := d.Get("port_id").(string); portID == "" {
			request.PortID = vpc.Null[string]()
		} else {
			request.PortID = vpc.Value(portID)
		}
		needsUpdate = true
	}
	if d.HasChange("fixed_ip") {
		if fixedIP := d.Get("fixed_ip").(string); fixedIP == "" {
			request.FixedIPAddress = vpc.Null[string]()
		} else {
			request.FixedIPAddress = vpc.Value(fixedIP)
		}
		needsUpdate = true
	}

	if needsUpdate {
		log.Printf("[DEBUG] updating networking floating ip %s", d.Id())
		if _, err := floatingip.Update(ctx, client, d.Id(), request); err != nil {
			return diag.FromErr(errUpdatingObject(networkingFloatingIPV2Object, d.Id(), err))
		}
	}

	if d.HasChange("tags") {
		values := expandToStringSlice(d.Get("tags").(*schema.Set).List())
		if _, err := floatingip.TagOperations(client, d.Id()).Replace(ctx, values); err != nil {
			return diag.FromErr(errUpdatingObject(networkingFloatingIPV2Object, d.Id(), err))
		}
	}

	return resourceNetworkingFloatingIPV2Read(ctx, d, meta)
}

func resourceNetworkingFloatingIPV2Delete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, err := networkingV2Client(meta, d.Get("project_id").(string), d.Get("region").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[DEBUG] deleting networking floating ip %s", d.Id())
	if err := floatingip.Delete(ctx, client, d.Id()); err != nil {
		return networkingCheckDeletedOnDelete(d, err, networkingFloatingIPV2Object, func() error {
			_, confirmErr := floatingip.Get(ctx, client, d.Id())

			return confirmErr
		})
	}

	return nil
}

func resourceNetworkingFloatingIPV2ImportState(_ context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	if err := resourceNetworkingImportProjectRegion(d, meta); err != nil {
		return nil, err
	}

	return []*schema.ResourceData{d}, nil
}

// suppressNetworkingFloatingIPV2SubnetIDDiff hides the difference between an
// unknown stored subnet_id and a configured one.
//
// The API never returns subnet_id (REQ-086), so an empty stored value means the
// provider could not learn it — after an import — rather than that the subnet
// changed. REQ-089 scopes the replacement rule to a change between two known
// values, so only a change from one non-empty value to another replaces the
// floating IP.
func suppressNetworkingFloatingIPV2SubnetIDDiff(_, stored, configured string, _ *schema.ResourceData) bool {
	return stored == "" && configured != ""
}

// suppressNetworkingFloatingIPV2WriteOnlyDiff is the same rule for dns_name and
// dns_domain: the API accepts them on create and returns neither, verified against
// a live floating IP, so an empty stored value means the provider could not read
// the value rather than that it changed.
func suppressNetworkingFloatingIPV2WriteOnlyDiff(
	_, stored, configured string, _ *schema.ResourceData,
) bool {
	return stored == "" && configured != ""
}

// resolveNetworkingFloatingIPV2Pool maps an external network name to its
// identifier. An unknown name and an ambiguous one are distinct errors, so the
// caller learns which of the two happened.
func resolveNetworkingFloatingIPV2Pool(
	ctx context.Context, client *vpc.Client, pool string,
) (string, error) {
	found, err := network.List(ctx, client, vpc.ListOptions{
		SelectionOptions: vpc.SelectionOptions{Filters: map[string][]string{
			"name":            {pool},
			"router:external": {"true"},
		}},
	})
	if err != nil {
		return "", errGettingObjects(networkingNetworkV2Object, err)
	}

	return networkingFloatingIPV2PoolID(found, pool)
}

func networkingFloatingIPV2PoolID(found []network.Network, pool string) (string, error) {
	switch len(found) {
	case 0:
		return "", fmt.Errorf(
			"pool '%s' does not match any external network in this project and region", pool)
	case 1:
		return found[0].ID, nil
	default:
		return "", fmt.Errorf(
			"pool '%s' matches %d external networks, please use a pool name that is unique",
			pool, len(found))
	}
}

func flattenNetworkingFloatingIPV2Resource(d *schema.ResourceData, found *floatingip.FloatingIP) {
	flattenNetworkingFloatingIPV2(d, found)

	if err := d.Set("tags", found.Tags); err != nil {
		log.Print(errSettingComplexAttr("tags", err))
	}
}

// flattenNetworkingFloatingIPV2 writes an API floating IP into the state.
func flattenNetworkingFloatingIPV2(d *schema.ResourceData, found *floatingip.FloatingIP) {
	d.Set("address", found.FloatingIPAddress)
	d.Set("floating_network_id", found.FloatingNetworkID)
	d.Set("port_id", derefString(found.PortID))
	d.Set("fixed_ip", derefString(found.FixedIPAddress))
	d.Set("router_id", derefString(found.RouterID))
	d.Set("description", found.Description)
	d.Set("status", found.Status)
	d.Set("qos_policy_id", derefString(found.QoSPolicyID))
	// dns_name and dns_domain are deliberately not set: the API returns neither,
	// so writing what it sent would replace a configured value with an empty one.
	d.Set("blocked", found.Blocked)

	if err := d.Set("all_tags", found.Tags); err != nil {
		log.Print(errSettingComplexAttr("all_tags", err))
	}
}
