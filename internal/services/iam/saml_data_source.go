package iam

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	iam "github.com/scaleway/scaleway-sdk-go/api/iam/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
)

var (
	_ datasource.DataSource              = (*SamlDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*SamlDataSource)(nil)
)

func NewSamlDataSource() datasource.DataSource {
	return &SamlDataSource{}
}

type SamlDataSource struct {
	iamAPI *iam.API
	meta   *meta.Meta
}

type ServiceProviderModel struct {
	EntityID                    types.String `tfsdk:"entity_id"`
	AssertionConsumerServiceUrl types.String `tfsdk:"assertion_consumer_service_url"`
}

type samlDataSourceModel struct {
	OrganizationID types.String `tfsdk:"organization_id"`
	// Output
	ID              types.String `tfsdk:"id"`
	Status          types.String `tfsdk:"status"`
	ServiceProvider types.Object `tfsdk:"service_provider"`
	EntityID        types.String `tfsdk:"entity_id"`
	SingleSignOnURL types.String `tfsdk:"single_sign_on_url"`
}

func (d *SamlDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_iam_saml"
}

func (d *SamlDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "SAML configuration data source",
		Attributes: map[string]schema.Attribute{
			"organization_id": schema.StringAttribute{
				MarkdownDescription: "The organization ID",
				Optional:            true,
				Computed:            true,
			},
			"id": schema.StringAttribute{
				MarkdownDescription: "The ID of the SAML configuration",
				Computed:            true,
			},
			"status": schema.StringAttribute{
				MarkdownDescription: "The status of the SAML configuration",
				Computed:            true,
			},
			"service_provider": schema.ObjectAttribute{
				MarkdownDescription: "The Service Provider information",
				Computed:            true,
				AttributeTypes:      getServiceProviderAttrTypes(),
			},
			"entity_id": schema.StringAttribute{
				MarkdownDescription: "The entity ID of the SAML Identity Provider",
				Computed:            true,
			},
			"single_sign_on_url": schema.StringAttribute{
				MarkdownDescription: "The single sign-on URL of the SAML Identity Provider",
				Computed:            true,
			},
		},
	}
}

func (d *SamlDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
	d.iamAPI = iam.NewAPI(d.meta.ScwClient())
}

func (d *SamlDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state samlDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)

	// TODO: use helper func
	orgID := state.OrganizationID.ValueString()
	if orgID == "" {
		defaultOrgID, exists := d.meta.ScwClient().GetDefaultOrganizationID()
		if exists {
			orgID = defaultOrgID
		} else {
			resp.Diagnostics.AddAttributeError(
				path.Root("organization_id"),
				"Organization ID is required",
				"Either set organization_id or configure a default organization",
			)
			return
		}
	}

	res, err := d.iamAPI.GetOrganizationSaml(&iam.GetOrganizationSamlRequest{
		OrganizationID: orgID,
	}, scw.WithContext(ctx))
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to read SAML",
			err.Error(),
		)
		return
	}

	state.OrganizationID = types.StringValue(orgID)
	state.ID = types.StringValue(res.ID)
	state.Status = types.StringValue(string(res.Status))
	state.ServiceProvider = getServiceProviderObject(res.ServiceProvider, &resp.Diagnostics)
	state.EntityID = types.StringValue(res.EntityID)
	state.SingleSignOnURL = types.StringValue(res.SingleSignOnURL)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
