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
	_ resource.Resource                = (*ScimResource)(nil)
	_ resource.ResourceWithConfigure   = (*ScimResource)(nil)
	_ resource.ResourceWithImportState = (*ScimResource)(nil)
)

func NewScimResource() resource.Resource {
	return &ScimResource{}
}

type ScimResource struct {
	iamAPI *iam.API
	meta   *meta.Meta
}

type scimResourceModel struct {
	OrganizationID types.String `tfsdk:"organization_id"`
	// Output
	ID        types.String `tfsdk:"id"`
	CreatedAt types.String `tfsdk:"created_at"`
}

func (r *ScimResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_iam_scim"
}

//go:embed descriptions/scim_resource.md
var scimResourceDescription string

func (r *ScimResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: scimResourceDescription,
		Attributes: map[string]schema.Attribute{
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
				MarkdownDescription: "The ID of the SCIM configuration",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"created_at": schema.StringAttribute{
				MarkdownDescription: "The date and time of SCIM configuration creation",
				Computed:            true,
			},
		},
	}
}

func (r *ScimResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *ScimResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data scimResourceModel
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

	_, err := r.iamAPI.GetOrganizationScim(&iam.GetOrganizationScimRequest{
		OrganizationID: orgID,
	}, scw.WithContext(ctx))

	var state scimResourceModel

	if err != nil {
		if httperrors.Is404(err) {
			res, err := r.iamAPI.EnableOrganizationScim(&iam.EnableOrganizationScimRequest{
				OrganizationID: orgID,
			}, scw.WithContext(ctx))
			if err != nil {
				resp.Diagnostics.AddError(
					"Failed to enable SCIM",
					err.Error(),
				)

				return
			}

			state = convertScimToState(res, orgID)
			resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
		} else {
			resp.Diagnostics.AddError(
				"Failed to check SCIM status",
				err.Error(),
			)

			return
		}
	} else {
		resp.Diagnostics.AddError(
			"SCIM already enabled",
			"SCIM configuration is already enabled for this organization.",
		)

		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *ScimResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state scimResourceModel
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

	scim, err := r.iamAPI.GetOrganizationScim(&iam.GetOrganizationScimRequest{
		OrganizationID: orgID,
	}, scw.WithContext(ctx))
	if err != nil {
		if httperrors.Is404(err) {
			resp.State.RemoveResource(ctx)

			return
		}

		resp.Diagnostics.AddError(
			"Failed to read SCIM",
			err.Error(),
		)

		return
	}

	state = convertScimToState(scim, orgID)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *ScimResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError(
		"Update not supported",
		"SCIM configuration does not support updates. Use the resource to enable/disable SCIM.",
	)
}

func (r *ScimResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state scimResourceModel
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

	existingScim, err := r.iamAPI.GetOrganizationScim(&iam.GetOrganizationScimRequest{
		OrganizationID: orgID,
	}, scw.WithContext(ctx))
	if err != nil {
		if httperrors.Is404(err) {
			return
		} else {
			resp.Diagnostics.AddError(
				"Failed to check SCIM status",
				err.Error(),
			)

			return
		}
	}

	err = r.iamAPI.DeleteScim(&iam.DeleteScimRequest{
		ScimID: existingScim.ID,
	}, scw.WithContext(ctx))
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to disable SCIM",
			err.Error(),
		)

		return
	}
}

func (r *ScimResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func convertScimToState(scim *iam.Scim, orgID string) scimResourceModel {
	state := scimResourceModel{
		ID:             types.StringValue(scim.ID),
		OrganizationID: types.StringValue(orgID),
	}

	if scim.CreatedAt != nil {
		state.CreatedAt = types.StringValue(scim.CreatedAt.String())
	}

	return state
}
