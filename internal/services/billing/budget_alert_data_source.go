package billing

import (
	"context"
	_ "embed"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	billing "github.com/scaleway/scaleway-sdk-go/api/billing/v2"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/verify"
)

var (
	_ datasource.DataSource              = (*BudgetAlertDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*BudgetAlertDataSource)(nil)
)

func NewBudgetAlertDataSource() datasource.DataSource {
	return &BudgetAlertDataSource{}
}

type BudgetAlertDataSource struct {
	billingAPI *billing.API
	meta       *meta.Meta
}

type budgetAlertDataSourceModel struct {
	AlertID        types.String `tfsdk:"alert_id"`
	BudgetID       types.String `tfsdk:"budget_id"`
	OrganizationID types.String `tfsdk:"organization_id"`
	ID             types.String `tfsdk:"id"`
	CreatedAt      types.String `tfsdk:"created_at"`
	UpdatedAt      types.String `tfsdk:"updated_at"`
	Threshold      types.Int64  `tfsdk:"threshold"`
}

func (d *BudgetAlertDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_billing_budget_alert"
}

//go:embed descriptions/budget_alert_data_source.md
var budgetAlertDataSourceDescription string

func (d *BudgetAlertDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: budgetAlertDataSourceDescription,
		Attributes: map[string]schema.Attribute{
			"alert_id": schema.StringAttribute{
				MarkdownDescription: "The ID of the budget alert to retrieve.",
				Required:            true,
				Validators: []validator.String{
					verify.IsStringUUID(),
				},
			},
			"budget_id": schema.StringAttribute{
				MarkdownDescription: "The ID of the budget. If not provided, it will be retrieved from the alert.",
				Optional:            true,
				Computed:            true,
				Validators: []validator.String{
					verify.IsStringUUID(),
				},
			},
			"organization_id": schema.StringAttribute{
				MarkdownDescription: "The organization ID. If not provided, the default organization configured in the provider is used.",
				Optional:            true,
				Computed:            true,
				Validators: []validator.String{
					verify.IsStringUUID(),
				},
			},
			"threshold": schema.Int64Attribute{
				MarkdownDescription: "Threshold percentage above which the alert is sent",
				Computed:            true,
			},
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The ID of the budget alert",
			},
			"created_at": schema.StringAttribute{
				MarkdownDescription: "The date and time of budget alert creation",
				Computed:            true,
			},
			"updated_at": schema.StringAttribute{
				MarkdownDescription: "The date and time when the budget alert was last updated",
				Computed:            true,
			},
		},
	}
}

func (d *BudgetAlertDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
	d.billingAPI = billing.NewAPI(d.meta.ScwClient())
}

func (d *BudgetAlertDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state budgetAlertDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	alertID := state.AlertID.ValueString()
	if alertID == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("alert_id"),
			"Alert ID is required",
			"The alert_id attribute must be set",
		)

		return
	}

	listResp, err := d.billingAPI.ListBudgets(&billing.ListBudgetsRequest{
		OrganizationID: state.OrganizationID.ValueStringPointer(),
	}, scw.WithContext(ctx))
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to list budgets",
			fmt.Sprintf("Could not list budgets: %v", err),
		)

		return
	}

	var found bool

	for _, budget := range listResp.Budgets {
		for _, alert := range budget.Alerts {
			if alert.ID == alertID {
				found = true
				state.AlertID = types.StringValue(alert.ID)
				state.ID = types.StringValue(alert.ID)
				state.BudgetID = types.StringValue(budget.ID)
				state.OrganizationID = types.StringValue(budget.OrganizationID)
				state.Threshold = types.Int64Value(int64(alert.Threshold))

				if alert.CreatedAt != nil {
					state.CreatedAt = types.StringValue(alert.CreatedAt.String())
				}

				if alert.UpdatedAt != nil {
					state.UpdatedAt = types.StringValue(alert.UpdatedAt.String())
				}

				break
			}
		}

		if found {
			break
		}
	}

	if !found {
		resp.Diagnostics.AddError(
			"Budget alert not found",
			fmt.Sprintf("Budget alert %s was not found", alertID),
		)

		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
