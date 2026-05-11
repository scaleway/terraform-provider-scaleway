package opensearch

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/list"
	"github.com/hashicorp/terraform-plugin-framework/list/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-mux/tf5to6server/translate"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	searchdbapi "github.com/scaleway/scaleway-sdk-go/api/searchdb/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/identity"
	listscw "github.com/scaleway/terraform-provider-scaleway/v2/internal/list"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
)

var (
	_ list.ListResource                 = (*DeploymentListResource)(nil)
	_ list.ListResourceWithConfigure    = (*DeploymentListResource)(nil)
	_ list.ListResourceWithRawV6Schemas = (*DeploymentListResource)(nil)
)

type DeploymentListResource struct {
	meta        *meta.Meta
	searchdbAPI *searchdbapi.API
}

func (r *DeploymentListResource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	if request.ProviderData == nil {
		return
	}

	m, ok := request.ProviderData.(*meta.Meta)
	if !ok {
		response.Diagnostics.AddError(
			"Unexpected List Configure Type",
			fmt.Sprintf("Expected *meta.Meta, got: %T. Please report this issue to the provider developers.", request.ProviderData),
		)

		return
	}

	r.meta = m
	r.searchdbAPI = searchdbapi.NewAPI(meta.ExtractScwClient(m))
}

func NewDeploymentListResource() list.ListResource {
	return &DeploymentListResource{}
}

func (r *DeploymentListResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_opensearch_deployment"
}

func (r *DeploymentListResource) ListResourceConfigSchema(_ context.Context, _ list.ListResourceSchemaRequest, response *list.ListResourceSchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"name":            listscw.NameAttribute("Name of the OpenSearch deployment to filter on"),
			"tags":            listscw.TagsAttribute("Tags of the OpenSearch deployment to filter on"),
			"organization_id": listscw.OrganizationIDAttribute("Organization ID of the OpenSearch deployment to filter on"),
			"project_ids":     listscw.ProjectIDsAttribute("Project IDs of the OpenSearch deployment to filter on"),
			"regions":         listscw.RegionsAttribute("Regions of the OpenSearch deployment to filter on"),
			"version":         deploymentListVersionAttribute(),
		},
	}
}

func deploymentListVersionAttribute() schema.StringAttribute {
	return schema.StringAttribute{
		Description: "OpenSearch engine version to filter on (same value as the deployment `version` attribute, e.g. \"2.15\")",
		Optional:    true,
	}
}

func (r *DeploymentListResource) RawV6Schemas(ctx context.Context, _ list.RawV6SchemaRequest, resp *list.RawV6SchemaResponse) {
	deploymentResource := ResourceDeployment()

	resp.ProtoV6Schema = translate.Schema(deploymentResource.ProtoSchema(ctx)())
	resp.ProtoV6IdentitySchema = translate.ResourceIdentitySchema(deploymentResource.ProtoIdentitySchema(ctx)())
}

type DeploymentListResourceModel struct {
	Tags           types.List   `tfsdk:"tags"`
	Name           types.String `tfsdk:"name"`
	Version        types.String `tfsdk:"version"`
	OrganizationID types.String `tfsdk:"organization_id"`
	ProjectIDs     types.List   `tfsdk:"project_ids"`
	Regions        types.List   `tfsdk:"regions"`
}

func (m *DeploymentListResourceModel) GetTags() types.List {
	return m.Tags
}

func (m *DeploymentListResourceModel) GetRegions() types.List {
	return m.Regions
}

func (m *DeploymentListResourceModel) GetProjects() types.List {
	return m.ProjectIDs
}

