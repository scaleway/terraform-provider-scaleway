package annotations

import (
	"context"
	_ "embed"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	annotations "github.com/scaleway/scaleway-sdk-go/api/annotations/v1"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
)

var (
	_ datasource.DataSource              = (*AnnotationsBindingDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*AnnotationsBindingDataSource)(nil)
)

func NewAnnotationsBindingDataSource() datasource.DataSource {
	return &AnnotationsBindingDataSource{}
}

type AnnotationsBindingDataSource struct {
	annotationsAPI *annotations.API
	meta           *meta.Meta
}

type annotationsBindingDataSourceModel struct {
	ID      types.String `tfsdk:"id"`
	Srn     types.String `tfsdk:"srn"`
	ValueID types.String `tfsdk:"value_id"`
	KeyID   types.String `tfsdk:"key_id"`
}

func (d *AnnotationsBindingDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_annotations_binding"
}

//go:embed descriptions/binding_data_source.md
var annotationsBindingDataSourceDescription string

func (d *AnnotationsBindingDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: annotationsBindingDataSourceDescription,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The ID of the annotation binding to retrieve.",
			},
			"srn": schema.StringAttribute{
				MarkdownDescription: "Scaleway Resource Number associated to the binding.",
				Computed:            true,
			},
			"value_id": schema.StringAttribute{
				MarkdownDescription: "ID of the value associated to the binding.",
				Computed:            true,
			},
			"key_id": schema.StringAttribute{
				MarkdownDescription: "ID of the key associated to the binding.",
				Computed:            true,
			},
		},
	}
}

func (d *AnnotationsBindingDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	m, ok := req.ProviderData.(*meta.Meta)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *meta.Meta, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	d.meta = m
	d.annotationsAPI = annotations.NewAPI(d.meta.ScwClient())
}

func (d *AnnotationsBindingDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state annotationsBindingDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	bindingID := locality.ExpandID(state.ID.ValueString())
	if bindingID == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("id"),
			"Annotation Binding ID is required",
			"The id attribute must be set",
		)

		return
	}

	binding, err := getBindingByID(ctx, d.annotationsAPI, bindingID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to get annotation binding",
			fmt.Sprintf("Could not retrieve annotation binding %s: %v", bindingID, err),
		)

		return
	}

	state.ID = types.StringValue(binding.ID)
	state.Srn = types.StringValue(binding.Srn)
	state.ValueID = types.StringValue(binding.Value.ID)
	state.KeyID = types.StringValue(binding.Key.ID)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
