package selectel

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"log"
)

func resourceDedicatedSSHKeysV1() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceDedicatedSSHKeysV1Create,
		ReadContext:   resourceDedicatedSSHKeysV1Read,
		DeleteContext: resourceDedicatedSSHKeysV1Delete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"public_key": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"user_id": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			//computed
			"id": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceDedicatedSSHKeysV1Create(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	dsClient, diagErr := getDedicatedClient(d, meta, false)
	if diagErr != nil {
		return diagErr
	}

	name := d.Get("name").(string)
	publicKey := d.Get("public_key").(string)

	var userID string
	if v, ok := d.GetOk("user_id"); ok {
		userID = v.(string)
	}

	key, _, err := dsClient.CreateSSHKey(ctx, name, publicKey, userID)
	if err != nil {
		return diag.FromErr(fmt.Errorf("failed to retrieve SSH key: %w", err))
	}

	_ = d.Set("name", key.Name)
	_ = d.Set("public_key", key.PublicKey)
	_ = d.Set("user_id", key.SubUserID)
	d.SetId(key.ID)

	return nil
}

func resourceDedicatedSSHKeysV1Read(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	dsClient, diagErr := getDedicatedClient(d, meta, false)
	if diagErr != nil {
		return diagErr
	}

	log.Print(msgGet(objectKeypair, d.Id()))

	key, _, err := dsClient.GetSSHKey(ctx, d.Id())
	if err != nil {
		return diag.FromErr(fmt.Errorf("failed to retrieve SSH key: %w", err))
	}

	_ = d.Set("name", key.Name)
	_ = d.Set("public_key", key.PublicKey)
	_ = d.Set("user_id", key.SubUserID)

	return nil
}

func resourceDedicatedSSHKeysV1Delete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	dsClient, diagErr := getDedicatedClient(d, meta, false)
	if diagErr != nil {
		return diagErr
	}

	_, err := dsClient.DeleteSSHKey(ctx, d.Id())
	if err != nil {
		return diag.FromErr(fmt.Errorf("failed to retrieve SSH key: %w", err))
	}

	d.SetId("")

	return nil
}
