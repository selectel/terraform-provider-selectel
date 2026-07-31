package selectel

import (
	"context"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/selectel/vpc-go/pkg/v2/addressscope"
)

const networkingAddressScopeV2Object = "networking address scope"

func resourceNetworkingAddressScopeV2() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceNetworkingAddressScopeV2Create,
		ReadContext:   resourceNetworkingAddressScopeV2Read,
		UpdateContext: resourceNetworkingAddressScopeV2Update,
		DeleteContext: resourceNetworkingAddressScopeV2Delete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceNetworkingAddressScopeV2ImportState,
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
				Required: true,
			},
			// No local validation of the value: an unsupported one is rejected by
			// the API, and the diagnostic comes from its answer.
			"ip_version": {
				Type:     schema.TypeInt,
				Optional: true,
				ForceNew: true,
				Default:  4,
			},
			// Read-only: the policy reserves sharing an address scope for the
			// Keystone admin role, so it is reflected and never sent.
			"shared": {
				Type:     schema.TypeBool,
				Computed: true,
			},
		},
	}
}

func resourceNetworkingAddressScopeV2Create(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, err := networkingV2Client(meta, d.Get("project_id").(string), d.Get("region").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	log.Print("[DEBUG] creating networking address scope")
	created, err := addressscope.Create(ctx, client, addressscope.CreateRequest{
		Name:      vpcString(d.Get("name").(string)),
		IPVersion: d.Get("ip_version").(int),
	})
	if err != nil {
		return diag.FromErr(errCreatingObject(networkingAddressScopeV2Object, err))
	}

	d.SetId(created.ID)

	return resourceNetworkingAddressScopeV2Read(ctx, d, meta)
}

func resourceNetworkingAddressScopeV2Read(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, err := networkingV2Client(meta, d.Get("project_id").(string), d.Get("region").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[DEBUG] getting networking address scope %s", d.Id())
	found, err := addressscope.Get(ctx, client, d.Id())
	if err != nil {
		return networkingCheckDeleted(d, err, networkingAddressScopeV2Object)
	}

	flattenNetworkingAddressScopeV2(d, found)

	return nil
}

func resourceNetworkingAddressScopeV2Update(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, err := networkingV2Client(meta, d.Get("project_id").(string), d.Get("region").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	// name is the only updatable attribute; shared is never sent, so a value an
	// admin set outside Terraform is left alone.
	if d.HasChange("name") {
		log.Printf("[DEBUG] updating networking address scope %s", d.Id())
		if _, err := addressscope.Update(ctx, client, d.Id(), addressscope.UpdateRequest{
			Name: vpcString(d.Get("name").(string)),
		}); err != nil {
			return diag.FromErr(errUpdatingObject(networkingAddressScopeV2Object, d.Id(), err))
		}
	}

	return resourceNetworkingAddressScopeV2Read(ctx, d, meta)
}

func resourceNetworkingAddressScopeV2Delete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, err := networkingV2Client(meta, d.Get("project_id").(string), d.Get("region").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[DEBUG] deleting networking address scope %s", d.Id())
	if err := addressscope.Delete(ctx, client, d.Id()); err != nil {
		// A subnet pool still referencing the scope makes this a conflict, which
		// stays a diagnostic error rather than a silent state removal.
		return networkingCheckDeletedOnDelete(d, err, networkingAddressScopeV2Object, func() error {
			_, confirmErr := addressscope.Get(ctx, client, d.Id())

			return confirmErr
		})
	}

	return nil
}

func resourceNetworkingAddressScopeV2ImportState(_ context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	if err := resourceNetworkingImportProjectRegion(d, meta); err != nil {
		return nil, err
	}

	return []*schema.ResourceData{d}, nil
}

// flattenNetworkingAddressScopeV2 writes an API address scope into the state.
func flattenNetworkingAddressScopeV2(d *schema.ResourceData, found *addressscope.AddressScope) {
	d.Set("name", found.Name)
	d.Set("ip_version", found.IPVersion)
	d.Set("shared", found.Shared)
}
