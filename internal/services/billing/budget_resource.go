package billing

import (
	"context"
	_ "embed"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
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
	_ resource.Resource                = (*BudgetResource)(nil)
	_ resource.ResourceWithConfigure   = (*BudgetResource)(nil)
	_ resource.ResourceWithImportState = (*BudgetResource)(nil)
)

func NewBudgetResource() resource.Resource {
	return &BudgetResource{}
}

type BudgetResource struct {
	billingAPI *billing.API
	meta       *meta.Meta
}

type budgetResourceModel struct {
	ID               types.String `tfsdk:"id"`
	OrganizationID   types.String `tfsdk:"organization_id"`
	CreatedAt        types.String `tfsdk:"created_at"`
	UpdatedAt        types.String `tfsdk:"updated_at"`
	ConsumptionLimit types.Int64  `tfsdk:"consumption_limit"`
	Enabled          types.Bool   `tfsdk:"enabled"`
}

func (r *BudgetResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_billing_budget"
}

//go:embed descriptions/budget_resource.md
var budgetResourceDescription string

func (r *BudgetResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: budgetResourceDescription,
		Attributes: map[string]schema.Attribute{
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
				Required:            true,
			},
			"enabled": schema.BoolAttribute{
				MarkdownDescription: "Whether the budget is enabled or not.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(true),
			},
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The ID of the budget",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
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

func (r *BudgetResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *BudgetResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data budgetResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	orgID := data.OrganizationID.ValueString()
	if orgID == "" {
		defaultOrgID, exists := r.meta.ScwClient().GetDefaultOrganizationID()
		if exists {
			orgID = defaultOrgID
		} else {
			resp.Diagnostics.AddAttributeError(
				path.Root("organization_id"),
				"Organization ID is required",
				"Either set organization_id or configure a default organization",
			)

			return
		}
	}

	res, err := r.billingAPI.CreateBudget(&billing.CreateBudgetRequest{
		OrganizationID:   orgID,
		ConsumptionLimit: uint32(data.ConsumptionLimit.ValueInt64()),
		Enabled:          data.Enabled.ValueBool(),
	}, scw.WithContext(ctx))
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to create budget",
			err.Error(),
		)

		return
	}

	state := convertBudgetToState(res, data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *BudgetResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state budgetResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	budgetID := state.ID.ValueString()
	if budgetID == "" {
		resp.Diagnostics.AddError(
			"Budget ID not set",
			"Cannot refresh budget without an ID",
		)

		return
	}

	res, err := r.billingAPI.GetBudget(&billing.GetBudgetRequest{
		BudgetID: budgetID,
	}, scw.WithContext(ctx))
	if err != nil {
		if httperrors.Is404(err) {
			resp.State.RemoveResource(ctx)

			return
		}

		resp.Diagnostics.AddError(
			"Failed to get budget",
			err.Error(),
		)

		return
	}

	state = convertBudgetToState(res, state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *BudgetResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan budgetResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	consumptionLimit := uint32(plan.ConsumptionLimit.ValueInt64())

	_, err := r.billingAPI.UpdateBudget(&billing.UpdateBudgetRequest{
		BudgetID:         plan.ID.ValueString(),
		ConsumptionLimit: &consumptionLimit,
		Enabled:          plan.Enabled.ValueBoolPointer(),
	}, scw.WithContext(ctx))
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to update budget",
			err.Error(),
		)

		return
	}

	res, err := r.billingAPI.GetBudget(&billing.GetBudgetRequest{
		BudgetID: plan.ID.ValueString(),
	}, scw.WithContext(ctx))
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to get budget after update",
			err.Error(),
		)

		return
	}

	newState := convertBudgetToState(res, plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &newState)...)
}

func (r *BudgetResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state budgetResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	err := r.billingAPI.DeleteBudget(&billing.DeleteBudgetRequest{
		BudgetID: state.ID.ValueString(),
	}, scw.WithContext(ctx))
	if err != nil {
		if !httperrors.Is404(err) {
			resp.Diagnostics.AddError(
				"Failed to delete budget",
				err.Error(),
			)
		}
	}
}

func (r *BudgetResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	budgetID := req.ID
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), budgetID)...)
}

func convertBudgetToState(budget *billing.Budget, state budgetResourceModel) budgetResourceModel {
	state.ID = types.StringValue(budget.ID)
	state.OrganizationID = types.StringValue(budget.OrganizationID)

	if budget.ConsumptionLimit != nil {
		state.ConsumptionLimit = types.Int64Value(budget.ConsumptionLimit.Units)
	}

	state.Enabled = types.BoolValue(budget.Enabled)

	if budget.CreatedAt != nil {
		state.CreatedAt = types.StringValue(budget.CreatedAt.String())
	}

	if budget.UpdatedAt != nil {
		state.UpdatedAt = types.StringValue(budget.UpdatedAt.String())
	}

	return state
}
