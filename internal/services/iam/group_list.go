package iam

import (
	"context"

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
	_ list.ListResource                 = (*GroupListResource)(nil)
	_ list.ListResourceWithConfigure    = (*GroupListResource)(nil)
	_ list.ListResourceWithRawV6Schemas = (*GroupListResource)(nil)
)

type GroupListResource struct {
	meta   *meta.Meta
	iamAPI *iamSDK.API
}

func (r *GroupListResource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	m := listscw.ConfigureMeta(request, response)
	if m == nil {
		return
	}

	r.meta = m
	r.iamAPI = iamSDK.NewAPI(meta.ExtractScwClient(m))
}

func NewGroupListResource() list.ListResource {
	return &GroupListResource{}
}

func (r *GroupListResource) ListResourceConfigSchema(_ context.Context, _ list.ListResourceSchemaRequest, response *list.ListResourceSchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"organization_id": listscw.OrganizationIDAttribute("Organization ID to filter for"),
			"name":            listscw.NameAttribute("Name of the group to filter for"),
			"tag":             listscw.TagAttribute("Filter by tags containing a given string"),
			"user_ids": schema.ListAttribute{
				Description: "Filter groups by user IDs",
				ElementType: types.StringType,
				Optional:    true,
			},
			"application_ids": schema.ListAttribute{
				Description: "Filter groups by application IDs",
				ElementType: types.StringType,
				Optional:    true,
			},
		},
	}
}

func (r *GroupListResource) RawV6Schemas(ctx context.Context, req list.RawV6SchemaRequest, resp *list.RawV6SchemaResponse) {
	groupResource := ResourceGroup()

	resp.ProtoV6Schema = translate.Schema(groupResource.ProtoSchema(ctx)())
	resp.ProtoV6IdentitySchema = translate.ResourceIdentitySchema(groupResource.ProtoIdentitySchema(ctx)())
}

type GroupListResourceModel struct {
	OrganizationID types.String `tfsdk:"organization_id"`
	Name           types.String `tfsdk:"name"`
	Tag            types.String `tfsdk:"tag"`
	UserIDs        types.List   `tfsdk:"user_ids"`
	ApplicationIDs types.List   `tfsdk:"application_ids"`
}

func (r *GroupListResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_iam_group"
}

func (r *GroupListResource) FetchGroups(ctx context.Context, data GroupListResourceModel) ([]*iamSDK.Group, error) {
	request := &iamSDK.ListGroupsRequest{
		OrganizationID: data.OrganizationID.ValueString(),
		Name:           data.Name.ValueStringPointer(),
	}

	if !data.Tag.IsNull() && !data.Tag.IsUnknown() {
		tag := data.Tag.ValueString()
		request.Tag = &tag
	}

	if !data.UserIDs.IsNull() && !data.UserIDs.IsUnknown() {
		var userIDs []string
		data.UserIDs.ElementsAs(ctx, &userIDs, false)

		if len(userIDs) > 0 {
			request.UserIDs = userIDs
		}
	}

	if !data.ApplicationIDs.IsNull() && !data.ApplicationIDs.IsUnknown() {
		var applicationIDs []string
		data.ApplicationIDs.ElementsAs(ctx, &applicationIDs, false)

		if len(applicationIDs) > 0 {
			request.ApplicationIDs = applicationIDs
		}
	}

	response, err := r.iamAPI.ListGroups(request, scw.WithContext(ctx), scw.WithAllPages())
	if err != nil {
		return nil, err
	}

	return response.Groups, nil
}

func (r *GroupListResource) List(ctx context.Context, req list.ListRequest, stream *list.ListResultsStream) {
	var data GroupListResourceModel

	diags := req.Config.Get(ctx, &data)
	if diags.HasError() {
		stream.Results = list.ListResultsStreamDiagnostics(diags)

		return
	}

	allGroups, err := r.FetchGroups(ctx, data)
	if err != nil {
		stream.Results = list.ListResultsStreamDiagnostics(diag.Diagnostics{
			diag.NewErrorDiagnostic("Listing IAM Groups", "Failed to list IAM Groups: "+err.Error()),
		})

		return
	}

	stream.Results = func(push func(list.ListResult) bool) {
		for _, group := range allGroups {
			result := req.NewListResult(ctx)
			result.DisplayName = group.Name

			groupResource := ResourceGroup()
			resourceData := groupResource.Data(&terraform.InstanceState{})

			err = identity.SetGlobalIdentity(resourceData, group.ID)
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

			setGroupState(resourceData, group, false)

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
