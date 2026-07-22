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
	_ resource.Resource                = (*AnnotationsKeyResource)(nil)
	_ resource.ResourceWithConfigure   = (*AnnotationsKeyResource)(nil)
	_ resource.ResourceWithImportState = (*AnnotationsKeyResource)(nil)
)

func NewAnnotationsKeyResource() resource.Resource {
	return &AnnotationsKeyResource{}
}

type AnnotationsKeyResource struct {
	annotationsAPI *annotations.API
	meta           *meta.Meta
}

type annotationsKeyResourceModel struct {
	ID             types.String `tfsdk:"id"`
	Name           types.String `tfsdk:"name"`
	Description    types.String `tfsdk:"description"`
	OrganizationID types.String `tfsdk:"organization_id"`
}

func (r *AnnotationsKeyResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_annotations_key"
}

//go:embed descriptions/key_resource.md
var annotationsKeyResourceDescription string

func (r *AnnotationsKeyResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: annotationsKeyResourceDescription,
		Description:         annotationsKeyResourceDescription,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The ID of the annotation key resource.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Name of the annotation key.",
			},
			"description": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Description of the annotation key.",
			},
			"organization_id": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "ID of the organization. If not set, the organization ID is derived from the provider configuration.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *AnnotationsKeyResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *AnnotationsKeyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data annotationsKeyResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var orgID string
	if !data.OrganizationID.IsNull() && !data.OrganizationID.IsUnknown() {
		orgID = locality.ExpandID(data.OrganizationID.ValueString())
	} else {
		defaultOrgID, exists := r.meta.ScwClient().GetDefaultOrganizationID()
		if !exists {
			resp.Diagnostics.AddError(
				"Missing organization ID",
				"The organization_id attribute is required to create an annotation key. Please provide it explicitly or configure a default organization in the provider.",
			)

			return
		}

		orgID = defaultOrgID
	}

	createReq := &annotations.CreateKeyRequest{
		OrganizationID: orgID,
		Name:           data.Name.ValueString(),
	}

	if !data.Description.IsNull() && !data.Description.IsUnknown() {
		createReq.Description = data.Description.ValueString()
	}

	key, err := r.annotationsAPI.CreateKey(createReq, scw.WithContext(ctx))
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to create annotation key",
			fmt.Sprintf("Failed to create annotation key: %s", err),
		)

		return
	}

	data.ID = types.StringValue(key.ID)
	data.OrganizationID = types.StringValue(orgID)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *AnnotationsKeyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state annotationsKeyResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	keyID := locality.ExpandID(state.ID.ValueString())

	key, err := r.annotationsAPI.GetKey(&annotations.GetKeyRequest{
		KeyID: keyID,
	}, scw.WithContext(ctx))
	if err != nil {
		if httperrors.Is404(err) {
			resp.State.RemoveResource(ctx)

			return
		}

		resp.Diagnostics.AddError(
			"Failed to read annotation key",
			fmt.Sprintf("Failed to read annotation key: %s", err),
		)

		return
	}

	state.ID = types.StringValue(key.ID)
	state.Name = types.StringValue(key.Name)

	if key.Description != "" {
		state.Description = types.StringValue(key.Description)
	} else {
		state.Description = types.StringNull()
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *AnnotationsKeyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data annotationsKeyResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	keyID := locality.ExpandID(data.ID.ValueString())

	updateReq := &annotations.UpdateKeyRequest{
		KeyID: keyID,
	}

	if !data.Name.IsNull() && !data.Name.IsUnknown() {
		updateReq.Name = data.Name.ValueStringPointer()
	}

	if !data.Description.IsNull() && !data.Description.IsUnknown() {
		updateReq.Description = data.Description.ValueStringPointer()
	}

	key, err := r.annotationsAPI.UpdateKey(updateReq, scw.WithContext(ctx))
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to update annotation key",
			fmt.Sprintf("Failed to update annotation key: %s", err),
		)

		return
	}

	data.ID = types.StringValue(key.ID)
	data.Name = types.StringValue(key.Name)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *AnnotationsKeyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state annotationsKeyResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	keyID := locality.ExpandID(state.ID.ValueString())

	err := r.annotationsAPI.DeleteKey(&annotations.DeleteKeyRequest{
		KeyID: keyID,
	}, scw.WithContext(ctx))
	if err != nil {
		if httperrors.Is404(err) {
			return
		}

		resp.Diagnostics.AddError(
			"Failed to delete annotation key",
			fmt.Sprintf("Failed to delete annotation key: %s", err),
		)

		return
	}
}

func (r *AnnotationsKeyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	keyID := locality.ExpandID(req.ID)

	key, err := r.annotationsAPI.GetKey(&annotations.GetKeyRequest{
		KeyID: keyID,
	}, scw.WithContext(ctx))
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to import annotation key",
			fmt.Sprintf("Failed to import annotation key: %s", err),
		)

		return
	}

	orgID, exists := r.meta.ScwClient().GetDefaultOrganizationID()
	if !exists {
		resp.Diagnostics.AddError(
			"Missing default organization ID",
			"Cannot import annotation key without a default organization ID",
		)

		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), key.ID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("name"), key.Name)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("organization_id"), orgID)...)

	if key.Description != "" {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("description"), key.Description)...)
	} else {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("description"), types.StringNull())...)
	}
}
