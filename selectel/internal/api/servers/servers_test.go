package servers

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/terraform-providers/terraform-provider-selectel/selectel/internal/httptest"
)

func TestServiceClient_Servers(t *testing.T) {
	t.Run("Server_Success", func(t *testing.T) {
		// Prepare
		body := `{
			"result": [{
				"name": "server1"
			}]
		}`
		fakeResp := httptest.NewFakeResponse(200, body) //nolint:bodyclose
		fakeTransport := httptest.NewFakeTransport(fakeResp, nil)
		client := newFakeClient("http://fake", fakeTransport)

		// Execute
		svrs, respRes, err := client.Servers(context.Background(), false)

		// Analyse
		require.NoError(t, err)
		require.NotNil(t, respRes)
		require.Equal(t, 200, respRes.StatusCode)
		wantSvrs := Servers{
			&Server{Name: "server1"},
		}
		require.Equal(t, wantSvrs, svrs)
	})

	t.Run("Server_InvalidJSON", func(t *testing.T) {
		// Prepare
		body := invalidJSONBody
		fakeResp := httptest.NewFakeResponse(200, body) //nolint:bodyclose
		fakeTransport := httptest.NewFakeTransport(fakeResp, nil)
		client := newFakeClient("http://fake", fakeTransport)

		// Execute
		svrs, respRes, err := client.Servers(context.Background(), false)

		// Analyse
		require.Error(t, err)
		require.Nil(t, svrs)
		require.NotNil(t, respRes)
		require.Equal(t, 200, respRes.StatusCode)
	})

	t.Run("Server_HTTPError", func(t *testing.T) {
		// Prepare
		body := httpErrorBody
		fakeResp := httptest.NewFakeResponse(404, body) //nolint:bodyclose
		fakeTransport := httptest.NewFakeTransport(fakeResp, nil)
		client := newFakeClient("http://fake", fakeTransport)

		// Execute
		svrs, respRes, err := client.Servers(context.Background(), false)

		// Analyse
		require.Error(t, err)
		require.NotNil(t, respRes)
		require.NotNil(t, respRes.Err)
		require.Nil(t, svrs)
		require.EqualError(t, respRes.Err, httpErrorMessage)
	})

	t.Run("Server_DoRequestError", func(t *testing.T) {
		// Prepare
		fakeTransport := httptest.NewFakeTransport(nil, errors.New("network failure"))
		client := newFakeClient("http://fake", fakeTransport)

		// Execute
		svrs, respRes, err := client.Servers(context.Background(), false)

		// Analyse
		require.Error(t, err)
		require.Nil(t, svrs)
		require.Nil(t, respRes)
	})

	t.Run("ServerChip_Success", func(t *testing.T) {
		// Prepare
		body := `{
			"result": [{
				"name": "chip1"
			}]
		}`
		fakeResp := httptest.NewFakeResponse(200, body) //nolint:bodyclose
		fakeTransport := httptest.NewFakeTransport(fakeResp, nil)
		client := newFakeClient("http://fake", fakeTransport)

		// Execute
		svrs, respRes, err := client.Servers(context.Background(), true)

		// Analyse
		require.NoError(t, err)
		require.NotNil(t, respRes)
		require.Equal(t, 200, respRes.StatusCode)
		wantSvrs := Servers{
			&Server{Name: "chip1"},
		}
		require.Equal(t, wantSvrs, svrs)
	})

	t.Run("ServerChip_InvalidJSON", func(t *testing.T) {
		// Prepare
		body := invalidJSONBody
		fakeResp := httptest.NewFakeResponse(200, body) //nolint:bodyclose
		fakeTransport := httptest.NewFakeTransport(fakeResp, nil)
		client := newFakeClient("http://fake", fakeTransport)

		// Execute
		svrs, respRes, err := client.Servers(context.Background(), true)

		// Analyse
		require.Error(t, err)
		require.Nil(t, svrs)
		require.NotNil(t, respRes)
		require.Equal(t, 200, respRes.StatusCode)
	})

	t.Run("ServerChip_HTTPError", func(t *testing.T) {
		// Prepare
		body := httpErrorBody
		fakeResp := httptest.NewFakeResponse(404, body) //nolint:bodyclose
		fakeTransport := httptest.NewFakeTransport(fakeResp, nil)
		client := newFakeClient("http://fake", fakeTransport)

		// Execute
		svrs, respRes, err := client.Servers(context.Background(), true)

		// Analyse
		require.Error(t, err)
		require.NotNil(t, respRes)
		require.NotNil(t, respRes.Err)
		require.Nil(t, svrs)
		require.EqualError(t, respRes.Err, httpErrorMessage)
	})

	t.Run("ServerChip_DoRequestError", func(t *testing.T) {
		// Prepare
		fakeTransport := httptest.NewFakeTransport(nil, errors.New("network failure"))
		client := newFakeClient("http://fake", fakeTransport)

		// Execute
		svrs, respRes, err := client.Servers(context.Background(), true)

		// Analyse
		require.Error(t, err)
		require.Nil(t, svrs)
		require.Nil(t, respRes)
	})
}

