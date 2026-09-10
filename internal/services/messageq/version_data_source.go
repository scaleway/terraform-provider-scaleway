package messageq

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	messageqapi "github.com/scaleway/scaleway-sdk-go/api/messageq/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
)

var (
	_ datasource.DataSource              = (*versionDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*versionDataSource)(nil)
)

func NewVersionDataSource() datasource.DataSource {
	return &versionDataSource{}
}

type versionDataSource struct {
	api  *messageqapi.API
	meta *meta.Meta
}

type versionDataSourceModel struct {
	Name      types.String `tfsdk:"name"`
	Region    types.String `tfsdk:"region"`
	ID        types.String `tfsdk:"id"`
	Version   types.String `tfsdk:"version"`
	EndOfLife types.String `tfsdk:"end_of_life"`
	Disabled  types.Bool   `tfsdk:"disabled"`
	Beta      types.Bool   `tfsdk:"beta"`
}

func (d *versionDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_messageq_version"
}

func (d *versionDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "The `scaleway_messageq_version` data source is used to retrieve information about an available MessageQ version.\n\n" +
			"Refer to the [MessageQ API documentation](https://www.scaleway.com/en/developers/api/messageq) for more information.",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The MessageQ version name. Use `latest` to retrieve the most recent available non-disabled version.",
			},
			"region": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "The region the MessageQ version is available in.",
			},
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The ID of the version, in the `{region}/{version}` format.",
			},
			"version": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The MessageQ version string.",
			},
			"end_of_life": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The end-of-life date of the version (RFC 3339 format).",
			},
			"disabled": schema.BoolAttribute{
				Computed:            true,
				MarkdownDescription: "Whether the version is disabled.",
			},
			"beta": schema.BoolAttribute{
				Computed:            true,
				MarkdownDescription: "Whether the version is in beta.",
			},
		},
	}
}

func (d *versionDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *versionDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config versionDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)

	if resp.Diagnostics.HasError() {
		return
	}

	region, err := meta.ExtractFrameworkRegion(config.Region, d.meta.ScwClient())
	if err != nil {
		resp.Diagnostics.AddError("Failed to resolve region", err.Error())

		return
	}

	name := config.Name.ValueString()

	var version *messageqapi.Version

	if name == "latest" {
		res, err := d.api.ListVersions(&messageqapi.ListVersionsRequest{
			Region:  region,
			OrderBy: messageqapi.ListVersionsRequestOrderByVersionDesc,
		}, scw.WithContext(ctx), scw.WithAllPages())
		if err != nil {
			resp.Diagnostics.AddError("Failed to list MessageQ versions", err.Error())

			return
		}

		for _, candidate := range res.Versions {
			if candidate.Disabled {
				continue
			}

			version = candidate

			break
		}

		if version == nil {
			resp.Diagnostics.AddError("No MessageQ versions found", "could not find the latest available version")

			return
		}
	} else {
		res, err := d.api.ListVersions(&messageqapi.ListVersionsRequest{
			Region:  region,
			Version: &name,
		}, scw.WithContext(ctx), scw.WithAllPages())
		if err != nil {
			resp.Diagnostics.AddError("Failed to list MessageQ versions", err.Error())

			return
		}

		for _, candidate := range res.Versions {
			if candidate.Version == name {
				version = candidate

				break
			}
		}

		if version == nil {
			resp.Diagnostics.AddError("MessageQ version not found", fmt.Sprintf("could not find version %q", name))

			return
		}
	}

	state := versionDataSourceModel{
		Name:     types.StringValue(name),
		Region:   types.StringValue(region.String()),
		ID:       types.StringValue(regional.NewIDString(region, version.Version)),
		Version:  types.StringValue(version.Version),
		Disabled: types.BoolValue(version.Disabled),
		Beta:     types.BoolValue(version.Beta),
	}

	if version.EndOfLife != nil {
		state.EndOfLife = types.StringValue(version.EndOfLife.Format("2006-01-02T15:04:05Z07:00"))
	} else {
		state.EndOfLife = types.StringNull()
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
