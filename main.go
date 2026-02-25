package main

import (
	"os"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/plugin"
	"github.com/terraform-providers/terraform-provider-selectel/selectel"
	"github.com/terraform-providers/terraform-provider-selectel/version"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: func() *schema.Provider {
			return selectel.Provider(version.Version)
		},
		Debug: os.Getenv("TF_DEBUG") == "1",
	})
}
