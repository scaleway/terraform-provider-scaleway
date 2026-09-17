package block

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/list"
	"github.com/hashicorp/terraform-plugin-framework/list/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	blockSDK "github.com/scaleway/scaleway-sdk-go/api/block/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/identity/framework"
	listscw "github.com/scaleway/terraform-provider-scaleway/v2/internal/list"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/zonal"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/verify"
)

var (
	_ list.ListResource              = (*SnapshotListResource)(nil)
	_ list.ListResourceWithConfigure = (*SnapshotListResource)(nil)
)

type SnapshotListResource struct {
	meta     *meta.Meta
	blockAPI *blockSDK.API
}

func (r *SnapshotListResource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	m := listscw.ConfigureMeta(request, response)
	if m == nil {
		return
	}

	r.meta = m
	r.blockAPI = blockSDK.NewAPI(meta.ExtractScwClient(m))
}

func NewSnapshotListResource() list.ListResource {
	return &SnapshotListResource{}
}

func (r *SnapshotListResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_block_snapshot"
}

func (r *SnapshotListResource) ListResourceConfigSchema(_ context.Context, _ list.ListResourceSchemaRequest, response *list.ListResourceSchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"zones": listscw.ZonesAttribute("Zones of the block snapshot to filter on"),
			"project_ids": listscw.ProjectIDsAttribute(
				"Project IDs of the block snapshot to filter on",
			),
			"volume_ids": schema.ListAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Description: "Volume IDs to list snapshots from. Use [\"*\"] only to include " +
					"all volumes in each selected zone and project. Otherwise each value must " +
					"be a zonal ID (`zone/uuid`) or a bare volume UUID.",
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
					listvalidator.ValueStringsAre(
						stringvalidator.Any(
							stringvalidator.OneOf("*"),
							verify.IsStringUUIDOrUUIDWithZone(),
						),
					),
				},
			},
			"name":            listscw.NameAttribute("Name of the snapshot to filter on"),
			"tags":            listscw.TagsAttribute("Tags of the snapshot to filter on"),
			"organization_id": listscw.OrganizationIDAttribute("Organization ID of the snapshot to filter on"),
			"include_deleted": schema.BoolAttribute{
				Description: "Display deleted snapshots not erased yet",
				Optional:    true,
			},
		},
	}
}

type SnapshotListResourceModel struct {
	VolumeIDs      types.List   `tfsdk:"volume_ids"`
	ProjectIDs     types.List   `tfsdk:"project_ids"`
	Zones          types.List   `tfsdk:"zones"`
	Name           types.String `tfsdk:"name"`
	Tags           types.List   `tfsdk:"tags"`
	OrganizationID types.String `tfsdk:"organization_id"`
	IncludeDeleted types.Bool   `tfsdk:"include_deleted"`
}

func (m *SnapshotListResourceModel) GetZones() types.List {
	return m.Zones
}

func (m *SnapshotListResourceModel) GetProjects() types.List {
	return m.ProjectIDs
}

func (m *SnapshotListResourceModel) GetTags() types.List {
	return m.Tags
}

type blockSnapshotRow struct {
	Snapshot  *blockSDK.Snapshot
	Zone      scw.Zone
	ProjectID string
}

type snapshotListTarget struct {
	Zone      scw.Zone
	ProjectID string
	VolumeID  string
}

func (r *SnapshotListResource) List(ctx context.Context, req list.ListRequest, stream *list.ListResultsStream) {
	var data SnapshotListResourceModel

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

	var volumeIDElems []string

	if !data.VolumeIDs.IsNull() && !data.VolumeIDs.IsUnknown() {
		diags = data.VolumeIDs.ElementsAs(ctx, &volumeIDElems, true)
		if diags.HasError() {
			stream.Results = list.ListResultsStreamDiagnostics(diags)

			return
		}
	}

	if len(volumeIDElems) == 0 {
		stream.Results = list.ListResultsStreamDiagnostics(diag.Diagnostics{
			diag.NewErrorDiagnostic("Invalid volume_ids", "`volume_ids` must contain at least one element."),
		})

		return
	}

	if slices.Contains(volumeIDElems, "*") && len(volumeIDElems) != 1 {
		stream.Results = list.ListResultsStreamDiagnostics(diag.Diagnostics{
			diag.NewErrorDiagnostic("Invalid volume_ids", `When using "*", volume_ids must be exactly ["*"].`),
		})

		return
	}

	targets, targetDiags := r.buildSnapshotListTargets(ctx, volumeIDElems, zones, projects)
	if targetDiags.HasError() {
		stream.Results = list.ListResultsStreamDiagnostics(targetDiags)

		return
	}

	allRows, err := listscw.FetchConcurrently(ctx, targets,
		func(ctx context.Context, target snapshotListTarget) ([]blockSnapshotRow, error) {
			return r.fetchSnapshotRows(ctx, target, data)
		},
		func(a, b blockSnapshotRow) int {
			if a.ProjectID != b.ProjectID {
				return strings.Compare(a.ProjectID, b.ProjectID)
			}

			if a.Zone != b.Zone {
				return strings.Compare(string(a.Zone), string(b.Zone))
			}

			if a.Snapshot.ID != b.Snapshot.ID {
				return strings.Compare(a.Snapshot.ID, b.Snapshot.ID)
			}

			return 0
		},
	)
	if err != nil {
		stream.Results = list.ListResultsStreamDiagnostics(diag.Diagnostics{
			diag.NewErrorDiagnostic("Listing block snapshots", "Failed to list block snapshots: "+err.Error()),
		})

		return
	}

	stream.Results = func(push func(list.ListResult) bool) {
		for _, row := range allRows {
			result := req.NewListResult(ctx)
			result.DisplayName = row.Snapshot.Name

			identityDiags := result.Identity.Set(ctx, framework.SetZonalIdentity(row.Zone, row.Snapshot.ID))
			result.Diagnostics.Append(identityDiags...)

			if req.IncludeResource {
				resourceModel := flattenSnapshotForList(ctx, row.Snapshot, &result.Diagnostics)
				resourceDiags := result.Resource.Set(ctx, resourceModel)
				result.Diagnostics.Append(resourceDiags...)
			}

			if !push(result) {
				return
			}
		}
	}
}

