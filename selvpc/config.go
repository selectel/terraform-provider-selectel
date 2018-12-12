package selvpc

import (
	"errors"
	"strings"

	"github.com/selectel/go-selvpcclient/selvpcclient"
	"github.com/selectel/go-selvpcclient/selvpcclient/resell"
	"github.com/selectel/go-selvpcclient/selvpcclient/resell/v2"
)

// Config contains all available configuration options.
type Config struct {
	Token    string
	Endpoint string
}

// Validate performs config validation.
func (c *Config) Validate() error {
	if c.Token == "" {
		return errors.New("token must be specified")
	}
	if c.Endpoint == "" {
		c.Endpoint = strings.Join([]string{resell.Endpoint, v2.APIVersion}, "/")
	}
	return nil
}

func (c *Config) resellV2Client() *selvpcclient.ServiceClient {
	return v2.NewV2ResellClientWithEndpoint(c.Token, c.Endpoint)
}
