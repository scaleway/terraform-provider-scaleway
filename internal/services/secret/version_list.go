package secret

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strconv"

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
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
)

var (
	_ list.ListResource                 = (*VersionListResource)(nil)
	_ list.ListResourceWithConfigure    = (*VersionListResource)(nil)
	_ list.ListResourceWithRawV6Schemas = (*VersionListResource)(nil)
)

type VersionListResource struct {
	meta      *meta.Meta
	secretAPI *secret.API
}

func (r *VersionListResource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	m := listscw.ConfigureMeta(request, response)
	if m == nil {
		return
	}

	r.meta = m
	r.secretAPI = secret.NewAPI(meta.ExtractScwClient(m))
}

func NewVersionListResource() list.ListResource {
	return &VersionListResource{}
}

func (r *VersionListResource) ListResourceConfigSchema(_ context.Context, _ list.ListResourceSchemaRequest, response *list.ListResourceSchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"regions":         listscw.RegionsAttribute("Regions to target. Use '*' to list from all regions"),
			"project_ids":     listscw.ProjectIDsAttribute("Project IDs to filter for. Use '*' to list across all projects"),
			"organization_id": listscw.OrganizationIDAttribute("Organization ID to filter for"),
			"secret_ids": schema.ListAttribute{
				Description: "IDs of the secrets to list versions for. Use '*' to list versions from all secrets. If empty, returns an error.",
				ElementType: types.StringType,
				Required:    true,
			},
			"status": schema.ListAttribute{
				Description: "Filter by status",
				ElementType: types.StringType,
				Optional:    true,
			},
		},
	}
}

func (r *VersionListResource) RawV6Schemas(ctx context.Context, req list.RawV6SchemaRequest, resp *list.RawV6SchemaResponse) {
	versionResource := ResourceVersion()

	resp.ProtoV6Schema = translate.Schema(versionResource.ProtoSchema(ctx)())
	resp.ProtoV6IdentitySchema = translate.ResourceIdentitySchema(versionResource.ProtoIdentitySchema(ctx)())
}

type VersionListResourceModel struct {
	Regions        types.List   `tfsdk:"regions"`
	ProjectIDs     types.List   `tfsdk:"project_ids"`
	OrganizationID types.String `tfsdk:"organization_id"`
	SecretIDs      types.List   `tfsdk:"secret_ids"`
	Status         types.List   `tfsdk:"status"`
}

type versionListTarget struct {
	Region    scw.Region
	ProjectID string
	SecretID  string
}

func (m *VersionListResourceModel) GetRegions() types.List {
	return m.Regions
}

func (m *VersionListResourceModel) GetProjects() types.List {
	return m.ProjectIDs
}

func (r *VersionListResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_secret_version"
}

