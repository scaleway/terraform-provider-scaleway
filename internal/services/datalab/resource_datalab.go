package datalab

import (
	"context"
	"errors"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	datalab "github.com/scaleway/scaleway-sdk-go/api/datalab/v1beta1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
)

var (
	_ resource.Resource                = (*DatalabResource)(nil)
	_ resource.ResourceWithConfigure   = (*DatalabResource)(nil)
	_ resource.ResourceWithImportState = (*DatalabResource)(nil)
)

func NewDatalabResource() resource.Resource {
	return &DatalabResource{}
}

type DatalabResource struct {
	api  *datalab.API
	meta *meta.Meta
}

type datalabResourceModel struct {
	Tags              types.List   `tfsdk:"tags"`
	Status            types.String `tfsdk:"status"`
	Worker            types.Object `tfsdk:"worker"`
	Region            types.String `tfsdk:"region"`
	Description       types.String `tfsdk:"description"`
	Name              types.String `tfsdk:"name"`
	SparkVersion      types.String `tfsdk:"spark_version"`
	ProjectID         types.String `tfsdk:"project_id"`
	PrivateNetworkID  types.String `tfsdk:"private_network_id"`
	Main              types.Object `tfsdk:"main"`
	NotebookMasterURL types.String `tfsdk:"notebook_master_url"`
	TotalStorage      types.Object `tfsdk:"total_storage"`
	ID                types.String `tfsdk:"id"`
	CreatedAt         types.String `tfsdk:"created_at"`
	UpdatedAt         types.String `tfsdk:"updated_at"`
	NotebookURL       types.String `tfsdk:"notebook_url"`
	HasNotebook       types.Bool   `tfsdk:"has_notebook"`
}

type sparkMainModel struct {
	NodeType       types.String `tfsdk:"node_type"`
	SparkUIURL     types.String `tfsdk:"spark_ui_url"`
	SparkMasterURL types.String `tfsdk:"spark_master_url"`
	RootVolume     types.Object `tfsdk:"root_volume"`
}

type sparkWorkerModel struct {
	NodeType   types.String `tfsdk:"node_type"`
	RootVolume types.Object `tfsdk:"root_volume"`
	NodeCount  types.Int64  `tfsdk:"node_count"`
}

type volumeModel struct {
	Type types.String `tfsdk:"type"`
	Size types.Int64  `tfsdk:"size"`
}

func volumeAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"type": types.StringType,
		"size": types.Int64Type,
	}
}

func sparkMainAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"node_type":        types.StringType,
		"spark_ui_url":     types.StringType,
		"spark_master_url": types.StringType,
		"root_volume":      types.ObjectType{AttrTypes: volumeAttrTypes()},
	}
}

func sparkWorkerAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"node_type":   types.StringType,
		"node_count":  types.Int64Type,
		"root_volume": types.ObjectType{AttrTypes: volumeAttrTypes()},
	}
}

func (r *DatalabResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_datalab"
}