func (r *DeploymentListResource) List(ctx context.Context, req list.ListRequest, stream *list.ListResultsStream) {
	var data DeploymentListResourceModel

	diags := req.Config.Get(ctx, &data)
	if diags.HasError() {
		stream.Results = list.ListResultsStreamDiagnostics(diags)

		return
	}

	tags, diags := listscw.ExtractTags(ctx, &data)
	if diags.HasError() {
		stream.Results = list.ListResultsStreamDiagnostics(diags)

		return
	}

	regions, err := listscw.ExtractRegions(ctx, &data, r.meta)
	if err != nil {
		stream.Results = list.ListResultsStreamDiagnostics(diag.Diagnostics{
			diag.NewErrorDiagnostic("Listing regions", "An error was encountered when listing regions: "+err.Error()),
		})

		return
	}

	projects, err := listscw.ExtractProjects(ctx, &data, r.meta)
	if err != nil {
		stream.Results = list.ListResultsStreamDiagnostics(diag.Diagnostics{
			diag.NewErrorDiagnostic("Listing projects", "An error was encountered when listing projects: "+err.Error()),
		})

		return
	}

	var targets []listscw.RegionalFetchTarget

	for _, region := range regions {
		for _, project := range projects {
			targets = append(targets, listscw.RegionalFetchTarget{Region: region, ProjectID: project})
		}
	}

	allDeployments, err := listscw.FetchConcurrently(ctx, targets,
		func(ctx context.Context, target listscw.RegionalFetchTarget) ([]*searchdbapi.Deployment, error) {
			return r.fetchDeployments(ctx, target, tags, data)
		},
		func(a, b *searchdbapi.Deployment) int {
			if a.ProjectID != b.ProjectID {
				return strings.Compare(a.ProjectID, b.ProjectID)
			}

			if a.Region != b.Region {
				return strings.Compare(string(a.Region), string(b.Region))
			}

			return strings.Compare(a.ID, b.ID)
		},
	)
	if err != nil {
		stream.Results = list.ListResultsStreamDiagnostics(diag.Diagnostics{
			diag.NewErrorDiagnostic("Listing OpenSearch deployments", "Failed to list OpenSearch deployments: "+err.Error()),
		})

		return
	}

	stream.Results = func(push func(list.ListResult) bool) {
		for _, deployment := range allDeployments {
			result := req.NewListResult(ctx)
			result.DisplayName = deployment.Name

			deploymentResource := ResourceDeployment()
			resourceData := deploymentResource.Data(&terraform.InstanceState{})

			err := identity.SetRegionalIdentity(resourceData, deployment.Region, deployment.ID)
			if err != nil {
				result.Diagnostics.AddError(
					"Retrieving identity data",
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

			diagsState := setDeploymentState(resourceData, deployment)
			if diagsState.HasError() {
				tflog.Error(ctx, "error from setting setDeploymentState")

				if !push(result) {
					return
				}

				continue
			}

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

func (r *DeploymentListResource) fetchDeployments(ctx context.Context, target listscw.RegionalFetchTarget, tags []string, data DeploymentListResourceModel) ([]*searchdbapi.Deployment, error) {
	versionFilter := deploymentListVersionFilter(data)

	req := &searchdbapi.ListDeploymentsRequest{
		Region:         target.Region,
		Name:           data.Name.ValueStringPointer(),
		Tags:           tags,
		OrganizationID: data.OrganizationID.ValueStringPointer(),
		ProjectID:      &target.ProjectID,
		OrderBy:        searchdbapi.ListDeploymentsRequestOrderByCreatedAtAsc,
		// Do not set Version on the API request: the SDK validates it with a pattern that
		// rejects typical OpenSearch version strings (they contain dots). Filter on
		// Deployment.Version after listing instead.
	}

	response, err := r.searchdbAPI.ListDeployments(req, scw.WithContext(ctx), scw.WithAllPages())
	if err != nil {
		return nil, err
	}

	// Defensive filter: keep only deployments for the requested project/region target,
	// and apply optional version filter (client-side; see Version comment above).
	filtered := make([]*searchdbapi.Deployment, 0, len(response.Deployments))
	for _, dep := range response.Deployments {
		if dep == nil {
			continue
		}

		if dep.ProjectID != target.ProjectID || dep.Region != target.Region {
			continue
		}

		if versionFilter != nil && dep.Version != *versionFilter {
			continue
		}

		filtered = append(filtered, dep)
	}

	return filtered, nil
}

func deploymentListVersionFilter(data DeploymentListResourceModel) *string {
	if data.Version.IsNull() || data.Version.IsUnknown() {
		return nil
	}

	v := strings.TrimSpace(data.Version.ValueString())
	if v == "" {
		return nil
	}

	return &v
}
