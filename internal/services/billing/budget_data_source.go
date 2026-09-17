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
	_ datasource.DataSource              = (*BudgetDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*BudgetDataSource)(nil)
)

func NewBudgetDataSource() datasource.DataSource {
	return &BudgetDataSource{}
}

type BudgetDataSource struct {
	billingAPI *billing.API
	meta       *meta.Meta
}

type budgetDataSourceModel struct {
	ID               types.String `tfsdk:"id"`
	BudgetID         types.String `tfsdk:"budget_id"`
	OrganizationID   types.String `tfsdk:"organization_id"`
	CreatedAt        types.String `tfsdk:"created_at"`
	UpdatedAt        types.String `tfsdk:"updated_at"`
	ConsumptionLimit types.Int64  `tfsdk:"consumption_limit"`
	Enabled          types.Bool   `tfsdk:"enabled"`
}

func (d *BudgetDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_billing_budget"
}

//go:embed descriptions/budget_data_source.md
var budgetDataSourceDescription string

func (d *BudgetDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: budgetDataSourceDescription,
		Attributes: map[string]schema.Attribute{
			"budget_id": schema.StringAttribute{
				MarkdownDescription: "The ID of the budget to retrieve.",
				Required:            true,
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
			"consumption_limit": schema.Int64Attribute{
				MarkdownDescription: "Cost limit for the budget in cents.",
				Computed:            true,
			},
			"enabled": schema.BoolAttribute{
				MarkdownDescription: "Whether the budget is enabled or not.",
				Computed:            true,
			},
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The ID of the budget",
			},
			"created_at": schema.StringAttribute{
				MarkdownDescription: "The date and time of budget creation",
				Computed:            true,
			},
			"updated_at": schema.StringAttribute{
				MarkdownDescription: "The date and time when the budget was last updated",
				Computed:            true,
			},
		},
	}
}

func (d *BudgetDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *BudgetDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state budgetDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	budgetID := state.BudgetID.ValueString()
	if budgetID == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("budget_id"),
			"Budget ID is required",
			"The budget_id attribute must be set",
		)

		return
	}

	res, err := d.billingAPI.GetBudget(&billing.GetBudgetRequest{
		BudgetID: budgetID,
	}, scw.WithContext(ctx))
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to get budget",
			fmt.Sprintf("Could not retrieve budget %s: %v", budgetID, err),
		)

		return
	}

	state.BudgetID = types.StringValue(res.ID)
	state.ID = types.StringValue(res.ID)
	state.OrganizationID = types.StringValue(res.OrganizationID)

	if res.ConsumptionLimit != nil {
		state.ConsumptionLimit = types.Int64Value(res.ConsumptionLimit.Units)
	}

	state.Enabled = types.BoolValue(res.Enabled)

	if res.CreatedAt != nil {
		state.CreatedAt = types.StringValue(res.CreatedAt.String())
	}

	if res.UpdatedAt != nil {
		state.UpdatedAt = types.StringValue(res.UpdatedAt.String())
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
