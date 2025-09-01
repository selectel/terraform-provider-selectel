package servers

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/terraform-providers/terraform-provider-selectel/selectel/internal/httptest"
)

func TestServiceClient_Networks(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// Prepare
		body := `{
			"result": [
				{
					"uuid": "net1",
					"telematic_type": "INET"
				}
			]
		}`
		fakeResp := httptest.NewFakeResponse(200, body) //nolint:bodyclose
		client := newFakeClient("http://fake", httptest.NewFakeTransport(fakeResp, nil))

		// Execute
		nets, respRes, err := client.Networks(context.Background(), "locid", "inet")

		// Validate
		require.NoError(t, err)
		require.NotNil(t, respRes)
		require.Equal(t, "net1", nets[0].UUID)
	})

	t.Run("InvalidJSON", func(t *testing.T) {
		// Prepare
		body := invalidJSONBody
		fakeResp := httptest.NewFakeResponse(200, body) //nolint:bodyclose
		client := newFakeClient("http://fake", httptest.NewFakeTransport(fakeResp, nil))

		// Execute
		nets, respRes, err := client.Networks(context.Background(), "locid", "inet")

		// Validate
		require.Error(t, err)
		require.Nil(t, nets)
		require.NotNil(t, respRes)
	})

	t.Run("HTTPError", func(t *testing.T) {
		// Prepare
		body := httpErrorBody
		fakeResp := httptest.NewFakeResponse(404, body) //nolint:bodyclose
		client := newFakeClient("http://fake", httptest.NewFakeTransport(fakeResp, nil))

		// Execute
		nets, respRes, err := client.Networks(context.Background(), "locid", "inet")

		// Validate
		require.Error(t, err)
		require.NotNil(t, respRes)
		require.NotNil(t, respRes.Err)
		require.Nil(t, nets)
		require.EqualError(t, respRes.Err, httpErrorMessage)
	})

	t.Run("DoRequestError", func(t *testing.T) {
		// Prepare
		client := newFakeClient("http://fake", httptest.NewFakeTransport(nil, errors.New("network failure")))

		// Execute
		nets, respRes, err := client.Networks(context.Background(), "locid", "inet")

		// Validate
		require.Error(t, err)
		require.Nil(t, nets)
		require.Nil(t, respRes)
	})
}

func TestServiceClient_NetworkSubnets(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		body := `{
			"result": [
				{
					"network_uuid": "net1",
					"subnet": "192.168.1.0/24",
					"free": 1
				}
			]
		}`
		fakeResp := httptest.NewFakeResponse(200, body) //nolint:bodyclose
		client := newFakeClient("http://fake", httptest.NewFakeTransport(fakeResp, nil))

		subnets, respRes, err := client.NetworkSubnets(context.Background(), "locid")
		require.NoError(t, err)
		require.NotNil(t, respRes)
		require.Equal(t, "net1", subnets[0].NetworkUUID)
	})

	t.Run("InvalidJSON", func(t *testing.T) {
		body := invalidJSONBody
		fakeResp := httptest.NewFakeResponse(200, body) //nolint:bodyclose
		client := newFakeClient("http://fake", httptest.NewFakeTransport(fakeResp, nil))

		subnets, respRes, err := client.NetworkSubnets(context.Background(), "locid")
		require.Error(t, err)
		require.Nil(t, subnets)
		require.NotNil(t, respRes)
	})

	t.Run("HTTPError", func(t *testing.T) {
		// Prepare
		body := httpErrorBody
		fakeResp := httptest.NewFakeResponse(404, body) //nolint:bodyclose
		client := newFakeClient("http://fake", httptest.NewFakeTransport(fakeResp, nil))

		// Execute
		subnets, respRes, err := client.NetworkSubnets(context.Background(), "locid")

		// Validate
		require.Error(t, err)
		require.NotNil(t, respRes)
		require.NotNil(t, respRes.Err)
		require.Nil(t, subnets)
		require.EqualError(t, respRes.Err, httpErrorMessage)
	})

	t.Run("DoRequestError", func(t *testing.T) {
		client := newFakeClient("http://fake", httptest.NewFakeTransport(nil, errors.New("network failure")))

		subnets, respRes, err := client.NetworkSubnets(context.Background(), "locid")
		require.Error(t, err)
		require.Nil(t, subnets)
		require.Nil(t, respRes)
	})
}

