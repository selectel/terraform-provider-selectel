package servers

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/terraform-providers/terraform-provider-selectel/selectel/internal/httptest"
)

func TestPricePlans_FindOneByName(t *testing.T) {
	plans := PricePlans{
		&PricePlan{UUID: "1", Name: "Plan1"},
		&PricePlan{UUID: "2", Name: "Plan2"},
	}
	tests := []struct {
		name   string
		search string
		want   *PricePlan
	}{
		{"Found", "Plan1", &PricePlan{UUID: "1", Name: "Plan1"}},
		{"NotFound", "Unknown", nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := plans.FindOneByName(tt.search)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestServiceClient_PricePlans(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// Prepare
		body := `{
			"result": [
				{
					"uuid": "1",
					"name": "Plan1"
				}
			]
		}`
		fakeResp := httptest.NewFakeResponse(200, body) //nolint:bodyclose
		fakeTransport := httptest.NewFakeTransport(fakeResp, nil)
		client := newFakeClient("http://fake", fakeTransport)

		// Execute
		plans, respRes, err := client.PricePlans(context.Background())

		// Analyse
		require.NoError(t, err)
		require.NotNil(t, respRes)
		require.Equal(t, 200, respRes.StatusCode)
		wantPlans := PricePlans{
			&PricePlan{UUID: "1", Name: "Plan1"},
		}
		require.Equal(t, wantPlans, plans)
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
		plans, respRes, err := client.PricePlans(context.Background())

		// Analyse
		require.Error(t, err)
		require.Nil(t, plans)
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
		plans, respRes, err := client.PricePlans(context.Background())

		// Analyse
		require.Error(t, err)
		require.NotNil(t, respRes)
		require.NotNil(t, respRes.Err)
		require.Nil(t, plans)
		expectedErrMsg := httpErrorMessage
		require.EqualError(t, respRes.Err, expectedErrMsg)
	})

	t.Run("DoRequestError", func(t *testing.T) {
		// Prepare
		fakeTransport := httptest.NewFakeTransport(nil, errors.New("network failure"))
		client := newFakeClient("http://fake", fakeTransport)

		// Execute
		plans, respRes, err := client.PricePlans(context.Background())

		// Analyse
		require.Error(t, err)
		require.Nil(t, plans)
		require.Nil(t, respRes)
	})
}
