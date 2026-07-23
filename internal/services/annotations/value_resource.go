package annotations

import (
	"context"
	_ "embed"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	annotations "github.com/scaleway/scaleway-sdk-go/api/annotations/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
)

var (
	_ resource.Resource                = (*AnnotationsValueResource)(nil)
	_ resource.ResourceWithConfigure   = (*AnnotationsValueResource)(nil)
	_ resource.ResourceWithImportState = (*AnnotationsValueResource)(nil)
)

func NewAnnotationsValueResource() resource.Resource {
	return &AnnotationsValueResource{}
}

type AnnotationsValueResource struct {
	annotationsAPI *annotations.API
	meta           *meta.Meta
}

type annotationsValueResourceModel struct {
	ID          types.String `tfsdk:"id"`
	KeyID       types.String `tfsdk:"key_id"`
	Name        types.String `tfsdk:"name"`
	ValueID     types.String `tfsdk:"value_id"`
	Description types.String `tfsdk:"description"`
}

func (r *AnnotationsValueResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_annotations_value"
}

//go:embed descriptions/value_resource.md
var annotationsValueResourceDescription string

func (r *AnnotationsValueResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: annotationsValueResourceDescription,
		Description:         annotationsValueResourceDescription,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The ID of the annotation value resource.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"key_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "ID of the key the value is associated to.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Name of the annotation value.",
			},
			"description": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Description of the annotation value.",
			},
			"value_id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The ID of the annotation value.",
			},
		},
	}
}

func (r *AnnotationsValueResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
	r.annotationsAPI = annotations.NewAPI(r.meta.ScwClient())
}

func (r *AnnotationsValueResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data annotationsValueResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	keyID := locality.ExpandID(data.KeyID.ValueString())

	createReq := &annotations.CreateValueRequest{
		KeyID: keyID,
		Name:  data.Name.ValueString(),
	}

	if !data.Description.IsNull() && !data.Description.IsUnknown() {
		createReq.Description = data.Description.ValueString()
	}

	value, err := r.annotationsAPI.CreateValue(createReq, scw.WithContext(ctx))
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to create annotation value",
			fmt.Sprintf("Failed to create annotation value: %s", err),
		)

		return
	}

	data.ID = types.StringValue(value.ID)
	data.ValueID = types.StringValue(value.ID)
	data.KeyID = types.StringValue(value.KeyID)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *AnnotationsValueResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state annotationsValueResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	valueID := locality.ExpandID(state.ID.ValueString())

	value, err := r.annotationsAPI.GetValue(&annotations.GetValueRequest{
		ValueID: valueID,
	}, scw.WithContext(ctx))
	if err != nil {
		if httperrors.Is404(err) {
			resp.State.RemoveResource(ctx)

			return
		}

		resp.Diagnostics.AddError(
			"Failed to read annotation value",
			fmt.Sprintf("Failed to read annotation value: %s", err),
		)

		return
	}

	state.ID = types.StringValue(value.ID)
	state.ValueID = types.StringValue(value.ID)
	state.KeyID = types.StringValue(value.KeyID)
	state.Name = types.StringValue(value.Name)

	if value.Description != "" {
		state.Description = types.StringValue(value.Description)
	} else {
		state.Description = types.StringNull()
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *AnnotationsValueResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data annotationsValueResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	valueID := locality.ExpandID(data.ID.ValueString())

	updateReq := &annotations.UpdateValueRequest{
		ValueID: valueID,
	}

	if !data.Name.IsNull() && !data.Name.IsUnknown() {
		updateReq.Name = data.Name.ValueStringPointer()
	}

	if !data.Description.IsNull() && !data.Description.IsUnknown() {
		updateReq.Description = data.Description.ValueStringPointer()
	}

	value, err := r.annotationsAPI.UpdateValue(updateReq, scw.WithContext(ctx))
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to update annotation value",
			fmt.Sprintf("Failed to update annotation value: %s", err),
		)

		return
	}

	data.ID = types.StringValue(value.ID)
	data.ValueID = types.StringValue(value.ID)
	data.KeyID = types.StringValue(value.KeyID)
	data.Name = types.StringValue(value.Name)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *AnnotationsValueResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state annotationsValueResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	valueID := locality.ExpandID(state.ID.ValueString())

	err := r.annotationsAPI.DeleteValue(&annotations.DeleteValueRequest{
		ValueID: valueID,
	}, scw.WithContext(ctx))
	if err != nil {
		if httperrors.Is404(err) {
			return
		}

		resp.Diagnostics.AddError(
			"Failed to delete annotation value",
			fmt.Sprintf("Failed to delete annotation value: %s", err),
		)

		return
	}
}

func (r *AnnotationsValueResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	valueID := locality.ExpandID(req.ID)

	value, err := r.annotationsAPI.GetValue(&annotations.GetValueRequest{
		ValueID: valueID,
	}, scw.WithContext(ctx))
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to import annotation value",
			fmt.Sprintf("Failed to import annotation value: %s", err),
		)

		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), value.ID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("value_id"), value.ID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("key_id"), value.KeyID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("name"), value.Name)...)

	if value.Description != "" {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("description"), value.Description)...)
	} else {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("description"), types.StringNull())...)
	}
}
