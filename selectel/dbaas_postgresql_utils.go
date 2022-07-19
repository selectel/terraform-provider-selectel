package selectel

import (
	"context"
	"errors"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/selectel/dbaas-go"
)

func resourceDBaaSPostgreSQLDatastoreV1PoolerFromSet(poolerSet *schema.Set) (*dbaas.Pooler, error) {
	if poolerSet.Len() == 0 {
		return nil, nil
	}
	var resourceModeRaw, resourceSizeRaw interface{}
	var ok bool

	resourcePoolerMap := poolerSet.List()[0].(map[string]interface{})
	if resourceModeRaw, ok = resourcePoolerMap["mode"]; !ok {
		return &dbaas.Pooler{}, errors.New("pooler.mode value isn't provided")
	}
	if resourceSizeRaw, ok = resourcePoolerMap["size"]; !ok {
		return &dbaas.Pooler{}, errors.New("pooler.size value isn't provided")
	}

	resourceMode := resourceModeRaw.(string)
	resourceSize := resourceSizeRaw.(int)

	pooler := &dbaas.Pooler{
		Mode: resourceMode,
		Size: resourceSize,
	}

	return pooler, nil
}

func resourceDBaaSPostgreSQLDatastoreV1PoolerOptsFromSet(poolerSet *schema.Set) (dbaas.DatastorePoolerOpts, error) {
	if poolerSet.Len() == 0 {
		return dbaas.DatastorePoolerOpts{}, nil
	}
	var resourceModeRaw, resourceSizeRaw interface{}
	var ok bool

	resourcePoolerMap := poolerSet.List()[0].(map[string]interface{})
	if resourceModeRaw, ok = resourcePoolerMap["mode"]; !ok {
		return dbaas.DatastorePoolerOpts{}, errors.New("pooler.mode value isn't provided")
	}
	if resourceSizeRaw, ok = resourcePoolerMap["size"]; !ok {
		return dbaas.DatastorePoolerOpts{}, errors.New("pooler.size value isn't provided")
	}

	resourceMode := resourceModeRaw.(string)
	resourceSize := resourceSizeRaw.(int)

	pooler := dbaas.DatastorePoolerOpts{
		Mode: resourceMode,
		Size: resourceSize,
	}

	return pooler, nil
}

func updatePostgreSQLDatastorePooler(ctx context.Context, d *schema.ResourceData, client *dbaas.API) error {
	poolerSet := d.Get("pooler").(*schema.Set)
	poolerOpts, err := resourceDBaaSPostgreSQLDatastoreV1PoolerOptsFromSet(poolerSet)
	if err != nil {
		return errParseDatastoreV1Pooler(err)
	}

	log.Print(msgUpdate(objectDatastore, d.Id(), poolerOpts))
	_, err = client.PoolerDatastore(ctx, d.Id(), poolerOpts)
	if err != nil {
		return errUpdatingObject(objectDatastore, d.Id(), err)
	}

	log.Printf("[DEBUG] waiting for datastore %s to become 'ACTIVE'", d.Id())
	timeout := d.Timeout(schema.TimeoutUpdate)
	err = waitForDBaaSDatastoreV1ActiveState(ctx, client, d.Id(), timeout)
	if err != nil {
		return errUpdatingObject(objectDatastore, d.Id(), err)
	}

	return nil
}
