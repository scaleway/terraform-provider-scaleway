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
	_ list.ListResource                 = (*ApplicationListResource)(nil)
	_ list.ListResourceWithConfigure    = (*ApplicationListResource)(nil)
	_ list.ListResourceWithRawV6Schemas = (*ApplicationListResource)(nil)
)

type ApplicationListResource struct {
	meta   *meta.Meta
	iamAPI *iamSDK.API
}

func (r *ApplicationListResource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	m := listscw.ConfigureMeta(request, response)
	if m == nil {
		return
	}

	r.meta = m
	r.iamAPI = iamSDK.NewAPI(meta.ExtractScwClient(m))
}

func NewApplicationListResource() list.ListResource {
	return &ApplicationListResource{}
}

func (r *ApplicationListResource) ListResourceConfigSchema(_ context.Context, _ list.ListResourceSchemaRequest, response *list.ListResourceSchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"organization_id": listscw.OrganizationIDAttribute("Organization ID to filter for"),
			"name":            listscw.NameAttribute("Name of the application to filter for"),
			"tag":             listscw.TagAttribute("Filter by tags containing a given string"),
			"editable": schema.BoolAttribute{
				Description: "Filter by editable status",
				Optional:    true,
			},
			"application_ids": schema.ListAttribute{
				Description: "Filter applications by application IDs",
				ElementType: types.StringType,
				Optional:    true,
			},
		},
	}
}

func (r *ApplicationListResource) RawV6Schemas(ctx context.Context, req list.RawV6SchemaRequest, resp *list.RawV6SchemaResponse) {
	applicationResource := ResourceApplication()

	resp.ProtoV6Schema = translate.Schema(applicationResource.ProtoSchema(ctx)())
	resp.ProtoV6IdentitySchema = translate.ResourceIdentitySchema(applicationResource.ProtoIdentitySchema(ctx)())
}

type ApplicationListResourceModel struct {
	ApplicationIDs types.List   `tfsdk:"application_ids"`
	OrganizationID types.String `tfsdk:"organization_id"`
	Name           types.String `tfsdk:"name"`
	Tag            types.String `tfsdk:"tag"`
	Editable       types.Bool   `tfsdk:"editable"`
}

func (r *ApplicationListResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_iam_application"
}

func (r *ApplicationListResource) FetchApplications(ctx context.Context, data ApplicationListResourceModel) ([]*iamSDK.Application, error) {
	request := &iamSDK.ListApplicationsRequest{
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

	if !data.Name.IsNull() && !data.Name.IsUnknown() {
		name := data.Name.ValueString()
		request.Name = &name
	}

	if !data.Tag.IsNull() && !data.Tag.IsUnknown() {
		tag := data.Tag.ValueString()
		request.Tag = &tag
	}

	if !data.Editable.IsNull() && !data.Editable.IsUnknown() {
		editable := data.Editable.ValueBool()
		request.Editable = &editable
	}

	if !data.ApplicationIDs.IsNull() && !data.ApplicationIDs.IsUnknown() {
		var applicationIDs []string
		data.ApplicationIDs.ElementsAs(ctx, &applicationIDs, false)

		if len(applicationIDs) > 0 {
			request.ApplicationIDs = applicationIDs
		}
	}

	response, err := r.iamAPI.ListApplications(request, scw.WithContext(ctx), scw.WithAllPages())
	if err != nil {
		return nil, err
	}

	return response.Applications, nil
}

func (r *ApplicationListResource) List(ctx context.Context, req list.ListRequest, stream *list.ListResultsStream) {
	var data ApplicationListResourceModel

	diags := req.Config.Get(ctx, &data)
	if diags.HasError() {
		stream.Results = list.ListResultsStreamDiagnostics(diags)

		return
	}

	allApplications, err := r.FetchApplications(ctx, data)
	if err != nil {
		stream.Results = list.ListResultsStreamDiagnostics(diag.Diagnostics{
			diag.NewErrorDiagnostic("Listing IAM Applications", "Failed to list IAM Applications: "+err.Error()),
		})

		return
	}

	stream.Results = func(push func(list.ListResult) bool) {
		for _, application := range allApplications {
			result := req.NewListResult(ctx)
			result.DisplayName = application.Name

			applicationResource := ResourceApplication()
			resourceData := applicationResource.Data(&terraform.InstanceState{})

			err = identity.SetGlobalIdentity(resourceData, application.ID)
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

			setApplicationState(resourceData, application)

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
