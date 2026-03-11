package iam

import (
	"context"
	_ "embed"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	iam "github.com/scaleway/scaleway-sdk-go/api/iam/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/verify"
)

var (
	_ resource.Resource                = (*SamlResource)(nil)
	_ resource.ResourceWithConfigure   = (*SamlResource)(nil)
	_ resource.ResourceWithImportState = (*SamlResource)(nil)
)

func NewSamlResource() resource.Resource {
	return &SamlResource{}
}

type SamlResource struct {
	iamAPI *iam.API
	meta   *meta.Meta
}

type samlResourceModel struct {
	OrganizationID types.String `tfsdk:"organization_id"`
	// Output
	ID              types.String `tfsdk:"id"`
	EntityID        types.String `tfsdk:"entity_id"`
	SingleSignOnURL types.String `tfsdk:"single_sign_on_url"`
	Status          types.String `tfsdk:"status"`
	ServiceProvider types.Object `tfsdk:"service_provider"`
}

func (r *SamlResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_iam_saml"
}

func getServiceProviderAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"entity_id":                      types.StringType,
		"assertion_consumer_service_url": types.StringType,
	}
}

func getServiceProviderObject(data *iam.SamlServiceProvider, diags *diag.Diagnostics) types.Object {
	if data == nil {
		return types.ObjectNull(getServiceProviderAttrTypes())
	}

	serviceProviderModel := ServiceProviderModel{
		EntityID:                    types.StringValue(data.EntityID),
		AssertionConsumerServiceUrl: types.StringValue(data.AssertionConsumerServiceURL),
	}

	attrValues := map[string]attr.Value{
		"entity_id":                      serviceProviderModel.EntityID,
		"assertion_consumer_service_url": serviceProviderModel.AssertionConsumerServiceUrl,
	}

	obj, d := types.ObjectValue(getServiceProviderAttrTypes(), attrValues)
	if diags != nil {
		diags.Append(d...)
	}

	return obj
}

//go:embed descriptions/saml_resource.md
var samlResourceDescription string

func (r *SamlResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: samlResourceDescription,
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
				Computed:            true,
			},
			"single_sign_on_url": schema.StringAttribute{
				MarkdownDescription: "The single sign-on URL of the SAML Identity Provider",
				Optional:            true,
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
		},
	}
}

func (r *SamlResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *SamlResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
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

	_, err := r.iamAPI.GetOrganizationSaml(&iam.GetOrganizationSamlRequest{
		OrganizationID: orgID,
	}, scw.WithContext(ctx))
	if err != nil {
		if httperrors.Is404(err) {
			res, err := r.iamAPI.EnableOrganizationSaml(&iam.EnableOrganizationSamlRequest{
				OrganizationID: orgID,
			}, scw.WithContext(ctx))
			if err != nil {
				resp.Diagnostics.AddError(
					"Failed to enable SAML",
					err.Error(),
				)

				return
			}

			state := r.convertToState(res, orgID, &resp.Diagnostics)
			resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
		} else {
			resp.Diagnostics.AddError(
				"Failed to check SAML status",
				err.Error(),
			)
		}
	} else {
		resp.Diagnostics.AddError(
			"SAML already enabled",
			"SAML configuration is already enabled for this organization.",
		)
	}
	// The read is deliberately skipped since all computed attributes have been set in the create.
}

func (r *SamlResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
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
			"Failed to read SAML",
			err.Error(),
		)

		return
	}

	state = r.convertToState(saml, orgID, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *SamlResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
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
				"Failed to update SAML",
				err.Error(),
			)

			return
		}

		updatedSaml := r.convertToState(saml, orgID, &resp.Diagnostics)
		resp.Diagnostics.Append(resp.State.Set(ctx, &updatedSaml)...)
	}
}

func (r *SamlResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state samlResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// TODO: use helper func
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

	existingSaml, err := r.iamAPI.GetOrganizationSaml(&iam.GetOrganizationSamlRequest{
		OrganizationID: orgID,
	}, scw.WithContext(ctx))
	if err != nil {
		if httperrors.Is404(err) {
			return
		} else {
			resp.Diagnostics.AddError(
				"Failed to check SAML status",
				err.Error(),
			)

			return
		}
	}

	err = r.iamAPI.DeleteSaml(&iam.DeleteSamlRequest{
		SamlID: existingSaml.ID,
	}, scw.WithContext(ctx))
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to disable SAML",
			err.Error(),
		)

		return
	}
}

func (r *SamlResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *SamlResource) convertToState(saml *iam.Saml, orgID string, diags *diag.Diagnostics) samlResourceModel {
	return samlResourceModel{
		ID:              types.StringValue(saml.ID),
		EntityID:        types.StringValue(saml.EntityID),
		SingleSignOnURL: types.StringValue(saml.SingleSignOnURL),
		Status:          types.StringValue(string(saml.Status)),
		OrganizationID:  types.StringValue(orgID),
		ServiceProvider: getServiceProviderObject(saml.ServiceProvider, diags),
	}
}
