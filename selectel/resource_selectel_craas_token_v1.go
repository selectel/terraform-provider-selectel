package selectel

import (
	"context"
	"log"
	"net/http"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/selectel/craas-go/pkg/v1/token"
	"github.com/terraform-providers/terraform-provider-selectel/selectel/internal/hashcode"
)

func resourceCRaaSTokenV1() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceCRaaSTokenV1Create,
		ReadContext:   resourceCRaaSTokenV1Read,
		DeleteContext: resourceCRaaSTokenV1Delete,
		Schema: map[string]*schema.Schema{
			"project_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"token_ttl": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				ValidateFunc: validation.StringInSlice([]string{
					string(token.TTL12Hours),
					string(token.TTL1Year),
				}, false),
				Default: string(token.TTL1Year),
			},
			"username": {
				Type:      schema.TypeString,
				Sensitive: true,
				Computed:  true,
			},
			"token": {
				Type:      schema.TypeString,
				Sensitive: true,
				Computed:  true,
			},
		},
	}
}

func resourceCRaaSTokenV1Create(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	craasClient, diagErr := getCRaaSClient(d, meta)
	if diagErr != nil {
		return diagErr
	}

	tokenTTL := d.Get("token_ttl").(string)
	createOpts := &token.CreateOpts{
		TokenTTL: token.TTL(tokenTTL),
	}

	log.Print(msgCreate(objectRegistryToken, createOpts))
	newToken, _, err := token.Create(ctx, craasClient, createOpts)
	if err != nil {
		return diag.FromErr(errCreatingObject(objectRegistryToken, err))
	}

	tokenID := strconv.Itoa(hashcode.String(newToken.Token))
	d.SetId(tokenID)
	d.Set("token", newToken.Token)

	return resourceCRaaSTokenV1Read(ctx, d, meta)
}

func resourceCRaaSTokenV1Read(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	craasClient, diagErr := getCRaaSClient(d, meta)
	if diagErr != nil {
		return diagErr
	}

	log.Print(msgGet(objectRegistryToken, d.Id()))
	craasToken, response, err := token.Get(ctx, craasClient, d.Get("token").(string))
	if err != nil {
		if response != nil {
			if response.StatusCode == http.StatusNotFound {
				d.SetId("")
				return nil
			}
		}

		return diag.FromErr(errGettingObject(objectRegistryToken, d.Id(), err))
	}

	d.Set("username", craasV1TokenUsername)
	d.Set("token", craasToken.Token)

	return nil
}

func resourceCRaaSTokenV1Delete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	craasClient, diagErr := getCRaaSClient(d, meta)
	if diagErr != nil {
		return diagErr
	}

	log.Print(msgDelete(objectRegistryToken, d.Id()))
	_, err := token.Revoke(ctx, craasClient, d.Get("token").(string))
	if err != nil {
		return diag.FromErr(errDeletingObject(objectRegistryToken, d.Id(), err))
	}

	return nil
}
