package selectel

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/selectel/secretsmanager-go/service/secrets"
)

func resourceSecretsManagerSecretV1() *schema.Resource {
	return &schema.Resource{
		Description: "represents a Secret — entity from SecretsManager service",

		CreateContext: resourceSecretsManagerSecretV1Create,
		ReadContext:   resourceSecretsManagerSecretV1Read,
		UpdateContext: resourceSecretsManagerSecretV1Update,
		DeleteContext: resourceSecretsManagerSecretV1Delete,

		Importer: &schema.ResourceImporter{
			StateContext: resourceSecretsManagerSecretV1ImportState,
		},

		Schema: map[string]*schema.Schema{
			"key": {
				Description: "unique key,name of the secret",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"description": {
				Description: "optional description of the secret",
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    false,
			},
			"value": {
				Description: "secret value, e.g. password, API key, certificate key, or other",
				Type:        schema.TypeString,
				Required:    true,
				Sensitive:   true,
				ForceNew:    false, // otherwise, will replace existing secret if you import it
			},
			"project_id": {
				Description: "id of a project where secret is used",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"name": {
				Description: "computed name of the secret same as key",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"created_at": {
				Description: "time when the secret was created",
				Type:        schema.TypeString,
				Computed:    true,
			},
		},
	}
}

func resourceSecretsManagerSecretV1Create(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cl, diagErr := getSecretsManagerClient(d, meta)
	if diagErr != nil {
		return diagErr
	}

	key := d.Get("key").(string)
	desc := d.Get("description").(string)
	value := d.Get("value").(string)

	secret := secrets.UserSecret{
		Key:         key,
		Description: desc,
		Value:       value,
	}

	log.Print(msgCreate(objectSecret, secret.Key))

	errCr := cl.Secrets.Create(ctx, secret)
	if errCr != nil {
		return diag.FromErr(errCreatingObject(objectSecret, errCr))
	}

	projectID := d.Get("project_id").(string)
	d.SetId(resourceSecretV1BuildID(projectID, key))

	return resourceSecretsManagerSecretV1Read(ctx, d, meta)
}

func resourceSecretsManagerSecretV1Read(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cl, diagErr := getSecretsManagerClient(d, meta)
	if diagErr != nil {
		return diagErr
	}

	log.Print(msgGet(objectSecret, d.Id()))

	_, key, err := resourceSecretsManagerSecretV1ParseID(d.Id())
	if err != nil {
		return diag.FromErr(errParseID(objectSecret, d.Id()))
	}

	secret, errGet := cl.Secrets.Get(ctx, key)
	if errGet != nil {
		return diag.FromErr(errGettingObject(objectSecret, d.Id(), errGet))
	}

	d.Set("name", secret.Name)
	d.Set("key", secret.Name)
	d.Set("description", secret.Description)
	if _, ok := d.GetOk("value"); !ok {
		d.Set("value", "UNKNOWN")
	}
	d.Set("created_at", secret.Version.CreatedAt)

	return nil
}

func resourceSecretsManagerSecretV1Delete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cl, diagErr := getSecretsManagerClient(d, meta)
	if diagErr != nil {
		return diagErr
	}

	key := d.Get("key").(string)

	log.Print(msgDelete(objectSecret, d.Id()))

	errDel := cl.Secrets.Delete(ctx, key)
	if errDel != nil {
		return diag.FromErr(errDeletingObject(objectSecret, d.Id(), errDel))
	}

	return nil
}

func resourceSecretsManagerSecretV1Update(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cl, diagErr := getSecretsManagerClient(d, meta)
	if diagErr != nil {
		return diagErr
	}

	key := d.Get("key").(string)
	desc := d.Get("description").(string)

	secret := secrets.UserSecret{
		Key:         key,
		Description: desc,
	}

	log.Print(msgUpdate(objectSecret, d.Id(), secret))

	errUpd := cl.Secrets.Update(ctx, secret)
	if errUpd != nil {
		return diag.FromErr(errUpdatingObject(objectSecret, d.Id(), errUpd))
	}

	return resourceSecretsManagerSecretV1Read(ctx, d, meta)
}

// resourceSecretsManagerSecretV1ImportState —  helper used in Importer: &schema.ResourceImporter
// to avoid difficulties occurred with required INFRA_PROJECT_ID env in
// resourceSecretsManagerSecretV1Read when uising schema.ImportStatePassthroughContext.
func resourceSecretsManagerSecretV1ImportState(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	config := meta.(*Config)
	if config.ProjectID == "" {
		return nil, errors.New("INFRA_PROJECT_ID must be set for the resource import")
	}

	d.Set("project_id", config.ProjectID)

	_, key, err := resourceSecretsManagerSecretV1ParseID(d.Id())
	if err != nil {
		return nil, errParseID(objectSecret, d.Id())
	}

	log.Print(msgImport(objectSecret, key))
	resourceSecretsManagerSecretV1Read(ctx, d, meta)

	return []*schema.ResourceData{d}, nil
}

// resourceSecretV1BuildID — helper that builds ID that is going to be set
// in Secret resource, to prevent cases where several VPC projects has same key.
func resourceSecretV1BuildID(projectID, key string) string {
	return fmt.Sprintf("%s/%s", projectID, key)
}

// resourceSecretsManagerSecretV1ParseID — helper that separates Project ID and key
// from resource ID that was set using resourceSecretV1BuildID.
func resourceSecretsManagerSecretV1ParseID(id string) (string, string, error) {
	idParts := strings.Split(id, "/")
	if len(idParts) != 2 {
		return "", "", errParseID(objectSecret, id)
	}

	return idParts[0], idParts[1], nil
}
