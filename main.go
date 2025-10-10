package main

import (
	"context"
	"flag"
	"log"

	"github.com/hashicorp/terraform-plugin-go/tfprotov6/tf6server"
	"github.com/hashicorp/terraform-plugin-mux/tf6muxserver"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/provider"
)

func main() {
	ctx := context.Background()

	var debugMode bool

	flag.BoolVar(&debugMode, "debuggable", false, "set to true to run the provider with support for debuggers like delve")
	flag.Parse()

	var serveOpts []tf6server.ServeOpt
	if debugMode {
		serveOpts = append(serveOpts, tf6server.WithManagedDebug())
	}

	providers, err := provider.NewProviderList(ctx, nil)
	if err != nil {
		log.Fatal(err)
	}

	muxServer, err := tf6muxserver.NewMuxServer(ctx, providers...)
	if err != nil {
		log.Fatal(err)
	}

	err = tf6server.Serve(
		"registry.terraform.io/scaleway/scaleway",
		muxServer.ProviderServer,
		serveOpts...,
	)
	if err != nil {
		log.Fatal(err)
	}
}
