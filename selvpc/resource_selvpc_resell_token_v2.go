package selvpc

import (
	"context"
	"log"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/selectel/go-selvpcclient/selvpcclient/resell/v2/tokens"
)

func resourceResellTokenV2() *schema.Resource {
	return &schema.Resource{
		Create: resourceResellTokenV2Create,
		Read:   resourceResellTokenV2Read,
		Delete: resourceResellTokenV2Delete,
		Schema: map[string]*schema.Schema{
			"project_id": {
				Type:          schema.TypeString,
				ConflictsWith: []string{"account_name"},
				Optional:      true,
				ForceNew:      true,
			},
			"account_name": {
				Type:          schema.TypeString,
				ConflictsWith: []string{"project_id"},
				Optional:      true,
				ForceNew:      true,
			},
		},
	}
}

func resourceResellTokenV2Create(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	resellV2Client := config.resellV2Client()
	ctx := context.Background()

	opts := tokens.TokenOpts{
		ProjectID:  d.Get("project_id").(string),
		DomainName: d.Get("account_name").(string),
	}

	log.Printf("[DEBUG] Creating token with options: %v\n", opts)
	token, _, err := tokens.Create(ctx, resellV2Client, opts)
	if err != nil {
		return errCreatingObject("token", err)
	}

	d.SetId(token.ID)

	return resourceResellTokenV2Read(d, meta)
}

func resourceResellTokenV2Read(d *schema.ResourceData, meta interface{}) error {
	// There is no API support for getting a token yet.

	return nil
}

func resourceResellTokenV2Delete(d *schema.ResourceData, meta interface{}) error {
	// There is no API support for deleting a token yet.

	return nil
}
