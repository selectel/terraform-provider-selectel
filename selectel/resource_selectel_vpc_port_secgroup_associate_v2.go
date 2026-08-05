package selectel

import (
	"context"
	"log"
	"sort"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/selectel/vpc-go/pkg/v2/port"
)

const networkingPortSecGroupAssociateV2Object = "networking port security group association"

func resourceVPCPortSecGroupAssociateV2() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceVPCPortSecGroupAssociateV2Create,
		ReadContext:   resourceVPCPortSecGroupAssociateV2Read,
		UpdateContext: resourceVPCPortSecGroupAssociateV2Update,
		DeleteContext: resourceVPCPortSecGroupAssociateV2Delete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceVPCPortSecGroupAssociateV2ImportState,
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
			// The resource addresses an existing port and never creates or
			// deletes it, nor owns any of its other attributes.
			"port_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"security_group_ids": {
				Type:     schema.TypeSet,
				Required: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			// false adds the listed groups to whatever the port already has;
			// true makes this resource the sole authority over the port's list.
			"enforce": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			// Always the port's full actual list, whatever enforce is set to.
			"all_security_group_ids": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func resourceVPCPortSecGroupAssociateV2Create(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, err := networkingV2Client(meta, d.Get("project_id").(string), d.Get("region").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	portID := d.Get("port_id").(string)
	desired := expandToStringSlice(d.Get("security_group_ids").(*schema.Set).List())

	// Reading and writing the port's whole list is a read-modify-write, so
	// concurrent associations on the same port must not interleave.
	selMutexKV.Lock(portID)
	defer selMutexKV.Unlock(portID)

	groups := desired
	if !d.Get("enforce").(bool) {
		// Additive: keep every group the port already carries.
		found, err := port.Get(ctx, client, portID)
		if err != nil {
			return diag.FromErr(errGettingObject(networkingPortV2Object, portID, err))
		}

		groups = unionNetworkingSecGroupIDs(found.SecurityGroups, desired)
	}

	log.Printf("[DEBUG] setting %d security group(s) on networking port %s", len(groups), portID)
	if _, err := port.Update(ctx, client, portID, port.UpdateRequest{
		SecurityGroups: &groups,
	}); err != nil {
		return diag.FromErr(errCreatingObject(networkingPortSecGroupAssociateV2Object, err))
	}

	d.SetId(portID)

	return resourceVPCPortSecGroupAssociateV2Read(ctx, d, meta)
}

func resourceVPCPortSecGroupAssociateV2Read(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, err := networkingV2Client(meta, d.Get("project_id").(string), d.Get("region").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[DEBUG] getting security groups of networking port %s", d.Id())
	found, err := port.Get(ctx, client, d.Id())
	if err != nil {
		// The association cannot outlive the port it addresses.
		return networkingCheckDeleted(d, err, networkingPortV2Object)
	}

	d.Set("port_id", found.ID)
	if err := d.Set("all_security_group_ids", found.SecurityGroups); err != nil {
		log.Print(errSettingComplexAttr("all_security_group_ids", err))
	}

	// In enforce mode the resource owns the whole list, so it reports the whole
	// list. In additive mode it owns only its own subset: an externally removed
	// group disappears from security_group_ids, while a group added by someone
	// else never shows up there.
	managed := found.SecurityGroups
	if !d.Get("enforce").(bool) {
		managed = intersectNetworkingSecGroupIDs(
			expandToStringSlice(d.Get("security_group_ids").(*schema.Set).List()),
			found.SecurityGroups)
	}

	if err := d.Set("security_group_ids", managed); err != nil {
		log.Print(errSettingComplexAttr("security_group_ids", err))
	}

	return nil
}

func resourceVPCPortSecGroupAssociateV2Update(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, err := networkingV2Client(meta, d.Get("project_id").(string), d.Get("region").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	if !d.HasChange("security_group_ids") && !d.HasChange("enforce") {
		return resourceVPCPortSecGroupAssociateV2Read(ctx, d, meta)
	}

	portID := d.Get("port_id").(string)
	desired := expandToStringSlice(d.Get("security_group_ids").(*schema.Set).List())

	selMutexKV.Lock(portID)
	defer selMutexKV.Unlock(portID)

	groups := desired
	// Switching enforce on, like staying in enforce mode, makes the configured
	// set the port's whole list. Only a resource that stays additive keeps the
	// groups it does not own.
	stayingAdditive := !d.Get("enforce").(bool)
	if stayingAdditive {
		found, err := port.Get(ctx, client, portID)
		if err != nil {
			return diag.FromErr(errGettingObject(networkingPortV2Object, portID, err))
		}

		// Only the subset this resource had claimed is replaced; anything else on
		// the port survives, including groups that appeared externally.
		previous, _ := d.GetChange("security_group_ids")
		kept := subtractNetworkingSecGroupIDs(
			found.SecurityGroups,
			expandToStringSlice(previous.(*schema.Set).List()))

		groups = unionNetworkingSecGroupIDs(kept, desired)
	}

	log.Printf("[DEBUG] updating security groups of networking port %s to %d entrie(s)", portID, len(groups))
	if _, err := port.Update(ctx, client, portID, port.UpdateRequest{
		SecurityGroups: &groups,
	}); err != nil {
		return diag.FromErr(errUpdatingObject(networkingPortSecGroupAssociateV2Object, portID, err))
	}

	return resourceVPCPortSecGroupAssociateV2Read(ctx, d, meta)
}

func resourceVPCPortSecGroupAssociateV2Delete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, err := networkingV2Client(meta, d.Get("project_id").(string), d.Get("region").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	portID := d.Get("port_id").(string)

	selMutexKV.Lock(portID)
	defer selMutexKV.Unlock(portID)

	found, err := port.Get(ctx, client, portID)
	if err != nil {
		if networkingIsNotFound(err) {
			// The port is gone, so its security groups went with it.
			return nil
		}

		return diag.FromErr(errGettingObject(networkingPortV2Object, portID, err))
	}

	// Additive mode releases only what it claimed. Enforce mode owned the whole
	// list, so it clears the whole list.
	groups := []string{}
	if !d.Get("enforce").(bool) {
		groups = subtractNetworkingSecGroupIDs(
			found.SecurityGroups,
			expandToStringSlice(d.Get("security_group_ids").(*schema.Set).List()))
	}

	log.Printf("[DEBUG] releasing security groups of networking port %s", portID)
	if _, err := port.Update(ctx, client, portID, port.UpdateRequest{
		SecurityGroups: &groups,
	}); err != nil {
		return diag.FromErr(errDeletingObject(networkingPortSecGroupAssociateV2Object, portID, err))
	}

	return nil
}

func resourceVPCPortSecGroupAssociateV2ImportState(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	if err := resourceNetworkingImportProjectRegion(d, meta); err != nil {
		return nil, err
	}

	client, err := networkingV2Client(meta, d.Get("project_id").(string), d.Get("region").(string))
	if err != nil {
		return nil, err
	}

	// A port_id that matches nothing must fail the import rather than produce an
	// empty resource in the state.
	found, err := port.Get(ctx, client, d.Id())
	if err != nil {
		return nil, errGettingObject(networkingPortV2Object, d.Id(), err)
	}

	d.Set("port_id", found.ID)

	return []*schema.ResourceData{d}, nil
}

// unionNetworkingSecGroupIDs returns the identifiers of both sets, without
// duplicates and in a stable order.
func unionNetworkingSecGroupIDs(left, right []string) []string {
	seen := make(map[string]struct{}, len(left)+len(right))
	result := make([]string, 0, len(left)+len(right))
	for _, group := range append(append([]string{}, left...), right...) {
		if _, exists := seen[group]; exists {
			continue
		}

		seen[group] = struct{}{}
		result = append(result, group)
	}

	sort.Strings(result)

	return result
}

// intersectNetworkingSecGroupIDs returns the identifiers present in both sets.
func intersectNetworkingSecGroupIDs(left, right []string) []string {
	present := make(map[string]struct{}, len(right))
	for _, group := range right {
		present[group] = struct{}{}
	}

	result := make([]string, 0, len(left))
	for _, group := range left {
		if _, exists := present[group]; exists {
			result = append(result, group)
		}
	}

	sort.Strings(result)

	return result
}

// subtractNetworkingSecGroupIDs returns the identifiers of left that are absent
// from right.
func subtractNetworkingSecGroupIDs(left, right []string) []string {
	removed := make(map[string]struct{}, len(right))
	for _, group := range right {
		removed[group] = struct{}{}
	}

	result := make([]string, 0, len(left))
	for _, group := range left {
		if _, exists := removed[group]; exists {
			continue
		}

		result = append(result, group)
	}

	sort.Strings(result)

	return result
}
