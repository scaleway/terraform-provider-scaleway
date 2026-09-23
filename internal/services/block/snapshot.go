package block

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/objectvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/scaleway/scaleway-sdk-go/api/block/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	identityfw "github.com/scaleway/terraform-provider-scaleway/v2/internal/identity/framework"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/zonal"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
	providertypes "github.com/scaleway/terraform-provider-scaleway/v2/internal/types"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/verify"
)

var (
	_ resource.Resource                = (*SnapshotResource)(nil)
	_ resource.ResourceWithConfigure   = (*SnapshotResource)(nil)
	_ resource.ResourceWithImportState = (*SnapshotResource)(nil)
	_ resource.ResourceWithIdentity    = (*SnapshotResource)(nil)
)

func NewSnapshotResource() resource.Resource {
	return &SnapshotResource{}
}

type SnapshotResource struct {
	api  *block.API
	meta *meta.Meta
}

type snapshotResourceModel struct {
	ID        types.String `tfsdk:"id"`
	Name      types.String `tfsdk:"name"`
	VolumeID  types.String `tfsdk:"volume_id"`
	Tags      types.List   `tfsdk:"tags"`
	Import    types.Object `tfsdk:"import"`
	Export    types.Object `tfsdk:"export"`
	Srn       types.String `tfsdk:"srn"`
	Zone      types.String `tfsdk:"zone"`
	ProjectID types.String `tfsdk:"project_id"`
}

type snapshotResourceIdentityModel = identityfw.ZonalIdentity

type snapshotImportModel struct {
	Bucket types.String `tfsdk:"bucket"`
	Key    types.String `tfsdk:"key"`
}

type snapshotExportModel struct {
	Bucket types.String `tfsdk:"bucket"`
	Key    types.String `tfsdk:"key"`
}

func (r *SnapshotResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_block_snapshot"
}

func (r *SnapshotResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Scaleway Block Snapshot",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The ID of the snapshot",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "The snapshot name",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"volume_id": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "ID of the volume from which creates a snapshot",
				Validators: []validator.String{
					verify.IsStringUUIDOrUUIDWithZone(),
				},
			},
			"tags": schema.ListAttribute{
				Optional:            true,
				ElementType:         types.StringType,
				MarkdownDescription: "The tags associated with the snapshot",
			},
			"import": schema.SingleNestedAttribute{
				Optional:            true,
				MarkdownDescription: "Import snapshot from a qcow",
				PlanModifiers: []planmodifier.Object{
					objectplanmodifier.RequiresReplace(),
				},
				Validators: []validator.Object{
					objectvalidator.ConflictsWith(path.MatchRoot("volume_id")),
				},
				Attributes: map[string]schema.Attribute{
					"bucket": schema.StringAttribute{
						Required:            true,
						MarkdownDescription: "Bucket containing qcow",
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.RequiresReplace(),
						},
					},
					"key": schema.StringAttribute{
						Required:            true,
						MarkdownDescription: "Key of the qcow file in the specified bucket",
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.RequiresReplace(),
						},
					},
				},
			},
			"export": schema.SingleNestedAttribute{
				Optional:            true,
				MarkdownDescription: "Export snapshot to a qcow",
				Attributes: map[string]schema.Attribute{
					"bucket": schema.StringAttribute{
						Required:            true,
						MarkdownDescription: "Bucket containing qcow",
					},
					"key": schema.StringAttribute{
						Required:            true,
						MarkdownDescription: "Key of the qcow file in the specified bucket",
					},
				},
			},
			"srn": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The Scaleway Resource Name (SRN) of the snapshot",
			},
			"zone": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "The zone you want to attach the resource to",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
				Validators: []validator.String{
					verify.IsStringOneOfWithWarning(zonal.AllZones()),
				},
			},
			"project_id": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "The project_id you want to attach the resource to",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
				Validators: []validator.String{
					verify.IsStringUUID(),
				},
			},
		},
	}
}

func (r *SnapshotResource) IdentitySchema(_ context.Context, _ resource.IdentitySchemaRequest, resp *resource.IdentitySchemaResponse) {
	resp.IdentitySchema = identityfw.DefaultZonal()
}

func (r *SnapshotResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	m, ok := req.ProviderData.(*meta.Meta)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *meta.Meta, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.meta = m
	r.api = block.NewAPI(r.meta.ScwClient())
}

