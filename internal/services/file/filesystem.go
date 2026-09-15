package file

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	file "github.com/scaleway/scaleway-sdk-go/api/file/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
	scwtypes "github.com/scaleway/terraform-provider-scaleway/v2/internal/types"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/verify"
)

var (
	_ resource.Resource                = (*FileSystemResource)(nil)
	_ resource.ResourceWithConfigure   = (*FileSystemResource)(nil)
	_ resource.ResourceWithImportState = (*FileSystemResource)(nil)
)

func NewFileSystemResource() resource.Resource {
	return &FileSystemResource{}
}

type FileSystemResource struct {
	api  *file.API
	meta *meta.Meta
}

type fileSystemResourceModel struct {
	Tags                types.List   `tfsdk:"tags"`
	ID                  types.String `tfsdk:"id"`
	Name                types.String `tfsdk:"name"`
	ProjectID           types.String `tfsdk:"project_id"`
	OrganizationID      types.String `tfsdk:"organization_id"`
	Region              types.String `tfsdk:"region"`
	Status              types.String `tfsdk:"status"`
	CreatedAt           types.String `tfsdk:"created_at"`
	UpdatedAt           types.String `tfsdk:"updated_at"`
	SRN                 types.String `tfsdk:"srn"`
	SizeInGB            types.Int64  `tfsdk:"size_in_gb"`
	NumberOfAttachments types.Int64  `tfsdk:"number_of_attachments"`
}

func (r *FileSystemResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_file_filesystem"
}

func (r *FileSystemResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Scaleway File FileSystem.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The ID of the filesystem, in the `{region}/{id}` format.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Computed:    true,
				Optional:    true,
				Description: "The name of the filesystem",
			},
			"size_in_gb": schema.Int64Attribute{
				Required: true,
				Validators: []validator.Int64{
					int64validator.Between(25, 50000),
				},
				Description: "The filesystem size in GB. Minimum 25GB, maximum 50TB",
			},
			"tags": schema.ListAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Description: "The list of tags assigned to the filesystem",
			},
			"project_id": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "The project ID the filesystem belongs to. Defaults to the provider's project ID.",
				Validators: []validator.String{
					verify.IsStringUUID(),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"organization_id": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "ID of the organization. If not set, the organization ID is derived from the provider configuration.",
				Validators: []validator.String{
					verify.IsStringUUID(),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"region": regional.SchemaAttributeComputed(),
			"status": schema.StringAttribute{
				Computed:    true,
				Description: "The Current status of the filesystem (e.g. creating, available, ...)",
			},
			"number_of_attachments": schema.Int64Attribute{
				Computed:    true,
				Description: "The current number of attachments (mounts) that the filesystem has",
			},
			"created_at": schema.StringAttribute{
				Computed:    true,
				Description: "The creation date of the filesystem",
			},
			"updated_at": schema.StringAttribute{
				Computed:    true,
				Description: "The last update date of the properties of the filesystem",
			},
			"srn": schema.StringAttribute{
				Computed:    true,
				Description: "The Scaleway Resource Name (SRN) of the filesystem",
			},
		},
	}
}

func (r *FileSystemResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
	r.api = file.NewAPI(r.meta.ScwClient())
}

