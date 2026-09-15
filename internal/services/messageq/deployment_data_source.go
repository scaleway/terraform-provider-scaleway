package messageq

import (
	"context"
	_ "embed"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	messageqapi "github.com/scaleway/scaleway-sdk-go/api/messageq/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	scwdatasource "github.com/scaleway/terraform-provider-scaleway/v2/internal/datasource"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/verify"
)

//go:embed descriptions/deployment_data_source.md
var deploymentDataSourceDescription string

var (
	_ datasource.DataSource              = (*DeploymentDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*DeploymentDataSource)(nil)
)

func NewDeploymentDataSource() datasource.DataSource {
	return &DeploymentDataSource{}
}

type DeploymentDataSource struct {
	api  *messageqapi.API
	meta *meta.Meta
}

type deploymentDataSourceModel struct {
	Tags         types.List   `tfsdk:"tags"`
	Endpoints    types.List   `tfsdk:"endpoints"`
	Volume       types.Object `tfsdk:"volume"`
	ID           types.String `tfsdk:"id"`
	DeploymentID types.String `tfsdk:"deployment_id"`
	Region       types.String `tfsdk:"region"`
	ProjectID    types.String `tfsdk:"project_id"`
	Name         types.String `tfsdk:"name"`
	Version      types.String `tfsdk:"version"`
	NodeType     types.String `tfsdk:"node_type"`
	Status       types.String `tfsdk:"status"`
	CreatedAt    types.String `tfsdk:"created_at"`
	UpdatedAt    types.String `tfsdk:"updated_at"`
	NodeCount    types.Int64  `tfsdk:"node_count"`
}

func (d *DeploymentDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_messageq_deployment"
}

func (d *DeploymentDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: deploymentDataSourceDescription,
		Attributes: map[string]schema.Attribute{
			"deployment_id": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "The ID of the MessageQ deployment",
				Validators: []validator.String{
					verify.IsStringUUIDOrUUIDWithRegion(),
					stringvalidator.ConflictsWith(path.MatchRoot("name")),
				},
			},
			"name": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Name of the MessageQ deployment",
				Validators: []validator.String{
					stringvalidator.ConflictsWith(path.MatchRoot("deployment_id")),
				},
			},
			"region": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "The region the MessageQ deployment is in.",
			},
			"project_id": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "The project ID the MessageQ deployment belongs to.",
			},
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The ID of the MessageQ deployment, in the `{region}/{id}` format.",
			},
			"tags": schema.ListAttribute{
				Computed:            true,
				ElementType:         types.StringType,
				MarkdownDescription: "List of tags applied to the deployment",
			},
			"version": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "MessageQ version",
			},
			"node_count": schema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: "Number of nodes",
			},
			"node_type": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Type of node",
			},
			"volume": schema.SingleNestedAttribute{
				Computed:            true,
				MarkdownDescription: "Volume configuration",
				Attributes: map[string]schema.Attribute{
					"type": schema.StringAttribute{
						Computed:            true,
						MarkdownDescription: "Volume type (sbs_5k, sbs_15k)",
					},
					"size_in_gb": schema.Int64Attribute{
						Computed:            true,
						MarkdownDescription: "Volume size in GB",
					},
				},
			},
			"endpoints": schema.ListNestedAttribute{
				Computed:            true,
				MarkdownDescription: "List of endpoints",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Endpoint ID",
						},
						"services": schema.ListNestedAttribute{
							Computed:            true,
							MarkdownDescription: "List of services",
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"name": schema.StringAttribute{
										Computed:            true,
										MarkdownDescription: "Service name",
									},
									"port": schema.Int64Attribute{
										Computed:            true,
										MarkdownDescription: "Service port",
									},
									"url": schema.StringAttribute{
										Computed:            true,
										MarkdownDescription: "Service URL",
									},
								},
							},
						},
						"public": schema.BoolAttribute{
							Computed:            true,
							MarkdownDescription: "Whether the endpoint is public",
						},
						"private_network_id": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Private network ID if applicable",
						},
					},
				},
			},
			"status": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The status of the deployment",
			},
			"created_at": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Date and time of deployment creation (RFC 3339 format)",
			},
			"updated_at": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Date and time of deployment last update (RFC 3339 format)",
			},
		},
	}
}

