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

func getBindingByID(ctx context.Context, api *annotations.API, bindingID string) (*annotations.Binding, error) {
	resp, err := api.ListBindings(&annotations.ListBindingsRequest{}, scw.WithContext(ctx))
	if err != nil {
		return nil, err
	}

	for _, binding := range resp.Bindings {
		if binding.ID == bindingID {
			return binding, nil
		}
	}

	return nil, &scw.ResourceNotFoundError{Resource: "binding", ResourceID: bindingID}
}

var (
	_ resource.Resource                = (*AnnotationsBindingResource)(nil)
	_ resource.ResourceWithConfigure   = (*AnnotationsBindingResource)(nil)
	_ resource.ResourceWithImportState = (*AnnotationsBindingResource)(nil)
)

func NewAnnotationsBindingResource() resource.Resource {
	return &AnnotationsBindingResource{}
}

type AnnotationsBindingResource struct {
	annotationsAPI *annotations.API
	meta           *meta.Meta
}

type annotationsBindingResourceModel struct {
	ID      types.String `tfsdk:"id"`
	Srn     types.String `tfsdk:"srn"`
	ValueID types.String `tfsdk:"value_id"`
	KeyID   types.String `tfsdk:"key_id"`
}

func (r *AnnotationsBindingResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_annotations_binding"
}

//go:embed descriptions/binding_resource.md
var annotationsBindingResourceDescription string

func (r *AnnotationsBindingResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: annotationsBindingResourceDescription,
		Description:         annotationsBindingResourceDescription,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The ID of the annotation binding resource.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"srn": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Scaleway Resource Number to associate.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"value_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "ID of the value to associate.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"key_id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "ID of the key associated to the binding.",
			},
		},
	}
}

func (r *AnnotationsBindingResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *AnnotationsBindingResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data annotationsBindingResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	createReq := &annotations.CreateBindingRequest{
		Srn:     data.Srn.ValueString(),
		ValueID: locality.ExpandID(data.ValueID.ValueString()),
	}

	binding, err := r.annotationsAPI.CreateBinding(createReq, scw.WithContext(ctx))
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to create annotation binding",
			fmt.Sprintf("Failed to create annotation binding: %s", err),
		)

		return
	}

	data.ID = types.StringValue(binding.ID)
	data.Srn = types.StringValue(binding.Srn)
	data.ValueID = types.StringValue(binding.Value.ID)
	data.KeyID = types.StringValue(binding.Key.ID)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *AnnotationsBindingResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state annotationsBindingResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	bindingID := locality.ExpandID(state.ID.ValueString())

	binding, err := getBindingByID(ctx, r.annotationsAPI, bindingID)
	if err != nil {
		if httperrors.Is404(err) {
			resp.State.RemoveResource(ctx)

			return
		}

		resp.Diagnostics.AddError(
			"Failed to read annotation binding",
			fmt.Sprintf("Failed to read annotation binding: %s", err),
		)

		return
	}

	state.ID = types.StringValue(binding.ID)
	state.Srn = types.StringValue(binding.Srn)
	state.ValueID = types.StringValue(binding.Value.ID)
	state.KeyID = types.StringValue(binding.Key.ID)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *AnnotationsBindingResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError(
		"Update not supported",
		"Annotation bindings cannot be updated. Changes to srn or value_id require resource replacement.",
	)
}

func (r *AnnotationsBindingResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state annotationsBindingResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	bindingID := locality.ExpandID(state.ID.ValueString())

	err := r.annotationsAPI.DeleteBinding(&annotations.DeleteBindingRequest{
		BindingID: bindingID,
	}, scw.WithContext(ctx))
	if err != nil {
		if httperrors.Is404(err) {
			return
		}

		resp.Diagnostics.AddError(
			"Failed to delete annotation binding",
			fmt.Sprintf("Failed to delete annotation binding: %s", err),
		)

		return
	}
}

func (r *AnnotationsBindingResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	bindingID := locality.ExpandID(req.ID)

	binding, err := getBindingByID(ctx, r.annotationsAPI, bindingID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to import annotation binding",
			fmt.Sprintf("Failed to import annotation binding: %s", err),
		)

		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), binding.ID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("srn"), binding.Srn)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("value_id"), binding.Value.ID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("key_id"), binding.Key.ID)...)
}
