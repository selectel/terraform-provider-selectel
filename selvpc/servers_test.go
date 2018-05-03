package selvpc

import (
	"testing"

	"github.com/selectel/go-selvpcclient/selvpcclient/resell/v2/servers"
	"github.com/stretchr/testify/assert"
)

func TestServersMapsFromStructs(t *testing.T) {
	serverStructs := []servers.Server{
		{
			ID:     "a208023f-69fe-4a9e-8285-dd44e94a854a",
			Name:   "fake",
			Status: "ACTIVE",
		},
	}
	expectedServersMaps := []map[string]interface{}{
		{
			"id":     "a208023f-69fe-4a9e-8285-dd44e94a854a",
			"name":   "fake",
			"status": "ACTIVE",
		},
	}

	actualServersMaps := serversMapsFromStructs(serverStructs)

	assert.Equal(t, expectedServersMaps, actualServersMaps)
}
