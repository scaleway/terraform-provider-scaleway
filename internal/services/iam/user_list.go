package iam

import (
	"context"
	"errors"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/list"
	"github.com/hashicorp/terraform-plugin-framework/list/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-mux/tf5to6server/translate"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	iamSDK "github.com/scaleway/scaleway-sdk-go/api/iam/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/identity"
	listscw "github.com/scaleway/terraform-provider-scaleway/v2/internal/list"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
)

var (
	_ list.ListResource                 = (*UserListResource)(nil)
	_ list.ListResourceWithConfigure    = (*UserListResource)(nil)
	_ list.ListResourceWithRawV6Schemas = (*UserListResource)(nil)
)

type UserListResource struct {
	meta   *meta.Meta
	iamAPI *iamSDK.API
}

func (r *UserListResource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	m := listscw.ConfigureMeta(request, response)
	if m == nil {
		return
	}

	r.meta = m
	r.iamAPI = iamSDK.NewAPI(meta.ExtractScwClient(m))
}

func NewUserListResource() list.ListResource {
	return &UserListResource{}
}

func (r *UserListResource) ListResourceConfigSchema(_ context.Context, _ list.ListResourceSchemaRequest, response *list.ListResourceSchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"organization_id": listscw.OrganizationIDAttribute("Organization ID to filter for"),
			"tag":             listscw.TagAttribute("Filter by tags containing a given string"),
			"mfa": schema.BoolAttribute{
				Description: "Filter by MFA status",
				Optional:    true,
			},
			"type": schema.StringAttribute{
				Description: "Filter by user type",
				Optional:    true,
			},
			"user_ids": schema.ListAttribute{
				Description: "Filter users by user IDs",
				ElementType: types.StringType,
				Optional:    true,
			},
		},
	}
}

func (r *UserListResource) RawV6Schemas(ctx context.Context, req list.RawV6SchemaRequest, resp *list.RawV6SchemaResponse) {
	userResource := ResourceUser()

	resp.ProtoV6Schema = translate.Schema(userResource.ProtoSchema(ctx)())
	resp.ProtoV6IdentitySchema = translate.ResourceIdentitySchema(userResource.ProtoIdentitySchema(ctx)())
}

type UserListResourceModel struct {
	UserIDs        types.List   `tfsdk:"user_ids"`
	OrganizationID types.String `tfsdk:"organization_id"`
	Tag            types.String `tfsdk:"tag"`
	Type           types.String `tfsdk:"type"`
	Mfa            types.Bool   `tfsdk:"mfa"`
}

func (r *UserListResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_iam_user"
}

func (r *UserListResource) FetchUsers(ctx context.Context, data UserListResourceModel) ([]*iamSDK.User, error) {
	request := &iamSDK.ListUsersRequest{
		OrganizationID: data.OrganizationID.ValueStringPointer(),
	}

	if request.OrganizationID == nil || *request.OrganizationID == "" {
		defaultOrgID, exists := r.meta.ScwClient().GetDefaultOrganizationID()
		if exists {
			request.OrganizationID = &defaultOrgID
		} else {
			return nil, errors.New("organization ID is required. Either set organization_id or configure a default organization")
		}
	}

	if !data.Tag.IsNull() && !data.Tag.IsUnknown() {
		tag := data.Tag.ValueString()
		request.Tag = &tag
	}

	if !data.Mfa.IsNull() && !data.Mfa.IsUnknown() {
		mfa := data.Mfa.ValueBool()
		request.Mfa = &mfa
	}

	if !data.Type.IsNull() && !data.Type.IsUnknown() {
		userType := data.Type.ValueString()
		request.Type = iamSDK.UserType(userType)
	}

	if !data.UserIDs.IsNull() && !data.UserIDs.IsUnknown() {
		var userIDs []string
		data.UserIDs.ElementsAs(ctx, &userIDs, false)

		if len(userIDs) > 0 {
			request.UserIDs = userIDs
		}
	}

	response, err := r.iamAPI.ListUsers(request, scw.WithContext(ctx), scw.WithAllPages())
	if err != nil {
		return nil, err
	}

	return response.Users, nil
}

func (r *UserListResource) List(ctx context.Context, req list.ListRequest, stream *list.ListResultsStream) {
	var data UserListResourceModel

	diags := req.Config.Get(ctx, &data)
	if diags.HasError() {
		stream.Results = list.ListResultsStreamDiagnostics(diags)

		return
	}

	allUsers, err := r.FetchUsers(ctx, data)
	if err != nil {
		stream.Results = list.ListResultsStreamDiagnostics(diag.Diagnostics{
			diag.NewErrorDiagnostic("Listing IAM Users", "Failed to list IAM Users: "+err.Error()),
		})

		return
	}

	stream.Results = func(push func(list.ListResult) bool) {
		for _, user := range allUsers {
			result := req.NewListResult(ctx)
			result.DisplayName = user.Email

			userResource := ResourceUser()
			resourceData := userResource.Data(&terraform.InstanceState{})

			err = identity.SetGlobalIdentity(resourceData, user.ID)
			if err != nil {
				result.Diagnostics.AddError("Retrieving identity data",
					"An error was encountered when retrieving the identity data: "+err.Error(),
				)

				if !push(result) {
					return
				}

				continue
			}

			tfTypeIdentity, errIdentityState := resourceData.TfTypeIdentityState()
			if errIdentityState != nil {
				result.Diagnostics.AddError(
					"Converting identity data",
					"An error was encountered when converting the identity data: "+errIdentityState.Error(),
				)
			}

			identitySetDiags := result.Identity.Set(ctx, *tfTypeIdentity)
			result.Diagnostics.Append(identitySetDiags...)

			setUserState(resourceData, user)

			tfTypeResource, errTfTypeResourceState := resourceData.TfTypeResourceState()
			if errTfTypeResourceState != nil {
				result.Diagnostics.AddError(
					"Converting resource state",
					"An error was encountered when converting the resource state: "+errTfTypeResourceState.Error(),
				)
			}

			resourceSetDiags := result.Resource.Set(ctx, *tfTypeResource)
			result.Diagnostics.Append(resourceSetDiags...)

			if !push(result) {
				return
			}
		}
	}
}
