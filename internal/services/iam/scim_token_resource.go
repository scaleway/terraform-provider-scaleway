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
	"github.com/hashicorp/terraform-plugin-framework/types"
	iam "github.com/scaleway/scaleway-sdk-go/api/iam/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/verify"
)

var (
	_ resource.Resource                = (*ScimTokenResource)(nil)
	_ resource.ResourceWithConfigure   = (*ScimTokenResource)(nil)
	_ resource.ResourceWithImportState = (*ScimTokenResource)(nil)
)

func NewScimTokenResource() resource.Resource {
	return &ScimTokenResource{}
}

type ScimTokenResource struct {
	iamAPI *iam.API
	meta   *meta.Meta
}

type scimTokenResourceModel struct {
	ScimID         types.String `tfsdk:"scim_id"`
	OrganizationID types.String `tfsdk:"organization_id"`
	// Output
	ID          types.String `tfsdk:"id"`
	BearerToken types.String `tfsdk:"bearer_token"`
	CreatedAt   types.String `tfsdk:"created_at"`
	ExpiresAt   types.String `tfsdk:"expires_at"`
}

func (r *ScimTokenResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_iam_scim_token"
}

//go:embed descriptions/scim_token_resource.md
var scimTokenResourceDescription string

func (r *ScimTokenResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: scimTokenResourceDescription,
		Attributes: map[string]schema.Attribute{
			"scim_id": schema.StringAttribute{
				MarkdownDescription: "The SCIM configuration ID for which to create the token.",
				Required:            true,
				Validators: []validator.String{
					verify.IsStringUUID(),
				},
			},
			"organization_id": schema.StringAttribute{
				MarkdownDescription: "The organization ID. If not provided, the default organization configured in the provider is used.",
				Optional:            true,
				Computed:            true,
				Validators: []validator.String{
					verify.IsStringUUID(),
				},
			},
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The ID of the SCIM token",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"bearer_token": schema.StringAttribute{
				MarkdownDescription: "The Bearer Token to use to authenticate to SCIM endpoints.",
				Computed:            true,
				Sensitive:           true,
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

func (r *ScimTokenResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *ScimTokenResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data scimTokenResourceModel
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

	res, err := r.iamAPI.CreateScimToken(&iam.CreateScimTokenRequest{
		ScimID: data.ScimID.ValueString(),
	}, scw.WithContext(ctx))
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to create SCIM token",
			err.Error(),
		)

		return
	}

	state := convertScimTokenToState(res, orgID, data.ScimID.ValueString())
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *ScimTokenResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state scimTokenResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	tokenID := state.ID.ValueString()
	if tokenID == "" {
		resp.Diagnostics.AddError(
			"Token ID not set",
			"Cannot refresh SCIM token without an ID",
		)

		return
	}

	if state.ScimID.ValueString() == "" {
		orgID := state.OrganizationID.ValueString()
		if orgID == "" {
			defaultOrgID, exists := r.meta.ScwClient().GetDefaultOrganizationID()
			if exists {
				orgID = defaultOrgID
			} else {
				resp.Diagnostics.AddError(
					"Organization ID required",
					"Cannot determine SCIM configuration without organization_id",
				)

				return
			}
		}

		scimResp, err := r.iamAPI.GetOrganizationScim(&iam.GetOrganizationScimRequest{
			OrganizationID: orgID,
		}, scw.WithContext(ctx))
		if err != nil {
			resp.Diagnostics.AddError(
				"Failed to get SCIM configuration",
				fmt.Sprintf("Could not retrieve SCIM configuration for organization %s: %v", orgID, err),
			)

			return
		}

		state.ScimID = types.StringValue(scimResp.ID)
	}

	listResp, err := r.iamAPI.ListScimTokens(&iam.ListScimTokensRequest{
		ScimID: state.ScimID.ValueString(),
	}, scw.WithContext(ctx))
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to list SCIM tokens",
			fmt.Sprintf("Could not list SCIM tokens for SCIM configuration %s: %v", state.ScimID.ValueString(), err),
		)

		return
	}

	found := false

	for _, token := range listResp.ScimTokens {
		if token.ID == tokenID {
			found = true

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
		resp.Diagnostics.AddWarning(
			"SCIM token not found in API",
			fmt.Sprintf("SCIM token %s was not found in the API. It may have been deleted. The state will be returned as-is, but you may need to recreate the token.", tokenID),
		)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *ScimTokenResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError(
		"Update not supported",
		"SCIM token does not support updates. To rotate the token, create a new one.",
	)
}

func (r *ScimTokenResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state scimTokenResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	err := r.iamAPI.DeleteScimToken(&iam.DeleteScimTokenRequest{
		TokenID: state.ID.ValueString(),
	}, scw.WithContext(ctx))
	if err != nil {
		if httperrors.Is404(err) {
			return
		}

		resp.Diagnostics.AddError(
			"Failed to delete SCIM token",
			err.Error(),
		)

		return
	}
}

func (r *ScimTokenResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)

	if orgID, exists := r.meta.ScwClient().GetDefaultOrganizationID(); exists {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("organization_id"), orgID)...)
	}
}

func convertScimTokenToState(res *iam.CreateScimTokenResponse, orgID, scimID string) scimTokenResourceModel {
	state := scimTokenResourceModel{
		ID:             types.StringValue(res.Token.ID),
		ScimID:         types.StringValue(scimID),
		OrganizationID: types.StringValue(orgID),
		BearerToken:    types.StringValue(res.BearerToken),
	}

	if res.Token.CreatedAt != nil {
		state.CreatedAt = types.StringValue(res.Token.CreatedAt.String())
	}

	if res.Token.ExpiresAt != nil {
		state.ExpiresAt = types.StringValue(res.Token.ExpiresAt.String())
	}

	return state
}