func (d *DeploymentDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
	d.api = messageqapi.NewAPI(d.meta.ScwClient())
}

func (d *DeploymentDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config deploymentDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)

	if resp.Diagnostics.HasError() {
		return
	}

	region, err := meta.ExtractFrameworkRegion(config.Region, d.meta.ScwClient())
	if err != nil {
		resp.Diagnostics.AddError("Failed to resolve region", err.Error())

		return
	}

	var deploymentID string

	deploymentIDAttr := types.StringNull()
	hasDeploymentID := !config.DeploymentID.IsNull() && !config.DeploymentID.IsUnknown() && config.DeploymentID.ValueString() != ""
	hasName := !config.Name.IsNull() && !config.Name.IsUnknown() && config.Name.ValueString() != ""

	switch {
	case hasDeploymentID:
		var parseErr error

		region, deploymentID, parseErr = RegionAndIDFromAttr(config.DeploymentID.ValueString(), region)
		if parseErr != nil {
			resp.Diagnostics.AddError("Failed to parse deployment_id", parseErr.Error())

			return
		}

		deploymentIDAttr = config.DeploymentID
	case hasName:
		deploymentName := config.Name.ValueString()
		listReq := &messageqapi.ListDeploymentsRequest{
			Region: region,
			Name:   &deploymentName,
		}

		if !config.ProjectID.IsNull() && !config.ProjectID.IsUnknown() && config.ProjectID.ValueString() != "" {
			projectID := config.ProjectID.ValueString()
			listReq.ProjectID = &projectID
		}

		res, listErr := d.api.ListDeployments(listReq, scw.WithContext(ctx))
		if listErr != nil {
			resp.Diagnostics.AddError("Failed to list MessageQ deployments", listErr.Error())

			return
		}

		foundDeployment, findErr := scwdatasource.FindExact(
			res.Deployments,
			func(s *messageqapi.Deployment) bool { return s.Name == deploymentName },
			deploymentName,
		)
		if findErr != nil {
			resp.Diagnostics.AddError("Failed to find MessageQ deployment", findErr.Error())

			return
		}

		deploymentID = foundDeployment.ID
		deploymentIDAttr = types.StringValue(regional.NewIDString(region, deploymentID))
	default:
		resp.Diagnostics.AddError(
			"Missing lookup attribute",
			"Either deployment_id or name must be specified.",
		)

		return
	}

	deployment, err := waitForDeployment(ctx, d.api, region, deploymentID, defaultDeploymentReadTimeout)
	if err != nil {
		resp.Diagnostics.AddError("Failed to read MessageQ deployment", err.Error())

		return
	}

	flat := flattenDeployment(ctx, deployment, types.ObjectNull(privateNetworkAttrTypes()), nil, &resp.Diagnostics)
	// Data sources must report every endpoint the API returns; resource filtering
	// against private_network would hide private endpoints after a PN is attached.
	flat.Endpoints = flattenEndpointsList(deployment.Endpoints, &resp.Diagnostics)

	state := deploymentDataSourceModel{
		ID:           flat.ID,
		DeploymentID: deploymentIDAttr,
		Region:       flat.Region,
		ProjectID:    flat.ProjectID,
		Name:         flat.Name,
		Tags:         flat.Tags,
		Version:      flat.Version,
		NodeCount:    flat.NodeCount,
		NodeType:     flat.NodeType,
		Volume:       flat.Volume,
		Endpoints:    flat.Endpoints,
		Status:       flat.Status,
		CreatedAt:    flat.CreatedAt,
		UpdatedAt:    flat.UpdatedAt,
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