func TestServiceClient_ServerByID(t *testing.T) {
	t.Run("Server_Success", func(t *testing.T) {
		body := `{"result": {"uuid": "123", "name": "server1"}}`
		fakeResp := httptest.NewFakeResponse(200, body) //nolint:bodyclose
		client := newFakeClient("http://fake", httptest.NewFakeTransport(fakeResp, nil))

		svr, respRes, err := client.ServerByID(context.Background(), "123", false)
		require.NoError(t, err)
		require.NotNil(t, respRes)
		require.Equal(t, "123", svr.ID)
		require.Equal(t, "server1", svr.Name)
	})

	t.Run("Server_InvalidJSON", func(t *testing.T) {
		body := invalidJSONBody
		fakeResp := httptest.NewFakeResponse(200, body) //nolint:bodyclose
		client := newFakeClient("http://fake", httptest.NewFakeTransport(fakeResp, nil))

		svr, respRes, err := client.ServerByID(context.Background(), "123", false)
		require.Error(t, err)
		require.Nil(t, svr)
		require.NotNil(t, respRes)
	})

	t.Run("Server_HTTPError", func(t *testing.T) {
		// Prepare
		body := httpErrorBody
		fakeResp := httptest.NewFakeResponse(404, body) //nolint:bodyclose
		fakeTransport := httptest.NewFakeTransport(fakeResp, nil)
		client := newFakeClient("http://fake", fakeTransport)

		// Execute
		svr, respRes, err := client.ServerByID(context.Background(), "123", false)

		// Analyse
		require.Error(t, err)
		require.NotNil(t, respRes)
		require.NotNil(t, respRes.Err)
		require.Nil(t, svr)
		require.EqualError(t, respRes.Err, httpErrorMessage)
	})

	t.Run("Server_DoRequestError", func(t *testing.T) {
		client := newFakeClient("http://fake", httptest.NewFakeTransport(nil, errors.New("network failure")))

		svr, respRes, err := client.ServerByID(context.Background(), "123", false)
		require.Error(t, err)
		require.Nil(t, svr)
		require.Nil(t, respRes)
	})

	t.Run("ServerChip_Success", func(t *testing.T) {
		body := `{"result": {"uuid": "chipid", "name": "chip1"}}`
		fakeResp := httptest.NewFakeResponse(200, body) //nolint:bodyclose
		client := newFakeClient("http://fake", httptest.NewFakeTransport(fakeResp, nil))

		svr, respRes, err := client.ServerByID(context.Background(), "chipid", true)
		require.NoError(t, err)
		require.NotNil(t, respRes)
		require.Equal(t, "chipid", svr.ID)
		require.Equal(t, "chip1", svr.Name)
	})

	t.Run("ServerChip_InvalidJSON", func(t *testing.T) {
		body := invalidJSONBody
		fakeResp := httptest.NewFakeResponse(200, body) //nolint:bodyclose
		client := newFakeClient("http://fake", httptest.NewFakeTransport(fakeResp, nil))

		svr, respRes, err := client.ServerByID(context.Background(), "chipid", true)
		require.Error(t, err)
		require.Nil(t, svr)
		require.NotNil(t, respRes)
	})

	t.Run("ServerChip_HTTPError", func(t *testing.T) {
		// Prepare
		body := httpErrorBody
		fakeResp := httptest.NewFakeResponse(404, body) //nolint:bodyclose
		fakeTransport := httptest.NewFakeTransport(fakeResp, nil)
		client := newFakeClient("http://fake", fakeTransport)

		// Execute
		svr, respRes, err := client.ServerByID(context.Background(), "chipid", true)

		// Analyse
		require.Error(t, err)
		require.NotNil(t, respRes)
		require.NotNil(t, respRes.Err)
		require.Nil(t, svr)
		require.EqualError(t, respRes.Err, httpErrorMessage)
	})

	t.Run("ServerChip_DoRequestError", func(t *testing.T) {
		client := newFakeClient("http://fake", httptest.NewFakeTransport(nil, errors.New("network failure")))

		svr, respRes, err := client.ServerByID(context.Background(), "chipid", true)
		require.Error(t, err)
		require.Nil(t, svr)
		require.Nil(t, respRes)
	})
}

