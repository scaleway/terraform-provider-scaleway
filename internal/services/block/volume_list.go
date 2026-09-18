package block

import (
	"context"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/list"
	"github.com/hashicorp/terraform-plugin-framework/list/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	blockSDK "github.com/scaleway/scaleway-sdk-go/api/block/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/identity/framework"
	listscw "github.com/scaleway/terraform-provider-scaleway/v2/internal/list"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/zonal"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
)

var (
	_ list.ListResource              = (*VolumeListResource)(nil)
	_ list.ListResourceWithConfigure = (*VolumeListResource)(nil)
)

type VolumeListResource struct {
	meta     *meta.Meta
	blockAPI *blockSDK.API
}

func (r *VolumeListResource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	m := listscw.ConfigureMeta(request, response)
	if m == nil {
		return
	}

	r.meta = m
	r.blockAPI = blockSDK.NewAPI(meta.ExtractScwClient(m))
}

func NewVolumeListResource() list.ListResource {
	return &VolumeListResource{}
}

func (r *VolumeListResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_block_volume"
}

func (r *VolumeListResource) ListResourceConfigSchema(_ context.Context, _ list.ListResourceSchemaRequest, response *list.ListResourceSchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"zones": listscw.ZonesAttribute("Zones of the block volume to filter on"),
			"project_ids": listscw.ProjectIDsAttribute(
				"Project IDs of the block volume to filter on",
			),
			"name":            listscw.NameAttribute("Name of the volume to filter on"),
			"tags":            listscw.TagsAttribute("Tags of the volume to filter on"),
			"organization_id": listscw.OrganizationIDAttribute("Organization ID of the volume to filter on"),
		},
	}
}

type VolumeListResourceModel struct {
	ProjectIDs     types.List   `tfsdk:"project_ids"`
	Zones          types.List   `tfsdk:"zones"`
	Name           types.String `tfsdk:"name"`
	Tags           types.List   `tfsdk:"tags"`
	OrganizationID types.String `tfsdk:"organization_id"`
}

func (m *VolumeListResourceModel) GetZones() types.List {
	return m.Zones
}

func (m *VolumeListResourceModel) GetProjects() types.List {
	return m.ProjectIDs
}

func (m *VolumeListResourceModel) GetTags() types.List {
	return m.Tags
}

type blockVolumeRow struct {
	Volume    *blockSDK.Volume
	Zone      scw.Zone
	ProjectID string
}

type volumeListTarget struct {
	Zone      scw.Zone
	ProjectID string
}

func (r *VolumeListResource) List(ctx context.Context, req list.ListRequest, stream *list.ListResultsStream) {
	var data VolumeListResourceModel

	diags := req.Config.Get(ctx, &data)
	if diags.HasError() {
		stream.Results = list.ListResultsStreamDiagnostics(diags)

		return
	}

	zones, err := listscw.ExtractZones(ctx, &data, r.meta)
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

	targets := make([]volumeListTarget, 0, len(zones)*len(projects))
	for _, zone := range zones {
		for _, project := range projects {
			targets = append(targets, volumeListTarget{
				Zone:      zone,
				ProjectID: project,
			})
		}
	}

	allRows, err := listscw.FetchConcurrently(ctx, targets,
		func(ctx context.Context, target volumeListTarget) ([]blockVolumeRow, error) {
			return r.fetchVolumeRows(ctx, target, data)
		},
		func(a, b blockVolumeRow) int {
			if a.ProjectID != b.ProjectID {
				return strings.Compare(a.ProjectID, b.ProjectID)
			}

			if a.Zone != b.Zone {
				return strings.Compare(string(a.Zone), string(b.Zone))
			}

			if a.Volume.ID != b.Volume.ID {
				return strings.Compare(a.Volume.ID, b.Volume.ID)
			}

			return 0
		},
	)
	if err != nil {
		stream.Results = list.ListResultsStreamDiagnostics(diag.Diagnostics{
			diag.NewErrorDiagnostic("Listing block volumes", "Failed to list block volumes: "+err.Error()),
		})

		return
	}

	stream.Results = func(push func(list.ListResult) bool) {
		for _, row := range allRows {
			result := req.NewListResult(ctx)
			result.DisplayName = row.Volume.Name

			identityDiags := result.Identity.Set(ctx, framework.SetZonalIdentity(row.Zone, row.Volume.ID))
			result.Diagnostics.Append(identityDiags...)

			if req.IncludeResource {
				resourceModel := flattenVolumeForList(ctx, row.Volume, &result.Diagnostics)
				resourceDiags := result.Resource.Set(ctx, resourceModel)
				result.Diagnostics.Append(resourceDiags...)
			}

			if !push(result) {
				return
			}
		}
	}
}

func (r *VolumeListResource) fetchVolumeRows(ctx context.Context, target volumeListTarget, data VolumeListResourceModel) ([]blockVolumeRow, error) {
	listReq := &blockSDK.ListVolumesRequest{
		Zone:      target.Zone,
		ProjectID: &target.ProjectID,
	}

	if !data.Name.IsNull() && !data.Name.IsUnknown() {
		n := strings.TrimSpace(data.Name.ValueString())
		if n != "" {
			listReq.Name = &n
		}
	}

	if !data.Tags.IsNull() && !data.Tags.IsUnknown() {
		var tagStrings []string
		data.Tags.ElementsAs(ctx, &tagStrings, false)

		if len(tagStrings) > 0 {
			listReq.Tags = tagStrings
		}
	}

	if !data.OrganizationID.IsNull() && !data.OrganizationID.IsUnknown() {
		orgID := data.OrganizationID.ValueString()
		listReq.OrganizationID = &orgID
	}

	resp, err := r.blockAPI.ListVolumes(listReq, scw.WithContext(ctx), scw.WithAllPages())
	if err != nil {
		return nil, err
	}

	rows := make([]blockVolumeRow, 0, len(resp.Volumes))
	for _, volume := range resp.Volumes {
		if volume == nil {
			continue
		}

		rows = append(rows, blockVolumeRow{
			Zone:      target.Zone,
			ProjectID: target.ProjectID,
			Volume:    volume,
		})
	}

	return rows, nil
}

func flattenVolumeForList(ctx context.Context, volume *blockSDK.Volume, diags *diag.Diagnostics) volumeResourceModel {
	model := volumeResourceModel{
		ID:        types.StringValue(zonal.NewIDString(volume.Zone, volume.ID)),
		Name:      types.StringValue(volume.Name),
		SizeInGB:  types.Int64Value(int64(volume.Size / scw.GB)),
		ProjectID: types.StringValue(volume.ProjectID),
		Zone:      types.StringValue(volume.Zone.String()),
		SRN:       types.StringValue(volume.Srn),
	}

	tagsList, d := types.ListValueFrom(ctx, types.StringType, volume.Tags)
	diags.Append(d...)

	model.Tags = tagsList

	if volume.Specs != nil && volume.Specs.PerfIops != nil {
		model.Iops = types.Int64Value(int64(*volume.Specs.PerfIops))
	}

	if volume.ParentSnapshotID != nil {
		model.SnapshotID = types.StringValue(zonal.NewIDString(volume.Zone, *volume.ParentSnapshotID))
	} else {
		model.SnapshotID = types.StringNull()
	}

	return model
}
