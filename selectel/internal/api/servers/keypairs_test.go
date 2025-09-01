package servers

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/terraform-providers/terraform-provider-selectel/selectel/internal/httptest"
)

func TestServiceClient_SSHKeys(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// Prepare
		body := `{
			"result": [{
				"name_public_key": "key1",
				"public_key": "ssh-rsa AAA..."
			}]
		}`
		fakeResp := httptest.NewFakeResponse(200, body) //nolint:bodyclose
		fakeTransport := httptest.NewFakeTransport(fakeResp, nil)
		client := newFakeClient("http://fake", fakeTransport)

		// Execute
		keys, respRes, err := client.SSHKeys(context.Background())

		// Analyse
		require.NoError(t, err)
		require.NotNil(t, respRes)
		require.Equal(t, 200, respRes.StatusCode)
		wantKeys := SSHKeys{
			&SSHKey{
				Name:      "key1",
				PublicKey: "ssh-rsa AAA...",
			},
		}
		require.Equal(t, wantKeys, keys)
	})

	t.Run("InvalidJSON", func(t *testing.T) {
		// Prepare
		body := invalidJSONBody
		fakeResp := httptest.NewFakeResponse(200, body) //nolint:bodyclose
		fakeTransport := httptest.NewFakeTransport(fakeResp, nil)
		client := newFakeClient("http://fake", fakeTransport)

		// Execute
		keys, respRes, err := client.SSHKeys(context.Background())

		// Analyse
		require.Error(t, err)
		require.Nil(t, keys)
		require.NotNil(t, respRes)
		require.Equal(t, 200, respRes.StatusCode)
	})

	t.Run("HTTPError", func(t *testing.T) {
		// Prepare
		body := httpErrorBody
		fakeResp := httptest.NewFakeResponse(404, body) //nolint:bodyclose
		fakeTransport := httptest.NewFakeTransport(fakeResp, nil)
		client := newFakeClient("http://fake", fakeTransport)

		// Execute
		keys, respRes, err := client.SSHKeys(context.Background())

		// Analyse
		require.Error(t, err)
		require.NotNil(t, respRes)
		require.NotNil(t, respRes.Err)
		require.Nil(t, keys)
		require.EqualError(t, respRes.Err, httpErrorMessage)
	})

	t.Run("DoRequestError", func(t *testing.T) {
		// Prepare
		fakeTransport := httptest.NewFakeTransport(nil, errors.New("network failure"))
		client := newFakeClient("http://fake", fakeTransport)

		// Execute
		keys, respRes, err := client.SSHKeys(context.Background())

		// Analyse
		require.Error(t, err)
		require.Nil(t, keys)
		require.Nil(t, respRes)
	})
}

func TestSSHKeys_FindOneByName(t *testing.T) {
	t.Run("KeyFound", func(t *testing.T) {
		keys := SSHKeys{
			&SSHKey{Name: "key1"},
			&SSHKey{Name: "key2"},
		}

		key := keys.FindOneByName("key1")
		require.NotNil(t, key)
		require.Equal(t, "key1", key.Name)
	})

	t.Run("KeyNotFound", func(t *testing.T) {
		keys := SSHKeys{
			&SSHKey{Name: "key1"},
			&SSHKey{Name: "key2"},
		}

		key := keys.FindOneByName("key3")
		require.Nil(t, key)
	})
}
