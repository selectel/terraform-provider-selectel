package selectel

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceDedicatedSSHKeysV1() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceDedicatedSSHKeysV1Create,
		ReadContext:   resourceDedicatedSSHKeysV1Read,
		DeleteContext: resourceDedicatedSSHKeysV1Delete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceDedicatedSSHKeysV1Import,
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringIsNotEmpty,
			},
			"public_key": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringIsNotEmpty,
				DiffSuppressFunc: func(_, oldVal, newVal string, _ *schema.ResourceData) bool {
					// Suppress diff if keys are equal after trimming whitespace
					return strings.TrimSpace(oldVal) == strings.TrimSpace(newVal)
				},
			},
			"user_id": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.IsUUID,
			},
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

func resourceDedicatedSSHKeysV1Import(
	ctx context.Context,
	d *schema.ResourceData,
	meta any,
) ([]*schema.ResourceData, error) {
	client, diagErr := getDedicatedClient(d, meta, false)
	if diagErr != nil {
		return nil, fmt.Errorf("failed to get dedicated client")
	}

	keys, _, err := client.SSHKeys(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list SSH keys: %w", err)
	}

	for _, k := range keys {
		if k.Name == d.Id() {
			d.SetId(k.ID)
			_ = d.Set("name", k.Name)
			_ = d.Set("public_key", k.PublicKey)
			_ = d.Set("user_id", k.SubUserID)

			return []*schema.ResourceData{d}, nil
		}
	}

	return nil, fmt.Errorf("SSH key with name %q not found", d.Id())
}