func (r *SnapshotResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan snapshotResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	zone, err := meta.ExtractFrameworkZone(plan.Zone, r.meta.ScwClient())
	if err != nil {
		resp.Diagnostics.AddError("Failed to resolve zone", err.Error())

		return
	}

	projectID, err := meta.ExtractFrameworkProjectID(plan.ProjectID, r.meta.ScwClient())
	if err != nil {
		resp.Diagnostics.AddError("Failed to resolve project ID", err.Error())

		return
	}

	name := plan.Name.ValueString()
	if name == "" {
		name = providertypes.NewRandomName("snapshot")
	}

	tags := providertypes.ExpandStringList(ctx, plan.Tags, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	var snapshot *block.Snapshot

	if plan.Import.IsNull() || plan.Import.IsUnknown() {
		snapshot, err = r.api.CreateSnapshot(&block.CreateSnapshotRequest{
			Zone:      zone,
			ProjectID: projectID,
			Name:      name,
			VolumeID:  locality.ExpandID(plan.VolumeID.ValueString()),
			Tags:      tags,
		}, scw.WithContext(ctx))
		if err != nil {
			resp.Diagnostics.AddError("Failed to create block snapshot", err.Error())

			return
		}
	} else {
		var importData snapshotImportModel
		resp.Diagnostics.Append(plan.Import.As(ctx, &importData, basetypes.ObjectAsOptions{})...)

		if resp.Diagnostics.HasError() {
			return
		}

		snapshot, err = r.api.ImportSnapshotFromObjectStorage(&block.ImportSnapshotFromObjectStorageRequest{
			Zone:      zone,
			ProjectID: projectID,
			Name:      name,
			Bucket:    regional.ExpandID(importData.Bucket.ValueString()).ID,
			Key:       importData.Key.ValueString(),
			Tags:      tags,
		}, scw.WithContext(ctx))
		if err != nil {
			resp.Diagnostics.AddError("Failed to import block snapshot", err.Error())

			return
		}
	}

	snapshot, err = waitForBlockSnapshot(ctx, r.api, zone, snapshot.ID, defaultBlockTimeout)
	if err != nil {
		resp.Diagnostics.AddError("Failed waiting for block snapshot", err.Error())

		return
	}

	if !plan.Export.IsNull() && !plan.Export.IsUnknown() {
		var exportData snapshotExportModel
		resp.Diagnostics.Append(plan.Export.As(ctx, &exportData, basetypes.ObjectAsOptions{})...)

		if resp.Diagnostics.HasError() {
			return
		}

		_, err = r.api.ExportSnapshotToObjectStorage(&block.ExportSnapshotToObjectStorageRequest{
			Zone:       zone,
			SnapshotID: snapshot.ID,
			Bucket:     regional.ExpandID(exportData.Bucket.ValueString()).ID,
			Key:        exportData.Key.ValueString(),
		}, scw.WithContext(ctx))
		if err != nil {
			resp.Diagnostics.AddError("Failed to export block snapshot", err.Error())

			return
		}
	}

	state := flattenBlockSnapshot(ctx, snapshot, &plan, &resp.Diagnostics)
	state.VolumeID = plan.VolumeID
	state.Import = plan.Import
	state.Export = plan.Export

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	resp.Diagnostics.Append(
		resp.Identity.Set(ctx, identityfw.SetZonalIdentity(snapshot.Zone, snapshot.ID))...,
	)
}

func (r *SnapshotResource) Read(
	ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse,
) {
	var (
		state    snapshotResourceModel
		identity snapshotResourceIdentityModel
	)

	resp.Diagnostics.Append(req.Identity.Get(ctx, &identity)...)
	identityAvailable := !resp.Diagnostics.HasError() &&
		!identity.ID.IsNull() &&
		!identity.ID.IsUnknown()

	if !identityAvailable && resp.Diagnostics.HasError() {
		resp.Diagnostics = nil
	}

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	resourceID := state.ID.ValueString()
	if identityAvailable {
		resourceID = identity.ID.ValueString()
	}

	zone, id, err := zonal.ParseID(locality.ExpandID(resourceID))
	if err != nil {
		resp.Diagnostics.AddError("Failed to parse block snapshot ID", err.Error())

		return
	}

	snapshot, err := waitForBlockSnapshot(ctx, r.api, zone, id, defaultBlockTimeout)
	if err != nil {
		if httperrors.Is404(err) {
			resp.State.RemoveResource(ctx)

			return
		}

		resp.Diagnostics.AddError("Failed to read block snapshot", err.Error())

		return
	}

	newState := flattenBlockSnapshot(ctx, snapshot, &state, &resp.Diagnostics)

	// Preserve the user-provided volume_id (which may carry a zone locality prefix)
	// so it does not drift against the config. Fall back to the API value on import.
	if state.VolumeID.IsNull() || state.VolumeID.IsUnknown() {
		if snapshot.ParentVolume != nil {
			newState.VolumeID = types.StringValue(snapshot.ParentVolume.ID)
		} else {
			newState.VolumeID = types.StringNull()
		}
	} else {
		newState.VolumeID = state.VolumeID
	}

	newState.Import = state.Import
	newState.Export = state.Export

	resp.Diagnostics.Append(resp.State.Set(ctx, &newState)...)
	resp.Diagnostics.Append(
		resp.Identity.Set(ctx, identityfw.SetZonalIdentity(snapshot.Zone, snapshot.ID))...,
	)
}

func (r *SnapshotResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var (
		plan  snapshotResourceModel
		state snapshotResourceModel
	)

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	zone, id, err := zonal.ParseID(locality.ExpandID(state.ID.ValueString()))
	if err != nil {
		resp.Diagnostics.AddError("Failed to parse block snapshot ID", err.Error())

		return
	}

	snapshot, err := waitForBlockSnapshot(ctx, r.api, zone, id, defaultBlockTimeout)
	if err != nil {
		if httperrors.Is404(err) {
			resp.State.RemoveResource(ctx)

			return
		}

		resp.Diagnostics.AddError("Failed to read block snapshot", err.Error())

		return
	}

	updateReq := &block.UpdateSnapshotRequest{
		Zone:       snapshot.Zone,
		SnapshotID: snapshot.ID,
	}

	shouldUpdate := false

	if !plan.Name.Equal(state.Name) {
		name := plan.Name.ValueString()
		updateReq.Name = &name
		shouldUpdate = true
	}

	if !plan.Tags.Equal(state.Tags) {
		tags := providertypes.ExpandUpdatedStringList(ctx, plan.Tags, &resp.Diagnostics)
		if resp.Diagnostics.HasError() {
			return
		}

		updateReq.Tags = &tags
		shouldUpdate = true
	}

	if shouldUpdate {
		if _, err := r.api.UpdateSnapshot(updateReq, scw.WithContext(ctx)); err != nil {
			resp.Diagnostics.AddError("Failed to update block snapshot", err.Error())

			return
		}
	}

	if !plan.Export.Equal(state.Export) && !plan.Export.IsNull() && !plan.Export.IsUnknown() {
		var exportData snapshotExportModel
		resp.Diagnostics.Append(plan.Export.As(ctx, &exportData, basetypes.ObjectAsOptions{})...)

		if resp.Diagnostics.HasError() {
			return
		}

		_, err = r.api.ExportSnapshotToObjectStorage(&block.ExportSnapshotToObjectStorageRequest{
			Zone:       snapshot.Zone,
			SnapshotID: snapshot.ID,
			Bucket:     regional.ExpandID(exportData.Bucket.ValueString()).ID,
			Key:        exportData.Key.ValueString(),
		}, scw.WithContext(ctx))
		if err != nil {
			resp.Diagnostics.AddError("Failed to export block snapshot", err.Error())

			return
		}
	}

	snapshot, err = waitForBlockSnapshot(ctx, r.api, zone, id, defaultBlockTimeout)
	if err != nil {
		resp.Diagnostics.AddError("Failed to read block snapshot after update", err.Error())

		return
	}

	newState := flattenBlockSnapshot(ctx, snapshot, &plan, &resp.Diagnostics)
	newState.VolumeID = plan.VolumeID
	newState.Import = plan.Import
	newState.Export = plan.Export

	resp.Diagnostics.Append(resp.State.Set(ctx, &newState)...)
	resp.Diagnostics.Append(resp.Identity.Set(ctx, identityfw.SetZonalIdentity(snapshot.Zone, snapshot.ID))...)
}

func (r *SnapshotResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state snapshotResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	zone, id, err := zonal.ParseID(locality.ExpandID(state.ID.ValueString()))
	if err != nil {
		resp.Diagnostics.AddError("Failed to parse block snapshot ID", err.Error())

		return
	}

	_, err = waitForBlockSnapshotToBeAvailable(ctx, r.api, zone, id, defaultBlockTimeout)
	if err != nil {
		resp.Diagnostics.AddError("Failed waiting for block snapshot before delete", err.Error())

		return
	}

	err = r.api.DeleteSnapshot(&block.DeleteSnapshotRequest{
		Zone:       zone,
		SnapshotID: id,
	}, scw.WithContext(ctx))
	if err != nil {
		if httperrors.Is404(err) {
			return
		}

		resp.Diagnostics.AddError("Failed to delete block snapshot", err.Error())

		return
	}

	_, err = waitForBlockSnapshot(ctx, r.api, zone, id, defaultBlockTimeout)
	if err != nil && !httperrors.Is404(err) {
		resp.Diagnostics.AddError("Failed waiting for block snapshot deletion", err.Error())
	}
}

func (r *SnapshotResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughWithIdentity(ctx, path.Root("id"), path.Root("id"), req, resp)
}

func flattenBlockSnapshot(ctx context.Context, snapshot *block.Snapshot, reference any, diags *diag.Diagnostics) snapshotResourceModel {
	if snapshot == nil {
		return snapshotResourceModel{}
	}

	state := snapshotResourceModel{
		ID:        types.StringValue(zonal.NewIDString(snapshot.Zone, snapshot.ID)),
		Name:      types.StringValue(snapshot.Name),
		ProjectID: types.StringValue(snapshot.ProjectID),
		Zone:      types.StringValue(snapshot.Zone.String()),
		Srn:       types.StringValue(snapshot.Srn),
	}

	tags, tagsDiags := providertypes.FlattenStringList(ctx, "tags", snapshot.Tags, reference)
	diags.Append(tagsDiags...)

	state.Tags = tags

	return state
}
