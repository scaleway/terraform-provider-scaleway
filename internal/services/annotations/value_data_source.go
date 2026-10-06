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
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
)

var (
	_ datasource.DataSource              = (*AnnotationsValueDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*AnnotationsValueDataSource)(nil)
)

func NewAnnotationsValueDataSource() datasource.DataSource {
	return &AnnotationsValueDataSource{}
}

type AnnotationsValueDataSource struct {
	annotationsAPI *annotations.API
	meta           *meta.Meta
}

type annotationsValueDataSourceModel struct {
	ID          types.String `tfsdk:"id"`
	ValueID     types.String `tfsdk:"value_id"`
	KeyID       types.String `tfsdk:"key_id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
}

func (d *AnnotationsValueDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_annotations_value"
}

//go:embed descriptions/value_data_source.md
var annotationsValueDataSourceDescription string

func (d *AnnotationsValueDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: annotationsValueDataSourceDescription,
		Attributes: map[string]schema.Attribute{
			"value_id": schema.StringAttribute{
				MarkdownDescription: "The ID of the annotation value to retrieve.",
				Required:            true,
			},
			"key_id": schema.StringAttribute{
				MarkdownDescription: "ID of the key the value is associated to.",
				Computed:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Name of the annotation value.",
				Computed:            true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "Description of the annotation value.",
				Computed:            true,
			},
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The ID of the annotation value.",
			},
		},
	}
}

func (d *AnnotationsValueDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *AnnotationsValueDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state annotationsValueDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	valueID := locality.ExpandID(state.ValueID.ValueString())
	if valueID == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("value_id"),
			"Annotation Value ID is required",
			"The value_id attribute must be set",
		)

		return
	}

	value, err := d.annotationsAPI.GetValue(&annotations.GetValueRequest{
		ValueID: valueID,
	}, scw.WithContext(ctx))
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to get annotation value",
			fmt.Sprintf("Could not retrieve annotation value %s: %v", valueID, err),
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
