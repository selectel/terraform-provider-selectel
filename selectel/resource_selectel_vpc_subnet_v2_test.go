package selectel

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDeprecatedVPCSubnetV2(t *testing.T) {
	deprecated := resourceDeprecatedVPCSubnetV2()
	canonical := resourceVPCPublicSubnetV2()

	assert.NotEmpty(t, deprecated.DeprecationMessage)
	assert.Equal(t, len(canonical.Schema), len(deprecated.Schema))
	for name := range canonical.Schema {
		assert.Contains(t, deprecated.Schema, name)
	}
}
