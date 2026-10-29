package iam

import (
	"context"
	_ "embed"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	iam "github.com/scaleway/scaleway-sdk-go/api/iam/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/verify"
)

var (
	_ datasource.DataSource              = (*ScimTokenDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*ScimTokenDataSource)(nil)
)

func NewScimTokenDataSource() datasource.DataSource {
	return &ScimTokenDataSource{}
}

type ScimTokenDataSource struct {
	iamAPI *iam.API
	meta   *meta.Meta
}

type scimTokenDataSourceModel struct {
	TokenID        types.String `tfsdk:"token_id"`
	ScimID         types.String `tfsdk:"scim_id"`
	OrganizationID types.String `tfsdk:"organization_id"`
	// Output
	ID        types.String `tfsdk:"id"`
	CreatedAt types.String `tfsdk:"created_at"`
	ExpiresAt types.String `tfsdk:"expires_at"`
}

func (d *ScimTokenDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_iam_scim_token"
}

//go:embed descriptions/scim_token_data_source.md
var scimTokenDataSourceDescription string

func (d *ScimTokenDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: scimTokenDataSourceDescription,
		Attributes: map[string]schema.Attribute{
			"token_id": schema.StringAttribute{
				MarkdownDescription: "The ID of the SCIM token to retrieve.",
				Required:            true,
				Validators: []validator.String{
					verify.IsStringUUID(),
				},
			},
			"scim_id": schema.StringAttribute{
				MarkdownDescription: "The SCIM configuration ID. If not provided, the SCIM configuration for the organization is used.",
				Optional:            true,
				Computed:            true,
				Validators: []validator.String{
					verify.IsStringUUID(),
				},
			},
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The ID of the SCIM token",
			},
			"organization_id": schema.StringAttribute{
				MarkdownDescription: "The organization ID. If not provided, the default organization configured in the provider is used.",
				Optional:            true,
				Computed:            true,
				Validators: []validator.String{
					verify.IsStringUUID(),
				},
			},
			"created_at": schema.StringAttribute{
				MarkdownDescription: "The date and time of SCIM token creation",
				Computed:            true,
			},
			"expires_at": schema.StringAttribute{
				MarkdownDescription: "The date and time when the SCIM token expires",
				Computed:            true,
			},
		},
	}
}

func (d *ScimTokenDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *ScimTokenDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state scimTokenDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

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

	tokenID := state.TokenID.ValueString()
	if tokenID == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("token_id"),
			"Token ID is required",
			"The token_id attribute must be set",
		)

		return
	}

	scimID := state.ScimID.ValueString()
	if scimID == "" {
		scimResp, err := d.iamAPI.GetOrganizationScim(&iam.GetOrganizationScimRequest{
			OrganizationID: orgID,
		}, scw.WithContext(ctx))
		if err != nil {
			resp.Diagnostics.AddError(
				"Failed to get SCIM configuration",
				fmt.Sprintf("Could not retrieve SCIM configuration for organization %s: %v", orgID, err),
			)

			return
		}

		scimID = scimResp.ID
	}

	listResp, err := d.iamAPI.ListScimTokens(&iam.ListScimTokensRequest{
		ScimID: scimID,
	}, scw.WithContext(ctx))
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to list SCIM tokens",
			fmt.Sprintf("Could not list SCIM tokens for SCIM configuration %s: %v", scimID, err),
		)

		return
	}

	found := false

	for _, token := range listResp.ScimTokens {
		if token.ID == tokenID {
			found = true
			state.TokenID = types.StringValue(token.ID)
			state.ID = types.StringValue(token.ID)
			state.OrganizationID = types.StringValue(orgID)
			state.ScimID = types.StringValue(scimID)

			if token.CreatedAt != nil {
				state.CreatedAt = types.StringValue(token.CreatedAt.String())
			}

			if token.ExpiresAt != nil {
				state.ExpiresAt = types.StringValue(token.ExpiresAt.String())
			}

			break
		}
	}

	if !found {
		resp.Diagnostics.AddError(
			"SCIM token not found",
			fmt.Sprintf("SCIM token %s was not found in SCIM configuration %s", tokenID, scimID),
		)

		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
