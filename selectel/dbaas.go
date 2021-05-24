package selectel

import (
	"context"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/selectel/dbaas-go"
	"github.com/selectel/go-selvpcclient/selvpcclient/resell/v2/tokens"
)

const (
	ru1DBaaSV1Endpoint = "https://ru-1.dbaas.selcloud.ru/v1"
	ru2DBaaSV1Endpoint = "https://ru-2.dbaas.selcloud.ru/v1"
	ru3DBaaSV1Endpoint = "https://ru-3.dbaas.selcloud.ru/v1"
	ru7DBaaSV1Endpoint = "https://ru-7.dbaas.selcloud.ru/v1"
	ru8DBaaSV1Endpoint = "https://ru-8.dbaas.selcloud.ru/v1"
	ru9DBaaSV1Endpoint = "https://ru-9.dbaas.selcloud.ru/v1"
)

func getDBaaSV1Endpoint(region string) (endpoint string) {
	switch region {
	case ru1Region:
		endpoint = ru1DBaaSV1Endpoint
	case ru2Region:
		endpoint = ru2DBaaSV1Endpoint
	case ru3Region:
		endpoint = ru3DBaaSV1Endpoint
	case ru7Region:
		endpoint = ru7DBaaSV1Endpoint
	case ru8Region:
		endpoint = ru8DBaaSV1Endpoint
	case ru9Region:
		endpoint = ru9DBaaSV1Endpoint
	}

	return
}

func getDBaaSClient(ctx context.Context, d *schema.ResourceData, meta interface{}) (*dbaas.API, diag.Diagnostics) {
	config := meta.(*Config)
	resellV2Client := config.resellV2Client()
	tokenOpts := tokens.TokenOpts{
		ProjectID: d.Get("project_id").(string),
	}

	log.Print(msgCreate(objectToken, tokenOpts))
	token, _, err := tokens.Create(ctx, resellV2Client, tokenOpts)
	if err != nil {
		return nil, diag.FromErr((errCreatingObject(objectToken, err)))
	}

	region := d.Get("region").(string)
	endpoint := getDBaaSV1Endpoint(region)
	client, err := dbaas.NewDBAASClient(token.ID, endpoint)
	if err != nil {
		return nil, diag.FromErr(err)
	}
	return client, nil
}
