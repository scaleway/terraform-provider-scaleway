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
	_ list.ListResource                 = (*PolicyListResource)(nil)
	_ list.ListResourceWithConfigure    = (*PolicyListResource)(nil)
	_ list.ListResourceWithRawV6Schemas = (*PolicyListResource)(nil)
)

type PolicyListResource struct {
	meta   *meta.Meta
	iamAPI *iamSDK.API
}

func (r *PolicyListResource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	m := listscw.ConfigureMeta(request, response)
	if m == nil {
		return
	}

	r.meta = m
	r.iamAPI = iamSDK.NewAPI(meta.ExtractScwClient(m))
}

func NewPolicyListResource() list.ListResource {
	return &PolicyListResource{}
}

func (r *PolicyListResource) ListResourceConfigSchema(_ context.Context, _ list.ListResourceSchemaRequest, response *list.ListResourceSchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"organization_id": listscw.OrganizationIDAttribute("Organization ID to filter for"),
			"tag":             listscw.TagAttribute("Filter by tags containing a given string"),
			"editable": schema.BoolAttribute{
				Description: "Filter by editable status",
				Optional:    true,
			},
			"policy_ids": schema.ListAttribute{
				Description: "Filter policies by policy IDs",
				ElementType: types.StringType,
				Optional:    true,
			},
			"user_ids": schema.ListAttribute{
				Description: "Filter policies by user IDs",
				ElementType: types.StringType,
				Optional:    true,
			},
			"group_ids": schema.ListAttribute{
				Description: "Filter policies by group IDs",
				ElementType: types.StringType,
				Optional:    true,
			},
			"application_ids": schema.ListAttribute{
				Description: "Filter policies by application IDs",
				ElementType: types.StringType,
				Optional:    true,
			},
			"no_principal": schema.BoolAttribute{
				Description: "Filter by policies with no principal",
				Optional:    true,
			},
			"policy_name": schema.StringAttribute{
				Description: "Filter by policy name",
				Optional:    true,
			},
		},
	}
}

func (r *PolicyListResource) RawV6Schemas(ctx context.Context, req list.RawV6SchemaRequest, resp *list.RawV6SchemaResponse) {
	policyResource := ResourcePolicy()

	resp.ProtoV6Schema = translate.Schema(policyResource.ProtoSchema(ctx)())
	resp.ProtoV6IdentitySchema = translate.ResourceIdentitySchema(policyResource.ProtoIdentitySchema(ctx)())
}

type PolicyListResourceModel struct {
	PolicyIDs      types.List   `tfsdk:"policy_ids"`
	UserIDs        types.List   `tfsdk:"user_ids"`
	GroupIDs       types.List   `tfsdk:"group_ids"`
	ApplicationIDs types.List   `tfsdk:"application_ids"`
	OrganizationID types.String `tfsdk:"organization_id"`
	Tag            types.String `tfsdk:"tag"`
	PolicyName     types.String `tfsdk:"policy_name"`
	Editable       types.Bool   `tfsdk:"editable"`
	NoPrincipal    types.Bool   `tfsdk:"no_principal"`
}

func (r *PolicyListResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_iam_policy"
}

func (r *PolicyListResource) FetchPolicies(ctx context.Context, data PolicyListResourceModel) ([]*iamSDK.Policy, error) {
	request := &iamSDK.ListPoliciesRequest{
		OrganizationID: data.OrganizationID.ValueString(),
	}

	if request.OrganizationID == "" {
		defaultOrgID, exists := r.meta.ScwClient().GetDefaultOrganizationID()
		if exists {
			request.OrganizationID = defaultOrgID
		} else {
			return nil, errors.New("organization ID is required. Either set organization_id or configure a default organization")
		}
	}

	if !data.Tag.IsNull() && !data.Tag.IsUnknown() {
		tag := data.Tag.ValueString()
		request.Tag = &tag
	}

	if !data.Editable.IsNull() && !data.Editable.IsUnknown() {
		editable := data.Editable.ValueBool()
		request.Editable = &editable
	}

	if !data.PolicyName.IsNull() && !data.PolicyName.IsUnknown() {
		policyName := data.PolicyName.ValueString()
		request.PolicyName = &policyName
	}

	if !data.NoPrincipal.IsNull() && !data.NoPrincipal.IsUnknown() {
		noPrincipal := data.NoPrincipal.ValueBool()
		request.NoPrincipal = &noPrincipal
	}

	if !data.PolicyIDs.IsNull() && !data.PolicyIDs.IsUnknown() {
		var policyIDs []string
		data.PolicyIDs.ElementsAs(ctx, &policyIDs, false)

		if len(policyIDs) > 0 {
			request.PolicyIDs = policyIDs
		}
	}

	if !data.UserIDs.IsNull() && !data.UserIDs.IsUnknown() {
		var userIDs []string
		data.UserIDs.ElementsAs(ctx, &userIDs, false)

		if len(userIDs) > 0 {
			request.UserIDs = userIDs
		}
	}

	if !data.GroupIDs.IsNull() && !data.GroupIDs.IsUnknown() {
		var groupIDs []string
		data.GroupIDs.ElementsAs(ctx, &groupIDs, false)

		if len(groupIDs) > 0 {
			request.GroupIDs = groupIDs
		}
	}

	if !data.ApplicationIDs.IsNull() && !data.ApplicationIDs.IsUnknown() {
		var applicationIDs []string
		data.ApplicationIDs.ElementsAs(ctx, &applicationIDs, false)

		if len(applicationIDs) > 0 {
			request.ApplicationIDs = applicationIDs
		}
	}

	response, err := r.iamAPI.ListPolicies(request, scw.WithContext(ctx), scw.WithAllPages())
	if err != nil {
		return nil, err
	}

	return response.Policies, nil
}

func (r *PolicyListResource) List(ctx context.Context, req list.ListRequest, stream *list.ListResultsStream) {
	var data PolicyListResourceModel

	diags := req.Config.Get(ctx, &data)
	if diags.HasError() {
		stream.Results = list.ListResultsStreamDiagnostics(diags)

		return
	}

	allPolicies, err := r.FetchPolicies(ctx, data)
	if err != nil {
		stream.Results = list.ListResultsStreamDiagnostics(diag.Diagnostics{
			diag.NewErrorDiagnostic("Listing IAM Policies", "Failed to list IAM Policies: "+err.Error()),
		})

		return
	}

	stream.Results = func(push func(list.ListResult) bool) {
		for _, policy := range allPolicies {
			result := req.NewListResult(ctx)
			result.DisplayName = policy.Name

			policyResource := ResourcePolicy()
			resourceData := policyResource.Data(&terraform.InstanceState{})

			err = identity.SetGlobalIdentity(resourceData, policy.ID)
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

			setPolicyState(resourceData, policy, nil)

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
