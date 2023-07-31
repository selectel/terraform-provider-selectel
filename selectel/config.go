package selectel

import (
	"context"
	"fmt"
	"sync"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	domainsV1 "github.com/selectel/domains-go/pkg/v1"
	"github.com/selectel/go-selvpcclient/v3/selvpcclient"
)

var (
	cfgSingletone *Config
	once          sync.Once
)

// Config contains all available configuration options.
type Config struct {
	Token     string
	Region    string
	ProjectID string

	Context        context.Context
	AuthURL        string
	Username       string
	Password       string
	UserDomainName string
	DomainName     string
	clientsCache   map[string]*selvpcclient.Client
	lock           sync.Mutex
}

func getConfig(d *schema.ResourceData) (*Config, diag.Diagnostics) {
	var err error

	once.Do(func() {
		cfgSingletone = &Config{
			Username:   d.Get("username").(string),
			Password:   d.Get("password").(string),
			DomainName: d.Get("domain_name").(string),
		}
		if v, ok := d.GetOk("token"); ok {
			cfgSingletone.Token = v.(string)
		}
		if v, ok := d.GetOk("auth_url"); ok {
			cfgSingletone.AuthURL = v.(string)
		}
		if v, ok := d.GetOk("user_domain_name"); ok {
			cfgSingletone.UserDomainName = v.(string)
		}
		if v, ok := d.GetOk("project_id"); ok {
			cfgSingletone.ProjectID = v.(string)
		}
		if v, ok := d.GetOk("region"); ok {
			cfgSingletone.Region = v.(string)
		}
	})
	if err != nil {
		return nil, diag.FromErr(err)
	}

	return cfgSingletone, nil
}

func (c *Config) GetSelVPCClient() (*selvpcclient.Client, error) {
	return c.GetSelVPCClientWithProjectScope("")
}

func (c *Config) GetSelVPCClientWithProjectScope(projectID string) (*selvpcclient.Client, error) {
	c.lock.Lock()
	defer c.lock.Unlock()

	clientsCacheKey := fmt.Sprintf("client_%s", projectID)

	if client, ok := c.clientsCache[clientsCacheKey]; ok {
		return client, nil
	}

	opts := &selvpcclient.ClientOptions{
		DomainName:     c.DomainName,
		Username:       c.Username,
		Password:       c.Password,
		ProjectID:      projectID,
		AuthURL:        c.AuthURL,
		UserDomainName: c.UserDomainName,
	}

	client, err := selvpcclient.NewClient(opts)
	if err != nil {
		return nil, err
	}

	if c.clientsCache == nil {
		c.clientsCache = map[string]*selvpcclient.Client{}
	}

	c.clientsCache[clientsCacheKey] = client

	return client, nil
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
