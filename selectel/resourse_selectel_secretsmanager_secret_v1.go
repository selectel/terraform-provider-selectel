package selectel

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/selectel/secretsmanager-go/secretsmanagererrors"
	"github.com/selectel/secretsmanager-go/service/secrets"
)

func resourceSecretsManagerSecretV1() *schema.Resource {
	return &schema.Resource{
		Description: "Creates and manages a secret in Secrets Manager using public API v1.",

		CreateContext: resourceSecretsManagerSecretV1Create,
		ReadContext:   resourceSecretsManagerSecretV1Read,
		UpdateContext: resourceSecretsManagerSecretV1Update,
		DeleteContext: resourceSecretsManagerSecretV1Delete,

		Importer: &schema.ResourceImporter{
			StateContext: resourceSecretsManagerSecretV1ImportState,
		},

		Schema: map[string]*schema.Schema{
			"key": {
				Description: "Secret name, unique within the project. Must contain 1 to 255 ASCII letters, digits, underscores, or hyphens. Changing this value recreates the secret and deletes all its versions.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"description": {
				Description: "Secret description. Defaults to an empty string. Maximum length is 1024 characters; each character must use at most three UTF-8 bytes. Changing only the description does not create a version.",
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    false,
			},
			"value": {
				Description: "Sensitive secret value as non-empty plain UTF-8 text, up to 262144 bytes. The SDK handles Base64 encoding. Changing this value creates a version; a positive current_version leaves the uploaded version inactive. Refresh and rollback preserve the last uploaded value. Import does not retrieve plaintext and sets an absent value to UNKNOWN; supply value in the configuration before applying.",
				Type:        schema.TypeString,
				Required:    true,
				Sensitive:   true,
				ForceNew:    false, // otherwise, will replace existing secret if you import it
			},
			"project_id": {
				Description: "Unique identifier of the associated project. Changing this value recreates the secret and deletes all its versions.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"name": {
				Description: "Secret name, equal to key.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"created_at": {
				Description: "Creation time of the active secret version in RFC3339 format.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"version": {
				Description: "Active secret version number, not necessarily the latest uploaded version.",
				Type:        schema.TypeInt,
				Computed:    true,
			},
			"current_version": {
				Description:  "Desired active version. Defaults to 0, which automatically activates new versions when value changes. Positive values pin an existing version. Must be non-negative; only 0 or 1 is valid on creation. Resetting to 0 does not immediately change the active version.",
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      0,
				ValidateFunc: validation.IntAtLeast(0),
			},
		},
	}
}

func resourceSecretsManagerSecretV1Create(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	if d.Get("current_version").(int) > 1 {
		return diag.Errorf("current_version must be 0 or 1 when creating a secret")
	}

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
	d.Set("version", 1)

	return resourceSecretsManagerSecretV1Read(ctx, d, meta)
}

func resourceSecretsManagerSecretV1Read(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
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
	d.Set("version", secret.Version.VersionID)
	if d.Get("current_version").(int) > 0 {
		d.Set("current_version", secret.Version.VersionID)
	}

	return nil
}

func resourceSecretsManagerSecretV1Delete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
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

func resourceSecretsManagerSecretV1Update(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	cl, diagErr := getSecretsManagerClient(d, meta)
	if diagErr != nil {
		d.Partial(true)
		return diagErr
	}

	key := d.Get("key").(string)
	if d.HasChange("description") {
		secret := secrets.UserSecret{
			Key:         key,
			Description: d.Get("description").(string),
		}
		log.Print(msgUpdate(objectSecret, d.Id(), "updated description"))
		if err := cl.Secrets.Update(ctx, secret); err != nil {
			d.Partial(true)
			return diag.FromErr(errUpdatingObject(objectSecret, d.Id(), err))
		}
	}

	currentVersion := d.Get("current_version").(int)
	if d.HasChange("value") {
		version, err := cl.Secrets.CreateVersion(ctx, key, d.Get("value").(string), currentVersion == 0)
		if err != nil {
			restoreSecretsManagerInputs(d, "value", "current_version")
			return diag.FromErr(errUpdatingObject(objectSecret, d.Id(), err))
		}
		if currentVersion == 0 {
			d.Set("version", version.VersionID)
			d.Set("created_at", version.CreatedAt)
		}
	}

	if currentVersion > 0 {
		secret, err := cl.Secrets.Get(ctx, key)
		if err != nil {
			restoreSecretsManagerInputs(d, "current_version")
			return diag.FromErr(errGettingObject(objectSecret, d.Id(), err))
		}
		d.Set("version", secret.Version.VersionID)
		d.Set("current_version", secret.Version.VersionID)
		d.Set("created_at", secret.Version.CreatedAt)
		if secret.Version.VersionID != uint(currentVersion) {
			activated, err := cl.Secrets.ActivateVersion(ctx, key, uint64(currentVersion))
			if errors.Is(err, secretsmanagererrors.ErrConflictStatusText) {
				refreshed := resourceSecretsManagerSecretV1Read(ctx, d, meta)
				if !refreshed.HasError() && d.Get("version").(int) == currentVersion {
					return nil
				}
			}
			if err != nil {
				return diag.FromErr(errUpdatingObject(objectSecret, d.Id(), err))
			}
			d.Set("version", activated.VersionID)
			d.Set("current_version", activated.VersionID)
			d.Set("created_at", activated.CreatedAt)
		}
	}

	return resourceSecretsManagerSecretV1Read(ctx, d, meta)
}

// resourceSecretsManagerSecretV1ImportState —  helper used in Importer: &schema.ResourceImporter
// to avoid difficulties occurred with required INFRA_PROJECT_ID env in
// resourceSecretsManagerSecretV1Read when uising schema.ImportStatePassthroughContext.
func resourceSecretsManagerSecretV1ImportState(ctx context.Context, d *schema.ResourceData, meta any) ([]*schema.ResourceData, error) {
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
