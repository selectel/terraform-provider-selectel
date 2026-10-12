package main

import (
	"context"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/plugin"
	"github.com/terraform-providers/terraform-provider-selectel/selectel"
	"github.com/terraform-providers/terraform-provider-selectel/version"
)

func main() {
	providerServer, err := selectel.ProviderServer(context.Background(), version.Version)
	if err != nil {
		log.Fatal(err)
	}

	plugin.Serve(&plugin.ServeOpts{
		GRPCProviderV6Func: providerServer,
	})
}