func (r *SnapshotListResource) buildSnapshotListTargets(
	ctx context.Context,
	volumeIDElems []string,
	zones []scw.Zone,
	projects []string,
) ([]snapshotListTarget, diag.Diagnostics) {
	var diags diag.Diagnostics

	if volumeIDElems[0] == "*" {
		targets := make([]snapshotListTarget, 0, len(zones)*len(projects))
		for _, zone := range zones {
			for _, project := range projects {
				targets = append(targets, snapshotListTarget{
					Zone:      zone,
					ProjectID: project,
					VolumeID:  "",
				})
			}
		}

		return targets, diags
	}

	projectSet := make(map[string]struct{}, len(projects))
	for _, p := range projects {
		projectSet[p] = struct{}{}
	}

	targets := make([]snapshotListTarget, 0, len(volumeIDElems))

	for _, rawID := range volumeIDElems {
		zone, volumeUUID, err := zonal.ParseID(rawID)
		if err != nil {
			diags.AddError("Invalid volume_ids", fmt.Sprintf("Could not parse volume id %q: %v", rawID, err))

			continue
		}

		if !slices.Contains(zones, zone) {
			diags.AddError(
				"Invalid volume_ids",
				fmt.Sprintf("Volume %q is in zone %q which is not included in the configured zones for this list.", rawID, zone),
			)

			continue
		}

		vol, err := r.blockAPI.GetVolume(&blockSDK.GetVolumeRequest{
			Zone:     zone,
			VolumeID: volumeUUID,
		}, scw.WithContext(ctx))
		if err != nil {
			diags.AddError("Listing block snapshots", fmt.Sprintf("Could not load volume %q: %v", rawID, err))

			continue
		}

		if _, ok := projectSet[vol.ProjectID]; !ok {
			diags.AddError(
				"Invalid volume_ids",
				fmt.Sprintf("Volume %q belongs to project %q which is not included in the configured project_ids for this list.", rawID, vol.ProjectID),
			)

			continue
		}

		targets = append(targets, snapshotListTarget{
			Zone:      zone,
			ProjectID: vol.ProjectID,
			VolumeID:  volumeUUID,
		})
	}

	return targets, diags
}

func (r *SnapshotListResource) fetchSnapshotRows(ctx context.Context, target snapshotListTarget, data SnapshotListResourceModel) ([]blockSnapshotRow, error) {
	if target.VolumeID == "" {
		return r.fetchSnapshotRowsForAllVolumes(ctx, target, data)
	}

	return r.fetchSnapshotRowsForProject(ctx, target.Zone, target.ProjectID, &target.VolumeID, data)
}

func (r *SnapshotListResource) fetchSnapshotRowsForAllVolumes(
	ctx context.Context,
	target snapshotListTarget,
	data SnapshotListResourceModel,
) (
	[]blockSnapshotRow, error,
) {
	return r.fetchSnapshotRowsForProject(ctx, target.Zone, target.ProjectID, nil, data)
}

func (r *SnapshotListResource) fetchSnapshotRowsForProject(
	ctx context.Context,
	zone scw.Zone,
	projectID string,
	volumeID *string,
	data SnapshotListResourceModel,
) ([]blockSnapshotRow, error) {
	listReq := &blockSDK.ListSnapshotsRequest{
		Zone:      zone,
		ProjectID: &projectID,
		VolumeID:  volumeID,
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

	if !data.IncludeDeleted.IsNull() && !data.IncludeDeleted.IsUnknown() {
		listReq.IncludeDeleted = data.IncludeDeleted.ValueBool()
	}

	resp, err := r.blockAPI.ListSnapshots(listReq, scw.WithContext(ctx), scw.WithAllPages())
	if err != nil {
		return nil, err
	}

	rows := make([]blockSnapshotRow, 0, len(resp.Snapshots))
	for _, snapshot := range resp.Snapshots {
		if snapshot == nil {
			continue
		}

		rows = append(rows, blockSnapshotRow{
			Zone:      zone,
			ProjectID: projectID,
			Snapshot:  snapshot,
		})
	}

	return rows, nil
}

func flattenSnapshotForList(ctx context.Context, snapshot *blockSDK.Snapshot, diags *diag.Diagnostics) snapshotResourceModel {
	model := snapshotResourceModel{
		ID:        types.StringValue(zonal.NewIDString(snapshot.Zone, snapshot.ID)),
		Name:      types.StringValue(snapshot.Name),
		ProjectID: types.StringValue(snapshot.ProjectID),
		Zone:      types.StringValue(snapshot.Zone.String()),
		SRN:       types.StringValue(snapshot.Srn),
	}

	tagsList, d := types.ListValueFrom(ctx, types.StringType, snapshot.Tags)
	diags.Append(d...)
	model.Tags = tagsList

	if snapshot.ParentVolume != nil {
		model.VolumeID = types.StringValue(zonal.NewIDString(snapshot.Zone, snapshot.ParentVolume.ID))
	} else {
		model.VolumeID = types.StringNull()
	}

	return model
}
