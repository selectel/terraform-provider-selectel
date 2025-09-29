package cloudbackup

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

type PlansQuery struct {
	Name       string
	VolumeName string
}

func (q *PlansQuery) queryParamsRaw() string {
	params := url.Values{}
	if q == nil {
		return params.Encode()
	}

	if q.Name != "" {
		params.Add("name", q.Name)
	}

	if q.VolumeName != "" {
		params.Add("volume_name", q.VolumeName)
	}

	return params.Encode()
}

func (client *ServiceClient) Plans(ctx context.Context, q *PlansQuery) ([]*Plan, *ResponseResult, error) {
	queryParams := ""
	if qRaw := q.queryParamsRaw(); q != nil && qRaw != "" {
		queryParams = "?" + qRaw
	}

	u := fmt.Sprintf("%s/plans/%s", client.Endpoint, queryParams)

	responseResult, err := client.DoRequest(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, nil, err
	}
	if responseResult.Err != nil {
		return nil, responseResult, responseResult.Err
	}

	var result struct {
		Plans []*Plan `json:"plans"`
	}
	err = responseResult.ExtractResult(&result)
	if err != nil {
		return nil, responseResult, err
	}

	return result.Plans, responseResult, nil
}

func (client *ServiceClient) Plan(ctx context.Context, planID string) (*Plan, *ResponseResult, error) {
	u := fmt.Sprintf("%s/plans/%s", client.Endpoint, planID)

	responseResult, err := client.DoRequest(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, nil, err
	}
	if responseResult.Err != nil {
		return nil, responseResult, responseResult.Err
	}

	var result *Plan
	err = responseResult.ExtractResult(&result)
	if err != nil {
		return nil, responseResult, err
	}

	return result, responseResult, nil
}

func (client *ServiceClient) PlanCreate(ctx context.Context, req *Plan) (*Plan, *ResponseResult, error) {
	u := fmt.Sprintf("%s/plans/", client.Endpoint)

	body, err := json.Marshal(req)
	if err != nil {
		return nil, nil, err
	}

	responseResult, err := client.DoRequest(ctx, http.MethodPost, u, bytes.NewReader(body))
	if err != nil {
		return nil, nil, err
	}
	if responseResult.Err != nil {
		return nil, responseResult, responseResult.Err
	}

	var result *Plan
	err = responseResult.ExtractResult(&result)
	if err != nil {
		return nil, responseResult, err
	}

	return result, responseResult, nil
}

func (client *ServiceClient) PlanUpdate(ctx context.Context, planID string, req *Plan) (*Plan, *ResponseResult, error) {
	u := fmt.Sprintf("%s/plans/%s", client.Endpoint, planID)

	body, err := json.Marshal(req)
	if err != nil {
		return nil, nil, err
	}

	responseResult, err := client.DoRequest(ctx, http.MethodPatch, u, bytes.NewReader(body))
	if err != nil {
		return nil, nil, err
	}
	if responseResult.Err != nil {
		return nil, responseResult, responseResult.Err
	}

	var result *Plan
	err = responseResult.ExtractResult(&result)
	if err != nil {
		return nil, responseResult, err
	}

	return result, responseResult, nil
}

func (client *ServiceClient) PlanDelete(ctx context.Context, planID string) (*ResponseResult, error) {
	u := fmt.Sprintf("%s/plans/%s", client.Endpoint, planID)

	responseResult, err := client.DoRequest(ctx, http.MethodDelete, u, nil)
	if err != nil {
		return nil, err
	}
	if responseResult.Err != nil {
		return responseResult, responseResult.Err
	}

	return responseResult, nil
}
