package datalab

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	datalab "github.com/scaleway/scaleway-sdk-go/api/datalab/v1beta1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
	scwtypes "github.com/scaleway/terraform-provider-scaleway/v2/internal/types"
)

var (
	_ datasource.DataSource              = (*DatalabsDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*DatalabsDataSource)(nil)
)

func NewDatalabsDataSource() datasource.DataSource {
	return &DatalabsDataSource{}
}

type DatalabsDataSource struct {
	api  *datalab.API
	meta *meta.Meta
}

type datalabsDataSourceModel struct {
	ProjectID      types.String `tfsdk:"project_id"`
	OrganizationID types.String `tfsdk:"organization_id"`
	Region         types.String `tfsdk:"region"`
	Name           types.String `tfsdk:"name"`
	Tags           types.List   `tfsdk:"tags"`
	Datalabs       types.List   `tfsdk:"datalabs"`
}

func datalabsItemAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"id":            types.StringType,
		"name":          types.StringType,
		"description":   types.StringType,
		"status":        types.StringType,
		"tags":          types.ListType{ElemType: types.StringType},
		"region":        types.StringType,
		"project_id":    types.StringType,
		"spark_version": types.StringType,
		"has_notebook":  types.BoolType,
		"created_at":    types.StringType,
		"updated_at":    types.StringType,
	}
}

func (d *DatalabsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_datalabs"
}

func (d *DatalabsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Lists Scaleway Datalab instances.",
		Attributes: map[string]schema.Attribute{
			"project_id": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "The project ID to filter Datalabs by.",
			},
			"organization_id": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "The organization ID to filter Datalabs by.",
			},
			"region": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "The region to list Datalabs from.",
			},
			"name": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "The name to filter Datalabs by.",
			},
			"tags": schema.ListAttribute{
				Optional:            true,
				ElementType:         types.StringType,
				MarkdownDescription: "The tags to filter Datalabs by.",
			},
			"datalabs": schema.ListNestedAttribute{
				Computed:            true,
				MarkdownDescription: "The list of Datalab instances.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The unique identifier of the Datalab instance.",
						},
						"name": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The name of the Datalab instance.",
						},
						"description": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The description of the Datalab instance.",
						},
						"status": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The current status of the Datalab instance.",
						},
						"tags": schema.ListAttribute{
							Computed:            true,
							ElementType:         types.StringType,
							MarkdownDescription: "Tags associated with the Datalab instance.",
						},
						"region": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The region of the Datalab instance.",
						},
						"project_id": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The project ID of the Datalab instance.",
						},
						"spark_version": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The Spark version of the Datalab instance.",
						},
						"has_notebook": schema.BoolAttribute{
							Computed:            true,
							MarkdownDescription: "Whether a JupyterLab notebook is associated with the Datalab.",
						},
						"created_at": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The creation timestamp of the Datalab instance.",
						},
						"updated_at": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The last update timestamp of the Datalab instance.",
						},
					},
				},
			},
		},
	}
}

func (d *DatalabsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
	d.api = datalab.NewAPI(d.meta.ScwClient())
}

func (d *DatalabsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config datalabsDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)

	if resp.Diagnostics.HasError() {
		return
	}

	region, err := resolveRegion(config.Region, d.meta.ScwClient())
	if err != nil {
		resp.Diagnostics.AddError("Failed to resolve region", err.Error())

		return
	}

	listReq := &datalab.ListDatalabsRequest{
		Region: region,
	}

	if !config.ProjectID.IsNull() && !config.ProjectID.IsUnknown() && config.ProjectID.ValueString() != "" {
		projectID := config.ProjectID.ValueString()
		listReq.ProjectID = &projectID
	}

	if !config.OrganizationID.IsNull() && !config.OrganizationID.IsUnknown() && config.OrganizationID.ValueString() != "" {
		orgID := config.OrganizationID.ValueString()
		listReq.OrganizationID = &orgID
	}

	if !config.Name.IsNull() && !config.Name.IsUnknown() && config.Name.ValueString() != "" {
		name := config.Name.ValueString()
		listReq.Name = &name
	}

	if !config.Tags.IsNull() && !config.Tags.IsUnknown() {
		listReq.Tags = expandTags(ctx, config.Tags, &resp.Diagnostics)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	listResp, listErr := d.api.ListDatalabs(listReq, scw.WithContext(ctx), scw.WithAllPages())
	if listErr != nil {
		resp.Diagnostics.AddError("Failed to list Datalabs", listErr.Error())

		return
	}

	state := config
	state.Datalabs = flattenDatalabsList(ctx, listResp.Datalabs, &resp.Diagnostics)
	state.Region = types.StringValue(region.String())

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func flattenDatalabsList(ctx context.Context, datalabs []*datalab.Datalab, diags *diag.Diagnostics) types.List {
	itemType := types.ObjectType{AttrTypes: datalabsItemAttrTypes()}

	if len(datalabs) == 0 {
		emptyList, d := types.ListValue(itemType, []attr.Value{})
		diags.Append(d...)

		return emptyList
	}

	items := make([]attr.Value, len(datalabs))

	for i, dl := range datalabs {
		tagList, d := scwtypes.FlattenFrameworkStringList(ctx, dl.Tags)
		diags.Append(d...)

		createdAt := types.StringNull()
		if dl.CreatedAt != nil {
			createdAt = types.StringValue(dl.CreatedAt.Format("2006-01-02T15:04:05Z07:00"))
		}

		updatedAt := types.StringNull()
		if dl.UpdatedAt != nil {
			updatedAt = types.StringValue(dl.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"))
		}

		attrValues := map[string]attr.Value{
			"id":            types.StringValue(dl.ID),
			"name":          types.StringValue(dl.Name),
			"description":   scwtypes.FlattenFrameworkStringValue(dl.Description),
			"status":        types.StringValue(string(dl.Status)),
			"tags":          tagList,
			"region":        types.StringValue(dl.Region.String()),
			"project_id":    types.StringValue(dl.ProjectID),
			"spark_version": types.StringValue(dl.SparkVersion),
			"has_notebook":  types.BoolValue(dl.HasNotebook),
			"created_at":    createdAt,
			"updated_at":    updatedAt,
		}

		obj, d := types.ObjectValue(datalabsItemAttrTypes(), attrValues)
		diags.Append(d...)

		items[i] = obj
	}

	list, d := types.ListValue(itemType, items)
	diags.Append(d...)

	return list
}