func (r *DatalabResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The ID of the Datalab instance, in the `{region}/{id}` format.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "The name of the Datalab instance. If not provided, a random name is generated.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"project_id": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "The project ID the Datalab belongs to. Defaults to the provider's project ID.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"region": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "The region the Datalab is in. Only `fr-par` is currently supported.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"description": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "A description for the Datalab instance.",
			},
			"tags": schema.ListAttribute{
				Optional:            true,
				ElementType:         types.StringType,
				MarkdownDescription: "Tags associated with the Datalab instance.",
			},
			"spark_version": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The Spark version to use for the Datalab instance. Available versions can be retrieved from `ListClusterVersions`.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"private_network_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The ID of the private network to attach the Datalab to.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"has_notebook": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Whether a JupyterLab notebook is associated with the Datalab.",
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplace(),
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"main": schema.SingleNestedAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "The Spark main node configuration.",
				PlanModifiers: []planmodifier.Object{
					objectplanmodifier.UseStateForUnknown(),
				},
				Attributes: map[string]schema.Attribute{
					"node_type": schema.StringAttribute{
						Required:            true,
						MarkdownDescription: "The node type for the main node.",
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.RequiresReplace(),
						},
					},
					"spark_ui_url": schema.StringAttribute{
						Computed:            true,
						MarkdownDescription: "The Spark UI URL.",
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
					"spark_master_url": schema.StringAttribute{
						Computed:            true,
						MarkdownDescription: "The Spark master URL within the VPC.",
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
					"root_volume": schema.SingleNestedAttribute{
						Computed:            true,
						MarkdownDescription: "Volume details for the main node.",
						PlanModifiers: []planmodifier.Object{
							objectplanmodifier.UseStateForUnknown(),
						},
						Attributes: map[string]schema.Attribute{
							"type": schema.StringAttribute{
								Computed:            true,
								MarkdownDescription: "The volume type.",
							},
							"size": schema.Int64Attribute{
								Computed:            true,
								MarkdownDescription: "The volume size in bytes.",
							},
						},
					},
				},
			},
			"worker": schema.SingleNestedAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "The Spark worker nodes configuration.",
				PlanModifiers: []planmodifier.Object{
					objectplanmodifier.UseStateForUnknown(),
				},
				Attributes: map[string]schema.Attribute{
					"node_type": schema.StringAttribute{
						Required:            true,
						MarkdownDescription: "The node type for worker nodes.",
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.RequiresReplace(),
						},
					},
					"node_count": schema.Int64Attribute{
						Required:            true,
						MarkdownDescription: "The number of worker nodes.",
					},
					"root_volume": schema.SingleNestedAttribute{
						Computed:            true,
						MarkdownDescription: "Volume details for worker nodes.",
						PlanModifiers: []planmodifier.Object{
							objectplanmodifier.UseStateForUnknown(),
						},
						Attributes: map[string]schema.Attribute{
							"type": schema.StringAttribute{
								Computed:            true,
								MarkdownDescription: "The volume type.",
							},
							"size": schema.Int64Attribute{
								Computed:            true,
								MarkdownDescription: "The volume size in bytes.",
							},
						},
					},
				},
			},
			"total_storage": schema.SingleNestedAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Persistent volume storage configuration.",
				PlanModifiers: []planmodifier.Object{
					objectplanmodifier.RequiresReplace(),
					objectplanmodifier.UseStateForUnknown(),
				},
				Attributes: map[string]schema.Attribute{
					"type": schema.StringAttribute{
						Optional:            true,
						Computed:            true,
						MarkdownDescription: "The volume type. Defaults to `sbs_5k`.",
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.RequiresReplace(),
						},
					},
					"size": schema.Int64Attribute{
						Optional:            true,
						Computed:            true,
						MarkdownDescription: "The volume size in bytes.",
						PlanModifiers: []planmodifier.Int64{
							int64planmodifier.RequiresReplace(),
						},
					},
				},
			},
			"status": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The current status of the Datalab instance.",
			},
			"created_at": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The creation timestamp of the Datalab instance.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"updated_at": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The last update timestamp of the Datalab instance.",
			},
			"notebook_url": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The URL of the JupyterLab notebook, if available.",
			},
			"notebook_master_url": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The URL used to reach the cluster from the notebook.",
			},
		},
	}
}

func (r *DatalabResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
	r.api = datalab.NewAPI(r.meta.ScwClient())
}

