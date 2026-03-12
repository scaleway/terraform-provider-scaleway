package iam

import (
	"context"
	_ "embed"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	iam "github.com/scaleway/scaleway-sdk-go/api/iam/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/verify"
)

var (
	_ resource.Resource                = (*SamlConfigurationResource)(nil)
	_ resource.ResourceWithConfigure   = (*SamlConfigurationResource)(nil)
	_ resource.ResourceWithImportState = (*SamlConfigurationResource)(nil)
)

func NewSamlConfigurationResource() resource.Resource {
	return &SamlConfigurationResource{}
}

type SamlConfigurationResource struct {
	iamAPI *iam.API
	meta   *meta.Meta
}

func (r *SamlConfigurationResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_iam_saml_configuration"
}

//go:embed descriptions/saml_configuration_resource.md
var samlConfigurationResourceDescription string

func (r *SamlConfigurationResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: samlConfigurationResourceDescription,
		Attributes: map[string]schema.Attribute{
			"organization_id": schema.StringAttribute{
				MarkdownDescription: "The organization ID",
				Optional:            true,
				Computed:            true,
				Validators: []validator.String{
					verify.IsStringUUID(),
				},
			},
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The ID of the SAML configuration",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
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
			"status": schema.StringAttribute{
				MarkdownDescription: "The status of the SAML configuration",
				Computed:            true,
			},
			"service_provider": schema.ObjectAttribute{
				MarkdownDescription: "The Service Provider information",
				Computed:            true,
				AttributeTypes:      getServiceProviderAttrTypes(),
			},
		},
	}
}

func (r *SamlConfigurationResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	m, ok := req.ProviderData.(*meta.Meta)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *meta.Meta, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.meta = m
	r.iamAPI = iam.NewAPI(r.meta.ScwClient())
}

func (r *SamlConfigurationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data samlResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	orgID := data.OrganizationID.ValueString()
	if orgID == "" {
		defaultOrgID, exists := r.meta.ScwClient().GetDefaultOrganizationID()
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

	saml, err := r.iamAPI.GetOrganizationSaml(&iam.GetOrganizationSamlRequest{
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

	updatedSaml, err := r.iamAPI.UpdateSaml(reqUpdate, scw.WithContext(ctx))
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to update SAML configuration",
			err.Error(),
		)

		return
	}

	state := convertSamlToState(updatedSaml, orgID, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *SamlConfigurationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state samlResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	orgID := state.OrganizationID.ValueString()
	if orgID == "" {
		defaultOrgID, exists := r.meta.ScwClient().GetDefaultOrganizationID()
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

	saml, err := r.iamAPI.GetOrganizationSaml(&iam.GetOrganizationSamlRequest{
		OrganizationID: orgID,
	}, scw.WithContext(ctx))
	if err != nil {
		if httperrors.Is404(err) {
			resp.State.RemoveResource(ctx)

			return
		}

		resp.Diagnostics.AddError(
			"Failed to read SAML configuration",
			err.Error(),
		)

		return
	}

	state = convertSamlToState(saml, orgID, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *SamlConfigurationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data samlResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	orgID := data.OrganizationID.ValueString()
	if orgID == "" {
		defaultOrgID, exists := r.meta.ScwClient().GetDefaultOrganizationID()
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

	existingSaml, err := r.iamAPI.GetOrganizationSaml(&iam.GetOrganizationSamlRequest{
		OrganizationID: orgID,
	}, scw.WithContext(ctx))
	if err != nil {
		if httperrors.Is404(err) {
			resp.Diagnostics.AddError(
				"SAML not enabled",
				"SAML configuration is not currently enabled for this organization.",
			)

			return
		} else {
			resp.Diagnostics.AddError(
				"Failed to check SAML status",
				err.Error(),
			)

			return
		}
	}

	hasChanged := false
	reqUpdate := &iam.UpdateSamlRequest{
		SamlID: existingSaml.ID,
	}

	if !data.EntityID.IsUnknown() && !data.EntityID.IsNull() && data.EntityID.ValueString() != existingSaml.EntityID {
		entityID := data.EntityID.ValueString()
		reqUpdate.EntityID = &entityID
		hasChanged = true
	}

	if !data.SingleSignOnURL.IsUnknown() && !data.SingleSignOnURL.IsNull() && data.SingleSignOnURL.ValueString() != existingSaml.SingleSignOnURL {
		ssoURL := data.SingleSignOnURL.ValueString()
		reqUpdate.SingleSignOnURL = &ssoURL
		hasChanged = true
	}

	if hasChanged {
		saml, err := r.iamAPI.UpdateSaml(reqUpdate, scw.WithContext(ctx))
		if err != nil {
			resp.Diagnostics.AddError(
				"Failed to update SAML configuration",
				err.Error(),
			)

			return
		}

		updatedSaml := convertSamlToState(saml, orgID, &resp.Diagnostics)
		resp.Diagnostics.Append(resp.State.Set(ctx, &updatedSaml)...)
	}
}

func (r *SamlConfigurationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// SAML configuration resource doesn't delete the SAML configuration itself
}

func (r *SamlConfigurationResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
