package redis

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/list"
	"github.com/hashicorp/terraform-plugin-framework/list/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-mux/tf5to6server/translate"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	redisapi "github.com/scaleway/scaleway-sdk-go/api/redis/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/identity"
	listscw "github.com/scaleway/terraform-provider-scaleway/v2/internal/list"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/zonal"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
)

var (
	_ list.ListResource                 = (*ClusterListResource)(nil)
	_ list.ListResourceWithConfigure    = (*ClusterListResource)(nil)
	_ list.ListResourceWithRawV6Schemas = (*ClusterListResource)(nil)
)

type ClusterListResource struct {
	meta     *meta.Meta
	redisAPI *redisapi.API
}

type zonalFetchTarget struct {
	Zone      scw.Zone
	ProjectID string
}

func (r *ClusterListResource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
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
	r.redisAPI = redisapi.NewAPI(meta.ExtractScwClient(m))
}

func NewClusterListResource() list.ListResource {
	return &ClusterListResource{}
}

func (r *ClusterListResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_redis_cluster"
}

func (r *ClusterListResource) ListResourceConfigSchema(_ context.Context, _ list.ListResourceSchemaRequest, response *list.ListResourceSchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"name":            listscw.NameAttribute("Name of the Redis cluster to filter on"),
			"tags":            listscw.TagsAttribute("Tags of the Redis cluster to filter on"),
			"organization_id": listscw.OrganizationIDAttribute("Organization ID of the Redis cluster to filter on"),
			"project_ids":     listscw.ProjectIDsAttribute("Project IDs of the Redis cluster to filter on"),
			"zones":           zonesAttribute("Zones of the Redis cluster to filter on"),
			"version":         versionAttribute(),
		},
	}
}

func zonesAttribute(description string) schema.ListAttribute {
	return schema.ListAttribute{
		Description: description + " Use '*' to list from all zones",
		Optional:    true,
		ElementType: types.StringType,
		Validators: []validator.List{
			listvalidator.ValueStringsAre(stringvalidator.OneOf(append(zonal.AllZones(), "*")...)),
		},
	}
}

func versionAttribute() schema.StringAttribute {
	return schema.StringAttribute{
		Description: "Redis engine version to filter on",
		Optional:    true,
	}
}

func (r *ClusterListResource) RawV6Schemas(ctx context.Context, _ list.RawV6SchemaRequest, resp *list.RawV6SchemaResponse) {
	clusterResource := ResourceCluster()

	resp.ProtoV6Schema = translate.Schema(clusterResource.ProtoSchema(ctx)())
	resp.ProtoV6IdentitySchema = translate.ResourceIdentitySchema(clusterResource.ProtoIdentitySchema(ctx)())
}

type ClusterListResourceModel struct {
	Tags           types.List   `tfsdk:"tags"`
	Name           types.String `tfsdk:"name"`
	Version        types.String `tfsdk:"version"`
	OrganizationID types.String `tfsdk:"organization_id"`
	ProjectIDs     types.List   `tfsdk:"project_ids"`
	Zones          types.List   `tfsdk:"zones"`
}

func (m *ClusterListResourceModel) GetTags() types.List {
	return m.Tags
}

func (m *ClusterListResourceModel) GetProjects() types.List {
	return m.ProjectIDs
}

func (r *ClusterListResource) List(ctx context.Context, req list.ListRequest, stream *list.ListResultsStream) {
	var data ClusterListResourceModel

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

	zones, err := extractZones(ctx, data, r.meta)
	if err != nil {
		stream.Results = list.ListResultsStreamDiagnostics(diag.Diagnostics{
			diag.NewErrorDiagnostic("Listing zones", "An error was encountered when listing zones: "+err.Error()),
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

	targets := make([]zonalFetchTarget, 0, len(zones)*len(projects))
	for _, zone := range zones {
		for _, project := range projects {
			targets = append(targets, zonalFetchTarget{Zone: zone, ProjectID: project})
		}
	}

	allClusters, err := listscw.FetchConcurrently(ctx, targets,
		func(ctx context.Context, target zonalFetchTarget) ([]*redisapi.Cluster, error) {
			return r.fetchClusters(ctx, target, tags, data)
		},
		func(a, b *redisapi.Cluster) int {
			if a.ProjectID != b.ProjectID {
				return strings.Compare(a.ProjectID, b.ProjectID)
			}

			if a.Zone != b.Zone {
				return strings.Compare(string(a.Zone), string(b.Zone))
			}

			return strings.Compare(a.ID, b.ID)
		},
	)
	if err != nil {
		stream.Results = list.ListResultsStreamDiagnostics(diag.Diagnostics{
			diag.NewErrorDiagnostic("Listing Redis clusters", "Failed to list Redis clusters: "+err.Error()),
		})

		return
	}

	stream.Results = func(push func(list.ListResult) bool) {
		for _, cluster := range allClusters {
			result := req.NewListResult(ctx)
			result.DisplayName = cluster.Name

			clusterResource := ResourceCluster()
			resourceData := clusterResource.Data(&terraform.InstanceState{})

			err := identity.SetZonalIdentity(resourceData, cluster.Zone, cluster.ID)
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

			diagsState := setClusterState(ctx, resourceData, r.redisAPI, cluster.Zone, cluster, r.meta)
			if diagsState.HasError() {
				tflog.Error(ctx, "error from setting setClusterState")

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

func (r *ClusterListResource) fetchClusters(ctx context.Context, target zonalFetchTarget, tags []string, data ClusterListResourceModel) ([]*redisapi.Cluster, error) {
	req := &redisapi.ListClustersRequest{
		Zone:           target.Zone,
		Name:           data.Name.ValueStringPointer(),
		Tags:           tags,
		OrganizationID: data.OrganizationID.ValueStringPointer(),
		ProjectID:      &target.ProjectID,
		OrderBy:        redisapi.ListClustersRequestOrderByCreatedAtAsc,
		Version:        versionFilter(data),
	}

	response, err := r.redisAPI.ListClusters(req, scw.WithContext(ctx), scw.WithAllPages())
	if err != nil {
		return nil, err
	}

	filtered := make([]*redisapi.Cluster, 0, len(response.Clusters))
	for _, cluster := range response.Clusters {
		if cluster == nil {
			continue
		}

		if cluster.ProjectID == target.ProjectID && cluster.Zone == target.Zone {
			filtered = append(filtered, cluster)
		}
	}

	return filtered, nil
}

func versionFilter(data ClusterListResourceModel) *string {
	if data.Version.IsNull() || data.Version.IsUnknown() {
		return nil
	}

	v := strings.TrimSpace(data.Version.ValueString())
	if v == "" {
		return nil
	}

	return &v
}

func extractZones(ctx context.Context, model ClusterListResourceModel, m *meta.Meta) ([]scw.Zone, error) {
	if model.Zones.IsNull() {
		defaultZone, exists := m.ScwClient().GetDefaultZone()
		if !exists {
			return nil, errors.New("no zones specified and no default zone configured")
		}

		return []scw.Zone{defaultZone}, nil
	}

	var zoneStrings []string

	diags := model.Zones.ElementsAs(ctx, &zoneStrings, false)
	if diags.HasError() {
		return nil, fmt.Errorf("converting zones: %s", diags.Errors()[0].Detail())
	}

	res := make([]scw.Zone, 0, len(zoneStrings))
	for _, zone := range zoneStrings {
		if zone == "*" {
			return scw.AllZones, nil
		}

		parsedZone, err := scw.ParseZone(zone)
		if err != nil {
			return nil, err
		}

		res = append(res, parsedZone)
	}

	return res, nil
}
