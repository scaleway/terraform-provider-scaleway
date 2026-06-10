package account

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
	accountSDK "github.com/scaleway/scaleway-sdk-go/api/account/v3"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/identity"
	listscw "github.com/scaleway/terraform-provider-scaleway/v2/internal/list"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
)

var (
	_ list.ListResource                 = (*ProjectListResource)(nil)
	_ list.ListResourceWithConfigure    = (*ProjectListResource)(nil)
	_ list.ListResourceWithRawV6Schemas = (*ProjectListResource)(nil)
)

type ProjectListResource struct {
	meta       *meta.Meta
	accountAPI *accountSDK.ProjectAPI
}

func (r *ProjectListResource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	m := listscw.ConfigureMeta(request, response)
	if m == nil {
		return
	}

	r.meta = m
	r.accountAPI = accountSDK.NewProjectAPI(meta.ExtractScwClient(m))
}

func NewProjectListResource() list.ListResource {
	return &ProjectListResource{}
}

func (r *ProjectListResource) ListResourceConfigSchema(_ context.Context, _ list.ListResourceSchemaRequest, response *list.ListResourceSchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"organization_id": listscw.OrganizationIDAttribute("Organization ID to filter for"),
			"name":            listscw.NameAttribute("Filter by project name containing a given string"),
			"project_ids": schema.ListAttribute{
				Description: "Filter projects by project IDs",
				ElementType: types.StringType,
				Optional:    true,
			},
		},
	}
}

func (r *ProjectListResource) RawV6Schemas(ctx context.Context, req list.RawV6SchemaRequest, resp *list.RawV6SchemaResponse) {
	projectResource := ResourceProject()

	resp.ProtoV6Schema = translate.Schema(projectResource.ProtoSchema(ctx)())
	resp.ProtoV6IdentitySchema = translate.ResourceIdentitySchema(projectResource.ProtoIdentitySchema(ctx)())
}

type ProjectListResourceModel struct {
	ProjectIDs     types.List   `tfsdk:"project_ids"`
	OrganizationID types.String `tfsdk:"organization_id"`
	Name           types.String `tfsdk:"name"`
}

func (r *ProjectListResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_account_project"
}

func (r *ProjectListResource) FetchProjects(ctx context.Context, data ProjectListResourceModel) ([]*accountSDK.Project, error) {
	request := &accountSDK.ProjectAPIListProjectsRequest{}

	if data.OrganizationID.IsNull() || data.OrganizationID.IsUnknown() {
		defaultOrgID, exists := r.meta.ScwClient().GetDefaultOrganizationID()
		if exists {
			request.OrganizationID = defaultOrgID
		} else {
			return nil, errors.New("organization ID is required. Either set organization_id or configure a default organization")
		}
	} else {
		request.OrganizationID = data.OrganizationID.ValueString()
	}

	if !data.Name.IsNull() && !data.Name.IsUnknown() {
		name := data.Name.ValueString()
		request.Name = &name
	}

	if !data.ProjectIDs.IsNull() && !data.ProjectIDs.IsUnknown() {
		var projectIDs []string
		data.ProjectIDs.ElementsAs(ctx, &projectIDs, false)

		if len(projectIDs) > 0 {
			request.ProjectIDs = projectIDs
		}
	}

	response, err := r.accountAPI.ListProjects(request, scw.WithContext(ctx), scw.WithAllPages())
	if err != nil {
		return nil, err
	}

	return response.Projects, nil
}

func (r *ProjectListResource) List(ctx context.Context, req list.ListRequest, stream *list.ListResultsStream) {
	var data ProjectListResourceModel

	diags := req.Config.Get(ctx, &data)
	if diags.HasError() {
		stream.Results = list.ListResultsStreamDiagnostics(diags)

		return
	}

	allProjects, err := r.FetchProjects(ctx, data)
	if err != nil {
		stream.Results = list.ListResultsStreamDiagnostics(diag.Diagnostics{
			diag.NewErrorDiagnostic("Listing Account Projects", "Failed to list Account Projects: "+err.Error()),
		})

		return
	}

	stream.Results = func(push func(list.ListResult) bool) {
		for _, project := range allProjects {
			result := req.NewListResult(ctx)
			result.DisplayName = project.Name

			projectResource := ResourceProject()
			resourceData := projectResource.Data(&terraform.InstanceState{})

			err = identity.SetGlobalIdentity(resourceData, project.ID)
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

			setProjectState(resourceData, project)

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
