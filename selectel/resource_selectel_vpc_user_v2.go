package selectel

import (
	"context"
	"log"
	"net/http"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/selectel/go-selvpcclient/selvpcclient/resell/v2/users"
)

func resourceVPCUserV2() *schema.Resource {
	return &schema.Resource{
		Create: resourceVPCUserV2Create,
		Read:   resourceVPCUserV2Read,
		Update: resourceVPCUserV2Update,
		Delete: resourceVPCUserV2Delete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: false,
			},
			"password": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: false,
			},
			"enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
				ForceNew: false,
			},
		},
	}
}

func resourceVPCUserV2Create(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	resellV2Client := config.resellV2Client()
	ctx := context.Background()

	opts := users.UserOpts{
		Name:     d.Get("name").(string),
		Password: d.Get("password").(string),
	}

	log.Print(msgCreate(objectUser, opts))
	user, _, err := users.Create(ctx, resellV2Client, opts)
	if err != nil {
		return errCreatingObject(objectUser, err)
	}

	d.SetId(user.ID)

	return resourceVPCUserV2Read(d, meta)
}

func resourceVPCUserV2Read(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	resellV2Client := config.resellV2Client()
	ctx := context.Background()

	log.Print(msgGet(objectUser, d.Id()))
	user, response, err := users.Get(ctx, resellV2Client, d.Id())
	if err != nil {
		if response.StatusCode == http.StatusNotFound {
			d.SetId("")
			return nil
		}

		return errGettingObject(objectUser, d.Id(), err)
	}

	d.Set("id", user.ID)
	d.Set("name", user.Name)
	d.Set("enabled", user.Enabled)

	return nil
}

func resourceVPCUserV2Update(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	resellV2Client := config.resellV2Client()
	ctx := context.Background()

	enabled := d.Get("enabled").(bool)
	opts := users.UserOpts{
		Name:     d.Get("name").(string),
		Password: d.Get("password").(string),
		Enabled:  &enabled,
	}

	log.Print(msgUpdate(objectUser, d.Id(), opts))
	_, _, err := users.Update(ctx, resellV2Client, d.Id(), opts)
	if err != nil {
		return errUpdatingObject(objectUser, d.Id(), err)
	}

	return resourceVPCUserV2Read(d, meta)
}

func resourceVPCUserV2Delete(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	resellV2Client := config.resellV2Client()
	ctx := context.Background()

	log.Print(msgDelete(objectUser, d.Id()))
	response, err := users.Delete(ctx, resellV2Client, d.Id())
	if err != nil {
		if response.StatusCode == http.StatusNotFound {
			d.SetId("")
			return nil
		}

		return errDeletingObject(objectUser, d.Id(), err)
	}

	return nil
}
