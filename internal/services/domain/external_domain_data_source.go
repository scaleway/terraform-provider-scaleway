package domain

import (
	"context"
	_ "embed"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	domainsdk "github.com/scaleway/scaleway-sdk-go/api/domain/v2beta1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
)

var (
	_ datasource.DataSource              = (*ExternalDomainDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*ExternalDomainDataSource)(nil)
)

func NewExternalDomainDataSource() datasource.DataSource {
	return &ExternalDomainDataSource{}
}

type ExternalDomainDataSource struct {
	api  *domainsdk.RegistrarAPI
	meta *meta.Meta
}

type externalDomainDataSourceModel struct {
	ID             types.String `tfsdk:"id"`
	Domain         types.String `tfsdk:"domain"`
	ProjectID      types.String `tfsdk:"project_id"`
	OrganizationID types.String `tfsdk:"organization_id"`
	NSServers      types.List   `tfsdk:"ns_servers"`
	Status         types.String `tfsdk:"status"`
}

//go:embed descriptions/domain_external_data_source.md
var externalDomainDataSourceDescription string

func (d *ExternalDomainDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_domain_external"
}

func (d *ExternalDomainDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: externalDomainDataSourceDescription,
		Description:         externalDomainDataSourceDescription,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The domain name",
			},
			"domain": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The domain name to look up",
			},
			"project_id": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "The project ID the domain belongs to",
			},
			"organization_id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The organization ID the domain is associated with",
			},
			"ns_servers": schema.ListAttribute{
				ElementType:         types.StringType,
				Computed:            true,
				MarkdownDescription: "List of NS servers for the domain",
			},
			"status": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The status of the domain",
			},
		},
	}
}

func (d *ExternalDomainDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	m, ok := req.ProviderData.(*meta.Meta)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *meta.Meta, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	d.meta = m
	d.api = domainsdk.NewRegistrarAPI(m.ScwClient())
}

func (d *ExternalDomainDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config externalDomainDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)

	if resp.Diagnostics.HasError() {
		return
	}

	domainName := config.Domain.ValueString()

	domainResp, err := d.api.GetDomain(&domainsdk.RegistrarAPIGetDomainRequest{
		Domain: domainName,
	}, scw.WithContext(ctx))
	if err != nil {
		if httperrors.Is404(err) {
			resp.Diagnostics.AddError(
				"External domain not found",
				fmt.Sprintf("No external domain found with name %q", domainName),
			)

			return
		}

		resp.Diagnostics.AddError("Failed to get external domain", err.Error())

		return
	}

	nsServers := []string{}
	if len(domainResp.DNSZones) > 0 {
		nsServers = domainResp.DNSZones[0].NsDefault
	}

	nsList, diags := types.ListValueFrom(ctx, types.StringType, nsServers)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	state := externalDomainDataSourceModel{
		ID:             types.StringValue(domainResp.Domain),
		Domain:         types.StringValue(domainResp.Domain),
		ProjectID:      types.StringValue(domainResp.ProjectID),
		OrganizationID: types.StringValue(domainResp.OrganizationID),
		NSServers:      nsList,
		Status:         types.StringValue(domainResp.Status.String()),
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
