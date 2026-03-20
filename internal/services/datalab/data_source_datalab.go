package datalab

import (
	"context"
	"errors"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	datalab "github.com/scaleway/scaleway-sdk-go/api/datalab/v1beta1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
)

var (
	_ datasource.DataSource              = (*DatalabDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*DatalabDataSource)(nil)
)

func NewDatalabDataSource() datasource.DataSource {
	return &DatalabDataSource{}
}

type DatalabDataSource struct {
	api  *datalab.API
	meta *meta.Meta
}

type datalabDataSourceModel struct {
	Tags              types.List   `tfsdk:"tags"`
	PrivateNetworkID  types.String `tfsdk:"private_network_id"`
	Main              types.Object `tfsdk:"main"`
	Region            types.String `tfsdk:"region"`
	ID                types.String `tfsdk:"id"`
	Description       types.String `tfsdk:"description"`
	Name              types.String `tfsdk:"name"`
	SparkVersion      types.String `tfsdk:"spark_version"`
	DatalabID         types.String `tfsdk:"datalab_id"`
	ProjectID         types.String `tfsdk:"project_id"`
	Worker            types.Object `tfsdk:"worker"`
	NotebookMasterURL types.String `tfsdk:"notebook_master_url"`
	TotalStorage      types.Object `tfsdk:"total_storage"`
	Status            types.String `tfsdk:"status"`
	CreatedAt         types.String `tfsdk:"created_at"`
	UpdatedAt         types.String `tfsdk:"updated_at"`
	NotebookURL       types.String `tfsdk:"notebook_url"`
	HasNotebook       types.Bool   `tfsdk:"has_notebook"`
}

func (d *DatalabDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_datalab"
}

