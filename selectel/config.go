package selectel

import (
	"context"
	"fmt"
	"sync"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/selectel/go-selvpcclient/v5/selvpcclient"
)

// Config contains all available configuration options.
type Config struct {
	Region    string
	ProjectID string

	Context        context.Context
	AuthURL        string
	AuthRegion     string
	Username       string
	Password       string
	UserDomainName string
	DomainName     string
	clientsCache   map[string]*selvpcclient.Client
	lock           sync.Mutex

	UserAgent string
}

// getConfig builds the Config of one provider configuration. A process that
// configures the provider again, such as an acceptance test run, gets the new
// project and region.
func getConfig(d *schema.ResourceData, userAgent string) (*Config, diag.Diagnostics) {
	config := &Config{
		Username:   d.Get("username").(string),
		Password:   d.Get("password").(string),
		DomainName: d.Get("domain_name").(string),
		AuthURL:    d.Get("auth_url").(string),
		AuthRegion: d.Get("auth_region").(string),
		UserAgent:  userAgent,
	}
	if v, ok := d.GetOk("user_domain_name"); ok {
		config.UserDomainName = v.(string)
	}
	if v, ok := d.GetOk("project_id"); ok {
		config.ProjectID = v.(string)
	}
	if v, ok := d.GetOk("region"); ok {
		config.Region = v.(string)
	}

	return config, nil
}

func (c *Config) selvpcClientOptions(projectID string) *selvpcclient.ClientOptions {
	return &selvpcclient.ClientOptions{
		DomainName:     c.DomainName,
		Username:       c.Username,
		Password:       c.Password,
		ProjectID:      projectID,
		AuthURL:        c.AuthURL,
		AuthRegion:     c.AuthRegion,
		UserDomainName: c.UserDomainName,
		UserAgent:      c.UserAgent,
	}
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

	client, err := selvpcclient.NewClient(c.selvpcClientOptions(projectID))
	if err != nil {
		return nil, err
	}

	if c.clientsCache == nil {
		c.clientsCache = map[string]*selvpcclient.Client{}
	}

	c.clientsCache[clientsCacheKey] = client

	return client, nil
}

// projectEndpoint returns the public endpoint of a service and a fresh token
// scoped to the project.
func (c *Config) projectEndpoint(ctx context.Context, projectID, region, serviceType string) (string, string, error) {
	opts := c.selvpcClientOptions(projectID)
	opts.Context = ctx
	client, err := selvpcclient.NewClient(opts)
	if err != nil {
		return "", "", fmt.Errorf("authenticate project %s: %w", projectID, err)
	}
	endpoint, err := client.Catalog.GetEndpoint(serviceType, region)
	if err != nil {
		return "", "", err
	}

	return endpoint.URL, client.GetXAuthToken(), nil
}