func (r *VersionListResource) List(ctx context.Context, req list.ListRequest, stream *list.ListResultsStream) {
	var data VersionListResourceModel

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

	targets, err := r.buildVersionListTargets(ctx, regions, projects, data)
	if err != nil {
		stream.Results = list.ListResultsStreamDiagnostics(diag.Diagnostics{
			diag.NewErrorDiagnostic("Building targets", "Failed to build version list targets: "+err.Error()),
		})

		return
	}

	allVersions, err := listscw.FetchConcurrently(ctx, targets,
		func(ctx context.Context, target versionListTarget) ([]*secret.SecretVersion, error) {
			return r.fetchVersionsForTarget(ctx, target, data)
		},
		func(a, b *secret.SecretVersion) int {
			return listscw.CompareRegionalProjectItems(a.SecretID, b.SecretID, a.Region, b.Region, strconv.FormatUint(uint64(a.Revision), 10), strconv.FormatUint(uint64(b.Revision), 10))
		},
	)
	if err != nil {
		stream.Results = list.ListResultsStreamDiagnostics(diag.Diagnostics{
			diag.NewErrorDiagnostic("Listing Secret Versions", "Failed to list Secret Versions: "+err.Error()),
		})

		return
	}

	stream.Results = func(push func(list.ListResult) bool) {
		for _, version := range allVersions {
			result := req.NewListResult(ctx)
			result.DisplayName = fmt.Sprintf("version-%d", version.Revision)

			versionResource := ResourceVersion()
			resourceData := versionResource.Data(&terraform.InstanceState{})

			err := identity.SetRegionalIdentity(resourceData, version.Region, fmt.Sprintf("%s/%d", version.SecretID, version.Revision))
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

			setVersionState(resourceData, version)

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

func (r *VersionListResource) parseSecretIDs(ctx context.Context, data VersionListResourceModel) ([]string, error) {
	var secretIDStrings []types.String

	if data.SecretIDs.IsNull() || data.SecretIDs.IsUnknown() {
		return nil, errors.New("secret_ids cannot be null or unknown")
	}

	diags := data.SecretIDs.ElementsAs(ctx, &secretIDStrings, false)
	if diags.HasError() {
		return nil, fmt.Errorf("failed to parse secret_ids: %v", diags)
	}

	if len(secretIDStrings) == 0 {
		return nil, errors.New("secret_ids list cannot be empty")
	}

	result := make([]string, len(secretIDStrings))
	for i, id := range secretIDStrings {
		result[i] = id.ValueString()
	}

	return result, nil
}

func (r *VersionListResource) parseStatuses(ctx context.Context, data VersionListResourceModel) ([]secret.SecretVersionStatus, error) {
	var statuses []secret.SecretVersionStatus

	if !data.Status.IsNull() && !data.Status.IsUnknown() {
		var statusStrings []types.String

		diags := data.Status.ElementsAs(ctx, &statusStrings, false)
		if diags.HasError() {
			return nil, fmt.Errorf("failed to parse status: %v", diags)
		}

		if len(statusStrings) > 0 {
			statuses = make([]secret.SecretVersionStatus, len(statusStrings))
			for i, s := range statusStrings {
				statuses[i] = secret.SecretVersionStatus(s.ValueString())
			}
		}
	}

	return statuses, nil
}

func (r *VersionListResource) buildVersionListTargets(ctx context.Context, regions []scw.Region, projects []string, data VersionListResourceModel) ([]versionListTarget, error) {
	secretIDStrings, err := r.parseSecretIDs(ctx, data)
	if err != nil {
		return nil, err
	}

	var targets []versionListTarget

	if slices.Contains(secretIDStrings, "*") && len(secretIDStrings) != 1 {
		return nil, errors.New("`When using \"*\", secret_ids must be exactly [\"*\"].`")
	}

	if len(secretIDStrings) == 1 && secretIDStrings[0] == "*" {
		for _, region := range regions {
			for _, projectID := range projects {
				secrets, err := r.listSecretsForProject(ctx, region, projectID)
				if err != nil {
					return nil, fmt.Errorf("failed to list secrets for project %s in region %s: %w", projectID, region, err)
				}

				for _, secret := range secrets {
					targets = append(targets, versionListTarget{
						Region:    region,
						ProjectID: projectID,
						SecretID:  secret.ID,
					})
				}
			}
		}
	} else {
		for _, region := range regions {
			for _, projectID := range projects {
				for _, secretID := range secretIDStrings {
					_, rawSecretID, err := regional.ParseID(secretID)
					if err != nil {
						return nil, fmt.Errorf("failed to parse secret ID %s: %w", secretID, err)
					}

					targets = append(targets, versionListTarget{
						Region:    region,
						ProjectID: projectID,
						SecretID:  rawSecretID,
					})
				}
			}
		}
	}

	return targets, nil
}

func (r *VersionListResource) listSecretsForProject(ctx context.Context, region scw.Region, projectID string) ([]*secret.Secret, error) {
	listSecretsRequest := &secret.ListSecretsRequest{
		Region: region,
	}

	if projectID != "" {
		listSecretsRequest.ProjectID = &projectID
	}

	secretsResponse, err := r.secretAPI.ListSecrets(listSecretsRequest, scw.WithContext(ctx), scw.WithAllPages())
	if err != nil {
		return nil, fmt.Errorf("failed to list secrets: %w", err)
	}

	return secretsResponse.Secrets, nil
}

func (r *VersionListResource) fetchVersionsForTarget(ctx context.Context, target versionListTarget, data VersionListResourceModel) ([]*secret.SecretVersion, error) {
	statuses, err := r.parseStatuses(ctx, data)
	if err != nil {
		return nil, err
	}

	return r.fetchVersionsForSecret(ctx, target.Region, target.SecretID, statuses)
}

func (r *VersionListResource) fetchVersionsForSecret(ctx context.Context, region scw.Region, secretID string, statuses []secret.SecretVersionStatus) ([]*secret.SecretVersion, error) {
	request := &secret.ListSecretVersionsRequest{
		Region:   region,
		SecretID: secretID,
	}

	if len(statuses) > 0 {
		request.Status = statuses
	}

	response, err := r.secretAPI.ListSecretVersions(request, scw.WithContext(ctx), scw.WithAllPages())
	if err != nil {
		return nil, err
	}

	return response.Versions, nil
}
