package selectel

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/selectel/iam-go/iamerrors"
	"github.com/selectel/iam-go/service/ec2"
)

func resourceIAMEC2V1() *schema.Resource {
	return &schema.Resource{
		Description:   "Represents a EC2 Credentials in IAM API. Access Key is used as a resource ID.",
		CreateContext: resourceIAMEC2V1Create,
		ReadContext:   resourceIAMEC2V1Read,
		DeleteContext: resourceIAMEC2V1Delete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"user_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Service User ID to assign EC2 Credentials to.",
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Name of the EC2 Credentials.",
			},
			"project_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Project ID to associate EC2 Credentials with.",
			},
			"secret_key": {
				Type:        schema.TypeString,
				Computed:    true,
				Sensitive:   true,
				Description: "Secret Key of the EC2 Credentials.",
			},
		},
	}
}

func resourceIAMEC2V1Create(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	iamClient, diagErr := getIAMClient(meta)
	if diagErr != nil {
		return diagErr
	}

	log.Print(msgCreate(objectEC2Credentials, d.Id()))
	credential, err := iamClient.EC2.Create(
		ctx,
		d.Get("user_id").(string),
		d.Get("name").(string),
		d.Get("project_id").(string),
	)
	if err != nil {
		return diag.FromErr(errCreatingObject(objectEC2Credentials, err))
	}

	d.SetId(credential.AccessKey)
	d.Set("secret_key", credential.SecretKey)
	d.Set("name", credential.Name)
	d.Set("project_id", credential.ProjectID)

	return nil
}

func resourceIAMEC2V1Read(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	iamClient, diagErr := getIAMClient(meta)
	if diagErr != nil {
		return diagErr
	}

	log.Print(msgGet(objectEC2Credentials, d.Id()))
	credentials, err := iamClient.EC2.List(ctx, d.Get("user_id").(string))
	if err != nil {
		return diag.FromErr(errGettingObject(objectEC2Credentials, d.Id(), err))
	}

	var credential ec2.Credential
	for _, c := range credentials {
		if d.Id() == c.AccessKey {
			credential = c
			break
		}
	}
	if credential.AccessKey == "" {
		d.SetId("")
		return diag.FromErr(errGettingObject(objectEC2Credentials, d.Id(), fmt.Errorf("EC2 Credentials with ID %s not found", d.Id())))
	}

	d.Set("name", credential.Name)
	d.Set("project_id", credential.ProjectID)
	if _, ok := d.GetOk("secret_key"); !ok {
		d.Set("secret_key", importIAMUndefined)
	}

	return nil
}

func resourceIAMEC2V1Delete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	iamClient, diagErr := getIAMClient(meta)
	if diagErr != nil {
		return diagErr
	}

	log.Print(msgDelete(objectEC2Credentials, d.Id()))
	err := iamClient.EC2.Delete(ctx, d.Get("user_id").(string), d.Id())
	if err != nil && !errors.Is(err, iamerrors.ErrCredentialNotFound) {
		return diag.FromErr(errDeletingObject(objectEC2Credentials, d.Id(), err))
	}

	return nil
}
