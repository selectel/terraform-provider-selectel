package servers

import (
	"context"
	"net/http"
)

type (
	SSHKey struct {
		Name      string `json:"name_public_key"`
		PublicKey string `json:"public_key"`
	}

	SSHKeys []*SSHKey
)

func (s SSHKeys) FindOneByName(name string) *SSHKey {
	for _, key := range s {
		if key.Name == name {
			return key
		}
	}

	return nil
}

func (s SSHKeys) FindOneByPK(pk string) *SSHKey {
	for _, key := range s {
		if key.PublicKey == pk {
			return key
		}
	}

	return nil
}

func (client *ServiceClient) SSHKeys(ctx context.Context) (SSHKeys, *ResponseResult, error) {
	url := client.Endpoint + "aux/ssh-keys/key"

	responseResult, err := client.DoRequest(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, nil, err
	}
	if responseResult.Err != nil {
		return nil, responseResult, responseResult.Err
	}

	var result struct {
		Keys SSHKeys `json:"result"`
	}
	err = responseResult.ExtractResult(&result)
	if err != nil {
		return nil, responseResult, err
	}

	return result.Keys, responseResult, nil
}
