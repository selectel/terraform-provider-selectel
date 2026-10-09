package selectel

import (
	"context"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/selectel/dbaas-go"
	waiters "github.com/terraform-providers/terraform-provider-selectel/selectel/waiters/dbaas"
)

func updateDBaaSUserV1Settings(ctx context.Context, d *schema.ResourceData, client *dbaas.API) error {
	user, err := client.User(ctx, d.Id())
	if err != nil {
		return err
	}

	settings := expandDBaaSUserV1Settings(d)
	if settings == nil {
		settings = map[string]any{}
	}
	for key := range user.Settings {
		if _, ok := settings[key]; !ok {
			settings[key] = nil
		}
	}

	// The API rejects an empty settings map, and there is nothing to
	// set or unset when both the desired state and the current
	// server-side settings are empty.
	if len(settings) == 0 {
		return nil
	}

	opts := dbaas.UserSettingsUpdateOpts{Settings: settings}
	log.Print(msgUpdate(objectUser, d.Id(), opts))
	_, err = client.UpdateUserSettings(ctx, d.Id(), opts)
	if err != nil {
		return errUpdatingObject(objectUser, d.Id(), err)
	}

	log.Printf("[DEBUG] waiting for user %s to become 'ACTIVE'", d.Id())
	timeout := d.Timeout(schema.TimeoutUpdate)
	err = waiters.WaitForDBaaSUserV1ActiveState(ctx, client, d.Id(), timeout)
	if err != nil {
		return errUpdatingObject(objectUser, d.Id(), err)
	}

	return nil
}

func expandDBaaSUserV1Settings(d *schema.ResourceData) map[string]any {
	if _, ok := d.GetOk("settings"); !ok {
		return nil
	}

	return d.Get("settings").(map[string]any)
}

func flattenDBaaSUserV1Settings(apiSettings map[string]any) map[string]any {
	settings := make(map[string]any, len(apiSettings))
	for key, value := range apiSettings {
		settings[key] = convertFieldToStringByType(value)
	}

	return settings
}
