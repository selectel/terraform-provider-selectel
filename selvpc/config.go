package selvpc

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/hashicorp/terraform/httpclient"
	"github.com/selectel/go-selvpcclient/selvpcclient"
	"github.com/selectel/go-selvpcclient/selvpcclient/resell"
	resellv2 "github.com/selectel/go-selvpcclient/selvpcclient/resell/v2"
)

// Config contains all available configuration options.
type Config struct {
	Token    string
	Endpoint string
}

// Validate performs config validation.
func (c *Config) Validate() error {
	if c.Token == "" {
		return fmt.Errorf("token must be specified")
	}
	if c.Endpoint == "" {
		c.Endpoint = selvpcclient.DefaultEndpoint
	}
	return nil
}

func (c *Config) resellV2Client() *selvpcclient.ServiceClient {
	endpoint := strings.Join([]string{c.Endpoint, resell.ServiceType, resellv2.APIVersion},
		"/")
	resellV2Client := &selvpcclient.ServiceClient{
		HTTPClient: &http.Client{},
		Endpoint:   endpoint,
		TokenID:    c.Token,
		UserAgent:  httpclient.UserAgentString(),
	}
	return resellV2Client
}
