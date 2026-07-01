package billing

import (
	"context"
	_ "embed"
	"errors"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
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
	_ resource.Resource                = (*BudgetAlertNotificationResource)(nil)
	_ resource.ResourceWithConfigure   = (*BudgetAlertNotificationResource)(nil)
	_ resource.ResourceWithImportState = (*BudgetAlertNotificationResource)(nil)
	_ resource.ResourceWithIdentity    = (*BudgetAlertNotificationResource)(nil)
)

func NewBudgetAlertNotificationResource() resource.Resource {
	return &BudgetAlertNotificationResource{}
}

type BudgetAlertNotificationResource struct {
	billingAPI *billing.API
	meta       *meta.Meta
}

type budgetAlertNotificationResourceModel struct {
	BudgetAlertID   types.String `tfsdk:"budget_alert_id"`
	SmsPhoneNumbers types.Set    `tfsdk:"sms_phone_numbers"`
	EmailAddresses  types.Set    `tfsdk:"email_addresses"`
	WebhookURLs     types.Set    `tfsdk:"webhook_urls"`
	// Output
	ID        types.String `tfsdk:"id"`
	CreatedAt types.String `tfsdk:"created_at"`
	UpdatedAt types.String `tfsdk:"updated_at"`
	Type      types.String `tfsdk:"type"`
}

type budgetAlertNotificationIdentityModel struct {
	ID types.String `tfsdk:"id"`
}

func (r *BudgetAlertNotificationResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_billing_budget_alert_notification"
}

//go:embed descriptions/budget_alert_notification_resource.md
var budgetAlertNotificationResourceDescription string

var budgetAlertNotificationTypeValidator = setvalidator.ExactlyOneOf(
	path.MatchRoot("sms_phone_numbers"),
	path.MatchRoot("email_addresses"),
	path.MatchRoot("webhook_urls"),
)

func (r *BudgetAlertNotificationResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: budgetAlertNotificationResourceDescription,
		Attributes: map[string]schema.Attribute{
			"budget_alert_id": schema.StringAttribute{
				MarkdownDescription: "The ID of the budget alert to create notification for.",
				Required:            true,
				Validators: []validator.String{
					verify.IsStringUUID(),
				},
			},
			"sms_phone_numbers": schema.SetAttribute{
				MarkdownDescription: "List of phone numbers to receive SMS notifications. Precisely one of sms_phone_numbers, email_addresses, or webhook_urls must be set.",
				Optional:            true,
				ElementType:         types.StringType,
				Validators: []validator.Set{
					budgetAlertNotificationTypeValidator,
				},
			},
			"email_addresses": schema.SetAttribute{
				MarkdownDescription: "List of email addresses to receive email notifications. Precisely one of sms_phone_numbers, email_addresses, or webhook_urls must be set.",
				Optional:            true,
				ElementType:         types.StringType,
				Validators: []validator.Set{
					budgetAlertNotificationTypeValidator,
				},
			},
			"webhook_urls": schema.SetAttribute{
				MarkdownDescription: "List of webhook URLs to receive webhook notifications. Precisely one of sms_phone_numbers, email_addresses, or webhook_urls must be set.",
				Optional:            true,
				ElementType:         types.StringType,
				Validators: []validator.Set{
					budgetAlertNotificationTypeValidator,
				},
			},
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The ID of the budget alert notification",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
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
		},
	}
}

func (r *BudgetAlertNotificationResource) IdentitySchema(ctx context.Context, req resource.IdentitySchemaRequest, resp *resource.IdentitySchemaResponse) {
	resp.IdentitySchema = identityschema.Schema{
		Attributes: map[string]identityschema.Attribute{
			"id": identityschema.StringAttribute{
				RequiredForImport: true,
			},
		},
	}
}