func TestServiceClient_ServerCalculateBilling(t *testing.T) {
	t.Run("Server_Success", func(t *testing.T) {
		body := `{"result": {"has_enough_balance": true}}`
		fakeResp := httptest.NewFakeResponse(200, body) //nolint:bodyclose
		client := newFakeClient("http://fake", httptest.NewFakeTransport(fakeResp, nil))

		billing, respRes, err := client.ServerCalculateBilling(context.Background(), "sid", "locid", "planid", "main", false)
		require.NoError(t, err)
		require.NotNil(t, respRes)
		require.True(t, billing.HasEnoughBalance)
	})

	t.Run("Server_InvalidJSON", func(t *testing.T) {
		body := invalidJSONBody
		fakeResp := httptest.NewFakeResponse(200, body) //nolint:bodyclose
		client := newFakeClient("http://fake", httptest.NewFakeTransport(fakeResp, nil))

		billing, respRes, err := client.ServerCalculateBilling(context.Background(), "sid", "locid", "planid", "main", false)
		require.Error(t, err)
		require.Nil(t, billing)
		require.NotNil(t, respRes)
	})

	t.Run("Server_HTTPError", func(t *testing.T) {
		// Prepare
		body := httpErrorBody
		fakeResp := httptest.NewFakeResponse(404, body) //nolint:bodyclose
		fakeTransport := httptest.NewFakeTransport(fakeResp, nil)
		client := newFakeClient("http://fake", fakeTransport)

		// Execute
		billing, respRes, err := client.ServerCalculateBilling(context.Background(), "sid", "locid", "planid", "main", false)

		// Analyse
		require.Error(t, err)
		require.NotNil(t, respRes)
		require.NotNil(t, respRes.Err)
		require.Nil(t, billing)
		require.EqualError(t, respRes.Err, httpErrorMessage)
	})

	t.Run("Server_DoRequestError", func(t *testing.T) {
		client := newFakeClient("http://fake", httptest.NewFakeTransport(nil, errors.New("network failure")))

		billing, respRes, err := client.ServerCalculateBilling(context.Background(), "sid", "locid", "planid", "main", false)
		require.Error(t, err)
		require.Nil(t, billing)
		require.Nil(t, respRes)
	})

	t.Run("ServerChip_Success", func(t *testing.T) {
		body := `{"result": {"has_enough_balance": false}}`
		fakeResp := httptest.NewFakeResponse(200, body) //nolint:bodyclose
		client := newFakeClient("http://fake", httptest.NewFakeTransport(fakeResp, nil))

		billing, respRes, err := client.ServerCalculateBilling(context.Background(), "sid", "locid", "planid", "main", true)
		require.NoError(t, err)
		require.NotNil(t, respRes)
		require.False(t, billing.HasEnoughBalance)
	})

	t.Run("ServerChip_InvalidJSON", func(t *testing.T) {
		body := invalidJSONBody
		fakeResp := httptest.NewFakeResponse(200, body) //nolint:bodyclose
		client := newFakeClient("http://fake", httptest.NewFakeTransport(fakeResp, nil))

		billing, respRes, err := client.ServerCalculateBilling(context.Background(), "sid", "locid", "planid", "main", true)
		require.Error(t, err)
		require.Nil(t, billing)
		require.NotNil(t, respRes)
	})

	t.Run("ServerChip_HTTPError", func(t *testing.T) {
		// Prepare
		body := httpErrorBody
		fakeResp := httptest.NewFakeResponse(404, body) //nolint:bodyclose
		fakeTransport := httptest.NewFakeTransport(fakeResp, nil)
		client := newFakeClient("http://fake", fakeTransport)

		// Execute
		billing, respRes, err := client.ServerCalculateBilling(context.Background(), "sid", "locid", "planid", "main", true)

		// Analyse
		require.Error(t, err)
		require.NotNil(t, respRes)
		require.NotNil(t, respRes.Err)
		require.Nil(t, billing)
		require.EqualError(t, respRes.Err, httpErrorMessage)
	})

	t.Run("ServerChip_DoRequestError", func(t *testing.T) {
		client := newFakeClient("http://fake", httptest.NewFakeTransport(nil, errors.New("network failure")))

		billing, respRes, err := client.ServerCalculateBilling(context.Background(), "sid", "locid", "planid", "main", true)
		require.Error(t, err)
		require.Nil(t, billing)
		require.Nil(t, respRes)
	})
}

