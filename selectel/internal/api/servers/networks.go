package servers

import (
	"context"
	"fmt"
	"net/http"
)

type NetworkType string

const (
	NetworkTypeInet  NetworkType = "inet"
	NetworkTypeLocal NetworkType = "local"
)

func (client *ServiceClient) Networks(ctx context.Context, locationID string, networkType NetworkType) (Networks, *ResponseResult, error) {
	url := fmt.Sprintf("%s/network?location_uuid=%s&network_type=%s", client.Endpoint, locationID, networkType)

	responseResult, err := client.DoRequest(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, nil, err
	}
	if responseResult.Err != nil {
		return nil, responseResult, responseResult.Err
	}

	var result struct {
		Result Networks `json:"result"`
	}
	err = responseResult.ExtractResult(&result)
	if err != nil {
		return nil, responseResult, err
	}

	return result.Result, responseResult, nil
}

func (client *ServiceClient) NetworkSubnets(ctx context.Context, locationID string) (Subnets, *ResponseResult, error) {
	url := fmt.Sprintf("%s/network/ipam/subnet?location_uuid=%s&is_master_shared=false", client.Endpoint, locationID)

	responseResult, err := client.DoRequest(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, nil, err
	}
	if responseResult.Err != nil {
		return nil, responseResult, responseResult.Err
	}

	var result struct {
		Result Subnets `json:"result"`
	}
	err = responseResult.ExtractResult(&result)
	if err != nil {
		return nil, responseResult, err
	}

	return result.Result, responseResult, nil
}

func (client *ServiceClient) NetworkLocalSubnets(ctx context.Context, networkID string) (Subnets, *ResponseResult, error) {
	url := fmt.Sprintf("%s/network/ipam/local_subnet?network_uuid=%s", client.Endpoint, networkID)

	responseResult, err := client.DoRequest(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, nil, err
	}
	if responseResult.Err != nil {
		return nil, responseResult, responseResult.Err
	}

	var result struct {
		Result Subnets `json:"result"`
	}
	err = responseResult.ExtractResult(&result)
	if err != nil {
		return nil, responseResult, err
	}

	return result.Result, responseResult, nil
}

func (client *ServiceClient) NetworkSubnet(ctx context.Context, subnetID string) (*Subnet, *ResponseResult, error) {
	url := fmt.Sprintf("%s/network/ipam/subnet/%s", client.Endpoint, subnetID)

	responseResult, err := client.DoRequest(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, nil, err
	}
	if responseResult.Err != nil {
		return nil, responseResult, responseResult.Err
	}

	var result struct {
		Result *Subnet `json:"result"`
	}
	err = responseResult.ExtractResult(&result)
	if err != nil {
		return nil, responseResult, err
	}

	return result.Result, responseResult, nil
}

func (client *ServiceClient) NetworkReservedIPs(ctx context.Context, locationID string) (ReservedIPs, *ResponseResult, error) {
	url := fmt.Sprintf("%s/network/ipam/ip?location_uuid=%s", client.Endpoint, locationID)

	responseResult, err := client.DoRequest(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, nil, err
	}
	if responseResult.Err != nil {
		return nil, responseResult, responseResult.Err
	}

	var result struct {
		Result ReservedIPs `json:"result"`
	}
	err = responseResult.ExtractResult(&result)
	if err != nil {
		return nil, responseResult, err
	}

	return result.Result, responseResult, nil
}

func (client *ServiceClient) NetworkSubnetLocalReservedIPs(ctx context.Context, subnetID string) (ReservedIPs, *ResponseResult, error) {
	url := fmt.Sprintf("%s/network/ipam/local_subnet/%s/local_ip", client.Endpoint, subnetID)

	responseResult, err := client.DoRequest(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, nil, err
	}
	if responseResult.Err != nil {
		return nil, responseResult, responseResult.Err
	}

	var result struct {
		Result ReservedIPs `json:"result"`
	}
	err = responseResult.ExtractResult(&result)
	if err != nil {
		return nil, responseResult, err
	}

	return result.Result, responseResult, nil
}
