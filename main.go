package main

import (
	"github.com/hashicorp/terraform/plugin"
	"github.com/terraform-providers/terraform-provider-scaleway/scaleway"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: scaleway.Provider})
}