func (r *DatalabResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data datalabResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	region, err := r.resolveRegion(data.Region)
	if err != nil {
		resp.Diagnostics.AddError("Failed to resolve region", err.Error())

		return
	}

	projectID, err := r.resolveProjectID(data.ProjectID)
	if err != nil {
		resp.Diagnostics.AddError("Failed to resolve project ID", err.Error())

		return
	}

	createReq := &datalab.CreateDatalabRequest{
		Region:           region,
		ProjectID:        projectID,
		Name:             data.Name.ValueString(),
		Description:      data.Description.ValueString(),
		SparkVersion:     data.SparkVersion.ValueString(),
		PrivateNetworkID: locality.ExpandID(data.PrivateNetworkID.ValueString()),
		HasNotebook:      data.HasNotebook.ValueBool(),
	}

	createReq.Tags = expandTags(ctx, data.Tags, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	if !data.Main.IsNull() && !data.Main.IsUnknown() {
		var mainData sparkMainModel
		resp.Diagnostics.Append(data.Main.As(ctx, &mainData, basetypes.ObjectAsOptions{})...)

		if resp.Diagnostics.HasError() {
			return
		}

		createReq.Main = &datalab.CreateDatalabRequestSparkMain{
			NodeType: mainData.NodeType.ValueString(),
		}
	}

	if !data.Worker.IsNull() && !data.Worker.IsUnknown() {
		var workerData sparkWorkerModel
		resp.Diagnostics.Append(data.Worker.As(ctx, &workerData, basetypes.ObjectAsOptions{})...)

		if resp.Diagnostics.HasError() {
			return
		}

		createReq.Worker = &datalab.CreateDatalabRequestSparkWorker{
			NodeType:  workerData.NodeType.ValueString(),
			NodeCount: uint32(workerData.NodeCount.ValueInt64()),
		}
	}

	if !data.TotalStorage.IsNull() && !data.TotalStorage.IsUnknown() {
		var storageData volumeModel
		resp.Diagnostics.Append(data.TotalStorage.As(ctx, &storageData, basetypes.ObjectAsOptions{})...)

		if resp.Diagnostics.HasError() {
			return
		}

		createReq.TotalStorage = &datalab.Volume{
			Type: datalab.VolumeType(storageData.Type.ValueString()),
			Size: scw.Size(storageData.Size.ValueInt64()),
		}
	}

	dl, err := r.api.CreateDatalab(createReq, scw.WithContext(ctx))
	if err != nil {
		resp.Diagnostics.AddError("Failed to create Datalab", err.Error())

		return
	}

	dl, err = r.api.WaitForDatalab(&datalab.WaitForDatalabRequest{
		Region:    region,
		DatalabID: dl.ID,
	}, scw.WithContext(ctx))
	if err != nil {
		resp.Diagnostics.AddError("Failed waiting for Datalab", err.Error())

		return
	}

	if dl.Status != datalab.DatalabStatusReady {
		resp.Diagnostics.AddError(
			"Datalab not ready",
			fmt.Sprintf("Datalab entered terminal status %q instead of ready", dl.Status),
		)

		return
	}

	state := flattenDatalab(ctx, dl, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *DatalabResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state datalabResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	region, id, err := regional.ParseID(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to parse Datalab ID", err.Error())

		return
	}

	dl, err := r.api.GetDatalab(&datalab.GetDatalabRequest{
		Region:    region,
		DatalabID: id,
	}, scw.WithContext(ctx))
	if err != nil {
		if httperrors.Is404(err) {
			resp.State.RemoveResource(ctx)

			return
		}

		resp.Diagnostics.AddError("Failed to read Datalab", err.Error())

		return
	}

	newState := flattenDatalab(ctx, dl, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &newState)...)
}

func (r *DatalabResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var (
		plan  datalabResourceModel
		state datalabResourceModel
	)

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	region, id, err := regional.ParseID(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to parse Datalab ID", err.Error())

		return
	}

	updateReq := &datalab.UpdateDatalabRequest{
		Region:    region,
		DatalabID: id,
	}

	updateReq.Tags = expandTags(ctx, plan.Tags, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	if !plan.Name.Equal(state.Name) {
		name := plan.Name.ValueString()
		updateReq.Name = &name
	}

	if !plan.Description.Equal(state.Description) {
		desc := plan.Description.ValueString()
		updateReq.Description = &desc
	}

	if !plan.Worker.IsNull() && !plan.Worker.IsUnknown() {
		var planWorker sparkWorkerModel
		resp.Diagnostics.Append(plan.Worker.As(ctx, &planWorker, basetypes.ObjectAsOptions{})...)

		if resp.Diagnostics.HasError() {
			return
		}

		if !state.Worker.IsNull() && !state.Worker.IsUnknown() {
			var stateWorker sparkWorkerModel
			resp.Diagnostics.Append(state.Worker.As(ctx, &stateWorker, basetypes.ObjectAsOptions{})...)

			if resp.Diagnostics.HasError() {
				return
			}

			if !planWorker.NodeCount.Equal(stateWorker.NodeCount) {
				nodeCount := uint32(planWorker.NodeCount.ValueInt64())
				updateReq.NodeCount = &nodeCount
			}
		}
	}

	dl, err := r.api.UpdateDatalab(updateReq, scw.WithContext(ctx))
	if err != nil {
		resp.Diagnostics.AddError("Failed to update Datalab", err.Error())

		return
	}

	dl, err = r.api.WaitForDatalab(&datalab.WaitForDatalabRequest{
		Region:    region,
		DatalabID: dl.ID,
	}, scw.WithContext(ctx))
	if err != nil {
		resp.Diagnostics.AddError("Failed waiting for Datalab update", err.Error())

		return
	}

	if dl.Status != datalab.DatalabStatusReady {
		resp.Diagnostics.AddError(
			"Datalab not ready after update",
			fmt.Sprintf("Datalab entered terminal status %q instead of ready", dl.Status),
		)

		return
	}

	newState := flattenDatalab(ctx, dl, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &newState)...)
}

func (r *DatalabResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state datalabResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	region, id, err := regional.ParseID(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to parse Datalab ID", err.Error())

		return
	}

	_, err = r.api.DeleteDatalab(&datalab.DeleteDatalabRequest{
		Region:    region,
		DatalabID: id,
	}, scw.WithContext(ctx))
	if err != nil {
		if httperrors.Is404(err) {
			return
		}

		resp.Diagnostics.AddError("Failed to delete Datalab", err.Error())

		return
	}

	dl, err := r.api.WaitForDatalab(&datalab.WaitForDatalabRequest{
		Region:    region,
		DatalabID: id,
	}, scw.WithContext(ctx))
	if err != nil {
		if httperrors.Is404(err) {
			return
		}

		resp.Diagnostics.AddError("Failed waiting for Datalab deletion", err.Error())

		return
	}

	if dl.Status != datalab.DatalabStatusDeleted {
		resp.Diagnostics.AddError(
			"Datalab deletion failed",
			fmt.Sprintf("Datalab entered terminal status %q instead of deleted", dl.Status),
		)
	}
}

func (r *DatalabResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	region, id, err := regional.ParseID(req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Failed to parse import ID", "Expected format: {region}/{id}. "+err.Error())

		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), regional.NewIDString(region, id))...)
}

