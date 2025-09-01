package servers

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/terraform-providers/terraform-provider-selectel/selectel/internal/httptest"
)

func TestLocations_FindOneByName(t *testing.T) {
	locs := Locations{
		&Location{UUID: "1", Name: "Loc1"},
		&Location{UUID: "2", Name: "Loc2"},
	}
	tests := []struct {
		name   string
		search string
		want   *Location
	}{
		{"Found", "Loc1", &Location{UUID: "1", Name: "Loc1"}},
		{"NotFound", "Unknown", nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := locs.FindOneByName(tt.search)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestServiceClient_Locations(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// Prepare
		body := `{
			"result": [
				{
					"uuid": "1",
					"name": "Loc1"
				}
			]
		}`
		fakeResp := httptest.NewFakeResponse(200, body) //nolint:bodyclose
		fakeTransport := httptest.NewFakeTransport(fakeResp, nil)
		client := newFakeClient("http://fake", fakeTransport)

		// Execute
		locs, respRes, err := client.Locations(context.Background())

		// Analyse
		require.NoError(t, err)
		require.NotNil(t, respRes)
		require.Equal(t, 200, respRes.StatusCode)
		wantLocs := Locations{
			&Location{UUID: "1", Name: "Loc1"},
		}
		require.Equal(t, wantLocs, locs)
	})

	t.Run("InvalidJSON", func(t *testing.T) {
		// Prepare
		body := `{
			"result": [
				invalid json
			]
		}`
		fakeResp := httptest.NewFakeResponse(200, body) //nolint:bodyclose
		fakeTransport := httptest.NewFakeTransport(fakeResp, nil)
		client := newFakeClient("http://fake", fakeTransport)

		// Execute
		locs, respRes, err := client.Locations(context.Background())

		// Analyse
		require.Error(t, err)
		require.Nil(t, locs)
		require.NotNil(t, respRes)
		require.Equal(t, 200, respRes.StatusCode)
	})

	t.Run("HTTPError", func(t *testing.T) {
		// Prepare
		body := `Not Found`
		fakeResp := httptest.NewFakeResponse(404, body) //nolint:bodyclose
		fakeTransport := httptest.NewFakeTransport(fakeResp, nil)
		client := newFakeClient("http://fake", fakeTransport)

		// Execute
		locs, respRes, err := client.Locations(context.Background())

		// Analyse
		require.Error(t, err)
		require.NotNil(t, respRes)
		require.NotNil(t, respRes.Err)
		require.Nil(t, locs)
		expectedErrMsg := httpErrorMessage
		require.EqualError(t, respRes.Err, expectedErrMsg)
	})

	t.Run("DoRequestError", func(t *testing.T) {
		// Prepare
		fakeTransport := httptest.NewFakeTransport(nil, errors.New("network failure"))
		client := newFakeClient("http://fake", fakeTransport)

		// Execute
		locs, respRes, err := client.Locations(context.Background())

		// Analyse
		require.Error(t, err)
		require.Nil(t, locs)
		require.Nil(t, respRes)
	})
}
