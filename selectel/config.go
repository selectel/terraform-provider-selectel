package selectel

import (
	"context"
	"fmt"
	"sync"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/selectel/go-selvpcclient/v3/selvpcclient"
)

var (
	cfgSingletone *Config
	once          sync.Once
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
}

func getConfig(d *schema.ResourceData) (*Config, diag.Diagnostics) {
	once.Do(func() {
		cfgSingletone = &Config{
			Username:   d.Get("username").(string),
			Password:   d.Get("password").(string),
			DomainName: d.Get("domain_name").(string),
		}
		if v, ok := d.GetOk("auth_url"); ok {
			cfgSingletone.AuthURL = v.(string)
		}
		if v, ok := d.GetOk("auth_region"); ok {
			cfgSingletone.AuthRegion = v.(string)
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
		AuthRegion:     c.AuthRegion,
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
