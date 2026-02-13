package scwconfig

import (
	"context"
	_ "embed"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/ephemeral"
	"github.com/hashicorp/terraform-plugin-framework/ephemeral/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
)

var (
	_ ephemeral.EphemeralResource              = (*ScwConfigEphemeralResource)(nil)
	_ ephemeral.EphemeralResourceWithConfigure = (*ScwConfigEphemeralResource)(nil)
)

type ScwConfigEphemeralResource struct {
	meta *meta.Meta
}

func NewScwConfigEphemeralResource() ephemeral.EphemeralResource {
	return &ScwConfigEphemeralResource{}
}

func (r *ScwConfigEphemeralResource) Configure(ctx context.Context, req ephemeral.ConfigureRequest, resp *ephemeral.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	m, ok := req.ProviderData.(*meta.Meta)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Ephemeral Resource Configure Type",
			fmt.Sprintf("Expected *meta.Meta, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.meta = m
}

func (r *ScwConfigEphemeralResource) Metadata(ctx context.Context, req ephemeral.MetadataRequest, resp *ephemeral.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_config"
}

type ScwConfigEphemeralResourceModel struct {
	ProjectID            types.String `tfsdk:"project_id"`
	ProjectIDSource      types.String `tfsdk:"project_id_source"`
	OrganizationID       types.String `tfsdk:"organization_id"`
	OrganizationIDSource types.String `tfsdk:"organization_id_source"`
	AccessKey            types.String `tfsdk:"access_key"`
	AccessKeySource      types.String `tfsdk:"access_key_source"`
	SecretKey            types.String `tfsdk:"secret_key"`
	SecretKeySource      types.String `tfsdk:"secret_key_source"`
	Zone                 types.String `tfsdk:"zone"`
	ZoneSource           types.String `tfsdk:"zone_source"`
	Region               types.String `tfsdk:"region"`
	RegionSource         types.String `tfsdk:"region_source"`
}

//go:embed descriptions/scw_config_ephemeral_resource.md
var scwConfigEphemeralResourceDescription string

func (r *ScwConfigEphemeralResource) Schema(ctx context.Context, req ephemeral.SchemaRequest, resp *ephemeral.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description:         scwConfigEphemeralResourceDescription,
		MarkdownDescription: scwConfigEphemeralResourceDescription,
		Attributes: map[string]schema.Attribute{
			"project_id": schema.StringAttribute{
				Computed:    true,
				Description: "Project ID used",
			},
			"project_id_source": schema.StringAttribute{
				Computed:    true,
				Description: "Where the project id definition comes from (Environment, configuration file, variable, ...)",
			},
			"organization_id": schema.StringAttribute{
				Computed:    true,
				Description: "Organization ID used",
			},
			"organization_id_source": schema.StringAttribute{
				Computed:    true,
				Description: "Where the organization id definition comes from (Environment, configuration file, variable, ...)",
			},
			"access_key": schema.StringAttribute{
				Computed:    true,
				Description: "Access Key used",
			},
			"access_key_source": schema.StringAttribute{
				Computed:    true,
				Description: "Where the access key definition comes from (Environment, configuration file, variable, ...)",
			},
			"secret_key": schema.StringAttribute{
				Computed:    true,
				Description: "Secret Key used",
				Sensitive:   true,
			},
			"secret_key_source": schema.StringAttribute{
				Computed:    true,
				Description: "Where the secret key definition comes from (Environment, configuration file, variable, ...)",
			},
			"zone": schema.StringAttribute{
				Computed:    true,
				Description: "Zone used",
			},
			"zone_source": schema.StringAttribute{
				Computed:    true,
				Description: "Where the zone definition comes from (Environment, configuration file, variable, ...)",
			},
			"region": schema.StringAttribute{
				Computed:    true,
				Description: "Region used",
			},
			"region_source": schema.StringAttribute{
				Computed:    true,
				Description: "Where the region definition comes from (Environment, configuration file, variable, ...)",
			},
		},
	}
}

func (r *ScwConfigEphemeralResource) Open(ctx context.Context, req ephemeral.OpenRequest, resp *ephemeral.OpenResponse) {
	var data ScwConfigEphemeralResourceModel

	if r.meta == nil {
		resp.Diagnostics.AddError(
			"Unconfigured ScwConfigEphemeralResource",
			"The ephemeral resource was not properly configured. The Scaleway client is missing. "+
				"This is usually a bug in the provider. Please report it to the maintainers.",
		)

		return
	}

	client := r.meta.ScwClient()

	accessKey, _ := client.GetAccessKey()
	data.AccessKey = types.StringValue(accessKey)
	data.AccessKeySource = types.StringValue(r.meta.AccessKeySource())

	secretKey, _ := client.GetSecretKey()
	data.SecretKey = types.StringValue(secretKey)
	data.SecretKeySource = types.StringValue(r.meta.SecretKeySource())

	projectID, _ := client.GetDefaultProjectID()
	data.ProjectID = types.StringValue(projectID)
	data.ProjectIDSource = types.StringValue(r.meta.ProjectIDSource())

	organizationID, _ := client.GetDefaultOrganizationID()
	data.OrganizationID = types.StringValue(organizationID)
	data.OrganizationIDSource = types.StringValue(r.meta.OrganizationIDSource())

	zone, _ := client.GetDefaultZone()
	data.Zone = types.StringValue(string(zone))
	data.ZoneSource = types.StringValue(r.meta.ZoneSource())

	region, _ := client.GetDefaultRegion()
	data.Region = types.StringValue(string(region))
	data.RegionSource = types.StringValue(r.meta.RegionSource())

	resp.Result.Set(ctx, &data)
}
