package main

import (
	"flag"

	"github.com/hashicorp/terraform-plugin-sdk/v2/plugin"
	"github.com/scaleway/terraform-provider-scaleway/v2/scaleway"
)

func main() {
	var debugMode bool

	flag.BoolVar(&debugMode, "debug", false, "set to true to run the provider with support for debuggers like delve")
	flag.Parse()

	options := &plugin.ServeOpts{
		ProviderFunc: scaleway.Provider(scaleway.DefaultProviderConfig()),
	}
	if debugMode {
		options.Debug = true
	}

	plugin.Serve(options)
}