func (r *BudgetAlertNotificationResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *BudgetAlertNotificationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data budgetAlertNotificationResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	req_create := &billing.CreateBudgetAlertNotificationRequest{
		BudgetAlertID: data.BudgetAlertID.ValueString(),
	}

	setNotificationTypeForRequest(ctx, &data, &req_create.SmsPhoneNumbers, &req_create.EmailAddresses, &req_create.WebhookURLs, &resp.Diagnostics)

	res, err := r.billingAPI.CreateBudgetAlertNotification(req_create, scw.WithContext(ctx))
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to create budget alert notification",
			err.Error(),
		)

		return
	}

	var diags diag.Diagnostics

	state := convertBudgetAlertNotificationToState(ctx, res, budgetAlertNotificationResourceModel{}, &diags)
	state.BudgetAlertID = data.BudgetAlertID

	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)

	identity := budgetAlertNotificationIdentityModel{
		ID: types.StringValue(res.ID),
	}
	resp.Diagnostics.Append(resp.Identity.Set(ctx, &identity)...)
}

func (r *BudgetAlertNotificationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state budgetAlertNotificationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	notificationID := state.ID.ValueString()
	if notificationID == "" {
		resp.Diagnostics.AddError(
			"Notification ID not set",
			"Cannot refresh budget alert notification without an ID",
		)

		return
	}

	foundWithBudget, err := findBudgetAlertNotification(r.billingAPI, r.meta, notificationID, "", "", ctx)
	if err != nil {
		if httperrors.Is404(err) {
			resp.State.RemoveResource(ctx)

			return
		}

		resp.Diagnostics.AddError(
			"Failed to get budget alert notification",
			err.Error(),
		)

		return
	}

	if foundWithBudget == nil {
		resp.State.RemoveResource(ctx)

		return
	}

	found := foundWithBudget.notification
	state.BudgetAlertID = types.StringValue(foundWithBudget.BudgetAlertID)

	var diags diag.Diagnostics

	state = convertBudgetAlertNotificationToState(ctx, found, state, &diags)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)

	identity := budgetAlertNotificationIdentityModel{
		ID: types.StringValue(found.ID),
	}
	resp.Diagnostics.Append(resp.Identity.Set(ctx, &identity)...)
}

func (r *BudgetAlertNotificationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan budgetAlertNotificationResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	var state budgetAlertNotificationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	req_update := &billing.UpdateBudgetAlertNotificationRequest{
		BudgetAlertNotificationID: state.ID.ValueString(),
	}

	setNotificationTypeForRequest(ctx, &plan, &req_update.SmsPhoneNumbers, &req_update.EmailAddresses, &req_update.WebhookURLs, &resp.Diagnostics)

	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.billingAPI.UpdateBudgetAlertNotification(req_update, scw.WithContext(ctx))
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to update budget alert notification",
			err.Error(),
		)

		return
	}

	// Re-read to get updated state
	found, err := findBudgetAlertNotification(r.billingAPI, r.meta, state.ID.ValueString(), "", "", ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to get budget alert notification after update",
			err.Error(),
		)

		return
	}

	if found == nil {
		resp.Diagnostics.AddError(
			"Budget alert notification not found after update",
			fmt.Sprintf("Budget alert notification %s was not found", state.ID.ValueString()),
		)

		return
	}

	var diags diag.Diagnostics

	newState := convertBudgetAlertNotificationToState(ctx, found.notification, budgetAlertNotificationResourceModel{}, &diags)
	newState.BudgetAlertID = types.StringValue(found.BudgetAlertID)

	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &newState)...)
}

func (r *BudgetAlertNotificationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state budgetAlertNotificationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	err := r.billingAPI.DeleteBudgetAlertNotification(&billing.DeleteBudgetAlertNotificationRequest{
		BudgetAlertNotificationID: state.ID.ValueString(),
	}, scw.WithContext(ctx))
	if err != nil {
		if !httperrors.Is404(err) {
			resp.Diagnostics.AddError(
				"Failed to delete budget alert notification",
				err.Error(),
			)
		}
	}
}

func (r *BudgetAlertNotificationResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	notificationID := req.ID

	found, err := findBudgetAlertNotification(r.billingAPI, r.meta, notificationID, "", "", ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to fetch budget alert notification during import",
			err.Error(),
		)

		return
	}

	if found == nil {
		resp.Diagnostics.AddError(
			"Budget alert notification not found during import",
			fmt.Sprintf("Budget alert notification %s was not found", notificationID),
		)

		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), notificationID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("budget_alert_id"), found.BudgetAlertID)...)
}