func (r *FileSystemResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data fileSystemResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	region, err := meta.ExtractFrameworkRegion(data.Region, r.meta.ScwClient())
	if err != nil {
		resp.Diagnostics.AddError("Failed to resolve region", err.Error())

		return
	}

	projectID, err := meta.ExtractFrameworkProjectID(data.ProjectID, r.meta.ScwClient())
	if err != nil {
		resp.Diagnostics.AddError("Failed to resolve project ID", err.Error())

		return
	}

	createReq := &file.CreateFileSystemRequest{
		Region:    region,
		ProjectID: projectID,
		Name:      scwtypes.ExpandOrGenerateString(data.Name.ValueString(), "file"),
		Size:      uint64(data.SizeInGB.ValueInt64() * int64(scw.GB)), // Translate from GB to Bytes
	}

	createReq.Tags = scwtypes.ExpandStringList(ctx, data.Tags, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	fs, err := r.api.CreateFileSystem(createReq, scw.WithContext(ctx))
	if err != nil {
		resp.Diagnostics.AddError("Failed to create File FileSystem", err.Error())

		return
	}

	fs, err = waitForFileSystem(ctx, r.api, region, fs.ID, defaultFileSystemTimeout)
	if err != nil {
		resp.Diagnostics.AddError("Failed to wait for File FileSystem during Create", err.Error())

		return
	}

	state := flattenFilesystem(ctx, fs, req, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *FileSystemResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state fileSystemResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	region, id, err := regional.ParseID(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to parse File FileSystem ID", err.Error())

		return
	}

	fs, err := waitForFileSystem(ctx, r.api, region, id, defaultFileSystemTimeout)
	if err != nil {
		if httperrors.Is404(err) {
			resp.State.RemoveResource(ctx)

			return
		}

		resp.Diagnostics.AddError("Failed to wait for File FileSystem during Read", err.Error())

		return
	}

	newState := flattenFilesystem(ctx, fs, req, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &newState)...)
}

func (r *FileSystemResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var (
		plan  fileSystemResourceModel
		state fileSystemResourceModel
	)

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	region, id, err := regional.ParseID(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to parse File FileSystem ID", err.Error())

		return
	}

	_, err = waitForFileSystem(ctx, r.api, region, id, defaultFileSystemTimeout)
	if err != nil {
		if httperrors.Is404(err) {
			resp.State.RemoveResource(ctx)

			return
		}

		resp.Diagnostics.AddError("Failed to wait for File FileSystem during Update", err.Error())

		return
	}

	updateReq := &file.UpdateFileSystemRequest{
		Region:       region,
		FilesystemID: id,
	}
	hasChanges := false

	if !plan.Name.Equal(state.Name) {
		updateReq.Name = new(plan.Name.ValueString())
		hasChanges = true
	}

	if !plan.Tags.Equal(state.Tags) {
		updateReq.Tags = new(scwtypes.ExpandUpdatedStringList(ctx, plan.Tags, &resp.Diagnostics))
		if resp.Diagnostics.HasError() {
			return
		}

		hasChanges = true
	}

	if !plan.SizeInGB.Equal(state.SizeInGB) {
		sizeInGB := plan.SizeInGB.ValueInt64()
		updateReq.Size = new(uint64(sizeInGB) * uint64(scw.GB))
		hasChanges = true
	}

	if !hasChanges {
		return
	}

	_, err = r.api.UpdateFileSystem(updateReq, scw.WithContext(ctx))
	if err != nil {
		resp.Diagnostics.AddError("Failed to update File FileSystem", err.Error())

		return
	}

	fs, err := waitForFileSystem(ctx, r.api, region, id, defaultFileSystemTimeout)
	if err != nil {
		// FIXME: Why this case?
		if httperrors.Is404(err) {
			resp.State.RemoveResource(ctx)

			return
		}

		resp.Diagnostics.AddError("Failed to wait for File FileSystem during Update", err.Error())

		return
	}

	newState := flattenFilesystem(ctx, fs, req, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &newState)...)
}

func (r *FileSystemResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state fileSystemResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	region, id, err := regional.ParseID(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to parse File FileSystem ID", err.Error())

		return
	}

	_, err = waitForFileSystem(ctx, r.api, region, id, defaultFileSystemTimeout)
	if err != nil {
		if httperrors.Is404(err) {
			return
		}

		resp.Diagnostics.AddError("Failed to wait for File FileSystem during Delete", err.Error())

		return
	}

	err = r.api.DeleteFileSystem(&file.DeleteFileSystemRequest{
		Region:       region,
		FilesystemID: id,
	}, scw.WithContext(ctx))
	if err != nil {
		if httperrors.Is404(err) {
			return
		}

		resp.Diagnostics.AddError("Failed to delete File FileSystem", err.Error())

		return
	}

	_, err = waitForFileSystem(ctx, r.api, region, id, defaultFileSystemTimeout)
	if err != nil && !httperrors.Is404(err) {
		resp.Diagnostics.AddError("Failed to wait for File FileSystem during Delete", err.Error())
	}
}

func (r *FileSystemResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	region, id, err := regional.ParseID(req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Failed to parse import ID", "Expected format: {region}/{id}. "+err.Error())

		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), regional.NewIDString(region, id))...)
}

func flattenFilesystem(ctx context.Context, fs *file.FileSystem, reference any, diags *diag.Diagnostics) any {
	model := fileSystemResourceModel{
		ID:                  types.StringValue(regional.NewIDString(fs.Region, fs.ID)),
		Name:                types.StringValue(fs.Name),
		SizeInGB:            types.Int64Value(int64(fs.Size / scw.GB)), // Translate from Bytes to GB
		ProjectID:           types.StringValue(fs.ProjectID),
		OrganizationID:      types.StringValue(fs.OrganizationID),
		Region:              types.StringValue(fs.Region.String()),
		Status:              types.StringValue(fs.Status.String()),
		NumberOfAttachments: types.Int64Value(int64(fs.NumberOfAttachments)),
		CreatedAt:           types.StringValue(fs.CreatedAt.Format(time.RFC3339)),
		UpdatedAt:           types.StringValue(fs.UpdatedAt.Format(time.RFC3339)),
		SRN:                 types.StringValue(fs.Srn),
	}

	tagsList, d := scwtypes.FlattenStringList(ctx, "tags", fs.Tags, reference)
	diags.Append(d...)

	model.Tags = tagsList

	return model
}
