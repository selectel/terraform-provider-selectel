package selectel

import (
	"errors"
	"strings"

	"github.com/hashicorp/go-retryablehttp"
	domainsV1 "github.com/selectel/domains-go/pkg/v1"
	"github.com/selectel/go-selvpcclient/selvpcclient"
	"github.com/selectel/go-selvpcclient/selvpcclient/resell"
	resellV2 "github.com/selectel/go-selvpcclient/selvpcclient/resell/v2"
)

// Config contains all available configuration options.
type Config struct {
	Token     string
	Endpoint  string
	ProjectID string
	Region    string
}

// Validate performs config validation.
func (c *Config) Validate() error {
	if c.Token == "" {
		return errors.New("token must be specified")
	}
	if c.Endpoint == "" {
		c.Endpoint = strings.Join([]string{resell.Endpoint, resellV2.APIVersion}, "/")
	}
	if c.Region != "" {
		if err := validateRegion(c.Region); err != nil {
			return err
		}
	}
	return nil
}

func (c *Config) resellV2Client() *selvpcclient.ServiceClient {
	return resellV2.NewV2ResellClientWithEndpoint(c.Token, c.Endpoint)
}

func (c *Config) domainsV1Client() *domainsV1.ServiceClient {
	domainsClient := domainsV1.NewDomainsClientV1WithDefaultEndpoint(c.Token)
	retryClient := retryablehttp.NewClient()
	retryClient.Logger = nil // Ignore retyablehttp client logs
	retryClient.RetryWaitMin = domainsV1DefaultRetryWaitMin
	retryClient.RetryWaitMax = domainsV1DefaultRetryWaitMax
	retryClient.RetryMax = domainsV1DefaultRetry
	domainsClient.HTTPClient = retryClient.StandardClient()
	return domainsClient
}
