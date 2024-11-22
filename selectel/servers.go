package selectel

import (
	"fmt"

	"github.com/selectel/go-selvpcclient/v4/selvpcclient/resell/v2/servers"
	"github.com/terraform-providers/terraform-provider-selectel/selectel/internal/hashcode"
)

// serversMapsFromStructs converts the provided license.Servers to
// the slice of maps correspondingly to the resource's schema.
func serversMapsFromStructs(serverStructs []servers.Server) []map[string]interface{} {
	associatedServers := make([]map[string]interface{}, len(serverStructs))

	for i, server := range serverStructs {
		associatedServers[i] = map[string]interface{}{
			"id":     server.ID,
			"name":   server.Name,
			"status": server.Status,
		}
	}

	return associatedServers
}

// hashServers is a hash function to use with the "servers" set.
func hashServers(v interface{}) int {
	m := v.(map[string]interface{})
	return hashcode.String(fmt.Sprintf("%s-", m["id"].(string)))
}
