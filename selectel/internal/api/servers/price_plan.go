package servers

import (
	"context"
	"fmt"
	"net/http"
)

type (
	PricePlan struct {
		UUID string `json:"uuid"`
		Name string `json:"name"`
	}

	PricePlans []*PricePlan
)

func (p PricePlans) FindOneByName(name string) *PricePlan {
	for _, pp := range p {
		if pp.Name == name {
			return pp
		}
	}

	return nil
}

func (client *ServiceClient) PricePlans(ctx context.Context) (PricePlans, *ResponseResult, error) {
	url := fmt.Sprintf("%s/plan", client.Endpoint)

	responseResult, err := client.DoRequest(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, nil, err
	}
	if responseResult.Err != nil {
		return nil, responseResult, responseResult.Err
	}

	var result struct {
		Result PricePlans `json:"result"`
	}
	err = responseResult.ExtractResult(&result)
	if err != nil {
		return nil, responseResult, err
	}

	return result.Result, responseResult, nil
}
