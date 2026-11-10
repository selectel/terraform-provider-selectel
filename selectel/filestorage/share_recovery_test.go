package filestorage

import (
	"net/http"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const proxyPage = `<html><body>502 Bad Gateway</body></html>`

func TestShareCreateRecovery(t *testing.T) {
	for name, lost := range map[string]answer{
		"server error":      {http.StatusInternalServerError, `{"computeFault": {"code": 500, "message": "boom"}}`},
		"rate limit":        {http.StatusTooManyRequests, `{"overLimit": {"code": 429, "message": "slow down"}}`},
		"unstructured 4xx":  {http.StatusBadRequest, proxyPage},
		"proxy error":       {http.StatusBadGateway, proxyPage},
		"empty 4xx":         {http.StatusBadRequest, ""},
		"listed after wait": {http.StatusGatewayTimeout, ""},
	} {
		t.Run(name+" finds the share by token", func(t *testing.T) {
			fake := newFakeAPI(t)
			fake.lostCreate = &lost
			if name == "listed after wait" {
				fake.hiddenLists = 2
			}
			fake.statuses = []string{"creating", "available"}
			r := fake.resource()

			state, diags := create(t, r, createPlan())

			require.False(t, diags.HasError(), "%v", diags)
			assert.Equal(t, shareID, readShareModel(t, state).ID.ValueString())
			assert.Equal(t, 1, fake.count("POST /shares"))
		})
	}

	t.Run("a share that does not appear stays pending", func(t *testing.T) {
		fake := newFakeAPI(t)
		fake.rejectCreate = &answer{http.StatusInternalServerError, ""}
		r := fake.resource()

		state, diags := create(t, r, createPlan())

		require.True(t, diags.HasError())
		got := readShareModel(t, state)
		token, pending := strings.CutPrefix(got.ID.ValueString(), pendingPrefix)
		require.True(t, pending, got.ID.ValueString())
		detail := diags.Errors()[0].Detail()
		assert.Contains(t, detail, token)
		assert.Contains(t, detail, "terraform untaint")
		assert.Contains(t, detail, "terraform state rm")
		assert.Equal(t, 1, fake.count("POST /shares"))
	})

	t.Run("a structured rejection does not search", func(t *testing.T) {
		fake := newFakeAPI(t)
		fake.rejectCreate = &answer{
			http.StatusRequestEntityTooLarge,
			`{"overLimit": {"code": 413, "message": "Requested share exceeds allowed project quota.", "retryAfter": "0"}}`,
		}
		r := fake.resource()

		state, diags := create(t, r, createPlan())

		require.True(t, diags.HasError())
		assert.Contains(t, diags.Errors()[0].Detail(), "quota")
		assert.True(t, state.Raw.IsNull())
		assert.Zero(t, fake.count("GET /shares/detail"))
	})
}

// pendingState is a share whose create request got no definite answer.
func pendingState(token string) shareModel {
	m := shareState()
	m.ID = types.StringValue(pendingPrefix + token)
	m.IP = types.StringValue("")
	m.Status = types.StringValue("")
	m.AvailabilityZone = types.StringValue("")
	m.AllMetadata = types.MapNull(types.StringType)
	m.ExportLocations = types.ListNull(types.ObjectType{AttrTypes: exportLocationTypes})

	return m
}

func TestSharePendingRecovery(t *testing.T) {
	const token = "token-1"

	t.Run("refresh replaces the pending ID", func(t *testing.T) {
		fake := newFakeAPI(t)
		fake.share["metadata"].(map[string]any)[createTokenKey] = token
		r := fake.resource()

		state, diags := read(t, r, pendingState(token))

		require.False(t, diags.HasError(), "%v", diags)
		got := readShareModel(t, state)
		assert.Equal(t, shareID, got.ID.ValueString())
		assert.Equal(t, "10.0.0.5", got.IP.ValueString())
		// The token is the provider's key: it never enters metadata, so the next
		// plan matches the configuration.
		assert.True(t, got.Metadata.IsNull())
	})

	t.Run("refresh without the share keeps the state", func(t *testing.T) {
		fake := newFakeAPI(t)
		r := fake.resource()

		state, diags := read(t, r, pendingState(token))

		require.True(t, diags.HasError())
		assert.Contains(t, diags.Errors()[0].Detail(), token)
		assert.Contains(t, diags.Errors()[0].Detail(), "terraform state rm")
		assert.Equal(t, pendingPrefix+token, readShareModel(t, state).ID.ValueString())
	})

	t.Run("destroy deletes the share found by token", func(t *testing.T) {
		fake := newFakeAPI(t)
		fake.share["metadata"].(map[string]any)[createTokenKey] = token
		fake.statuses = []string{"available", "deleted"}
		r := fake.resource()

		diags := destroy(t, r, pendingState(token))

		require.False(t, diags.HasError(), "%v", diags)
		assert.Equal(t, []string{"DELETE /shares/" + shareID}, fake.writes())
	})

	t.Run("destroy without the share deletes nothing", func(t *testing.T) {
		fake := newFakeAPI(t)
		r := fake.resource()

		diags := destroy(t, r, pendingState(token))

		require.True(t, diags.HasError())
		assert.Contains(t, diags.Errors()[0].Detail(), token)
		assert.Empty(t, fake.writes())
	})
}