func TestServiceClient_NetworkReservedIPs(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		body := `{
			"result": [
				{
					"ip": "192.168.1.10"
				}
			]
		}`
		fakeResp := httptest.NewFakeResponse(200, body) //nolint:bodyclose
		client := newFakeClient("http://fake", httptest.NewFakeTransport(fakeResp, nil))

		ips, respRes, err := client.NetworkReservedIPs(context.Background(), "locid")
		require.NoError(t, err)
		require.NotNil(t, respRes)
		require.Equal(t, "192.168.1.10", ips[0].IP.String())
	})

	t.Run("InvalidJSON", func(t *testing.T) {
		body := invalidJSONBody
		fakeResp := httptest.NewFakeResponse(200, body) //nolint:bodyclose
		client := newFakeClient("http://fake", httptest.NewFakeTransport(fakeResp, nil))

		ips, respRes, err := client.NetworkReservedIPs(context.Background(), "locid")
		require.Error(t, err)
		require.Nil(t, ips)
		require.NotNil(t, respRes)
	})

	t.Run("HTTPError", func(t *testing.T) {
		// Prepare
		body := httpErrorBody
		fakeResp := httptest.NewFakeResponse(404, body) //nolint:bodyclose
		client := newFakeClient("http://fake", httptest.NewFakeTransport(fakeResp, nil))

		// Execute
		ips, respRes, err := client.NetworkReservedIPs(context.Background(), "locid")

		// Validate
		require.Error(t, err)
		require.NotNil(t, respRes)
		require.NotNil(t, respRes.Err)
		require.Nil(t, ips)
		require.EqualError(t, respRes.Err, httpErrorMessage)
	})

	t.Run("DoRequestError", func(t *testing.T) {
		client := newFakeClient("http://fake", httptest.NewFakeTransport(nil, errors.New("network failure")))

		ips, respRes, err := client.NetworkReservedIPs(context.Background(), "locid")
		require.Error(t, err)
		require.Nil(t, ips)
		require.Nil(t, respRes)
	})
}

func TestServiceClient_NetworkLocalSubnets(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		body := `{
			"result": [
				{
					"network_uuid": "net1",
					"subnet": "192.168.1.0/24",
					"free": 1
				}
			]
		}`
		fakeResp := httptest.NewFakeResponse(200, body) //nolint:bodyclose
		client := newFakeClient("http://fake", httptest.NewFakeTransport(fakeResp, nil))

		subnets, respRes, err := client.NetworkLocalSubnets(context.Background(), "netid")
		require.NoError(t, err)
		require.NotNil(t, respRes)
		require.Equal(t, "net1", subnets[0].NetworkUUID)
	})

	t.Run("InvalidJSON", func(t *testing.T) {
		body := invalidJSONBody
		fakeResp := httptest.NewFakeResponse(200, body) //nolint:bodyclose
		client := newFakeClient("http://fake", httptest.NewFakeTransport(fakeResp, nil))

		subnets, respRes, err := client.NetworkLocalSubnets(context.Background(), "netid")
		require.Error(t, err)
		require.Nil(t, subnets)
		require.NotNil(t, respRes)
	})

	t.Run("HTTPError", func(t *testing.T) {
		// Prepare
		body := httpErrorBody
		fakeResp := httptest.NewFakeResponse(404, body) //nolint:bodyclose
		client := newFakeClient("http://fake", httptest.NewFakeTransport(fakeResp, nil))

		// Execute
		subnets, respRes, err := client.NetworkLocalSubnets(context.Background(), "netid")

		// Validate
		require.Error(t, err)
		require.NotNil(t, respRes)
		require.NotNil(t, respRes.Err)
		require.Nil(t, subnets)
		require.EqualError(t, respRes.Err, httpErrorMessage)
	})

	t.Run("DoRequestError", func(t *testing.T) {
		client := newFakeClient("http://fake", httptest.NewFakeTransport(nil, errors.New("network failure")))

		subnets, respRes, err := client.NetworkLocalSubnets(context.Background(), "netid")
		require.Error(t, err)
		require.Nil(t, subnets)
		require.Nil(t, respRes)
	})
}

