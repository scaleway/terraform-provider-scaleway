package provider

import (
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
)

func NewProviderList(config *Config) []func() tfprotov5.ProviderServer {
	return []func() tfprotov5.ProviderServer{
		// Provider using terraform-plugin-framework
		providerserver.NewProtocol5(&ScalewayProvider{}),

		// SDKProvider using terraform-plugin-sdk
		SDKProvider(config)().GRPCProvider,
	}
}
