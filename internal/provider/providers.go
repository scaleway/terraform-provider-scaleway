package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-mux/tf5to6server"
)

func NewProviderList(ctx context.Context, config *Config) ([]func() tfprotov6.ProviderServer, error) {
	// SDKProvider using terraform-plugin-sdk
	upgradedSdkProvider, err := tf5to6server.UpgradeServer(
		ctx,
		SDKProvider(config)().GRPCProvider,
	)
	if err != nil {
		return nil, err
	}

	return []func() tfprotov6.ProviderServer{
		// Provider using terraform-plugin-framework
		providerserver.NewProtocol6(&ScalewayProvider{}),

		func() tfprotov6.ProviderServer {
			return upgradedSdkProvider
		},
	}, nil
}