func TestServiceClient_NetworkSubnet(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		body := `{"result": {"uuid": "subnet1", "subnet": "192.168.1.0/24"}}`
		fakeResp := httptest.NewFakeResponse(200, body) //nolint:bodyclose
		client := newFakeClient("http://fake", httptest.NewFakeTransport(fakeResp, nil))

		subnet, respRes, err := client.NetworkSubnet(context.Background(), "subnetid")
		require.NoError(t, err)
		require.NotNil(t, respRes)
		require.Equal(t, "subnet1", subnet.UUID)
	})

	t.Run("InvalidJSON", func(t *testing.T) {
		body := `{"result": invalid}`
		fakeResp := httptest.NewFakeResponse(200, body) //nolint:bodyclose
		client := newFakeClient("http://fake", httptest.NewFakeTransport(fakeResp, nil))

		subnet, respRes, err := client.NetworkSubnet(context.Background(), "subnetid")
		require.Error(t, err)
		require.Nil(t, subnet)
		require.NotNil(t, respRes)
	})

	t.Run("HTTPError", func(t *testing.T) {
		// Prepare
		body := httpErrorBody
		fakeResp := httptest.NewFakeResponse(404, body) //nolint:bodyclose
		client := newFakeClient("http://fake", httptest.NewFakeTransport(fakeResp, nil))

		// Execute
		subnet, respRes, err := client.NetworkSubnet(context.Background(), "subnetid")

		// Analyse
		require.Error(t, err)
		require.NotNil(t, respRes)
		require.NotNil(t, respRes.Err)
		require.Nil(t, subnet)
		require.EqualError(t, respRes.Err, httpErrorMessage)
	})

	t.Run("DoRequestError", func(t *testing.T) {
		client := newFakeClient("http://fake", httptest.NewFakeTransport(nil, errors.New("network failure")))

		subnet, respRes, err := client.NetworkSubnet(context.Background(), "subnetid")
		require.Error(t, err)
		require.Nil(t, subnet)
		require.Nil(t, respRes)
	})
}

func TestServiceClient_NetworkSubnetLocalReservedIPs(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		body := `{
			"result": [
				{
					"ip": "192.168.1.10"
				}
			]
		}`
		fakeResp := httptest.NewFakeResponse(200, body) //nolint:bodyclose
		client := newFakeClient("http://fake", httptest.NewFakeTransport(fakeResp, nil))

		ips, respRes, err := client.NetworkSubnetLocalReservedIPs(context.Background(), "subnetid")
		require.NoError(t, err)
		require.NotNil(t, respRes)
		require.Equal(t, "192.168.1.10", ips[0].IP.String())
	})

	t.Run("InvalidJSON", func(t *testing.T) {
		body := invalidJSONBody
		fakeResp := httptest.NewFakeResponse(200, body) //nolint:bodyclose
		client := newFakeClient("http://fake", httptest.NewFakeTransport(fakeResp, nil))

		ips, respRes, err := client.NetworkSubnetLocalReservedIPs(context.Background(), "subnetid")
		require.Error(t, err)
		require.Nil(t, ips)
		require.NotNil(t, respRes)
	})

	t.Run("HTTPError", func(t *testing.T) {
		// Prepare
		body := httpErrorBody
		fakeResp := httptest.NewFakeResponse(404, body) //nolint:bodyclose
		client := newFakeClient("http://fake", httptest.NewFakeTransport(fakeResp, nil))

		// Execute
		ips, respRes, err := client.NetworkSubnetLocalReservedIPs(context.Background(), "subnetid")

		// Analyse
		require.Error(t, err)
		require.NotNil(t, respRes)
		require.NotNil(t, respRes.Err)
		require.Nil(t, ips)
		require.EqualError(t, respRes.Err, httpErrorMessage)
	})

	t.Run("DoRequestError", func(t *testing.T) {
		client := newFakeClient("http://fake", httptest.NewFakeTransport(nil, errors.New("network failure")))

		ips, respRes, err := client.NetworkSubnetLocalReservedIPs(context.Background(), "subnetid")
		require.Error(t, err)
		require.Nil(t, ips)
		require.Nil(t, respRes)
	})
}