func convertBudgetAlertNotificationToState(ctx context.Context, notification *billing.BudgetAlertNotification, state budgetAlertNotificationResourceModel, diags *diag.Diagnostics) budgetAlertNotificationResourceModel {
	state.ID = types.StringValue(notification.ID)

	if notification.CreatedAt != nil {
		state.CreatedAt = types.StringValue(notification.CreatedAt.String())
	}

	if notification.UpdatedAt != nil {
		state.UpdatedAt = types.StringValue(notification.UpdatedAt.String())
	}

	state.Type = types.StringValue(string(notification.Type))

	// Initialize all notification type sets to empty sets
	emptySet := types.SetNull(types.StringType)
	state.SmsPhoneNumbers = emptySet
	state.EmailAddresses = emptySet
	state.WebhookURLs = emptySet

	switch notification.Type {
	case billing.BudgetAlertNotificationTypeSms:
		if len(notification.Recipients) > 0 {
			phoneNumbers, setDiags := types.SetValueFrom(ctx, types.StringType, notification.Recipients)
			diags.Append(setDiags...)

			if !diags.HasError() {
				state.SmsPhoneNumbers = phoneNumbers
			}
		}
	case billing.BudgetAlertNotificationTypeEmail:
		if len(notification.Recipients) > 0 {
			emails, setDiags := types.SetValueFrom(ctx, types.StringType, notification.Recipients)
			diags.Append(setDiags...)

			if !diags.HasError() {
				state.EmailAddresses = emails
			}
		}
	case billing.BudgetAlertNotificationTypeWebhook:
		if len(notification.Recipients) > 0 {
			urls, setDiags := types.SetValueFrom(ctx, types.StringType, notification.Recipients)
			diags.Append(setDiags...)

			if !diags.HasError() {
				state.WebhookURLs = urls
			}
		}
	}

	return state
}

type notificationWithBudget struct {
	notification   *billing.BudgetAlertNotification
	BudgetAlertID  string
	OrganizationID string
}

func findBudgetAlertNotification(billingAPI *billing.API, meta *meta.Meta, notificationID, orgID, budgetAlertID string, ctx context.Context) (*notificationWithBudget, error) {
	if orgID == "" {
		defaultOrgID, exists := meta.ScwClient().GetDefaultOrganizationID()
		if exists {
			orgID = defaultOrgID
		} else {
			return nil, errors.New("could not determine default organization ID")
		}
	}

	listResp, err := billingAPI.ListBudgets(&billing.ListBudgetsRequest{
		OrganizationID: &orgID,
	}, scw.WithContext(ctx))
	if err != nil {
		return nil, err
	}

	for _, budget := range listResp.Budgets {
		for _, alert := range budget.Alerts {
			if budgetAlertID != "" && alert.ID != budgetAlertID {
				continue
			}

			for _, notification := range alert.Notifications {
				if notification.ID == notificationID {
					return &notificationWithBudget{
						notification:   notification,
						BudgetAlertID:  alert.ID,
						OrganizationID: budget.OrganizationID,
					}, nil
				}
			}
		}
	}

	return nil, nil
}

func setNotificationTypeForRequest(ctx context.Context, plan *budgetAlertNotificationResourceModel, smsPhoneNumbers, emailAddresses, webhookURLs **[]string, diags *diag.Diagnostics) {
	hasSms := !plan.SmsPhoneNumbers.IsNull() && !plan.SmsPhoneNumbers.IsUnknown()
	hasEmail := !plan.EmailAddresses.IsNull() && !plan.EmailAddresses.IsUnknown()
	hasWebhook := !plan.WebhookURLs.IsNull() && !plan.WebhookURLs.IsUnknown()

	if hasSms {
		phoneNumbers := []string{}
		diags.Append(plan.SmsPhoneNumbers.ElementsAs(ctx, &phoneNumbers, false)...)

		if !diags.HasError() {
			*smsPhoneNumbers = &phoneNumbers
		}
	}

	if hasEmail {
		emails := []string{}
		diags.Append(plan.EmailAddresses.ElementsAs(ctx, &emails, false)...)

		if !diags.HasError() {
			*emailAddresses = &emails
		}
	}

	if hasWebhook {
		urls := []string{}
		diags.Append(plan.WebhookURLs.ElementsAs(ctx, &urls, false)...)

		if !diags.HasError() {
			*webhookURLs = &urls
		}
	}
}
