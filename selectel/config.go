package selectel

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/openstack"
	"github.com/hashicorp/go-retryablehttp"
	domainsV1 "github.com/selectel/domains-go/pkg/v1"
	"github.com/selectel/go-selvpcclient/selvpcclient"
	"github.com/selectel/go-selvpcclient/selvpcclient/resell"
	resellV2 "github.com/selectel/go-selvpcclient/selvpcclient/resell/v2"
	"github.com/selectel/go-selvpcclient/selvpcclient/resell/v2/tokens"
)

const (
	DefaultIdentityEndpoint = "https://api.selvpc.ru/identity/v3"
)

// Config contains all available configuration options.
type Config struct {
	Token            string
	IdentityEndpoint string
	Endpoint         string
	ProjectID        string
	DomainName       string
	Region           string
	User             string
	Password         string
}

// Validate performs config validation.
func (c *Config) Validate() error {
	if c.Token == "" && !c.hasKeystoneCredentials() {
		return errors.New("token or credentials with domain name must be specified")
	}
	if c.Endpoint == "" {
		c.Endpoint = strings.Join([]string{resell.Endpoint, resellV2.APIVersion}, "/")
	}
	if c.Region != "" {
		if err := validateRegion(c.Region); err != nil {
			return err
		}
	}
	if c.IdentityEndpoint == "" {
		c.IdentityEndpoint = DefaultIdentityEndpoint
	}

	return nil
}

// Initialize Selectel resell client.
func (c *Config) resellV2Client() *selvpcclient.ServiceClient {
	return resellV2.NewV2ResellClientWithEndpoint(c.Token, c.Endpoint)
}

// Create Keystone token by Selectel token or Keystone credentials.
func (c *Config) getToken(ctx context.Context, projectID string) (string, error) {
	if !c.hasKeystoneCredentials() {
		return c.getTokenBySelectelToken(ctx, projectID)
	}

	return c.getTokenByCredentials(ctx, projectID)
}

// hasKeystoneCredentials determine Keystone credentials exists.
func (c *Config) hasKeystoneCredentials() bool {
	return c.User != "" && c.Password != "" && c.DomainName != ""
}

// Create Keystone token by Selectel token.
func (c *Config) getTokenBySelectelToken(ctx context.Context, projectID string) (string, error) {
	tokenOpts := tokens.TokenOpts{
		ProjectID: projectID,
	}
	resellV2Client := c.resellV2Client()
	log.Print(msgCreate(objectToken, tokenOpts))

	token, _, err := tokens.Create(ctx, resellV2Client, tokenOpts)
	if err != nil {
		return "", fmt.Errorf("create Keystone token by Selectel token: %w", err)
	}

	return token.ID, nil
}

// Create Keystone token by Keystone credentials.
func (c *Config) getTokenByCredentials(ctx context.Context, projectID string) (string, error) {
	providerOpts := gophercloud.AuthOptions{
		AllowReauth:      true,
		IdentityEndpoint: c.IdentityEndpoint,
		Username:         c.User,
		Password:         c.Password,
		DomainName:       c.DomainName,
		Scope: &gophercloud.AuthScope{
			ProjectID: projectID,
		},
	}

	newProvider, err := openstack.AuthenticatedClient(providerOpts)
	if err != nil {
		return "", fmt.Errorf("keystone auth: %w", err)
	}
	tokenID, err := newProvider.GetAuthResult().ExtractTokenID()
	if err != nil {
		return "", fmt.Errorf("extract token id: %w", err)
	}

	return tokenID, nil
}

// Initialize Selectel domains client.
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
