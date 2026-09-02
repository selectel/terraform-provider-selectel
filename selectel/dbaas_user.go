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

	settings := make(map[string]any)
	for _, key := range dbaasUserSettingsKeys {
		rawKey := "settings.0." + key
		if v, exists := d.GetOkExists(rawKey); exists { //nolint:staticcheck // GetOkExists distinguishes unset from zero values
			settings[key] = v
		}
	}

	return settings
}

func flattenDBaaSUserV1Settings(apiSettings map[string]any) []any {
	if len(apiSettings) == 0 {
		return nil
	}

	settings := make(map[string]any)
	knownKeys := make(map[string]struct{}, len(dbaasUserSettingsKeys))
	for _, key := range dbaasUserSettingsKeys {
		knownKeys[key] = struct{}{}
	}
	for key, value := range apiSettings {
		if _, ok := knownKeys[key]; !ok || value == nil {
			continue
		}
		settings[key] = coerceDBaaSUserV1SettingValue(value)
	}
	if len(settings) == 0 {
		return nil
	}

	return []any{settings}
}

func coerceDBaaSUserV1SettingValue(value any) any {
	switch typed := value.(type) {
	case float64:
		return int(typed)
	case float32:
		return int(typed)
	default:
		return value
	}
}
