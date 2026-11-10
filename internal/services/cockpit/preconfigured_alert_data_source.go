package cockpit

import (
	"context"
	_ "embed"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	cockpit "github.com/scaleway/scaleway-sdk-go/api/cockpit/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/account"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/transport"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/verify"
)

//go:embed descriptions/cockpit_preconfigured_alert_data_source.md
var cockpitPreconfiguredAlertDataSourceDescription string

var (
	_ datasource.DataSource              = (*PreconfiguredAlertDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*PreconfiguredAlertDataSource)(nil)
)

func NewPreconfiguredAlertDataSource() datasource.DataSource {
	return &PreconfiguredAlertDataSource{}
}

type PreconfiguredAlertDataSource struct {
	dataSourceBase
}

type preconfiguredAlertModel struct {
	Annotations        map[string]string `tfsdk:"annotations"`
	Name               types.String      `tfsdk:"name"`
	Rule               types.String      `tfsdk:"rule"`
	Duration           types.String      `tfsdk:"duration"`
	RuleStatus         types.String      `tfsdk:"rule_status"`
	State              types.String      `tfsdk:"state"`
	PreconfiguredRule  types.String      `tfsdk:"preconfigured_rule_id"`
	DisplayName        types.String      `tfsdk:"display_name"`
	DisplayDescription types.String      `tfsdk:"display_description"`
	ProductName        types.String      `tfsdk:"product_name"`
	ProductFamily      types.String      `tfsdk:"product_family"`
	DataSourceID       types.String      `tfsdk:"data_source_id"`
}

type preconfiguredAlertDataSourceModel struct {
	ID           types.String              `tfsdk:"id"`
	ProjectID    types.String              `tfsdk:"project_id"`
	Region       types.String              `tfsdk:"region"`
	DataSourceID types.String              `tfsdk:"data_source_id"`
	RuleStatus   types.String              `tfsdk:"rule_status"`
	Alerts       []preconfiguredAlertModel `tfsdk:"alerts"`
}

func (d *PreconfiguredAlertDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cockpit_preconfigured_alert"
}

func (d *PreconfiguredAlertDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: cockpitPreconfiguredAlertDataSourceDescription,
		Description:         cockpitPreconfiguredAlertDataSourceDescription,
		Attributes: map[string]schema.Attribute{
			"id":         schema.StringAttribute{Computed: true, MarkdownDescription: "The ID of the data source, in the `{region}/{project_id}` format."},
			"project_id": account.ProjectIDSchemaAttributeOptional("The project_id you want to attach the resource to"),
			"region":     regional.SchemaAttributeOptionalComputed(),
			"data_source_id": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Filter alerts by data source ID",
				Validators:          []validator.String{verify.IsStringUUID()},
			},
			"rule_status": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Filter alerts by rule status (enabled, disabled)",
				Validators:          []validator.String{verify.FrameworkValidateEnum[cockpit.AlertStatus]()},
			},
			"alerts": schema.ListNestedAttribute{
				Computed:            true,
				MarkdownDescription: "List of preconfigured alerts",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name":                  schema.StringAttribute{Computed: true, MarkdownDescription: "Name of the alert"},
						"rule":                  schema.StringAttribute{Computed: true, MarkdownDescription: "PromQL rule defining the alert condition"},
						"duration":              schema.StringAttribute{Computed: true, MarkdownDescription: "Duration for which the alert must be active before firing"},
						"rule_status":           schema.StringAttribute{Computed: true, MarkdownDescription: "Status of the alert (enabled, disabled, enabling, disabling)"},
						"state":                 schema.StringAttribute{Computed: true, MarkdownDescription: "Current state of the alert (inactive, pending, firing)"},
						"annotations":           schema.MapAttribute{Computed: true, ElementType: types.StringType, MarkdownDescription: "Annotations for the alert"},
						"preconfigured_rule_id": schema.StringAttribute{Computed: true, MarkdownDescription: "ID of the preconfigured rule"},
						"display_name":          schema.StringAttribute{Computed: true, MarkdownDescription: "Human readable name of the alert"},
						"display_description":   schema.StringAttribute{Computed: true, MarkdownDescription: "Human readable description of the alert"},
						"product_name":          schema.StringAttribute{Computed: true, MarkdownDescription: "Product associated with the alert"},
						"product_family":        schema.StringAttribute{Computed: true, MarkdownDescription: "Family of the product associated with the alert"},
						"data_source_id":        schema.StringAttribute{Computed: true, MarkdownDescription: "ID of the data source containing the alert rule"},
					},
				},
			},
		},
	}
}

func (d *PreconfiguredAlertDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data preconfiguredAlertDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	api, region, err := d.regionalAPI(data.Region)
	if err != nil {
		resp.Diagnostics.AddError("Failed to resolve region", err.Error())

		return
	}

	projectID, err := meta.ExtractFrameworkProjectID(data.ProjectID, d.meta.ScwClient())
	if err != nil {
		resp.Diagnostics.AddError("Failed to resolve project_id", err.Error())

		return
	}

	listReq := &cockpit.RegionalAPIListAlertsRequest{
		Region:          region,
		ProjectID:       projectID,
		IsPreconfigured: new(true),
	}

	if data.DataSourceID.ValueString() != "" {
		listReq.DataSourceID = data.DataSourceID.ValueStringPointer()
	}

	if data.RuleStatus.ValueString() != "" {
		listReq.RuleStatus = new(cockpit.AlertStatus(data.RuleStatus.ValueString()))
	}

	res, err := transport.RetryOnStatusValue(ctx, transport.IAMPropagation, func(ctx context.Context) (*cockpit.ListAlertsResponse, error) {
		return api.ListAlerts(listReq, scw.WithContext(ctx), scw.WithAllPages())
	})
	if err != nil {
		resp.Diagnostics.AddError("Failed to list preconfigured alerts", err.Error())

		return
	}

	data.Alerts = make([]preconfiguredAlertModel, 0, len(res.Alerts))

	for _, alert := range res.Alerts {
		data.Alerts = append(data.Alerts, flattenPreconfiguredAlert(alert))
	}

	data.ID = types.StringValue(regional.NewIDString(region, projectID))
	data.ProjectID = types.StringValue(projectID)
	data.Region = types.StringValue(region.String())

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func flattenPreconfiguredAlert(alert *cockpit.Alert) preconfiguredAlertModel {
	model := preconfiguredAlertModel{
		Name:               types.StringValue(alert.Name),
		Rule:               types.StringValue(alert.Rule),
		Duration:           types.StringValue(alert.Duration),
		RuleStatus:         types.StringValue(string(alert.RuleStatus)),
		Annotations:        alert.Annotations,
		DataSourceID:       types.StringValue(alert.DataSourceID),
		State:              types.StringNull(),
		PreconfiguredRule:  types.StringNull(),
		DisplayName:        types.StringNull(),
		DisplayDescription: types.StringNull(),
		ProductName:        types.StringNull(),
		ProductFamily:      types.StringNull(),
	}

	if alert.State != nil {
		model.State = types.StringValue(string(*alert.State))
	}

	if alert.PreconfiguredData != nil {
		model.PreconfiguredRule = types.StringValue(alert.PreconfiguredData.PreconfiguredRuleID)
		model.DisplayName = types.StringValue(alert.PreconfiguredData.DisplayName)
		model.DisplayDescription = types.StringValue(alert.PreconfiguredData.DisplayDescription)
		model.ProductName = types.StringValue(alert.PreconfiguredData.ProductName)
		model.ProductFamily = types.StringValue(alert.PreconfiguredData.ProductFamily)
	}

	return model
}
