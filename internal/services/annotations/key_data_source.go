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
	_ datasource.DataSource              = (*AnnotationsKeyDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*AnnotationsKeyDataSource)(nil)
)

func NewAnnotationsKeyDataSource() datasource.DataSource {
	return &AnnotationsKeyDataSource{}
}

type AnnotationsKeyDataSource struct {
	annotationsAPI *annotations.API
	meta           *meta.Meta
}

type annotationsKeyDataSourceModel struct {
	ID          types.String `tfsdk:"id"`
	KeyID       types.String `tfsdk:"key_id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
}

func (d *AnnotationsKeyDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_annotations_key"
}

//go:embed descriptions/key_data_source.md
var annotationsKeyDataSourceDescription string

func (d *AnnotationsKeyDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: annotationsKeyDataSourceDescription,
		Attributes: map[string]schema.Attribute{
			"key_id": schema.StringAttribute{
				MarkdownDescription: "The ID of the annotation key to retrieve.",
				Required:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Name of the annotation key.",
				Computed:            true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "Description of the annotation key.",
				Computed:            true,
			},
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The ID of the annotation key.",
			},
		},
	}
}

func (d *AnnotationsKeyDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *AnnotationsKeyDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state annotationsKeyDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	keyID := locality.ExpandID(state.KeyID.ValueString())
	if keyID == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("key_id"),
			"Annotation Key ID is required",
			"The key_id attribute must be set",
		)

		return
	}

	key, err := d.annotationsAPI.GetKey(&annotations.GetKeyRequest{
		KeyID: keyID,
	}, scw.WithContext(ctx))
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to get annotation key",
			fmt.Sprintf("Could not retrieve annotation key %s: %v", keyID, err),
		)

		return
	}

	state.ID = types.StringValue(key.ID)
	state.KeyID = types.StringValue(key.ID)
	state.Name = types.StringValue(key.Name)

	if key.Description != "" {
		state.Description = types.StringValue(key.Description)
	} else {
		state.Description = types.StringNull()
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
