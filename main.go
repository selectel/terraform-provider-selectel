package main

import (
	"github.com/hashicorp/terraform/plugin"
	"github.com/selectel/terraform-provider-selvpc/selvpc"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: selvpc.Provider,
	})
}
