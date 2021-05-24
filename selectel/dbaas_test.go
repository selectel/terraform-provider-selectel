package selectel

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetDBaaSDatastoreV1Endpoint(t *testing.T) {
	expectedEndpoints := map[string]string{
		ru1Region: ru1DBaaSV1Endpoint,
		ru2Region: ru2DBaaSV1Endpoint,
		ru3Region: ru3DBaaSV1Endpoint,
		ru7Region: ru7DBaaSV1Endpoint,
		ru8Region: ru8DBaaSV1Endpoint,
		ru9Region: ru9DBaaSV1Endpoint,
	}

	for region, expected := range expectedEndpoints {
		actual := getDBaaSV1Endpoint(region)
		assert.Equal(t, expected, actual)
	}
}
