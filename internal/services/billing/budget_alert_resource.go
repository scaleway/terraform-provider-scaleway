package billing

import (
	"context"
	_ "embed"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	billing "github.com/scaleway/scaleway-sdk-go/api/billing/v2"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/verify"
)

var (
	_ resource.Resource                = (*BudgetAlertResource)(nil)
	_ resource.ResourceWithConfigure   = (*BudgetAlertResource)(nil)
	_ resource.ResourceWithImportState = (*BudgetAlertResource)(nil)
)

func NewBudgetAlertResource() resource.Resource {
	return &BudgetAlertResource{}
}

type BudgetAlertResource struct {
	billingAPI *billing.API
	meta       *meta.Meta
}

type budgetAlertResourceModel struct {
	BudgetID       types.String `tfsdk:"budget_id"`
	OrganizationID types.String `tfsdk:"organization_id"`
	ID             types.String `tfsdk:"id"`
	CreatedAt      types.String `tfsdk:"created_at"`
	UpdatedAt      types.String `tfsdk:"updated_at"`
	Threshold      types.Int64  `tfsdk:"threshold"`
}

func (r *BudgetAlertResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_billing_budget_alert"
}

//go:embed descriptions/budget_alert_resource.md
var budgetAlertResourceDescription string

func (r *BudgetAlertResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: budgetAlertResourceDescription,
		Attributes: map[string]schema.Attribute{
			"budget_id": schema.StringAttribute{
				MarkdownDescription: "The ID of the budget to create alert for.",
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
			"threshold": schema.Int64Attribute{
				MarkdownDescription: "Threshold percentage above which the alert is sent (0-100).",
				Required:            true,
			},
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The ID of the budget alert",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
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

func (r *BudgetAlertResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	m, ok := req.ProviderData.(*meta.Meta)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *meta.Meta, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.meta = m
	r.billingAPI = billing.NewAPI(r.meta.ScwClient())
}

func (r *BudgetAlertResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data budgetAlertResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	res, err := r.billingAPI.CreateBudgetAlert(&billing.CreateBudgetAlertRequest{
		BudgetID:  data.BudgetID.ValueString(),
		Threshold: uint32(data.Threshold.ValueInt64()),
	}, scw.WithContext(ctx))
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to create budget alert",
			err.Error(),
		)

		return
	}

	state := convertBudgetAlertToState(res, data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *BudgetAlertResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state budgetAlertResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	alertID := state.ID.ValueString()
	if alertID == "" {
		resp.Diagnostics.AddError(
			"Alert ID not set",
			"Cannot refresh budget alert without an ID",
		)

		return
	}

	res, err := r.billingAPI.GetBudget(&billing.GetBudgetRequest{
		BudgetID: alertID,
	}, scw.WithContext(ctx))
	if err != nil {
		if httperrors.Is404(err) {
			resp.State.RemoveResource(ctx)

			return
		}

		resp.Diagnostics.AddError(
			"Failed to get budget alert",
			err.Error(),
		)

		return
	}

	var alert *billing.BudgetAlert

	for _, a := range res.Alerts {
		if a.ID == alertID {
			alert = a

			break
		}
	}

	if alert == nil {
		resp.Diagnostics.AddError(
			"Budget alert not found",
			fmt.Sprintf("Budget alert %s was not found", alertID),
		)

		return
	}

	state = convertBudgetAlertToState(alert, state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *BudgetAlertResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan budgetAlertResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	var state budgetAlertResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.billingAPI.UpdateBudgetAlert(&billing.UpdateBudgetAlertRequest{
		BudgetAlertID: state.ID.ValueString(),
		Threshold:     uint32(plan.Threshold.ValueInt64()),
	}, scw.WithContext(ctx))
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to update budget alert",
			err.Error(),
		)

		return
	}

	res, err := r.billingAPI.GetBudget(&billing.GetBudgetRequest{
		BudgetID: plan.BudgetID.ValueString(),
	}, scw.WithContext(ctx))
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to get budget after update",
			err.Error(),
		)

		return
	}

	var alert *billing.BudgetAlert

	for _, a := range res.Alerts {
		if a.ID == state.ID.ValueString() {
			alert = a

			break
		}
	}

	if alert == nil {
		resp.Diagnostics.AddError(
			"Budget alert not found after update",
			fmt.Sprintf("Budget alert %s was not found", state.ID.ValueString()),
		)

		return
	}

	newState := convertBudgetAlertToState(alert, plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &newState)...)
}

func (r *BudgetAlertResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state budgetAlertResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	err := r.billingAPI.DeleteBudgetAlert(&billing.DeleteBudgetAlertRequest{
		BudgetAlertID: state.ID.ValueString(),
	}, scw.WithContext(ctx))
	if err != nil {
		if !httperrors.Is404(err) {
			resp.Diagnostics.AddError(
				"Failed to delete budget alert",
				err.Error(),
			)
		}
	}
}

func (r *BudgetAlertResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	alertID := req.ID
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), alertID)...)
}

func convertBudgetAlertToState(alert *billing.BudgetAlert, state budgetAlertResourceModel) budgetAlertResourceModel {
	state.ID = types.StringValue(alert.ID)

	if alert.CreatedAt != nil {
		state.CreatedAt = types.StringValue(alert.CreatedAt.String())
	}

	if alert.UpdatedAt != nil {
		state.UpdatedAt = types.StringValue(alert.UpdatedAt.String())
	}

	state.Threshold = types.Int64Value(int64(alert.Threshold))

	return state
}
