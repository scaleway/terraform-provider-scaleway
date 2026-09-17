package block

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
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
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/identity/framework"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/zonal"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
	scwtypes "github.com/scaleway/terraform-provider-scaleway/v2/internal/types"
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
	blockAPI *block.API
	meta     *meta.Meta
}

type snapshotResourceModel struct {
	ID        types.String `tfsdk:"id"`
	Name      types.String `tfsdk:"name"`
	VolumeID  types.String `tfsdk:"volume_id"`
	Tags      types.List   `tfsdk:"tags"`
	Import    types.Object `tfsdk:"import"`
	Export    types.Object `tfsdk:"export"`
	SRN       types.String `tfsdk:"srn"`
	Zone      types.String `tfsdk:"zone"`
	ProjectID types.String `tfsdk:"project_id"`
}

type snapshotImportExportModel struct {
	Bucket types.String `tfsdk:"bucket"`
	Key    types.String `tfsdk:"key"`
}

type snapshotResourceIdentityModel = framework.ZonalIdentity

func (r *SnapshotResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_block_snapshot"
}

func (r *SnapshotResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Scaleway Block Snapshot.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The ID of the snapshot, in the `{zone}/{id}` format.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Computed:    true,
				Optional:    true,
				Description: "The snapshot name",
			},
			"volume_id": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "ID of the volume from which creates a snapshot",
				Validators: []validator.String{
					verify.IsStringUUIDOrUUIDWithZone(),
					stringvalidator.ConflictsWith(path.MatchRoot("import")),
				},
				PlanModifiers: []planmodifier.String{
					zonal.LocalityPlanModifier(),
				},
			},
			"tags": schema.ListAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Description: "The tags associated with the snapshot",
			},
			"srn": schema.StringAttribute{
				Computed:    true,
				Description: "The Scaleway Resource Name (SRN) of the snapshot",
			},
			"zone": zonal.SchemaAttributeComputed("The zone you want to attach the resource to"),
			"project_id": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "The project ID the snapshot belongs to. Defaults to the provider's project ID.",
				Validators: []validator.String{
					verify.IsStringUUID(),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
		Blocks: map[string]schema.Block{
			"import": schema.SingleNestedBlock{
				Description: "Import snapshot from a qcow",
				Attributes: map[string]schema.Attribute{
					"bucket": schema.StringAttribute{
						Optional:    true,
						Description: "Bucket containing qcow",
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.RequiresReplace(),
						},
					},
					"key": schema.StringAttribute{
						Optional:    true,
						Description: "Key of the qcow file in the specified bucket",
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.RequiresReplace(),
						},
					},
				},
				PlanModifiers: []planmodifier.Object{
					requireReplaceIfObjectChanged(),
				},
			},
			"export": schema.SingleNestedBlock{
				Description: "Export snapshot to a qcow",
				Attributes: map[string]schema.Attribute{
					"bucket": schema.StringAttribute{
						Optional:    true,
						Description: "Bucket containing qcow",
					},
					"key": schema.StringAttribute{
						Optional:    true,
						Description: "Key of the qcow file in the specified bucket",
					},
				},
			},
		},
	}
}

func requireReplaceIfObjectChanged() planmodifier.Object {
	return objectplanmodifier.RequiresReplaceIf(
		func(_ context.Context, _ planmodifier.ObjectRequest, resp *objectplanmodifier.RequiresReplaceIfFuncResponse) {
			resp.RequiresReplace = true
		},
		"Force replacement when the import block changes.",
		"Force replacement when the import block changes.",
	)
}

func (r *SnapshotResource) IdentitySchema(_ context.Context, _ resource.IdentitySchemaRequest, resp *resource.IdentitySchemaResponse) {
	resp.IdentitySchema = framework.DefaultZonal()
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
	r.blockAPI = block.NewAPI(r.meta.ScwClient())
}

