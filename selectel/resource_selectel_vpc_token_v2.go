package selectel

import (
	"context"
	"log"
	"net/http"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/selectel/go-selvpcclient/selvpcclient/resell/v2/tokens"
)

func resourceVPCTokenV2() *schema.Resource {
	return &schema.Resource{
		Create: resourceVPCTokenV2Create,
		Read:   resourceVPCTokenV2Read,
		Delete: resourceVPCTokenV2Delete,
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

func resourceVPCTokenV2Create(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	resellV2Client := config.resellV2Client()
	ctx := context.Background()

	opts := tokens.TokenOpts{
		ProjectID:  d.Get("project_id").(string),
		DomainName: d.Get("account_name").(string),
	}

	log.Print(msgCreate(objectToken, opts))
	token, _, err := tokens.Create(ctx, resellV2Client, opts)
	if err != nil {
		return errCreatingObject(objectToken, err)
	}

	d.SetId(token.ID)

	return resourceVPCTokenV2Read(d, meta)
}

func resourceVPCTokenV2Read(d *schema.ResourceData, meta interface{}) error {
	// There is no API support for getting a token yet.

	return nil
}

func resourceVPCTokenV2Delete(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	resellV2Client := config.resellV2Client()
	ctx := context.Background()

	log.Print(msgDelete(objectToken, d.Id()))
	response, err := tokens.Delete(ctx, resellV2Client, d.Id())
	if err != nil {
		if response != nil {
			if response.StatusCode == http.StatusNotFound {
				d.SetId("")
				return nil
			}
		}

		return errDeletingObject(objectToken, d.Id(), err)
	}

	return nil
}