func TestServiceClient_ServerBilling(t *testing.T) {
	t.Run("Server_Success", func(t *testing.T) {
		body := `{"result": [{"uuid": "some-uuid"}]}`
		fakeResp := httptest.NewFakeResponse(200, body) //nolint:bodyclose
		client := newFakeClient("http://fake", httptest.NewFakeTransport(fakeResp, nil))

		req := &ServerBillingPostPayload{
			LocationUUID:  "locid",
			PricePlanUUID: "planid",
			PayCurrency:   "main",
			Quantity:      1,
		}

		billings, respRes, err := client.ServerBilling(context.Background(), req, false)
		require.NoError(t, err)
		require.NotNil(t, respRes)
		require.Equal(t, "some-uuid", billings[0].UUID)
	})

	t.Run("Server_InvalidJSON", func(t *testing.T) {
		body := invalidJSONBody
		fakeResp := httptest.NewFakeResponse(200, body) //nolint:bodyclose
		client := newFakeClient("http://fake", httptest.NewFakeTransport(fakeResp, nil))

		req := &ServerBillingPostPayload{
			LocationUUID:  "locid",
			PricePlanUUID: "planid",
			PayCurrency:   "main",
			Quantity:      1,
		}

		billing, respRes, err := client.ServerBilling(context.Background(), req, false)
		require.Error(t, err)
		require.Nil(t, billing)
		require.NotNil(t, respRes)
	})

	t.Run("Server_HTTPError", func(t *testing.T) {
		// Prepare
		body := httpErrorBody
		fakeResp := httptest.NewFakeResponse(404, body) //nolint:bodyclose
		fakeTransport := httptest.NewFakeTransport(fakeResp, nil)
		client := newFakeClient("http://fake", fakeTransport)

		req := &ServerBillingPostPayload{
			LocationUUID:  "locid",
			PricePlanUUID: "planid",
			PayCurrency:   "main",
			Quantity:      1,
		}

		// Execute
		billing, respRes, err := client.ServerBilling(context.Background(), req, false)

		// Analyse
		require.Error(t, err)
		require.NotNil(t, respRes)
		require.NotNil(t, respRes.Err)
		require.Nil(t, billing)
		require.EqualError(t, respRes.Err, httpErrorMessage)
	})

	t.Run("Server_DoRequestError", func(t *testing.T) {
		client := newFakeClient("http://fake", httptest.NewFakeTransport(nil, errors.New("network failure")))

		req := &ServerBillingPostPayload{
			LocationUUID:  "locid",
			PricePlanUUID: "planid",
			PayCurrency:   "main",
			Quantity:      1,
		}

		billing, respRes, err := client.ServerBilling(context.Background(), req, false)
		require.Error(t, err)
		require.Nil(t, billing)
		require.Nil(t, respRes)
	})

	t.Run("ServerChip_Success", func(t *testing.T) {
		body := `{"result": [{"uuid": "some-uuid"}]}`
		fakeResp := httptest.NewFakeResponse(200, body) //nolint:bodyclose
		client := newFakeClient("http://fake", httptest.NewFakeTransport(fakeResp, nil))

		req := &ServerBillingPostPayload{
			LocationUUID:  "locid",
			PricePlanUUID: "planid",
			PayCurrency:   "main",
			Quantity:      1,
		}

		billings, respRes, err := client.ServerBilling(context.Background(), req, true)
		require.NoError(t, err)
		require.NotNil(t, respRes)
		require.Equal(t, "some-uuid", billings[0].UUID)
	})

	t.Run("ServerChip_InvalidJSON", func(t *testing.T) {
		body := invalidJSONBody
		fakeResp := httptest.NewFakeResponse(200, body) //nolint:bodyclose
		client := newFakeClient("http://fake", httptest.NewFakeTransport(fakeResp, nil))

		req := &ServerBillingPostPayload{
			LocationUUID:  "locid",
			PricePlanUUID: "planid",
			PayCurrency:   "main",
			Quantity:      1,
		}

		billing, respRes, err := client.ServerBilling(context.Background(), req, true)
		require.Error(t, err)
		require.Nil(t, billing)
		require.NotNil(t, respRes)
	})

	t.Run("ServerChip_HTTPError", func(t *testing.T) {
		// Prepare
		body := httpErrorBody
		fakeResp := httptest.NewFakeResponse(404, body) //nolint:bodyclose
		fakeTransport := httptest.NewFakeTransport(fakeResp, nil)
		client := newFakeClient("http://fake", fakeTransport)

		req := &ServerBillingPostPayload{
			LocationUUID:  "locid",
			PricePlanUUID: "planid",
			PayCurrency:   "main",
			Quantity:      1,
		}

		// Execute
		billing, respRes, err := client.ServerBilling(context.Background(), req, true)

		// Analyse
		require.Error(t, err)
		require.NotNil(t, respRes)
		require.NotNil(t, respRes.Err)
		require.Nil(t, billing)
		require.EqualError(t, respRes.Err, httpErrorMessage)
	})

	t.Run("ServerChip_DoRequestError", func(t *testing.T) {
		client := newFakeClient("http://fake", httptest.NewFakeTransport(nil, errors.New("network failure")))

		req := &ServerBillingPostPayload{
			LocationUUID:  "locid",
			PricePlanUUID: "planid",
			PayCurrency:   "main",
			Quantity:      1,
		}

		billing, respRes, err := client.ServerBilling(context.Background(), req, true)
		require.Error(t, err)
		require.Nil(t, billing)
		require.Nil(t, respRes)
	})
}

