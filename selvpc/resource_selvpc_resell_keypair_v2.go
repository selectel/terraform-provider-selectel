package selvpc

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/selectel/go-selvpcclient/selvpcclient/resell/v2/keypairs"
)

func resourceResellKeypairV2() *schema.Resource {
	return &schema.Resource{
		Create: resourceResellKeypairV2Create,
		Read:   resourceResellKeypairV2Read,
		Delete: resourceResellKeypairV2Delete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
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
			"regions": {
				Type:     schema.TypeSet,
				Optional: true,
				ForceNew: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"user_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceResellKeypairV2Create(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	resellV2Client := config.resellV2Client()
	ctx := context.Background()

	opts := keypairs.KeypairOpts{
		Name:      d.Get("name").(string),
		PublicKey: d.Get("public_key").(string),
		UserID:    d.Get("user_id").(string),
		Regions:   expandResellV2Regions(d.Get("regions").(*schema.Set)),
	}

	log.Printf("[DEBUG] Creating keypair with options: %v\n", opts)
	newKeypairs, _, err := keypairs.Create(ctx, resellV2Client, opts)
	if err != nil {
		return errCreatingObject("keypair", err)
	}

	// There can be several keypairs if user specified more than one region.
	// All those keypairs will have the same name and user ID attributes.
	if len(newKeypairs) == 0 {
		return errors.New("no keypairs were created")
	}
	// Retrieve same attributes to build ID of the resource.
	userID := newKeypairs[0].UserID
	keypairName := newKeypairs[0].Name

	d.SetId(resourceResellKeypairV2BuildID(userID, keypairName))

	return resourceResellKeypairV2Read(d, meta)
}

func resourceResellKeypairV2Read(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	resellV2Client := config.resellV2Client()
	ctx := context.Background()

	log.Printf("[DEBUG] Getting keypair %s", d.Id())
	userID, keypairName, err := resourceResellKeypairV2ParseID(d.Id())
	if err != nil {
		return err
	}
	existingKeypairs, _, err := keypairs.List(ctx, resellV2Client)
	if err != nil {
		return errSearchingKeypair(keypairName, err)
	}

	found := false
	for _, keypair := range existingKeypairs {
		if keypair.UserID == userID && keypair.Name == keypairName {
			found = true
			d.Set("name", keypair.Name)
			d.Set("public_key", keypair.PublicKey)
			d.Set("regions", keypair.Regions)
			d.Set("user_id", keypair.UserID)
		}
	}

	if !found {
		d.SetId("")
	}

	return nil
}

func resourceResellKeypairV2Delete(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	resellV2Client := config.resellV2Client()
	ctx := context.Background()

	userID, keypairName, err := resourceResellKeypairV2ParseID(d.Id())
	if err != nil {
		return err
	}

	log.Printf("[DEBUG] Deleting keypair %s\n", d.Id())
	response, err := keypairs.Delete(ctx, resellV2Client, keypairName, userID)
	if err != nil {
		if response.StatusCode == http.StatusNotFound {
			d.SetId("")
			return nil
		}

		return errDeletingObject("keypair", d.Id(), err)
	}

	return nil
}

func resourceResellKeypairV2BuildID(userID, keypairName string) string {
	return fmt.Sprintf("%s/%s", userID, keypairName)
}

func resourceResellKeypairV2ParseID(id string) (string, string, error) {
	idParts := strings.Split(id, "/")
	if len(idParts) != 2 {
		return "", "", errParseID(id)
	}

	return idParts[0], idParts[1], nil
}
