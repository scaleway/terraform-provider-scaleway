package billing

import (
	"context"
	_ "embed"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/identityschema"
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
	_ resource.Resource              = (*BudgetAlertResource)(nil)
	_ resource.ResourceWithConfigure = (*BudgetAlertResource)(nil)
	_ resource.ResourceWithIdentity  = (*BudgetAlertResource)(nil)
)

func NewBudgetAlertResource() resource.Resource {
	return &BudgetAlertResource{}
}

type BudgetAlertResource struct {
	billingAPI *billing.API
	meta       *meta.Meta
}

type budgetAlertResourceModel struct {
	BudgetID  types.String `tfsdk:"budget_id"`
	ID        types.String `tfsdk:"id"`
	CreatedAt types.String `tfsdk:"created_at"`
	UpdatedAt types.String `tfsdk:"updated_at"`
	Threshold types.Int64  `tfsdk:"threshold"`
}

type budgetAlertResourceIdentityModel struct {
	ID types.String `tfsdk:"id"`
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
			"threshold": schema.Int64Attribute{
				MarkdownDescription: "Threshold percentage above which the alert is sent (0-100).",
				Required:            true,
				Validators: []validator.Int64{
					int64validator.Between(0, 100),
				},
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

func (r *BudgetAlertResource) IdentitySchema(ctx context.Context, req resource.IdentitySchemaRequest, resp *resource.IdentitySchemaResponse) {
	resp.IdentitySchema = identityschema.Schema{
		Attributes: map[string]identityschema.Attribute{
			"id": identityschema.StringAttribute{
				RequiredForImport: true,
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

	state := convertBudgetAlertToState(res, budgetAlertResourceModel{
		BudgetID: data.BudgetID,
	})
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)

	identity := budgetAlertResourceIdentityModel{
		ID: types.StringValue(res.ID),
	}
	resp.Diagnostics.Append(resp.Identity.Set(ctx, &identity)...)
}

func (r *BudgetAlertResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state budgetAlertResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	orgID, _ := r.meta.ScwClient().GetDefaultOrganizationID()

	listResp, err := r.billingAPI.ListBudgets(&billing.ListBudgetsRequest{
		OrganizationID: &orgID,
	}, scw.WithContext(ctx))
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to list budgets",
			err.Error(),
		)

		return
	}

	foundAlert, foundBudgetID := findBudgetAlertInList(listResp.Budgets, state.ID.ValueString())

	if foundAlert == nil {
		resp.State.RemoveResource(ctx)

		return
	}

	state = convertBudgetAlertToState(foundAlert, state)
	state.BudgetID = types.StringValue(foundBudgetID)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)

	identity := budgetAlertResourceIdentityModel{
		ID: types.StringValue(foundAlert.ID),
	}
	resp.Diagnostics.Append(resp.Identity.Set(ctx, &identity)...)
}

func (r *BudgetAlertResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan budgetAlertResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	var state budgetAlertResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	res, err := r.billingAPI.UpdateBudgetAlert(&billing.UpdateBudgetAlertRequest{
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

	newState := convertBudgetAlertToState(res, plan)
	newState.BudgetID = plan.BudgetID
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

	orgID, _ := r.meta.ScwClient().GetDefaultOrganizationID()

	listResp, err := r.billingAPI.ListBudgets(&billing.ListBudgetsRequest{
		OrganizationID: &orgID,
	}, scw.WithContext(ctx))
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to list budgets during import",
			err.Error(),
		)

		return
	}

	foundAlert, foundBudgetID := findBudgetAlertInList(listResp.Budgets, alertID)

	if foundAlert == nil {
		resp.Diagnostics.AddError(
			"Budget alert not found during import",
			fmt.Sprintf("Budget alert %s was not found in any budget", alertID),
		)

		return
	}

	state := budgetAlertResourceModel{
		ID:       types.StringValue(alertID),
		BudgetID: types.StringValue(foundBudgetID),
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func convertBudgetAlertToState(alert *billing.BudgetAlert, state budgetAlertResourceModel) budgetAlertResourceModel {
	state.ID = types.StringValue(alert.ID)
	state.Threshold = types.Int64Value(int64(alert.Threshold))

	if alert.CreatedAt != nil {
		state.CreatedAt = types.StringValue(alert.CreatedAt.String())
	}

	if alert.UpdatedAt != nil {
		state.UpdatedAt = types.StringValue(alert.UpdatedAt.String())
	}

	return state
}

func findBudgetAlertInList(budgets []*billing.Budget, alertID string) (*billing.BudgetAlert, string) {
	for _, budget := range budgets {
		for _, alert := range budget.Alerts {
			if alert.ID == alertID {
				return alert, budget.ID
			}
		}
	}

	return nil, ""
}
