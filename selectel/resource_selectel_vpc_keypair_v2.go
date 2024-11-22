package selectel

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/selectel/go-selvpcclient/v4/selvpcclient/resell/v2/keypairs"
)

func resourceVPCKeypairV2() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceVPCKeypairV2Create,
		ReadContext:   resourceVPCKeypairV2Read,
		DeleteContext: resourceVPCKeypairV2Delete,
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

func resourceVPCKeypairV2Create(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*Config)
	selvpcClient, err := config.GetSelVPCClient()
	if err != nil {
		return diag.FromErr(fmt.Errorf("can't get selvpc client for keypairs object: %w", err))
	}

	opts := keypairs.KeypairOpts{
		Name:      d.Get("name").(string),
		PublicKey: d.Get("public_key").(string),
		UserID:    d.Get("user_id").(string),
		Regions:   expandVPCV2Regions(d.Get("regions").(*schema.Set)),
	}

	log.Print(msgCreate(objectKeypair, opts))
	newKeypairs, _, err := keypairs.Create(selvpcClient, opts)
	if err != nil {
		return diag.FromErr(errCreatingObject(objectKeypair, err))
	}

	// There can be several keypairs if user specified more than one region.
	// All those keypairs will have the same name and user ID attributes.
	if len(newKeypairs) == 0 {
		return diag.FromErr(errReadFromResponse(objectKeypair))
	}
	// Retrieve same attributes to build ID of the resource.
	userID := newKeypairs[0].UserID
	keypairName := newKeypairs[0].Name

	d.SetId(resourceVPCKeypairV2BuildID(userID, keypairName))

	return resourceVPCKeypairV2Read(ctx, d, meta)
}

func resourceVPCKeypairV2Read(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*Config)
	selvpcClient, err := config.GetSelVPCClient()
	if err != nil {
		return diag.FromErr(fmt.Errorf("can't get selvpc client for keypairs object: %w", err))
	}

	log.Print(msgGet(objectKeypair, d.Id()))
	userID, keypairName, err := resourceVPCKeypairV2ParseID(d.Id())
	if err != nil {
		return diag.FromErr(errParseID(objectKeypair, d.Id()))
	}
	existingKeypairs, _, err := keypairs.ListWithOpts(selvpcClient, keypairs.ListOpts{UserID: userID})
	if err != nil {
		return diag.FromErr(errSearchingKeypair(keypairName, err))
	}

	found := false
	for _, keypair := range existingKeypairs {
		if keypair.UserID == userID && keypair.Name == keypairName {
			found = true
			d.Set("name", keypair.Name)
			d.Set("public_key", keypair.PublicKey)
			d.Set("user_id", keypair.UserID)
		}
	}

	if !found {
		d.SetId("")
	}

	return nil
}

func resourceVPCKeypairV2Delete(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*Config)
	selvpcClient, err := config.GetSelVPCClient()
	if err != nil {
		return diag.FromErr(fmt.Errorf("can't get selvpc client for keypairs object: %w", err))
	}

	userID, keypairName, err := resourceVPCKeypairV2ParseID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	log.Print(msgDelete(objectKeypair, d.Id()))
	response, err := keypairs.Delete(selvpcClient, keypairName, userID)
	if err != nil {
		if response != nil {
			if response.StatusCode == http.StatusNotFound {
				d.SetId("")
				return nil
			}
		}

		return diag.FromErr(errDeletingObject(objectKeypair, d.Id(), err))
	}

	return nil
}

func resourceVPCKeypairV2BuildID(userID, keypairName string) string {
	return fmt.Sprintf("%s/%s", userID, keypairName)
}

func resourceVPCKeypairV2ParseID(id string) (string, string, error) {
	idParts := strings.Split(id, "/")
	if len(idParts) != 2 {
		return "", "", errParseID(objectKeypair, id)
	}

	return idParts[0], idParts[1], nil
}
