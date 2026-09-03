package kafka

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	kafkaapi "github.com/scaleway/scaleway-sdk-go/api/kafka/v1alpha1"
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
	api  *kafkaapi.API
	meta *meta.Meta
}

type versionDataSourceModel struct {
	Name              types.String `tfsdk:"name"`
	Region            types.String `tfsdk:"region"`
	ID                types.String `tfsdk:"id"`
	EndOfLifeAt       types.String `tfsdk:"end_of_life_at"`
	AvailableSettings types.List   `tfsdk:"available_settings"`
}

func (d *versionDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_kafka_version"
}

func (d *versionDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "The `scaleway_kafka_version` data source is used to retrieve information about an available Kafka version.\n\n" +
			"Refer to the [Kafka documentation](https://www.scaleway.com/en/docs/managed-databases/kafka/) and " +
			"[API documentation](https://www.scaleway.com/en/developers/api/kafka/) for more information.",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The Kafka version name. Use `latest` to retrieve the most recent available version.",
			},
			"region": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "The region the Kafka version is available in.",
			},
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The ID of the version, in the `{region}/{version}` format.",
			},
			"end_of_life_at": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The end-of-life date of the version (RFC 3339 format).",
			},
			"available_settings": schema.ListNestedAttribute{
				Computed:            true,
				MarkdownDescription: "The cluster configuration settings available for clusters running this version.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The setting name.",
						},
						"hot_configurable": schema.BoolAttribute{
							Computed:            true,
							MarkdownDescription: "Whether the setting can be applied without a restart.",
						},
						"description": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The setting description.",
						},
						"bool_property": schema.SingleNestedAttribute{
							Computed:            true,
							MarkdownDescription: "Boolean property, if the setting is a boolean.",
							Attributes: map[string]schema.Attribute{
								"default_value": schema.BoolAttribute{
									Computed:            true,
									MarkdownDescription: "The default value of the setting.",
								},
							},
						},
						"string_property": schema.SingleNestedAttribute{
							Computed:            true,
							MarkdownDescription: "String property, if the setting is a string.",
							Attributes: map[string]schema.Attribute{
								"default_value": schema.StringAttribute{
									Computed:            true,
									MarkdownDescription: "The default value of the setting.",
								},
								"string_constraint": schema.StringAttribute{
									Computed:            true,
									MarkdownDescription: "The string constraint of the setting (e.g. a regex).",
								},
							},
						},
						"int_property": schema.SingleNestedAttribute{
							Computed:            true,
							MarkdownDescription: "Integer property, if the setting is an integer.",
							Attributes: map[string]schema.Attribute{
								"min": schema.Int64Attribute{
									Computed:            true,
									MarkdownDescription: "The minimum value of the setting.",
								},
								"max": schema.Int64Attribute{
									Computed:            true,
									MarkdownDescription: "The maximum value of the setting.",
								},
								"default_value": schema.Int64Attribute{
									Computed:            true,
									MarkdownDescription: "The default value of the setting.",
								},
								"unit": schema.StringAttribute{
									Computed:            true,
									MarkdownDescription: "The unit of the setting.",
								},
							},
						},
						"float_property": schema.SingleNestedAttribute{
							Computed:            true,
							MarkdownDescription: "Float property, if the setting is a float.",
							Attributes: map[string]schema.Attribute{
								"min": schema.Float64Attribute{
									Computed:            true,
									MarkdownDescription: "The minimum value of the setting.",
								},
								"max": schema.Float64Attribute{
									Computed:            true,
									MarkdownDescription: "The maximum value of the setting.",
								},
								"default_value": schema.Float64Attribute{
									Computed:            true,
									MarkdownDescription: "The default value of the setting.",
								},
								"unit": schema.StringAttribute{
									Computed:            true,
									MarkdownDescription: "The unit of the setting.",
								},
							},
						},
					},
				},
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
	d.api = kafkaapi.NewAPI(d.meta.ScwClient())
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

	var version *kafkaapi.Version

	if name == "latest" {
		res, err := d.api.ListVersions(&kafkaapi.ListVersionsRequest{
			Region: region,
		}, scw.WithContext(ctx))
		if err != nil {
			resp.Diagnostics.AddError("Failed to list Kafka versions", err.Error())

			return
		}

		if len(res.Versions) == 0 {
			resp.Diagnostics.AddError("No Kafka versions found", "could not find the latest version")

			return
		}

		version = res.Versions[0]
	} else {
		res, err := d.api.ListVersions(&kafkaapi.ListVersionsRequest{
			Region:  region,
			Version: &name,
		}, scw.WithContext(ctx))
		if err != nil {
			resp.Diagnostics.AddError("Failed to list Kafka versions", err.Error())

			return
		}

		if len(res.Versions) == 0 {
			resp.Diagnostics.AddError("Kafka version not found", fmt.Sprintf("could not find version %q", name))

			return
		}

		version = res.Versions[0]
	}

	state := flattenKafkaVersion(ctx, version, region, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func flattenKafkaVersion(ctx context.Context, version *kafkaapi.Version, region scw.Region, diags *diag.Diagnostics) versionDataSourceModel {
	state := versionDataSourceModel{
		Name:              types.StringValue(version.Version),
		Region:            types.StringValue(region.String()),
		ID:                types.StringValue(regional.NewIDString(region, version.Version)),
		AvailableSettings: flattenKafkaVersionAvailableSettings(ctx, version.AvailableSettings, diags),
	}

	if version.EndOfLifeAt != nil {
		state.EndOfLifeAt = types.StringValue(version.EndOfLifeAt.Format("2006-01-02T15:04:05Z07:00"))
	} else {
		state.EndOfLifeAt = types.StringNull()
	}

	return state
}

func flattenKafkaVersionAvailableSettings(ctx context.Context, settings []*kafkaapi.VersionAvailableSetting, diags *diag.Diagnostics) types.List {
	itemType := types.ObjectType{AttrTypes: availableSettingAttrTypes()}

	if len(settings) == 0 {
		emptyList, d := types.ListValue(itemType, []attr.Value{})
		diags.Append(d...)

		return emptyList
	}

	items := make([]attr.Value, len(settings))

	for i, setting := range settings {
		attrValues := map[string]attr.Value{
			"name":             types.StringValue(setting.Name),
			"hot_configurable": types.BoolValue(setting.HotConfigurable),
			"description":      types.StringValue(setting.Description),
			"bool_property":    flattenKafkaSettingBoolProperty(setting.BoolProperty, diags),
			"string_property":  flattenKafkaSettingStringProperty(setting.StringProperty, diags),
			"int_property":     flattenKafkaSettingIntProperty(setting.IntProperty, diags),
			"float_property":   flattenKafkaSettingFloatProperty(setting.FloatProperty, diags),
		}

		obj, d := types.ObjectValue(availableSettingAttrTypes(), attrValues)
		diags.Append(d...)

		items[i] = obj
	}

	list, d := types.ListValue(itemType, items)
	diags.Append(d...)

	return list
}

func flattenKafkaSettingBoolProperty(prop *kafkaapi.VersionAvailableSettingBooleanProperty, diags *diag.Diagnostics) types.Object {
	if prop == nil {
		return types.ObjectNull(boolPropertyAttrTypes())
	}

	obj, d := types.ObjectValue(boolPropertyAttrTypes(), map[string]attr.Value{
		"default_value": types.BoolValue(prop.DefaultValue),
	})
	diags.Append(d...)

	return obj
}

func flattenKafkaSettingStringProperty(prop *kafkaapi.VersionAvailableSettingStringProperty, diags *diag.Diagnostics) types.Object {
	if prop == nil {
		return types.ObjectNull(stringPropertyAttrTypes())
	}

	attrValues := map[string]attr.Value{
		"default_value": types.StringValue(prop.DefaultValue),
	}

	if prop.StringConstraint != nil {
		attrValues["string_constraint"] = types.StringValue(*prop.StringConstraint)
	} else {
		attrValues["string_constraint"] = types.StringNull()
	}

	obj, d := types.ObjectValue(stringPropertyAttrTypes(), attrValues)
	diags.Append(d...)

	return obj
}

func flattenKafkaSettingIntProperty(prop *kafkaapi.VersionAvailableSettingIntegerProperty, diags *diag.Diagnostics) types.Object {
	if prop == nil {
		return types.ObjectNull(intPropertyAttrTypes())
	}

	attrValues := map[string]attr.Value{
		"min":           types.Int64Value(int64(prop.Min)),
		"max":           types.Int64Value(int64(prop.Max)),
		"default_value": types.Int64Value(int64(prop.DefaultValue)),
	}

	if prop.Unit != nil {
		attrValues["unit"] = types.StringValue(*prop.Unit)
	} else {
		attrValues["unit"] = types.StringNull()
	}

	obj, d := types.ObjectValue(intPropertyAttrTypes(), attrValues)
	diags.Append(d...)

	return obj
}

func flattenKafkaSettingFloatProperty(prop *kafkaapi.VersionAvailableSettingFloatProperty, diags *diag.Diagnostics) types.Object {
	if prop == nil {
		return types.ObjectNull(floatPropertyAttrTypes())
	}

	attrValues := map[string]attr.Value{
		"min":           types.Float64Value(float64(prop.Min)),
		"max":           types.Float64Value(float64(prop.Max)),
		"default_value": types.Float64Value(float64(prop.DefaultValue)),
	}

	if prop.Unit != nil {
		attrValues["unit"] = types.StringValue(*prop.Unit)
	} else {
		attrValues["unit"] = types.StringNull()
	}

	obj, d := types.ObjectValue(floatPropertyAttrTypes(), attrValues)
	diags.Append(d...)

	return obj
}

func availableSettingAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"name":             types.StringType,
		"hot_configurable": types.BoolType,
		"description":      types.StringType,
		"bool_property":    types.ObjectType{AttrTypes: boolPropertyAttrTypes()},
		"string_property":  types.ObjectType{AttrTypes: stringPropertyAttrTypes()},
		"int_property":     types.ObjectType{AttrTypes: intPropertyAttrTypes()},
		"float_property":   types.ObjectType{AttrTypes: floatPropertyAttrTypes()},
	}
}

func boolPropertyAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"default_value": types.BoolType,
	}
}

func stringPropertyAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"default_value":     types.StringType,
		"string_constraint": types.StringType,
	}
}

func intPropertyAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"min":           types.Int64Type,
		"max":           types.Int64Type,
		"default_value": types.Int64Type,
		"unit":          types.StringType,
	}
}

func floatPropertyAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"min":           types.Float64Type,
		"max":           types.Float64Type,
		"default_value": types.Float64Type,
		"unit":          types.StringType,
	}
}