func (r *DatalabResource) resolveRegion(regionAttr types.String) (scw.Region, error) {
	if !regionAttr.IsNull() && !regionAttr.IsUnknown() && regionAttr.ValueString() != "" {
		return scw.ParseRegion(regionAttr.ValueString())
	}

	region, exists := r.meta.ScwClient().GetDefaultRegion()
	if exists {
		return region, nil
	}

	return "", errors.New("region is required; set it on the resource or configure a default region on the provider")
}

func (r *DatalabResource) resolveProjectID(projectIDAttr types.String) (string, error) {
	if !projectIDAttr.IsNull() && !projectIDAttr.IsUnknown() && projectIDAttr.ValueString() != "" {
		return projectIDAttr.ValueString(), nil
	}

	projectID, exists := r.meta.ScwClient().GetDefaultProjectID()
	if exists {
		return projectID, nil
	}

	return "", errors.New("project_id is required; set it on the resource or configure a default project on the provider")
}

func flattenDatalab(ctx context.Context, dl *datalab.Datalab, diags *diag.Diagnostics) datalabResourceModel {
	model := datalabResourceModel{
		ID:               types.StringValue(regional.NewIDString(dl.Region, dl.ID)),
		Name:             types.StringValue(dl.Name),
		ProjectID:        types.StringValue(dl.ProjectID),
		Region:           types.StringValue(dl.Region.String()),
		Description:      types.StringValue(dl.Description),
		SparkVersion:     types.StringValue(dl.SparkVersion),
		PrivateNetworkID: types.StringValue(regional.NewIDString(dl.Region, dl.PrivateNetworkID)),
		HasNotebook:      types.BoolValue(dl.HasNotebook),
		Status:           types.StringValue(string(dl.Status)),
	}

	if dl.Tags != nil {
		tagValues := make([]attr.Value, len(dl.Tags))
		for i, t := range dl.Tags {
			tagValues[i] = types.StringValue(t)
		}

		tagList, d := types.ListValue(types.StringType, tagValues)
		diags.Append(d...)

		model.Tags = tagList
	} else {
		model.Tags = types.ListNull(types.StringType)
	}

	if dl.CreatedAt != nil {
		model.CreatedAt = types.StringValue(dl.CreatedAt.Format("2006-01-02T15:04:05Z07:00"))
	} else {
		model.CreatedAt = types.StringNull()
	}

	if dl.UpdatedAt != nil {
		model.UpdatedAt = types.StringValue(dl.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"))
	} else {
		model.UpdatedAt = types.StringNull()
	}

	if dl.NotebookURL != nil {
		model.NotebookURL = types.StringValue(*dl.NotebookURL)
	} else {
		model.NotebookURL = types.StringNull()
	}

	if dl.NotebookMasterURL != nil {
		model.NotebookMasterURL = types.StringValue(*dl.NotebookMasterURL)
	} else {
		model.NotebookMasterURL = types.StringNull()
	}

	model.Main = flattenSparkMain(dl.Main, diags)
	model.Worker = flattenSparkWorker(dl.Worker, diags)
	model.TotalStorage = flattenVolume(dl.TotalStorage, diags)

	return model
}

func flattenSparkMain(main *datalab.DatalabSparkMain, diags *diag.Diagnostics) types.Object {
	if main == nil {
		return types.ObjectNull(sparkMainAttrTypes())
	}

	rootVolume := flattenVolume(main.RootVolume, diags)

	attrValues := map[string]attr.Value{
		"node_type":        types.StringValue(main.NodeType),
		"spark_ui_url":     types.StringValue(main.SparkUIURL),
		"spark_master_url": types.StringValue(main.SparkMasterURL),
		"root_volume":      rootVolume,
	}

	obj, d := types.ObjectValue(sparkMainAttrTypes(), attrValues)
	diags.Append(d...)

	return obj
}

func flattenSparkWorker(worker *datalab.DatalabSparkWorker, diags *diag.Diagnostics) types.Object {
	if worker == nil {
		return types.ObjectNull(sparkWorkerAttrTypes())
	}

	rootVolume := flattenVolume(worker.RootVolume, diags)

	attrValues := map[string]attr.Value{
		"node_type":   types.StringValue(worker.NodeType),
		"node_count":  types.Int64Value(int64(worker.NodeCount)),
		"root_volume": rootVolume,
	}

	obj, d := types.ObjectValue(sparkWorkerAttrTypes(), attrValues)
	diags.Append(d...)

	return obj
}

func flattenVolume(vol *datalab.Volume, diags *diag.Diagnostics) types.Object {
	if vol == nil {
		return types.ObjectNull(volumeAttrTypes())
	}

	attrValues := map[string]attr.Value{
		"type": types.StringValue(string(vol.Type)),
		"size": types.Int64Value(int64(vol.Size)),
	}

	obj, d := types.ObjectValue(volumeAttrTypes(), attrValues)
	diags.Append(d...)

	return obj
}

func expandTags(ctx context.Context, tags types.List, diags *diag.Diagnostics) []string {
	if tags.IsNull() || tags.IsUnknown() {
		return nil
	}

	var result []string
	diags.Append(tags.ElementsAs(ctx, &result, false)...)

	return result
}