func (r *SnapshotResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data snapshotResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	zone, err := meta.ExtractFrameworkZone(data.Zone, r.meta.ScwClient())
	if err != nil {
		resp.Diagnostics.AddError("Failed to resolve zone", err.Error())

		return
	}

	projectID, err := meta.ExtractFrameworkProjectID(data.ProjectID, r.meta.ScwClient())
	if err != nil {
		resp.Diagnostics.AddError("Failed to resolve project ID", err.Error())

		return
	}

	var snapshot *block.Snapshot

	if data.Import.IsNull() {
		createReq := &block.CreateSnapshotRequest{
			Zone:      zone,
			ProjectID: projectID,
			Name:      scwtypes.ExpandOrGenerateString(data.Name.ValueString(), "snapshot"),
			VolumeID:  locality.ExpandID(data.VolumeID.ValueString()),
		}

		createReq.Tags = scwtypes.ExpandUpdatedStringList(ctx, data.Tags, &resp.Diagnostics)
		if resp.Diagnostics.HasError() {
			return
		}

		snapshot, err = r.blockAPI.CreateSnapshot(createReq, scw.WithContext(ctx))
		if err != nil {
			resp.Diagnostics.AddError("Failed to create block snapshot", err.Error())

			return
		}
	} else {
		var importData snapshotImportExportModel
		resp.Diagnostics.Append(data.Import.As(ctx, &importData, basetypes.ObjectAsOptions{})...)

		if resp.Diagnostics.HasError() {
			return
		}

		importReq := &block.ImportSnapshotFromObjectStorageRequest{
			Zone:      zone,
			ProjectID: projectID,
			Name:      scwtypes.ExpandOrGenerateString(data.Name.ValueString(), "snapshot"),
			Bucket:    regional.ExpandID(importData.Bucket.ValueString()).ID,
			Key:       importData.Key.ValueString(),
		}

		importReq.Tags = scwtypes.ExpandUpdatedStringList(ctx, data.Tags, &resp.Diagnostics)
		if resp.Diagnostics.HasError() {
			return
		}

		snapshot, err = r.blockAPI.ImportSnapshotFromObjectStorage(importReq, scw.WithContext(ctx))
		if err != nil {
			resp.Diagnostics.AddError("Failed to import block snapshot", err.Error())

			return
		}
	}

	snapshot, err = waitForBlockSnapshot(ctx, r.blockAPI, zone, snapshot.ID, defaultBlockTimeout)
	if err != nil {
		resp.Diagnostics.AddError("Failed to wait for block snapshot during Create", err.Error())

		return
	}

	if !data.Export.IsNull() {
		var exportData snapshotImportExportModel
		resp.Diagnostics.Append(data.Export.As(ctx, &exportData, basetypes.ObjectAsOptions{})...)

		if resp.Diagnostics.HasError() {
			return
		}

		_, err = r.blockAPI.ExportSnapshotToObjectStorage(&block.ExportSnapshotToObjectStorageRequest{
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

	state := flattenSnapshot(ctx, snapshot, req, &resp.Diagnostics)
	state.Import = data.Import
	state.Export = data.Export
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	resp.Diagnostics.Append(resp.Identity.Set(ctx, framework.SetZonalIdentity(snapshot.Zone, snapshot.ID))...)
}

func (r *SnapshotResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var (
		state    snapshotResourceModel
		identity snapshotResourceIdentityModel
	)

	resp.Diagnostics.Append(req.Identity.Get(ctx, &identity)...)
	identityAvailable := !resp.Diagnostics.HasError() && !identity.ID.IsNull() && !identity.ID.IsUnknown()

	if !identityAvailable && resp.Diagnostics.HasError() {
		resp.Diagnostics = nil
	}

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var (
		zone scw.Zone
		id   string
		err  error
	)

	if identityAvailable {
		zone, id, err = zonal.ParseID(identity.ID.ValueString())
	} else {
		zone, id, err = zonal.ParseID(state.ID.ValueString())
	}

	if err != nil {
		resp.Diagnostics.AddError("Failed to parse block snapshot ID", err.Error())

		return
	}

	snapshot, err := waitForBlockSnapshot(ctx, r.blockAPI, zone, id, defaultBlockTimeout)
	if err != nil {
		if httperrors.Is404(err) {
			resp.State.RemoveResource(ctx)

			return
		}

		resp.Diagnostics.AddError("Failed to wait for block snapshot during Read", err.Error())

		return
	}

	newState := flattenSnapshot(ctx, snapshot, req, &resp.Diagnostics)
	// Preserve config-only fields
	newState.Import = state.Import
	newState.Export = state.Export

	resp.Diagnostics.Append(resp.State.Set(ctx, &newState)...)
	resp.Diagnostics.Append(resp.Identity.Set(ctx, framework.SetZonalIdentity(snapshot.Zone, snapshot.ID))...)
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

	zone, id, err := zonal.ParseID(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to parse block snapshot ID", err.Error())

		return
	}

	snapshot, err := waitForBlockSnapshot(ctx, r.blockAPI, zone, id, defaultBlockTimeout)
	if err != nil {
		if httperrors.Is404(err) {
			resp.State.RemoveResource(ctx)

			return
		}

		resp.Diagnostics.AddError("Failed to wait for block snapshot during Update", err.Error())

		return
	}

	updateReq := &block.UpdateSnapshotRequest{
		Zone:       snapshot.Zone,
		SnapshotID: snapshot.ID,
	}

	hasChanges := false

	if !plan.Name.Equal(state.Name) {
		name := plan.Name.ValueString()
		updateReq.Name = &name
		hasChanges = true
	}

	if !plan.Tags.Equal(state.Tags) {
		tags := scwtypes.ExpandUpdatedStringList(ctx, plan.Tags, &resp.Diagnostics)
		if resp.Diagnostics.HasError() {
			return
		}

		updateReq.Tags = &tags
		hasChanges = true
	}

	if hasChanges {
		_, err = r.blockAPI.UpdateSnapshot(updateReq, scw.WithContext(ctx))
		if err != nil {
			resp.Diagnostics.AddError("Failed to update block snapshot", err.Error())

			return
		}
	}

	if !plan.Export.Equal(state.Export) && !plan.Export.IsNull() {
		var exportData snapshotImportExportModel
		resp.Diagnostics.Append(plan.Export.As(ctx, &exportData, basetypes.ObjectAsOptions{})...)

		if resp.Diagnostics.HasError() {
			return
		}

		_, err = r.blockAPI.ExportSnapshotToObjectStorage(&block.ExportSnapshotToObjectStorageRequest{
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

	snapshot, err = waitForBlockSnapshot(ctx, r.blockAPI, zone, id, defaultBlockTimeout)
	if err != nil {
		if httperrors.Is404(err) {
			resp.State.RemoveResource(ctx)

			return
		}

		resp.Diagnostics.AddError("Failed to wait for block snapshot during Update", err.Error())

		return
	}

	newState := flattenSnapshot(ctx, snapshot, req, &resp.Diagnostics)
	// Preserve config-only fields
	newState.Import = state.Import
	newState.Export = plan.Export

	resp.Diagnostics.Append(resp.State.Set(ctx, &newState)...)
	resp.Diagnostics.Append(resp.Identity.Set(ctx, framework.SetZonalIdentity(snapshot.Zone, snapshot.ID))...)
}

func (r *SnapshotResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state snapshotResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	zone, id, err := zonal.ParseID(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to parse block snapshot ID", err.Error())

		return
	}

	snapshot, err := waitForBlockSnapshotToBeAvailable(ctx, r.blockAPI, zone, id, defaultBlockTimeout)
	if err != nil {
		if httperrors.Is404(err) {
			return
		}

		resp.Diagnostics.AddError("Failed to wait for block snapshot during Delete", err.Error())

		return
	}

	err = r.blockAPI.DeleteSnapshot(&block.DeleteSnapshotRequest{
		Zone:       snapshot.Zone,
		SnapshotID: snapshot.ID,
	}, scw.WithContext(ctx))
	if err != nil {
		if httperrors.Is404(err) {
			return
		}

		resp.Diagnostics.AddError("Failed to delete block snapshot", err.Error())

		return
	}

	_, err = waitForBlockSnapshot(ctx, r.blockAPI, zone, id, defaultBlockTimeout)
	if err != nil && !httperrors.Is404(err) {
		resp.Diagnostics.AddError("Failed to wait for block snapshot during Delete", err.Error())
	}
}

func (r *SnapshotResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughWithIdentity(ctx, path.Root("id"), path.Root("id"), req, resp)
}

func flattenSnapshot(ctx context.Context, snapshot *block.Snapshot, reference any, diags *diag.Diagnostics) snapshotResourceModel {
	model := snapshotResourceModel{
		ID:        types.StringValue(zonal.NewIDString(snapshot.Zone, snapshot.ID)),
		Name:      types.StringValue(snapshot.Name),
		ProjectID: types.StringValue(snapshot.ProjectID),
		Zone:      types.StringValue(snapshot.Zone.String()),
		SRN:       types.StringValue(snapshot.Srn),
	}

	tagsList, d := scwtypes.FlattenStringList(ctx, "tags", snapshot.Tags, reference)
	diags.Append(d...)

	model.Tags = tagsList

	if snapshot.ParentVolume != nil {
		model.VolumeID = types.StringValue(zonal.NewIDString(snapshot.Zone, snapshot.ParentVolume.ID))
	} else {
		model.VolumeID = types.StringNull()
	}

	return model
}
