package servers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func (client *ServiceClient) Servers(ctx context.Context, isServerChip bool) (Servers, *ResponseResult, error) {
	if isServerChip {
		return client.serverChips(ctx)
	}

	return client.servers(ctx)
}

func (client *ServiceClient) servers(ctx context.Context) (Servers, *ResponseResult, error) {
	url := client.Endpoint + "/service/server"

	responseResult, err := client.DoRequest(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, nil, err
	}
	if responseResult.Err != nil {
		return nil, responseResult, responseResult.Err
	}

	var result struct {
		Servers Servers `json:"result"`
	}
	err = responseResult.ExtractResult(&result)
	if err != nil {
		return nil, responseResult, err
	}

	return result.Servers, responseResult, nil
}

func (client *ServiceClient) serverChips(ctx context.Context) (Servers, *ResponseResult, error) {
	url := client.Endpoint + "/service/serverchip"

	responseResult, err := client.DoRequest(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, nil, err
	}
	if responseResult.Err != nil {
		return nil, responseResult, responseResult.Err
	}

	var result struct {
		Servers Servers `json:"result"`
	}
	err = responseResult.ExtractResult(&result)
	if err != nil {
		return nil, responseResult, err
	}

	return result.Servers, responseResult, nil
}

func (client *ServiceClient) ServerByID(ctx context.Context, id string, isServerChip bool) (*Server, *ResponseResult, error) {
	if isServerChip {
		return client.serverChipByID(ctx, id)
	}

	return client.serverByID(ctx, id)
}

func (client *ServiceClient) serverByID(ctx context.Context, id string) (*Server, *ResponseResult, error) {
	url := fmt.Sprintf("%s/service/server/%s", client.Endpoint, id)

	responseResult, err := client.DoRequest(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, nil, err
	}
	if responseResult.Err != nil {
		return nil, responseResult, responseResult.Err
	}

	var result struct {
		Server *Server `json:"result"`
	}
	err = responseResult.ExtractResult(&result)
	if err != nil {
		return nil, responseResult, err
	}

	return result.Server, responseResult, nil
}

func (client *ServiceClient) serverChipByID(ctx context.Context, id string) (*Server, *ResponseResult, error) {
	url := fmt.Sprintf("%s/service/serverchip/%s", client.Endpoint, id)

	responseResult, err := client.DoRequest(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, nil, err
	}
	if responseResult.Err != nil {
		return nil, responseResult, responseResult.Err
	}

	var result struct {
		Server *Server `json:"result"`
	}
	err = responseResult.ExtractResult(&result)
	if err != nil {
		return nil, responseResult, err
	}

	result.Server.IsServerChip = true

	return result.Server, responseResult, nil
}

func (client *ServiceClient) ServerCalculateBilling(
	ctx context.Context, serviceID, locationID, pricePlanID, payCurrency string, isServerChip bool,
) (*ServiceBilling, *ResponseResult, error) {
	if isServerChip {
		return client.serverChipCalculateBilling(ctx, serviceID, locationID, pricePlanID, payCurrency)
	}

	return client.serverCalculateBilling(ctx, serviceID, locationID, pricePlanID, payCurrency)
}

func (client *ServiceClient) serverCalculateBilling(
	ctx context.Context, serviceID, locationID, pricePlanID, payCurrency string,
) (*ServiceBilling, *ResponseResult, error) {
	url := fmt.Sprintf("%s/service/server/%s/billing", client.Endpoint, serviceID)

	payload := struct {
		LocationUUID  string `json:"location_uuid"`
		PricePlanUUID string `json:"price_plan_uuid"`
		PayCurrency   string `json:"pay_currency"`
		Quantity      int    `json:"quantity"`
	}{
		LocationUUID:  locationID,
		PricePlanUUID: pricePlanID,
		PayCurrency:   payCurrency,
		Quantity:      1,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, nil, err
	}

	responseResult, err := client.DoRequest(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, nil, err
	}
	if responseResult.Err != nil {
		return nil, responseResult, responseResult.Err
	}

	var result struct {
		Result *ServiceBilling `json:"result"`
	}
	err = responseResult.ExtractResult(&result)
	if err != nil {
		return nil, responseResult, err
	}

	return result.Result, responseResult, nil
}

