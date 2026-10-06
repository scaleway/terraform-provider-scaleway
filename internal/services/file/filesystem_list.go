package file

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/list"
	"github.com/hashicorp/terraform-plugin-framework/list/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	file "github.com/scaleway/scaleway-sdk-go/api/file/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	identityfw "github.com/scaleway/terraform-provider-scaleway/v2/internal/identity/framework"
	listscw "github.com/scaleway/terraform-provider-scaleway/v2/internal/list"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
)

var (
	_ list.ListResource              = (*FileSystemListResource)(nil)
	_ list.ListResourceWithConfigure = (*FileSystemListResource)(nil)
)

type FileSystemListResource struct {
	meta    *meta.Meta
	fileAPI *file.API
}

func (r *FileSystemListResource) Configure(
	_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse,
) {
	m := listscw.ConfigureMeta(request, response)
	if m == nil {
		return
	}

	r.meta = m
	r.fileAPI = file.NewAPI(meta.ExtractScwClient(m))
}

func (r *FileSystemListResource) Metadata(
	ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_file_filesystem"
}

func NewFileSystemListResource() list.ListResource {
	return &FileSystemListResource{}
}

func (r *FileSystemListResource) ListResourceConfigSchema(
	_ context.Context,
	_ list.ListResourceSchemaRequest,
	response *list.ListResourceSchemaResponse,
) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"regions":         listscw.RegionsAttribute("Regions to filter on. For now, only \"fr-par\" is supported."),
			"project_ids":     listscw.ProjectIDsAttribute("Project IDs to filter on."),
			"organization_id": listscw.OrganizationIDAttribute("Organization ID to filter on."),
			"name":            listscw.NameAttribute("Name to filter on."),
			"tags":            listscw.TagsAttribute("Tags to filter on."),
			"filesystem_ids":  listscw.UUIDsAttribute("FileSystem IDs to filter on."),
		},
	}
}

type ListResourceModel struct {
	Tags           types.List   `tfsdk:"tags"`
	Name           types.String `tfsdk:"name"`
	OrganizationID types.String `tfsdk:"organization_id"`
	ProjectIDs     types.List   `tfsdk:"project_ids"`
	Regions        types.List   `tfsdk:"regions"`
	FileSystemIDs  types.List   `tfsdk:"filesystem_ids"`
}

func (m *ListResourceModel) GetTags() types.List {
	return m.Tags
}

func (m *ListResourceModel) GetRegions() types.List {
	return m.Regions
}

func (m *ListResourceModel) GetProjects() types.List {
	return m.ProjectIDs
}

func (m *ListResourceModel) GetFileSystemIDs() types.List {
	return m.FileSystemIDs
}

func (r *FileSystemListResource) FetchFileSystems(
	ctx context.Context,
	region scw.Region,
	project *string,
	tags []string,
	fileSystemIDs []string,
	data ListResourceModel,
) ([]*file.FileSystem, error) {
	listRequest := &file.ListFileSystemsRequest{
		Region:         region,
		Name:           data.Name.ValueStringPointer(),
		Tags:           tags,
		OrganizationID: data.OrganizationID.ValueStringPointer(),
		ProjectID:      project,
		FilesystemIDs:  fileSystemIDs,
	}

	response, err := r.fileAPI.ListFileSystems(
		listRequest, scw.WithContext(ctx), scw.WithAllPages(),
	)
	if err != nil {
		return nil, err
	}

	return response.Filesystems, nil
}

func (r *FileSystemListResource) List(ctx context.Context, req list.ListRequest, stream *list.ListResultsStream) {
	var data ListResourceModel

	// Read list config data into the model
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

	fileSystemIDs, diags := listscw.ExtractFileSystemIDs(ctx, &data)
	if diags.HasError() {
		stream.Results = list.ListResultsStreamDiagnostics(diags)

		return
	}

	allFileSystems, err := listscw.FetchConcurrently(ctx, listscw.RegionalProjectTargets(regions, projects),
		func(ctx context.Context, target listscw.RegionalFetchTarget) ([]*file.FileSystem, error) {
			return r.FetchFileSystems(ctx, target.Region, &target.ProjectID, tags, fileSystemIDs, data)
		},
		func(a, b *file.FileSystem) int {
			return listscw.CompareRegionalProjectItems(
				a.ProjectID, b.ProjectID, a.Region, b.Region, a.ID, b.ID,
			)
		},
	)
	if err != nil {
		stream.Results = list.ListResultsStreamDiagnostics(diag.Diagnostics{
			diag.NewErrorDiagnostic("Listing FileSystems", "Failed to list FileSystems: "+err.Error()),
		})

		return
	}

	stream.Results = func(push func(list.ListResult) bool) {
		for _, rawFS := range allFileSystems {
			result := req.NewListResult(ctx)
			result.DisplayName = rawFS.Name

			identityDiags := result.Identity.Set(ctx, identityfw.SetRegionalIdentity(
				rawFS.Region,
				rawFS.ID,
			))
			result.Diagnostics.Append(identityDiags...)

			if req.IncludeResource {
				resourceModel := flattenFilesystem(ctx, rawFS, &data, &result.Diagnostics)
				resourceDiags := result.Resource.Set(ctx, &resourceModel)
				result.Diagnostics.Append(resourceDiags...)
			}

			// Send the result to the stream.
			if !push(result) {
				return
			}
		}
	}
}
