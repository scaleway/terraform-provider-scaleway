package keymanager

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/list"
	"github.com/hashicorp/terraform-plugin-framework/list/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-mux/tf5to6server/translate"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	key_manager "github.com/scaleway/scaleway-sdk-go/api/key_manager/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/identity"
	listscw "github.com/scaleway/terraform-provider-scaleway/v2/internal/list"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
)

var (
	_ list.ListResource                 = (*KeyListResource)(nil)
	_ list.ListResourceWithConfigure    = (*KeyListResource)(nil)
	_ list.ListResourceWithRawV6Schemas = (*KeyListResource)(nil)
)

type KeyListResource struct {
	meta   *meta.Meta
	keyAPI *key_manager.API
}

func (r *KeyListResource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	m := listscw.ConfigureMeta(request, response)
	if m == nil {
		return
	}

	r.meta = m
	r.keyAPI = key_manager.NewAPI(meta.ExtractScwClient(m))
}

func NewKeyListResource() list.ListResource {
	return &KeyListResource{}
}

func (r *KeyListResource) ListResourceConfigSchema(_ context.Context, _ list.ListResourceSchemaRequest, response *list.ListResourceSchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"regions":     listscw.RegionsAttribute("Regions to filter for"),
			"project_ids": listscw.ProjectIDsAttribute("Project IDs to filter for"),
			"tags":        listscw.TagsAttribute("Tags to filter for"),
			"name":        listscw.NameAttribute("Name of the key to filter for"),
			"usage": schema.StringAttribute{
				Description: "Usage of the key to filter for (symmetric_encryption, asymmetric_encryption, asymmetric_signing)",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.OneOf("symmetric_encryption", "asymmetric_encryption", "asymmetric_signing"),
				},
			},
			"scheduled_for_deletion": schema.BoolAttribute{
				Description: "Filter keys by deletion status",
				Optional:    true,
			},
		},
	}
}

func (r *KeyListResource) RawV6Schemas(ctx context.Context, req list.RawV6SchemaRequest, resp *list.RawV6SchemaResponse) {
	keyResource := ResourceKeyManagerKey()

	resp.ProtoV6Schema = translate.Schema(keyResource.ProtoSchema(ctx)())
	resp.ProtoV6IdentitySchema = translate.ResourceIdentitySchema(keyResource.ProtoIdentitySchema(ctx)())
}

type KeyListResourceModel struct {
	Regions              types.List   `tfsdk:"regions"`
	ProjectIDs           types.List   `tfsdk:"project_ids"`
	Tags                 types.List   `tfsdk:"tags"`
	Name                 types.String `tfsdk:"name"`
	Usage                types.String `tfsdk:"usage"`
	ScheduledForDeletion types.Bool   `tfsdk:"scheduled_for_deletion"`
}

func (m *KeyListResourceModel) GetRegions() types.List  { return m.Regions }
func (m *KeyListResourceModel) GetProjects() types.List { return m.ProjectIDs }

func (r *KeyListResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_key_manager_key"
}

func (r *KeyListResource) FetchKeys(ctx context.Context, region scw.Region, projectID string, data KeyListResourceModel) ([]*key_manager.Key, error) {
	request := &key_manager.ListKeysRequest{
		Region:    region,
		ProjectID: &projectID,
	}

	if !data.Name.IsNull() && !data.Name.IsUnknown() {
		request.Name = data.Name.ValueStringPointer()
	}

	if !data.Tags.IsNull() && !data.Tags.IsUnknown() {
		var tags []string
		data.Tags.ElementsAs(ctx, &tags, false)
		request.Tags = tags
	}

	if !data.Usage.IsNull() && !data.Usage.IsUnknown() {
		usage := data.Usage.ValueString()
		request.Usage = key_manager.ListKeysRequestUsage(usage)
	}

	if !data.ScheduledForDeletion.IsNull() && !data.ScheduledForDeletion.IsUnknown() {
		request.ScheduledForDeletion = data.ScheduledForDeletion.ValueBool()
	}

	response, err := r.keyAPI.ListKeys(request, scw.WithContext(ctx), scw.WithAllPages())
	if err != nil {
		return nil, err
	}

	return response.Keys, nil
}

func (r *KeyListResource) List(ctx context.Context, req list.ListRequest, stream *list.ListResultsStream) {
	var data KeyListResourceModel

	diags := req.Config.Get(ctx, &data)
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

	targets := listscw.RegionalProjectTargets(regions, projects)

	allKeys, err := listscw.FetchConcurrently(ctx, targets,
		func(ctx context.Context, target listscw.RegionalFetchTarget) ([]*key_manager.Key, error) {
			return r.FetchKeys(ctx, target.Region, target.ProjectID, data)
		},
		func(a, b *key_manager.Key) int {
			return listscw.CompareRegionalProjectItems(a.ProjectID, b.ProjectID, a.Region, b.Region, a.ID, b.ID)
		},
	)
	if err != nil {
		stream.Results = list.ListResultsStreamDiagnostics(diag.Diagnostics{
			diag.NewErrorDiagnostic("Listing Key Manager Keys", "Failed to list Key Manager Keys: "+err.Error()),
		})

		return
	}

	stream.Results = func(push func(list.ListResult) bool) {
		for _, key := range allKeys {
			result := req.NewListResult(ctx)
			result.DisplayName = key.Name

			keyResource := ResourceKeyManagerKey()
			resourceData := keyResource.Data(&terraform.InstanceState{})

			err = identity.SetRegionalIdentity(resourceData, key.Region, key.ID)
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

			setKeyState(resourceData, key)

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