func (client *ServiceClient) serverChipCalculateBilling(
	ctx context.Context, serviceID, locationID, pricePlanID, payCurrency string,
) (*ServiceBilling, *ResponseResult, error) {
	url := fmt.Sprintf("%s/service/serverchip/%s/billing", client.Endpoint, serviceID)

	payload := struct {
		LocationUUID  string `json:"location_uuid"`
		PricePlanUUID string `json:"price_plan_uuid"`
		PayCurrency   string `json:"pay_currency"`
		Quantity      int    `json:"quantity"`
	}{
		LocationUUID:  locationID,
		PricePlanUUID: pricePlanID,
		PayCurrency:   payCurrency,
		Quantity:      1,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, nil, err
	}

	responseResult, err := client.DoRequest(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, nil, err
	}
	if responseResult.Err != nil {
		return nil, responseResult, responseResult.Err
	}

	var result struct {
		Result *ServiceBilling `json:"result"`
	}
	err = responseResult.ExtractResult(&result)
	if err != nil {
		return nil, responseResult, err
	}

	return result.Result, responseResult, nil
}

func (client *ServiceClient) ServerBilling(
	ctx context.Context, req *ServerBillingPostPayload, isServerChip bool,
) ([]*ServerBillingPostResult, *ResponseResult, error) {
	if isServerChip {
		return client.serverChipBilling(ctx, req)
	}

	return client.serverBilling(ctx, req)
}

func (client *ServiceClient) serverBilling(
	ctx context.Context, req *ServerBillingPostPayload,
) ([]*ServerBillingPostResult, *ResponseResult, error) {
	url := fmt.Sprintf("%s/resource/server/billing", client.Endpoint)

	body, err := json.Marshal(req)
	if err != nil {
		return nil, nil, err
	}

	responseResult, err := client.DoRequest(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, nil, err
	}
	if responseResult.Err != nil {
		return nil, responseResult, responseResult.Err
	}

	var result struct {
		Result []*ServerBillingPostResult `json:"result"`
	}
	err = responseResult.ExtractResult(&result)
	if err != nil {
		return nil, responseResult, err
	}

	return result.Result, responseResult, nil
}

func (client *ServiceClient) serverChipBilling(
	ctx context.Context, req *ServerBillingPostPayload,
) ([]*ServerBillingPostResult, *ResponseResult, error) {
	url := fmt.Sprintf("%s/resource/serverchip/billing", client.Endpoint)

	body, err := json.Marshal(req)
	if err != nil {
		return nil, nil, err
	}

	responseResult, err := client.DoRequest(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, nil, err
	}
	if responseResult.Err != nil {
		return nil, responseResult, responseResult.Err
	}

	var result struct {
		Result []*ServerBillingPostResult `json:"result"`
	}
	err = responseResult.ExtractResult(&result)
	if err != nil {
		return nil, responseResult, err
	}

	return result.Result, responseResult, nil
}

func (client *ServiceClient) ResourceDetails(
	ctx context.Context, id string,
) (*ResourceDetails, *ResponseResult, error) {
	url := fmt.Sprintf("%s/resource/%s", client.Endpoint, id)

	responseResult, err := client.DoRequest(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, nil, err
	}
	if responseResult.Err != nil {
		return nil, responseResult, responseResult.Err
	}

	var result struct {
		Result *ResourceDetails `json:"result"`
	}
	err = responseResult.ExtractResult(&result)
	if err != nil {
		return nil, responseResult, err
	}

	return result.Result, responseResult, nil
}

func (client *ServiceClient) DeleteResource(
	ctx context.Context, id string,
) (*ResponseResult, error) {
	url := fmt.Sprintf("%s/resource/billing/%s", client.Endpoint, id)

	payload := struct {
		Immediately bool `json:"immediately"`
	}{
		Immediately: false,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	responseResult, err := client.DoRequest(ctx, http.MethodDelete, url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	if responseResult.Err != nil {
		return responseResult, responseResult.Err
	}

	defer func() { _ = responseResult.Body.Close() }()

	return responseResult, nil
}
