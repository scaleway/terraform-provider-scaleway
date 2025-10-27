package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-mux/tf5to6server"
)

func NewProviderList(ctx context.Context, providerConfig *Config) ([]func() tfprotov6.ProviderServer, error) {
	// SDKProvider using terraform-plugin-sdk
	upgradedSdkProvider, err := tf5to6server.UpgradeServer(
		ctx,
		SDKProvider(providerConfig)().GRPCProvider,
	)
	if err != nil {
		return nil, err
	}

	var frameworkProvider func() provider.Provider
	if providerConfig != nil {
		frameworkProvider = NewFrameworkProvider(providerConfig.Meta)
	} else {
		frameworkProvider = NewFrameworkProvider(nil)
	}

	return []func() tfprotov6.ProviderServer{
		// Provider using terraform-plugin-framework
		providerserver.NewProtocol6(frameworkProvider()),

		func() tfprotov6.ProviderServer {
			return upgradedSdkProvider
		},
	}, nil
}
