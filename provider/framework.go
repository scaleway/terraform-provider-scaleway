package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/action"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/ephemeral"
	"github.com/hashicorp/terraform-plugin-framework/list"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/secret"
)

var _ provider.Provider = &ScalewayProvider{}

type ScalewayProvider struct{}

func NewFrameworkProvider() func() provider.Provider {
	return func() provider.Provider {
		return &ScalewayProvider{}
	}
}

func (p *ScalewayProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "scaleway"
}

func (p *ScalewayProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"access_key": schema.StringAttribute{
				Optional:    true,
				Description: "The Scaleway access key.",
			},
			"secret_key": schema.StringAttribute{
				Optional:    true,
				Description: "The Scaleway secret Key.",
			},
			"profile": schema.StringAttribute{
				Optional:    true,
				Description: "The Scaleway profile to use.",
			},
			"project_id": schema.StringAttribute{
				Optional:    true,
				Description: "The Scaleway project ID.",
			},
			"organization_id": schema.StringAttribute{
				Optional:    true,
				Description: "The Scaleway organization ID.",
			},
			"api_url": schema.StringAttribute{
				Optional:    true,
				Description: "The Scaleway API URL to use.",
			},
			"region": schema.StringAttribute{
				Optional:    true,
				Description: "The region you want to attach the resource to",
			},
			"zone": schema.StringAttribute{
				Description: "The zone you want to attach the resource to",
				Optional:    true,
			},
		},
	}
}

func (p *ScalewayProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
}

func (p *ScalewayProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		secret.NewResourceSecret,
	}
}

func (p *ScalewayProvider) EphemeralResources(_ context.Context) []func() ephemeral.EphemeralResource {
	return []func() ephemeral.EphemeralResource{}
}

func (p *ScalewayProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		secret.NewDataSourceSecret,
	}
}

func (p *ScalewayProvider) Actions(_ context.Context) []func() action.Action {
	return []func() action.Action{}
}

func (p *ScalewayProvider) ListResources(_ context.Context) []func() list.ListResource {
	return []func() list.ListResource{}
}
