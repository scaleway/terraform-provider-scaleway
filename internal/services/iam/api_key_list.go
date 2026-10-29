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
	_ list.ListResource                 = (*APIKeyListResource)(nil)
	_ list.ListResourceWithConfigure    = (*APIKeyListResource)(nil)
	_ list.ListResourceWithRawV6Schemas = (*APIKeyListResource)(nil)
)

type APIKeyListResource struct {
	meta   *meta.Meta
	iamAPI *iamSDK.API
}

func (r *APIKeyListResource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	m := listscw.ConfigureMeta(request, response)
	if m == nil {
		return
	}

	r.meta = m
	r.iamAPI = iamSDK.NewAPI(meta.ExtractScwClient(m))
}

func NewAPIKeyListResource() list.ListResource {
	return &APIKeyListResource{}
}

func (r *APIKeyListResource) ListResourceConfigSchema(_ context.Context, _ list.ListResourceSchemaRequest, response *list.ListResourceSchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"organization_id": listscw.OrganizationIDAttribute("Organization ID to filter for"),
			"editable": schema.BoolAttribute{
				Description: "Filter by editable status",
				Optional:    true,
			},
			"expired": schema.BoolAttribute{
				Description: "Filter by expired status",
				Optional:    true,
			},
			"description": schema.StringAttribute{
				Description: "Filter by description",
				Optional:    true,
			},
			"bearer_id": schema.StringAttribute{
				Description: "Filter by bearer ID",
				Optional:    true,
			},
			"bearer_type": schema.StringAttribute{
				Description: "Filter by type of bearer (user or application)",
				Optional:    true,
			},
			"access_keys": schema.ListAttribute{
				Description: "Filter by a list of access keys",
				ElementType: types.StringType,
				Optional:    true,
			},
		},
	}
}

func (r *APIKeyListResource) RawV6Schemas(ctx context.Context, req list.RawV6SchemaRequest, resp *list.RawV6SchemaResponse) {
	apiKeyResource := ResourceAPIKey()

	resp.ProtoV6Schema = translate.Schema(apiKeyResource.ProtoSchema(ctx)())
	resp.ProtoV6IdentitySchema = translate.ResourceIdentitySchema(apiKeyResource.ProtoIdentitySchema(ctx)())
}

type APIKeyListResourceModel struct {
	AccessKeys     types.List   `tfsdk:"access_keys"`
	OrganizationID types.String `tfsdk:"organization_id"`
	Description    types.String `tfsdk:"description"`
	BearerID       types.String `tfsdk:"bearer_id"`
	BearerType     types.String `tfsdk:"bearer_type"`
	Editable       types.Bool   `tfsdk:"editable"`
	Expired        types.Bool   `tfsdk:"expired"`
}

func (r *APIKeyListResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_iam_api_key"
}

func (r *APIKeyListResource) FetchAPIKeys(ctx context.Context, data APIKeyListResourceModel) ([]*iamSDK.APIKey, error) {
	request := &iamSDK.ListAPIKeysRequest{
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

	if !data.Editable.IsNull() && !data.Editable.IsUnknown() {
		editable := data.Editable.ValueBool()
		request.Editable = &editable
	}

	if !data.Expired.IsNull() && !data.Expired.IsUnknown() {
		expired := data.Expired.ValueBool()
		request.Expired = &expired
	}

	if !data.Description.IsNull() && !data.Description.IsUnknown() {
		description := data.Description.ValueString()
		request.Description = &description
	}

	if !data.BearerID.IsNull() && !data.BearerID.IsUnknown() {
		bearerID := data.BearerID.ValueString()
		request.BearerID = &bearerID
	}

	if !data.BearerType.IsNull() && !data.BearerType.IsUnknown() {
		bearerType := data.BearerType.ValueString()
		if bearerType != "" {
			request.BearerType = iamSDK.BearerType(bearerType)
		}
	}

	if !data.AccessKeys.IsNull() && !data.AccessKeys.IsUnknown() {
		var accessKeys []string
		data.AccessKeys.ElementsAs(ctx, &accessKeys, false)

		if len(accessKeys) > 0 {
			request.AccessKeys = accessKeys
		}
	}

	response, err := r.iamAPI.ListAPIKeys(request, scw.WithContext(ctx), scw.WithAllPages())
	if err != nil {
		return nil, err
	}

	return response.APIKeys, nil
}

func (r *APIKeyListResource) List(ctx context.Context, req list.ListRequest, stream *list.ListResultsStream) {
	var data APIKeyListResourceModel

	diags := req.Config.Get(ctx, &data)
	if diags.HasError() {
		stream.Results = list.ListResultsStreamDiagnostics(diags)

		return
	}

	allAPIKeys, err := r.FetchAPIKeys(ctx, data)
	if err != nil {
		stream.Results = list.ListResultsStreamDiagnostics(diag.Diagnostics{
			diag.NewErrorDiagnostic("Listing IAM API Keys", "Failed to list IAM API Keys: "+err.Error()),
		})

		return
	}

	stream.Results = func(push func(list.ListResult) bool) {
		for _, apiKey := range allAPIKeys {
			result := req.NewListResult(ctx)
			result.DisplayName = apiKey.AccessKey

			apiKeyResource := ResourceAPIKey()
			resourceData := apiKeyResource.Data(&terraform.InstanceState{})

			err = identity.SetGlobalIdentity(resourceData, apiKey.AccessKey)
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

			setAPIKeyState(resourceData, apiKey)

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
