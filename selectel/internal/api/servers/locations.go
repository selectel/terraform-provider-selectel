package servers

import (
	"context"
	"net/http"
)

type Location struct {
	UUID        string `json:"uuid"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Visibility  string `json:"visibility"`
}

type Locations []*Location

func (l Locations) FindOneByName(name string) *Location {
	for _, location := range l {
		if location.Name == name {
			return location
		}
	}

	return nil
}

func (client *ServiceClient) Locations(ctx context.Context) (Locations, *ResponseResult, error) {
	url := client.Endpoint + "/location"

	responseResult, err := client.DoRequest(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, nil, err
	}
	if responseResult.Err != nil {
		return nil, responseResult, responseResult.Err
	}

	var result struct {
		Locations Locations `json:"result"`
	}
	err = responseResult.ExtractResult(&result)
	if err != nil {
		return nil, responseResult, err
	}

	return result.Locations, responseResult, nil
}
