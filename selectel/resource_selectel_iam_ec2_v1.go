package selectel

import (
	"context"
	"errors"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/selectel/iam-go/iamerrors"
	"github.com/selectel/iam-go/service/ec2"
)

func resourceIAMEC2V1() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceIAMEC2V1Create,
		ReadContext:   resourceIAMEC2V1Read,
		DeleteContext: resourceIAMEC2V1Delete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"user_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"project_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"access_key": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"secret_key": {
				Type:     schema.TypeString,
				Computed: true,
				Sensitive: true,
			},
		},
	}
}

func resourceIAMEC2V1Create(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	iamClient, diagErr := getIAMClient(meta)
	if diagErr != nil {
		return diagErr
	}

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

	return resourceIAMEC2V1Read(ctx, d, meta)
}

func resourceIAMEC2V1Read(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	iamClient, diagErr := getIAMClient(meta)
	if diagErr != nil {
		return diagErr
	}

	log.Print(msgGet(objectEC2Credentials, d.Id()))
	credentials, _ := iamClient.EC2.List(ctx, d.Get("user_id").(string))

	var credential ec2.Credential
	for _, c := range credentials {
		if d.Id() == c.AccessKey {
			credential = c
			break
		}
	}
	d.Set("access_key", credential.AccessKey)

	return nil
}

func resourceIAMEC2V1Delete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	iamClient, diagErr := getIAMClient(meta)
	if diagErr != nil {
		return diagErr
	}

	log.Print(msgDelete(objectEC2Credentials, d.Id()))
	err := iamClient.EC2.Delete(ctx, d.Get("user_id").(string), d.Id())
	if err != nil {
		if errors.Is(err, iamerrors.ErrCredentialNotFound) {
			d.SetId("")
			return nil
		}

		return diag.FromErr(errDeletingObject(objectEC2Credentials, d.Id(), err))
	}

	return nil
}
