package selectel

import (
	"errors"
	"net/http"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	vpc "github.com/selectel/vpc-go/pkg/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testVPCObject = "network"

func newTestAPIError(status int, class vpc.ErrorClass) error {
	return &vpc.APIError{
		StatusCode: status,
		Type:       "NeutronError",
		Message:    "test",
		Class:      class,
	}
}

func newTestTransportError() error {
	return &vpc.TransportError{Err: errors.New("dial tcp: i/o timeout")}
}

func newVPCTestData(t *testing.T) *schema.ResourceData {
	t.Helper()

	d := schema.TestResourceDataRaw(t, vpcImportTestSchema(), map[string]any{})
	d.SetId("2a1b3c4d-5e6f-7081-92a3-b4c5d6e7f809")

	return d
}

func TestVPCCheckDeleted(t *testing.T) {
	t.Run("not found on Read removes the resource from state without an error", func(t *testing.T) {
		d := newVPCTestData(t)

		diags := vpcCheckDeleted(d, newTestAPIError(http.StatusNotFound, vpc.ErrorClassNotFound), testVPCObject)

		assert.False(t, diags.HasError())
		assert.Empty(t, d.Id())
	})

	for _, c := range []struct {
		name string
		err  error
	}{
		{name: "forbidden", err: newTestAPIError(http.StatusForbidden, vpc.ErrorClassForbidden)},
		{name: "server error", err: newTestAPIError(http.StatusInternalServerError, vpc.ErrorClassServer)},
		{name: "transport failure", err: newTestTransportError()},
	} {
		t.Run(c.name+" on Read keeps the resource in state and reports an error", func(t *testing.T) {
			d := newVPCTestData(t)
			id := d.Id()

			diags := vpcCheckDeleted(d, c.err, testVPCObject)

			require.True(t, diags.HasError())
			assert.Equal(t, id, d.Id())
		})
	}

	t.Run("no error leaves the resource untouched", func(t *testing.T) {
		d := newVPCTestData(t)
		id := d.Id()

		diags := vpcCheckDeleted(d, nil, testVPCObject)

		assert.False(t, diags.HasError())
		assert.Equal(t, id, d.Id())
	})
}