func (d *DatalabDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "",
		Attributes: map[string]schema.Attribute{
			"datalab_id": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "The ID of the Datalab instance to look up.",
			},
			"name": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "The name of the Datalab instance to look up.",
			},
			"project_id": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "The project ID the Datalab belongs to.",
			},
			"region": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "The region the Datalab is in.",
			},
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The ID of the Datalab instance, in the `{region}/{id}` format.",
			},
			"description": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "A description for the Datalab instance.",
			},
			"tags": schema.ListAttribute{
				Computed:            true,
				ElementType:         types.StringType,
				MarkdownDescription: "Tags associated with the Datalab instance.",
			},
			"spark_version": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The Spark version used by the Datalab instance.",
			},
			"private_network_id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The ID of the private network attached to the Datalab.",
			},
			"has_notebook": schema.BoolAttribute{
				Computed:            true,
				MarkdownDescription: "Whether a JupyterLab notebook is associated with the Datalab.",
			},
			"main": schema.SingleNestedAttribute{
				Computed:            true,
				MarkdownDescription: "The Spark main node configuration.",
				Attributes: map[string]schema.Attribute{
					"node_type": schema.StringAttribute{
						Computed:            true,
						MarkdownDescription: "The node type for the main node.",
					},
					"spark_ui_url": schema.StringAttribute{
						Computed:            true,
						MarkdownDescription: "The Spark UI URL.",
					},
					"spark_master_url": schema.StringAttribute{
						Computed:            true,
						MarkdownDescription: "The Spark master URL within the VPC.",
					},
					"root_volume": schema.SingleNestedAttribute{
						Computed:            true,
						MarkdownDescription: "Volume details for the main node.",
						Attributes: map[string]schema.Attribute{
							"type": schema.StringAttribute{
								Computed: true,
							},
							"size": schema.Int64Attribute{
								Computed: true,
							},
						},
					},
				},
			},
			"worker": schema.SingleNestedAttribute{
				Computed:            true,
				MarkdownDescription: "The Spark worker nodes configuration.",
				Attributes: map[string]schema.Attribute{
					"node_type": schema.StringAttribute{
						Computed:            true,
						MarkdownDescription: "The node type for worker nodes.",
					},
					"node_count": schema.Int64Attribute{
						Computed:            true,
						MarkdownDescription: "The number of worker nodes.",
					},
					"root_volume": schema.SingleNestedAttribute{
						Computed:            true,
						MarkdownDescription: "Volume details for worker nodes.",
						Attributes: map[string]schema.Attribute{
							"type": schema.StringAttribute{
								Computed: true,
							},
							"size": schema.Int64Attribute{
								Computed: true,
							},
						},
					},
				},
			},
			"total_storage": schema.SingleNestedAttribute{
				Computed:            true,
				MarkdownDescription: "Persistent volume storage configuration.",
				Attributes: map[string]schema.Attribute{
					"type": schema.StringAttribute{
						Computed: true,
					},
					"size": schema.Int64Attribute{
						Computed: true,
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

func (d *DatalabDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *DatalabDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config datalabDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)

	if resp.Diagnostics.HasError() {
		return
	}

	region, err := d.resolveRegion(config.Region)
	if err != nil {
		resp.Diagnostics.AddError("Failed to resolve region", err.Error())

		return
	}

	var dl *datalab.Datalab

	hasDatalabID := !config.DatalabID.IsNull() && !config.DatalabID.IsUnknown() && config.DatalabID.ValueString() != ""
	hasName := !config.Name.IsNull() && !config.Name.IsUnknown() && config.Name.ValueString() != ""

	switch {
	case hasDatalabID:
		datalabID := config.DatalabID.ValueString()

		parsedRegion, parsedID, parseErr := regional.ParseID(datalabID)
		if parseErr == nil {
			region = parsedRegion
			datalabID = parsedID
		}

		dl, err = d.api.GetDatalab(&datalab.GetDatalabRequest{
			Region:    region,
			DatalabID: datalabID,
		}, scw.WithContext(ctx))
		if err != nil {
			resp.Diagnostics.AddError("Failed to get Datalab", err.Error())

			return
		}
	case hasName:
		name := config.Name.ValueString()
		listReq := &datalab.ListDatalabsRequest{
			Region: region,
			Name:   &name,
		}

		if !config.ProjectID.IsNull() && !config.ProjectID.IsUnknown() && config.ProjectID.ValueString() != "" {
			projectID := config.ProjectID.ValueString()
			listReq.ProjectID = &projectID
		}

		listResp, listErr := d.api.ListDatalabs(listReq, scw.WithContext(ctx))
		if listErr != nil {
			resp.Diagnostics.AddError("Failed to list Datalabs", listErr.Error())

			return
		}

		if len(listResp.Datalabs) == 0 {
			resp.Diagnostics.AddError("Datalab not found", fmt.Sprintf("No Datalab found with name %q", name))

			return
		}

		if len(listResp.Datalabs) > 1 {
			resp.Diagnostics.AddError(
				"Multiple Datalabs found",
				fmt.Sprintf("Found %d Datalabs with name %q. Please use datalab_id to specify exactly which one.", len(listResp.Datalabs), name),
			)

			return
		}

		dl = listResp.Datalabs[0]
	default:
		resp.Diagnostics.AddError(
			"Missing lookup attribute",
			"Either datalab_id or name must be specified.",
		)

		return
	}

	state := flattenDatalabDataSource(ctx, dl, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (d *DatalabDataSource) resolveRegion(regionAttr types.String) (scw.Region, error) {
	if !regionAttr.IsNull() && !regionAttr.IsUnknown() && regionAttr.ValueString() != "" {
		return scw.ParseRegion(regionAttr.ValueString())
	}

	region, exists := d.meta.ScwClient().GetDefaultRegion()
	if exists {
		return region, nil
	}

	return "", errors.New("region is required; set it on the data source or configure a default region on the provider")
}

func flattenDatalabDataSource(ctx context.Context, dl *datalab.Datalab, diags *diag.Diagnostics) datalabDataSourceModel {
	flat := flattenDatalab(ctx, dl, diags)

	return datalabDataSourceModel{
		DatalabID:         types.StringValue(dl.ID),
		Name:              flat.Name,
		ProjectID:         flat.ProjectID,
		Region:            flat.Region,
		ID:                flat.ID,
		Description:       flat.Description,
		Tags:              flat.Tags,
		SparkVersion:      flat.SparkVersion,
		PrivateNetworkID:  flat.PrivateNetworkID,
		HasNotebook:       flat.HasNotebook,
		Main:              flat.Main,
		Worker:            flat.Worker,
		TotalStorage:      flat.TotalStorage,
		Status:            flat.Status,
		CreatedAt:         flat.CreatedAt,
		UpdatedAt:         flat.UpdatedAt,
		NotebookURL:       flat.NotebookURL,
		NotebookMasterURL: flat.NotebookMasterURL,
	}
}
