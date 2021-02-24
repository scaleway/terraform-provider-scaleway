package main

import (
	"context"
	"flag"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/plugin"
	"github.com/scaleway/terraform-provider-scaleway/scaleway"
)

func main() {
	var debugMode bool

	flag.BoolVar(&debugMode, "debug", false, "set to true to run the provider with support for debuggers like delve")
	flag.Parse()

	opts := &plugin.ServeOpts{ProviderFunc: scaleway.Provider(scaleway.DefaultProviderConfig())}

	if debugMode {
		err := plugin.Debug(context.Background(), "registry.terraform.io/scaleway/scaleway", opts)
		if err != nil {
			log.Println(err.Error())
		}
		return
	}

	plugin.Serve(opts)
}
