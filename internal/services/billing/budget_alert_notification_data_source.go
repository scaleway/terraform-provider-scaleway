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
	_ datasource.DataSource              = (*BudgetAlertNotificationDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*BudgetAlertNotificationDataSource)(nil)
)

func NewBudgetAlertNotificationDataSource() datasource.DataSource {
	return &BudgetAlertNotificationDataSource{}
}

type BudgetAlertNotificationDataSource struct {
	billingAPI *billing.API
	meta       *meta.Meta
}

type budgetAlertNotificationDataSourceModel struct {
	NotificationID types.String `tfsdk:"notification_id"`
	BudgetAlertID  types.String `tfsdk:"budget_alert_id"`
	// Output
	ID             types.String `tfsdk:"id"`
	OrganizationID types.String `tfsdk:"organization_id"`
	CreatedAt      types.String `tfsdk:"created_at"`
	UpdatedAt      types.String `tfsdk:"updated_at"`
	Type           types.String `tfsdk:"type"`
	Recipients     types.Set    `tfsdk:"recipients"`
}

func (d *BudgetAlertNotificationDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_billing_budget_alert_notification"
}

//go:embed descriptions/budget_alert_notification_data_source.md
var budgetAlertNotificationDataSourceDescription string

func (d *BudgetAlertNotificationDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: budgetAlertNotificationDataSourceDescription,
		Attributes: map[string]schema.Attribute{
			"notification_id": schema.StringAttribute{
				MarkdownDescription: "The ID of the budget alert notification to retrieve.",
				Required:            true,
				Validators: []validator.String{
					verify.IsStringUUID(),
				},
			},
			"budget_alert_id": schema.StringAttribute{
				MarkdownDescription: "The ID of the budget alert. If not provided, it will be retrieved from the notification.",
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
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The ID of the budget alert notification",
			},
			"created_at": schema.StringAttribute{
				MarkdownDescription: "The date and time of budget alert notification creation",
				Computed:            true,
			},
			"updated_at": schema.StringAttribute{
				MarkdownDescription: "The date and time when the budget alert notification was last updated",
				Computed:            true,
			},
			"type": schema.StringAttribute{
				MarkdownDescription: "The type of notification (sms, email, or webhook)",
				Computed:            true,
			},
			"recipients": schema.SetAttribute{
				MarkdownDescription: "List of recipients for this notification",
				Computed:            true,
				ElementType:         types.StringType,
			},
		},
	}
}

func (d *BudgetAlertNotificationDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *BudgetAlertNotificationDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state budgetAlertNotificationDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	notificationID := state.NotificationID.ValueString()
	if notificationID == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("notification_id"),
			"Notification ID is required",
			"The notification_id attribute must be set",
		)

		return
	}

	orgID := state.OrganizationID.ValueString()
	if orgID == "" {
		defaultOrgID, exists := d.meta.ScwClient().GetDefaultOrganizationID()
		if exists {
			orgID = defaultOrgID
		}
	}

	listResp, err := d.billingAPI.ListBudgets(&billing.ListBudgetsRequest{
		OrganizationID: &orgID,
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
			for _, notification := range alert.Notifications {
				if notification.ID == notificationID {
					found = true
					state.NotificationID = types.StringValue(notification.ID)
					state.ID = types.StringValue(notification.ID)
					state.BudgetAlertID = types.StringValue(alert.ID)
					state.OrganizationID = types.StringValue(budget.OrganizationID)
					state.Type = types.StringValue(string(notification.Type))

					if notification.CreatedAt != nil {
						state.CreatedAt = types.StringValue(notification.CreatedAt.String())
					}

					if notification.UpdatedAt != nil {
						state.UpdatedAt = types.StringValue(notification.UpdatedAt.String())
					}

					if len(notification.Recipients) > 0 {
						recipients, _ := types.SetValueFrom(ctx, types.StringType, notification.Recipients)
						state.Recipients = recipients
					}

					break
				}
			}

			if found {
				break
			}
		}

		if found {
			break
		}
	}

	if !found {
		resp.Diagnostics.AddError(
			"Budget alert notification not found",
			fmt.Sprintf("Budget alert notification %s was not found", notificationID),
		)

		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
