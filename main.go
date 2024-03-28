package main

import (
	"context"
	"flag"
	"log"

	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5/tf5server"
	"github.com/hashicorp/terraform-plugin-mux/tf5muxserver"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/provider"
)

func main() {
	ctx := context.Background()

	var debugMode bool
	flag.BoolVar(&debugMode, "debuggable", false, "set to true to run the provider with support for debuggers like delve")
	flag.Parse()
	var serveOpts []tf5server.ServeOpt
	if debugMode {
		serveOpts = append(serveOpts, tf5server.WithManagedDebug())
	}

	providers := []func() tfprotov5.ProviderServer{
		// Provider using terraform-plugin-sdk
		provider.Provider(provider.DefaultConfig())().GRPCProvider,
	}

	muxServer, err := tf5muxserver.NewMuxServer(ctx, providers...)
	if err != nil {
		log.Fatal(err)
	}

	err = tf5server.Serve(
		"registry.terraform.io/scaleway/scaleway",
		muxServer.ProviderServer,
		serveOpts...,
	)
	if err != nil {
		log.Fatal(err)
	}
}
