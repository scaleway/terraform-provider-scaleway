package iam

import (
	"context"
	_ "embed"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/action"
	"github.com/hashicorp/terraform-plugin-framework/action/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	iam "github.com/scaleway/scaleway-sdk-go/api/iam/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/verify"
)

var (
	_ action.Action              = (*SamlConfigurationAction)(nil)
	_ action.ActionWithConfigure = (*SamlConfigurationAction)(nil)
)

type SamlConfigurationAction struct {
	iamAPI *iam.API
	meta   *meta.Meta
}

type SamlConfigurationActionModel struct {
	OrganizationID  types.String `tfsdk:"organization_id"`
	EntityID        types.String `tfsdk:"entity_id"`
	SingleSignOnURL types.String `tfsdk:"single_sign_on_url"`
}

func NewSamlConfigurationAction() action.Action {
	return &SamlConfigurationAction{}
}

//go:embed descriptions/update_saml_configuration_action.md
var updateSamlConfigurationActionDescription string

func (a *SamlConfigurationAction) Metadata(ctx context.Context, req action.MetadataRequest, resp *action.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_iam_update_saml_configuration"
}

func (a *SamlConfigurationAction) Configure(ctx context.Context, req action.ConfigureRequest, resp *action.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	m, ok := req.ProviderData.(*meta.Meta)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Action Configure Type",
			fmt.Sprintf("Expected *meta.Meta, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	a.meta = m
	a.iamAPI = iam.NewAPI(a.meta.ScwClient())
}

func (a *SamlConfigurationAction) Schema(ctx context.Context, req action.SchemaRequest, resp *action.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: updateSamlConfigurationActionDescription,
		Description:         updateSamlConfigurationActionDescription,
		Attributes: map[string]schema.Attribute{
			"organization_id": schema.StringAttribute{
				MarkdownDescription: "The organization ID. If not provided, the default organization configured in the provider is used.",
				Optional:            true,
				Validators: []validator.String{
					verify.IsStringUUID(),
				},
			},
			"entity_id": schema.StringAttribute{
				MarkdownDescription: "The entity ID of the SAML Identity Provider",
				Optional:            true,
			},
			"single_sign_on_url": schema.StringAttribute{
				MarkdownDescription: "The single sign-on URL of the SAML Identity Provider",
				Optional:            true,
			},
		},
	}
}

func (a *SamlConfigurationAction) Invoke(ctx context.Context, req action.InvokeRequest, resp *action.InvokeResponse) {
	var data SamlConfigurationActionModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	orgID := data.OrganizationID.ValueString()
	if orgID == "" {
		defaultOrgID, exists := a.meta.ScwClient().GetDefaultOrganizationID()
		if exists {
			orgID = defaultOrgID
		} else {
			resp.Diagnostics.AddError(
				"Organization ID is required",
				"Either set organization_id or configure a default organization",
			)

			return
		}
	}

	saml, err := a.iamAPI.GetOrganizationSaml(&iam.GetOrganizationSamlRequest{
		OrganizationID: orgID,
	}, scw.WithContext(ctx))
	if err != nil {
		if httperrors.Is404(err) {
			resp.Diagnostics.AddError(
				"SAML not enabled",
				"SAML must be enabled before configuring it. Please use scaleway_iam_saml resource first.",
			)

			return
		}

		resp.Diagnostics.AddError(
			"Failed to check SAML status",
			err.Error(),
		)

		return
	}

	reqUpdate := &iam.UpdateSamlRequest{
		SamlID: saml.ID,
	}

	if !data.EntityID.IsUnknown() && !data.EntityID.IsNull() {
		entityID := data.EntityID.ValueString()
		reqUpdate.EntityID = &entityID
	}

	if !data.SingleSignOnURL.IsUnknown() && !data.SingleSignOnURL.IsNull() {
		ssoURL := data.SingleSignOnURL.ValueString()
		reqUpdate.SingleSignOnURL = &ssoURL
	}

	_, err = a.iamAPI.UpdateSaml(reqUpdate, scw.WithContext(ctx))
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to update SAML configuration",
			err.Error(),
		)

		return
	}
}