func TestServiceClient_ResourceDetails(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		body := `{"result": {"uuid": "some-uuid", "state": "some-state"}}`
		fakeResp := httptest.NewFakeResponse(200, body) //nolint:bodyclose
		client := newFakeClient("http://fake", httptest.NewFakeTransport(fakeResp, nil))

		details, respRes, err := client.ResourceDetails(context.Background(), "some-uuid")
		require.NoError(t, err)
		require.NotNil(t, respRes)
		require.Equal(t, &ResourceDetails{
			UUID:  "some-uuid",
			State: "some-state",
		}, details)
	})

	t.Run("InvalidJSON", func(t *testing.T) {
		body := invalidJSONBody
		fakeResp := httptest.NewFakeResponse(200, body) //nolint:bodyclose
		client := newFakeClient("http://fake", httptest.NewFakeTransport(fakeResp, nil))

		details, respRes, err := client.ResourceDetails(context.Background(), "resid")
		require.Error(t, err)
		require.Nil(t, details)
		require.NotNil(t, respRes)
	})

	t.Run("HTTPError", func(t *testing.T) {
		// Prepare
		body := httpErrorBody
		fakeResp := httptest.NewFakeResponse(404, body) //nolint:bodyclose
		fakeTransport := httptest.NewFakeTransport(fakeResp, nil)
		client := newFakeClient("http://fake", fakeTransport)

		// Execute
		details, respRes, err := client.ResourceDetails(context.Background(), "resid")

		// Analyse
		require.Error(t, err)
		require.NotNil(t, respRes)
		require.NotNil(t, respRes.Err)
		require.Nil(t, details)
		require.EqualError(t, respRes.Err, httpErrorMessage)
	})

	t.Run("DoRequestError", func(t *testing.T) {
		client := newFakeClient("http://fake", httptest.NewFakeTransport(nil, errors.New("network failure")))

		details, respRes, err := client.ResourceDetails(context.Background(), "resid")
		require.Error(t, err)
		require.Nil(t, details)
		require.Nil(t, respRes)
	})
}

func TestServiceClient_DeleteResource(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// Prepare
		body := `{"result": {}}`
		fakeResp := httptest.NewFakeResponse(200, body) //nolint:bodyclose
		client := newFakeClient("http://fake", httptest.NewFakeTransport(fakeResp, nil))

		// Execute
		respRes, err := client.DeleteResource(context.Background(), "resourceid")

		// Analyse
		require.NoError(t, err)
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
		respRes, err := client.DeleteResource(context.Background(), "resourceid")

		// Analyse
		require.Error(t, err)
		require.NotNil(t, respRes)
		require.NotNil(t, respRes.Err)
		require.EqualError(t, respRes.Err, httpErrorMessage)
	})

	t.Run("DoRequestError", func(t *testing.T) {
		// Prepare
		fakeTransport := httptest.NewFakeTransport(nil, errors.New("network failure"))
		client := newFakeClient("http://fake", fakeTransport)

		// Execute
		respRes, err := client.DeleteResource(context.Background(), "resourceid")

		// Analyse
		require.Error(t, err)
		require.Nil(t, respRes)
	})
}
