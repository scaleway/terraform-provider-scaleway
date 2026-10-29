package messageq

import (
	"context"
	_ "embed"
	"fmt"
	"io"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	messageqapi "github.com/scaleway/scaleway-sdk-go/api/messageq/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/verify"
)

//go:embed descriptions/certificate_authority_data_source.md
var certificateAuthorityDataSourceDescription string

var (
	_ datasource.DataSource              = (*CertificateAuthorityDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*CertificateAuthorityDataSource)(nil)
)

func NewCertificateAuthorityDataSource() datasource.DataSource {
	return &CertificateAuthorityDataSource{}
}

type CertificateAuthorityDataSource struct {
	api  *messageqapi.API
	meta *meta.Meta
}

type certificateAuthorityDataSourceModel struct {
	ID           types.String `tfsdk:"id"`
	DeploymentID types.String `tfsdk:"deployment_id"`
	Region       types.String `tfsdk:"region"`
	PEM          types.String `tfsdk:"pem"`
}

func (d *CertificateAuthorityDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_messageq_certificate_authority"
}

func (d *CertificateAuthorityDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: certificateAuthorityDataSourceDescription,
		Attributes: map[string]schema.Attribute{
			"deployment_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The ID of the MessageQ deployment",
				Validators: []validator.String{
					verify.IsStringUUIDOrUUIDWithRegion(),
				},
			},
			"region": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "The region the MessageQ deployment is in.",
			},
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The ID of the certificate authority data source, in the `{region}/{deployment_id}` format.",
			},
			"pem": schema.StringAttribute{
				Computed:            true,
				Sensitive:           true,
				MarkdownDescription: "PEM-encoded certificate authority content",
			},
		},
	}
}

func (d *CertificateAuthorityDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *CertificateAuthorityDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config certificateAuthorityDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)

	if resp.Diagnostics.HasError() {
		return
	}

	region, err := meta.ExtractFrameworkRegion(config.Region, d.meta.ScwClient())
	if err != nil {
		resp.Diagnostics.AddError("Failed to resolve region", err.Error())

		return
	}

	region, deploymentID, err := RegionAndIDFromAttr(config.DeploymentID.ValueString(), region)
	if err != nil {
		resp.Diagnostics.AddError("Failed to parse deployment_id", err.Error())

		return
	}

	file, err := d.api.DownloadDeploymentCertificateAuthority(&messageqapi.DownloadDeploymentCertificateAuthorityRequest{
		Region:       region,
		DeploymentID: deploymentID,
	}, scw.WithContext(ctx))
	if err != nil {
		resp.Diagnostics.AddError("Failed to download MessageQ certificate authority", err.Error())

		return
	}

	if file == nil || file.Content == nil {
		resp.Diagnostics.AddError(
			"Empty certificate authority",
			"certificate authority content is empty for deployment "+deploymentID,
		)

		return
	}

	content, err := io.ReadAll(file.Content)
	if err != nil {
		resp.Diagnostics.AddError("Failed to read certificate authority content", err.Error())

		return
	}

	state := certificateAuthorityDataSourceModel{
		ID:           types.StringValue(regional.NewIDString(region, deploymentID)),
		Region:       types.StringValue(region.String()),
		DeploymentID: config.DeploymentID,
		PEM:          types.StringValue(string(content)),
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
