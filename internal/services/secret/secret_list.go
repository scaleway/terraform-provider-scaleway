package secret

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/list"
	"github.com/hashicorp/terraform-plugin-framework/list/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-mux/tf5to6server/translate"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	secret "github.com/scaleway/scaleway-sdk-go/api/secret/v1beta1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/identity"
	listscw "github.com/scaleway/terraform-provider-scaleway/v2/internal/list"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
)

var (
	_ list.ListResource                 = (*SecretListResource)(nil)
	_ list.ListResourceWithConfigure    = (*SecretListResource)(nil)
	_ list.ListResourceWithRawV6Schemas = (*SecretListResource)(nil)
)

type SecretListResource struct {
	meta      *meta.Meta
	secretAPI *secret.API
}

func (r *SecretListResource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	m := listscw.ConfigureMeta(request, response)
	if m == nil {
		return
	}

	r.meta = m
	r.secretAPI = secret.NewAPI(meta.ExtractScwClient(m))
}

func NewSecretListResource() list.ListResource {
	return &SecretListResource{}
}

func (r *SecretListResource) ListResourceConfigSchema(_ context.Context, _ list.ListResourceSchemaRequest, response *list.ListResourceSchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"regions":         listscw.RegionsAttribute("Regions to target. Use '*' to list from all regions"),
			"project_ids":     listscw.ProjectIDsAttribute("Project IDs to filter for. Use '*' to list across all projects"),
			"organization_id": listscw.OrganizationIDAttribute("Organization ID to filter for"),
			"name":            listscw.NameAttribute("Filter by secret name"),
			"path": schema.StringAttribute{
				Description: "Filter by exact path",
				Optional:    true,
			},
			"tags": listscw.TagsAttribute("Filter by tags"),
			"type": schema.StringAttribute{
				Description: "Filter by secret type",
				Optional:    true,
			},
			"ephemeral": schema.BoolAttribute{
				Description: "Filter by ephemeral / not ephemeral",
				Optional:    true,
			},
			"scheduled_for_deletion": schema.BoolAttribute{
				Description: "Filter by whether the secret was scheduled for deletion / not scheduled for deletion",
				Optional:    true,
			},
		},
	}
}

func (r *SecretListResource) RawV6Schemas(ctx context.Context, req list.RawV6SchemaRequest, resp *list.RawV6SchemaResponse) {
	secretResource := ResourceSecret()

	resp.ProtoV6Schema = translate.Schema(secretResource.ProtoSchema(ctx)())
	resp.ProtoV6IdentitySchema = translate.ResourceIdentitySchema(secretResource.ProtoIdentitySchema(ctx)())
}

type SecretListResourceModel struct {
	Regions              types.List   `tfsdk:"regions"`
	ProjectIDs           types.List   `tfsdk:"project_ids"`
	OrganizationID       types.String `tfsdk:"organization_id"`
	Name                 types.String `tfsdk:"name"`
	Path                 types.String `tfsdk:"path"`
	Tags                 types.List   `tfsdk:"tags"`
	Type                 types.String `tfsdk:"type"`
	Ephemeral            types.Bool   `tfsdk:"ephemeral"`
	ScheduledForDeletion types.Bool   `tfsdk:"scheduled_for_deletion"`
}

func (m *SecretListResourceModel) GetRegions() types.List {
	return m.Regions
}

func (m *SecretListResourceModel) GetProjects() types.List {
	return m.ProjectIDs
}

func (m *SecretListResourceModel) GetTags() types.List {
	return m.Tags
}

func (r *SecretListResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_secret"
}

func (r *SecretListResource) FetchSecrets(ctx context.Context, region scw.Region, projectID string, tags []string, data SecretListResourceModel) ([]*secret.Secret, error) {
	request := &secret.ListSecretsRequest{
		Region:    region,
		ProjectID: &projectID,
	}

	if !data.OrganizationID.IsNull() && !data.OrganizationID.IsUnknown() {
		organizationID := data.OrganizationID.ValueString()
		request.OrganizationID = &organizationID
	}

	if !data.Name.IsNull() && !data.Name.IsUnknown() {
		name := data.Name.ValueString()
		request.Name = &name
	}

	if !data.Path.IsNull() && !data.Path.IsUnknown() {
		path := data.Path.ValueString()
		request.Path = &path
	}

	if len(tags) > 0 {
		request.Tags = tags
	}

	if !data.Type.IsNull() && !data.Type.IsUnknown() {
		secretType := data.Type.ValueString()
		request.Type = secret.SecretType(secretType)
	}

	if !data.Ephemeral.IsNull() && !data.Ephemeral.IsUnknown() {
		ephemeral := data.Ephemeral.ValueBool()
		request.Ephemeral = &ephemeral
	}

	if !data.ScheduledForDeletion.IsNull() && !data.ScheduledForDeletion.IsUnknown() {
		request.ScheduledForDeletion = data.ScheduledForDeletion.ValueBool()
	}

	response, err := r.secretAPI.ListSecrets(request, scw.WithContext(ctx), scw.WithAllPages())
	if err != nil {
		return nil, err
	}

	return response.Secrets, nil
}

func (r *SecretListResource) List(ctx context.Context, req list.ListRequest, stream *list.ListResultsStream) {
	var data SecretListResourceModel

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

	allSecrets, err := listscw.FetchConcurrently(ctx, listscw.RegionalProjectTargets(regions, projects),
		func(ctx context.Context, target listscw.RegionalFetchTarget) ([]*secret.Secret, error) {
			return r.FetchSecrets(ctx, target.Region, target.ProjectID, tags, data)
		},
		func(a, b *secret.Secret) int {
			return listscw.CompareRegionalProjectItems(a.ProjectID, b.ProjectID, a.Region, b.Region, a.ID, b.ID)
		},
	)
	if err != nil {
		stream.Results = list.ListResultsStreamDiagnostics(diag.Diagnostics{
			diag.NewErrorDiagnostic("Listing Secrets", "Failed to list Secrets: "+err.Error()),
		})

		return
	}

	stream.Results = func(push func(list.ListResult) bool) {
		for _, secret := range allSecrets {
			result := req.NewListResult(ctx)
			result.DisplayName = secret.Name

			secretResource := ResourceSecret()
			resourceData := secretResource.Data(&terraform.InstanceState{})

			err := identity.SetRegionalIdentity(resourceData, secret.Region, secret.ID)
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

			setSecretState(resourceData, secret, nil)

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
