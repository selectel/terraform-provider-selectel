package selectel

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	globalrouter "github.com/selectel/globalrouter-go/pkg/v1"
)

var (
	errNotFound        = errors.New("resource not found")
	errMultipleResults = errors.New(
		"your query returned more than one result. please try a more specific search criteria")
)

func getGlobalRouterClient(meta interface{}) (*globalrouter.ServiceClient, diag.Diagnostics) {
	config := meta.(*Config)
	selvpcClient, err := config.GetSelVPCClient()
	if err != nil {
		return nil, diag.FromErr(fmt.Errorf("can't get account-scope selvpc client for global router api: %w", err))
	}

	globalrouterClient, err := globalrouter.NewClientV1(
		selvpcClient.GetXAuthToken(),
		globalrouter.WithAPIUrl("https://api.selectel.ru/naas/v1"),
		globalrouter.WithClientUserAgent(config.UserAgent),
	)
	if err != nil {
		return nil, diag.FromErr(err)
	}

	return globalrouterClient, nil
}

func getZoneByParams(ctx context.Context, client *globalrouter.ServiceClient, zoneName string, service string) (*globalrouter.Zone, error) {
	opts := globalrouter.ZonesQueryParams{
		Filters: globalrouter.ZonesFilters{
			Name: zoneName,
			// use always only enabled zone
			Enable: true,
		},
	}

	zones, _, err := client.ListZones(ctx, &opts)
	if err != nil {
		return nil, err
	}

	if service != "" {
		// Search by service name, because filter supports only service_id
		zonesWithTargetService := (*zones)[:0]

		for _, zone := range *zones {
			if zone.Service == service {
				zonesWithTargetService = append(zonesWithTargetService, zone)
			}
		}
		*zones = zonesWithTargetService
	}

	if len(*zones) < 1 {
		return nil, errGettingObject(objectGlobalRouterZone, zoneName, errNotFound)
	}

	if len(*zones) > 1 {
		log.Printf("[DEBUG] Multiple results found: %#v", zones)
		return nil, errGettingObject(objectGlobalRouterZone, zoneName, errMultipleResults)
	}

	return &(*zones)[0], nil
}

func getServiceByParams(ctx context.Context, client *globalrouter.ServiceClient, serviceName string) (*globalrouter.Service, error) {
	opts := globalrouter.ServicesQueryParams{
		Filters: globalrouter.ServicesFilters{
			Name: serviceName,
		},
	}

	services, _, err := client.ListServices(ctx, &opts)
	if err != nil {
		return nil, err
	}

	if len(*services) < 1 {
		return nil, errGettingObject(objectGlobalRouterService, serviceName, errNotFound)
	}

	if len(*services) > 1 {
		log.Printf("[DEBUG] Multiple results found: %#v", services)
		return nil, errGettingObject(objectGlobalRouterService, serviceName, errMultipleResults)
	}

	return &(*services)[0], nil
}

func getQuotaByParams(ctx context.Context, client *globalrouter.ServiceClient, quotaName string, scope string, scopeValue string) (*globalrouter.Quota, error) {
	opts := globalrouter.QuotasQueryParams{
		Filters: globalrouter.QuotasFilters{
			Name:       quotaName,
			Scope:      scope,
			ScopeValue: scopeValue,
		},
	}

	quotas, _, err := client.ListQuotas(ctx, &opts)
	if err != nil {
		return nil, err
	}

	if len(*quotas) < 1 {
		return nil, errGettingObject(objectGlobalRouterQuota, quotaName, errNotFound)
	}

	if len(*quotas) > 1 {
		log.Printf("[DEBUG] Multiple results found: %#v", quotas)
		return nil, errGettingObject(objectGlobalRouterQuota, quotaName, errMultipleResults)
	}

	return &(*quotas)[0], nil
}

func getZoneGroupByParams(ctx context.Context, client *globalrouter.ServiceClient, zoneGroupName string) (*globalrouter.ZoneGroup, error) {
	opts := globalrouter.ZoneGroupsQueryParams{
		Filters: globalrouter.ZoneGroupsFilters{
			Name: zoneGroupName,
		},
	}

	zoneGroups, _, err := client.ListZoneGroups(ctx, &opts)
	if err != nil {
		return nil, err
	}

	if len(*zoneGroups) < 1 {
		return nil, errGettingObject(objectGlobalRouterZoneGroup, zoneGroupName, errNotFound)
	}

	if len(*zoneGroups) > 1 {
		log.Printf("[DEBUG] Multiple results found: %#v", zoneGroups)
		return nil, errGettingObject(objectGlobalRouterZoneGroup, zoneGroupName, errMultipleResults)
	}

	return &(*zoneGroups)[0], nil
}

func expandToStringSlice(v []any) []string {
	var s []string
	for _, val := range v {
		if strVal, ok := val.(string); ok {
			s = append(s, strVal)
		}
	}

	return s
}

func flattenZoneGroupsV1(zoneGroups []globalrouter.ZoneGroup) []map[string]any {
	zgs := make([]map[string]any, len(zoneGroups))

	for i, zoneGroup := range zoneGroups {
		zgs[i] = map[string]any{
			"id":          zoneGroup.ID,
			"name":        zoneGroup.Name,
			"description": zoneGroup.Description,
			"created_at":  zoneGroup.CreatedAt,
			"updated_at":  zoneGroup.CreatedAt,
		}
	}

	return zgs
}
