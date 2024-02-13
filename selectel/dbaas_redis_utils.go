package selectel

import (
	"context"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/selectel/dbaas-go"
	waiters "github.com/terraform-providers/terraform-provider-selectel/selectel/waiters/dbaas"
)

func updateRedisDatastorePassword(ctx context.Context, d *schema.ResourceData, client *dbaas.API) error {
	passwordOpts := dbaas.DatastorePasswordOpts{
		RedisPassword: d.Get("redis_password").(string),
	}

	log.Print(msgUpdate(objectDatastore, d.Id(), passwordOpts))
	_, err := client.PasswordDatastore(ctx, d.Id(), passwordOpts)
	if err != nil {
		return errUpdatingObject(objectDatastore, d.Id(), err)
	}

	log.Printf("[DEBUG] waiting for datastore %s to become 'ACTIVE'", d.Id())
	timeout := d.Timeout(schema.TimeoutUpdate)
	err = waiters.WaitForDBaaSDatastoreV1ActiveState(ctx, client, d.Id(), timeout)
	if err != nil {
		return errUpdatingObject(objectDatastore, d.Id(), err)
	}

	return nil
}

func resizeRedisDatastore(ctx context.Context, d *schema.ResourceData, client *dbaas.API) error {
	var resizeOpts dbaas.DatastoreResizeOpts
	nodeCount := d.Get("node_count").(int)
	resizeOpts.NodeCount = nodeCount

	flavorID := d.Get("flavor_id")

	resizeOpts.Flavor = nil
	resizeOpts.FlavorID = flavorID.(string)

	log.Print(msgUpdate(objectDatastore, d.Id(), resizeOpts))
	_, err := client.ResizeDatastore(ctx, d.Id(), resizeOpts)
	if err != nil {
		return errUpdatingObject(objectDatastore, d.Id(), err)
	}

	log.Printf("[DEBUG] waiting for datastore %s to become 'ACTIVE'", d.Id())
	timeout := d.Timeout(schema.TimeoutCreate)
	err = waiters.WaitForDBaaSDatastoreV1ActiveState(ctx, client, d.Id(), timeout)
	if err != nil {
		return errUpdatingObject(objectDatastore, d.Id(), err)
	}

	return nil
}
