package sdb

import (
	"context"
	_ "embed"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	sdbSDK "github.com/scaleway/scaleway-sdk-go/api/serverless_sqldb/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
	scwtypes "github.com/scaleway/terraform-provider-scaleway/v2/internal/types"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/verify"
)

//go:embed descriptions/versions_data_source.md
var versionsDataSourceDescription string

var (
	_ datasource.DataSource              = (*VersionsDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*VersionsDataSource)(nil)
)

func NewVersionsDataSource() datasource.DataSource {
	return &VersionsDataSource{}
}

type VersionsDataSource struct {
	api  *sdbSDK.API
	meta *meta.Meta
}

type versionsDataSourceModel struct {
	ID       types.String `tfsdk:"id"`
	Region   types.String `tfsdk:"region"`
	Name     types.String `tfsdk:"name"`
	Versions types.List   `tfsdk:"versions"`
}

func versionAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"name":           types.StringType,
		"end_of_life_at": types.StringType,
		"srn":            types.StringType,
	}
}

func (d *VersionsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_sdb_sql_versions"
}

func (d *VersionsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: versionsDataSourceDescription,
		Description:         versionsDataSourceDescription,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The ID of this data source",
			},
			"region": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "The region you want to attach the resource to",
				Validators: []validator.String{
					verify.IsStringOneOfWithWarning(regional.AllRegions()),
				},
			},
			"name": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Filter Serverless SQL Database versions that match a given name pattern (e.g. `16`).",
			},
			"versions": schema.ListNestedAttribute{
				Computed:            true,
				MarkdownDescription: "List of available Serverless SQL Database versions.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The major version of the PostgreSQL engine.",
						},
						"end_of_life_at": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "End of life date of the version (RFC3339).",
						},
						"srn": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The Scaleway Resource Name (SRN) of the version.",
						},
					},
				},
			},
		},
	}
}

func (d *VersionsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
	d.api = sdbSDK.NewAPI(d.meta.ScwClient())
}

func (d *VersionsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config versionsDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)

	if resp.Diagnostics.HasError() {
		return
	}

	region, err := meta.ExtractFrameworkRegion(config.Region, d.meta.ScwClient())
	if err != nil {
		resp.Diagnostics.AddError("Failed to resolve region", err.Error())

		return
	}

	listReq := &sdbSDK.ListVersionsRequest{
		Region: region,
	}

	if !config.Name.IsNull() && !config.Name.IsUnknown() && config.Name.ValueString() != "" {
		listReq.Version = scwtypes.ExpandStringPtr(config.Name.ValueString())
	}

	versions, err := d.api.ListVersions(listReq, scw.WithContext(ctx), scw.WithAllPages())
	if err != nil {
		resp.Diagnostics.AddError("Failed to list Serverless SQL Database versions", err.Error())

		return
	}

	state := versionsDataSourceModel{
		ID:     types.StringValue(region.String()),
		Region: types.StringValue(region.String()),
		Name:   config.Name,
	}

	state.Versions = flattenSDBVersions(versions.Versions, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func flattenSDBVersions(versions []*sdbSDK.Version, diags *diag.Diagnostics) types.List {
	elems := make([]attr.Value, 0, len(versions))

	for _, version := range versions {
		obj, d := types.ObjectValue(versionAttrTypes(), map[string]attr.Value{
			"name":           types.StringValue(version.Name),
			"end_of_life_at": types.StringValue(scwtypes.FlattenTime(version.EndOfLifeAt).(string)),
			"srn":            types.StringValue(version.Srn),
		})
		diags.Append(d...)

		elems = append(elems, obj)
	}

	list, d := types.ListValue(types.ObjectType{AttrTypes: versionAttrTypes()}, elems)
	diags.Append(d...)

	return list
}
