package selectel

import (
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	vpc "github.com/selectel/vpc-go/pkg/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testNetworkingObject = "network"

// newTestAPIError builds an SDK API error of the given class.
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

func TestNetworkingErrorPredicates(t *testing.T) {
	cases := []struct {
		name      string
		err       error
		notFound  bool
		forbidden bool
		conflict  bool
		badReq    bool
		server    bool
		transport bool
	}{
		{name: "not found", err: newTestAPIError(http.StatusNotFound, vpc.ErrorClassNotFound), notFound: true},
		{name: "forbidden", err: newTestAPIError(http.StatusForbidden, vpc.ErrorClassForbidden), forbidden: true},
		{
			name: "not authenticated", err: newTestAPIError(http.StatusUnauthorized, vpc.ErrorClassNotAuthenticated),
			forbidden: true,
		},
		{name: "conflict", err: newTestAPIError(http.StatusConflict, vpc.ErrorClassConflict), conflict: true},
		{name: "bad request", err: newTestAPIError(http.StatusBadRequest, vpc.ErrorClassBadRequest), badReq: true},
		{name: "server", err: newTestAPIError(http.StatusInternalServerError, vpc.ErrorClassServer), server: true},
		{name: "transport", err: newTestTransportError(), transport: true},
		{name: "plain error", err: errors.New("something else")},
		{
			name:     "wrapped not found",
			err:      fmt.Errorf("error getting network: %w", newTestAPIError(http.StatusNotFound, vpc.ErrorClassNotFound)),
			notFound: true,
		},
		{
			name:      "wrapped transport",
			err:       fmt.Errorf("read failed: %w", newTestTransportError()),
			transport: true,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			assert.Equal(t, c.notFound, networkingIsNotFound(c.err), "notFound")
			assert.Equal(t, c.forbidden, networkingIsForbidden(c.err), "forbidden")
			assert.Equal(t, c.conflict, networkingIsConflict(c.err), "conflict")
			assert.Equal(t, c.badReq, networkingIsBadRequest(c.err), "badRequest")
			assert.Equal(t, c.server, networkingIsServerError(c.err), "serverError")
			assert.Equal(t, c.transport, networkingIsTransportError(c.err), "transportError")
		})
	}

	t.Run("nil error matches no predicate", func(t *testing.T) {
		assert.False(t, networkingIsNotFound(nil))
		assert.False(t, networkingIsForbidden(nil))
		assert.False(t, networkingIsConflict(nil))
		assert.False(t, networkingIsBadRequest(nil))
		assert.False(t, networkingIsServerError(nil))
		assert.False(t, networkingIsTransportError(nil))
	})
}

func newNetworkingTestData(t *testing.T) *schema.ResourceData {
	t.Helper()

	d := schema.TestResourceDataRaw(t, networkingImportTestSchema(), map[string]interface{}{})
	d.SetId("2a1b3c4d-5e6f-7081-92a3-b4c5d6e7f809")

	return d
}

func TestNetworkingCheckDeleted(t *testing.T) {
	t.Run("not found on Read removes the resource from state without an error", func(t *testing.T) {
		d := newNetworkingTestData(t)

		diags := networkingCheckDeleted(d, newTestAPIError(http.StatusNotFound, vpc.ErrorClassNotFound), testNetworkingObject)

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
			d := newNetworkingTestData(t)
			id := d.Id()

			diags := networkingCheckDeleted(d, c.err, testNetworkingObject)

			require.True(t, diags.HasError())
			assert.Equal(t, id, d.Id())
		})
	}

	t.Run("no error leaves the resource untouched", func(t *testing.T) {
		d := newNetworkingTestData(t)
		id := d.Id()

		diags := networkingCheckDeleted(d, nil, testNetworkingObject)

		assert.False(t, diags.HasError())
		assert.Equal(t, id, d.Id())
	})
}

func TestNetworkingCheckDeletedOnDelete(t *testing.T) {
	notFound := newTestAPIError(http.StatusNotFound, vpc.ErrorClassNotFound)

	t.Run("not found confirmed absent is an idempotent success", func(t *testing.T) {
		d := newNetworkingTestData(t)
		confirmed := false

		diags := networkingCheckDeletedOnDelete(d, notFound, testNetworkingObject, func() error {
			confirmed = true

			return notFound
		})

		assert.False(t, diags.HasError())
		assert.True(t, confirmed, "the confirming read must run")
	})

	t.Run("not found while the object still exists is reported as an error", func(t *testing.T) {
		d := newNetworkingTestData(t)

		diags := networkingCheckDeletedOnDelete(d, notFound, testNetworkingObject, func() error {
			// A successful read means the object is still there, so the
			// not-found from Delete was a masked policy failure.
			return nil
		})

		require.True(t, diags.HasError())
		assert.Contains(t, diags[0].Summary, "denied by policy")
		assert.NotEmpty(t, d.Id())
	})

	t.Run("not found whose confirmation fails is reported as an error", func(t *testing.T) {
		d := newNetworkingTestData(t)

		diags := networkingCheckDeletedOnDelete(d, notFound, testNetworkingObject, func() error {
			return newTestAPIError(http.StatusForbidden, vpc.ErrorClassForbidden)
		})

		require.True(t, diags.HasError())
		assert.Contains(t, diags[0].Summary, "confirming that failed")
	})

	t.Run("conflict on Delete reports an error without confirming", func(t *testing.T) {
		d := newNetworkingTestData(t)
		confirmed := false

		diags := networkingCheckDeletedOnDelete(d,
			newTestAPIError(http.StatusConflict, vpc.ErrorClassConflict), testNetworkingObject,
			func() error {
				confirmed = true

				return nil
			})

		require.True(t, diags.HasError())
		assert.False(t, confirmed, "a conflict needs no confirming read")
	})

	t.Run("no error is a plain success", func(t *testing.T) {
		d := newNetworkingTestData(t)

		diags := networkingCheckDeletedOnDelete(d, nil, testNetworkingObject, func() error {
			return errors.New("must not be called")
		})

		assert.False(t, diags.HasError())
	})
}
