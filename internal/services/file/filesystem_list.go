package file

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/list"
	"github.com/hashicorp/terraform-plugin-framework/list/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-mux/tf5to6server/translate"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	file "github.com/scaleway/scaleway-sdk-go/api/file/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/identity"
	listscw "github.com/scaleway/terraform-provider-scaleway/v2/internal/list"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
)

var (
	_ list.ListResource                 = (*FileSystemListResource)(nil)
	_ list.ListResourceWithConfigure    = (*FileSystemListResource)(nil)
	_ list.ListResourceWithRawV6Schemas = (*FileSystemListResource)(nil)
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

func (r *FileSystemListResource) RawV6Schemas(ctx context.Context, req list.RawV6SchemaRequest, resp *list.RawV6SchemaResponse) {
	resourceVPC := ResourceFileSystem()

	resp.ProtoV6Schema = translate.Schema(resourceVPC.ProtoSchema(ctx)())
	resp.ProtoV6IdentitySchema = translate.ResourceIdentitySchema(resourceVPC.ProtoIdentitySchema(ctx)())
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

			fsResource := ResourceFileSystem()
			resourceData := fsResource.Data(&terraform.InstanceState{})

			err := identity.SetRegionalIdentity(resourceData, rawFS.Region, rawFS.ID)
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

			// Convert and set the identity and resource state into the result
			tfTypeIdentity, errIdentityState := resourceData.TfTypeIdentityState()
			if errIdentityState != nil {
				result.Diagnostics.AddError(
					"Converting identity data",
					"An error was encountered when converting the identity data: "+errIdentityState.Error(),
				)
			}

			identitySetDiags := result.Identity.Set(ctx, *tfTypeIdentity)
			result.Diagnostics.Append(identitySetDiags...)

			setFileSystemState(resourceData, rawFS)

			// Convert and set the resource state into the result
			tfTypeResource, errTfTypeResourceState := resourceData.TfTypeResourceState()
			if errTfTypeResourceState != nil {
				result.Diagnostics.AddError(
					"Converting resource state",
					"An error was encountered when converting the resource state: "+errTfTypeResourceState.Error(),
				)
			}

			resourceSetDiags := result.Resource.Set(ctx, *tfTypeResource)
			result.Diagnostics.Append(resourceSetDiags...)

			// Send the result to the stream.
			if !push(result) {
				return
			}
		}
	}
}
