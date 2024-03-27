package selectel

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/selectel/iam-go/iamerrors"
	"github.com/selectel/iam-go/service/s3credentials"
)

func resourceIAMS3CredentialsV1() *schema.Resource {
	return &schema.Resource{
		Description:   "Represents a S3 Credentials in IAM API. Access Key is used as a resource ID.",
		CreateContext: resourceIAMS3CredentialsV1Create,
		ReadContext:   resourceIAMS3CredentialsV1Read,
		DeleteContext: resourceIAMS3CredentialsV1Delete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"user_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Service User ID to assign S3 Credentials to.",
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Name of the S3 Credentials.",
			},
			"project_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Project ID to associate S3 Credentials with.",
			},
			"secret_key": {
				Type:        schema.TypeString,
				Computed:    true,
				Sensitive:   true,
				Description: "Secret Key of the S3 Credentials.",
			},
		},
	}
}

func resourceIAMS3CredentialsV1Create(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	iamClient, diagErr := getIAMClient(meta)
	if diagErr != nil {
		return diagErr
	}

	log.Print(msgCreate(objectS3Credentials, d.Id()))
	credentials, err := iamClient.S3Credentials.Create(
		ctx,
		d.Get("user_id").(string),
		d.Get("name").(string),
		d.Get("project_id").(string),
	)
	if err != nil {
		return diag.FromErr(errCreatingObject(objectS3Credentials, err))
	}

	d.SetId(credentials.AccessKey)
	d.Set("secret_key", credentials.SecretKey)
	d.Set("name", credentials.Name)
	d.Set("project_id", credentials.ProjectID)

	return nil
}

func resourceIAMS3CredentialsV1Read(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	iamClient, diagErr := getIAMClient(meta)
	if diagErr != nil {
		return diagErr
	}

	log.Print(msgGet(objectS3Credentials, d.Id()))
	credentials, err := iamClient.S3Credentials.List(ctx, d.Get("user_id").(string))
	if err != nil {
		return diag.FromErr(errGettingObject(objectS3Credentials, d.Id(), err))
	}

	var credential s3credentials.Credentials
	for _, c := range credentials {
		if d.Id() == c.AccessKey {
			credential = c
			break
		}
	}
	if credential.AccessKey == "" {
		d.SetId("")
		return diag.FromErr(errGettingObject(objectS3Credentials, d.Id(), fmt.Errorf("S3 Credentials with ID %s not found", d.Id())))
	}

	d.Set("name", credential.Name)
	d.Set("project_id", credential.ProjectID)
	if _, ok := d.GetOk("secret_key"); !ok {
		d.Set("secret_key", importIAMUndefined)
	}

	return nil
}

func resourceIAMS3CredentialsV1Delete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	iamClient, diagErr := getIAMClient(meta)
	if diagErr != nil {
		return diagErr
	}

	log.Print(msgDelete(objectS3Credentials, d.Id()))
	err := iamClient.S3Credentials.Delete(ctx, d.Get("user_id").(string), d.Id())
	if err != nil && !errors.Is(err, iamerrors.ErrCredentialNotFound) {
		return diag.FromErr(errDeletingObject(objectS3Credentials, d.Id(), err))
	}

	return nil
}
