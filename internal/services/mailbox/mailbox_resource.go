package mailbox

import (
	"context"
	_ "embed"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	mailboxsdk "github.com/scaleway/scaleway-sdk-go/api/mailbox/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/identity/framework"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
	providertypes "github.com/scaleway/terraform-provider-scaleway/v2/internal/types"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/verify"
)

var (
	_ resource.Resource                = (*MailboxResource)(nil)
	_ resource.ResourceWithConfigure   = (*MailboxResource)(nil)
	_ resource.ResourceWithImportState = (*MailboxResource)(nil)
	_ resource.ResourceWithIdentity    = (*MailboxResource)(nil)
)

func NewMailboxResource() resource.Resource {
	return &MailboxResource{}
}

type MailboxResource struct {
	api  *mailboxsdk.API
	meta *meta.Meta
}

type mailboxResourceModel struct {
	SubscriptionPeriod             types.String `tfsdk:"subscription_period"`
	NextSubscriptionPeriod         types.String `tfsdk:"next_subscription_period"`
	LocalPart                      types.String `tfsdk:"local_part"`
	Password                       types.String `tfsdk:"password"`
	PasswordWo                     types.String `tfsdk:"password_wo"`
	UpdatedAt                      types.String `tfsdk:"updated_at"`
	DomainID                       types.String `tfsdk:"domain_id"`
	Status                         types.String `tfsdk:"status"`
	ID                             types.String `tfsdk:"id"`
	SubscriptionPeriodStartedAt    types.String `tfsdk:"subscription_period_started_at"`
	Email                          types.String `tfsdk:"email"`
	NextSubscriptionPeriodStartsAt types.String `tfsdk:"next_subscription_period_starts_at"`
	DeletionScheduledAt            types.String `tfsdk:"deletion_scheduled_at"`
	CreatedAt                      types.String `tfsdk:"created_at"`
	PasswordWoVersion              types.Int64  `tfsdk:"password_wo_version"`
}

type mailboxResourceIdentityModel = framework.GlobalIdentity

func (r *MailboxResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_mailbox_mailbox"
}

//go:embed descriptions/mailbox_resource.md
var mailboxResourceDescription string

func (r *MailboxResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: mailboxResourceDescription,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "ID of the mailbox",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"domain_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "ID of the mailbox domain to which this mailbox belongs",
				Validators: []validator.String{
					verify.IsStringUUID(),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"local_part": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Local part of the email address (the part before the @). Changing this forces a new mailbox.",
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 64),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"password": schema.StringAttribute{
				Optional:            true,
				Sensitive:           true,
				MarkdownDescription: "Password for the mailbox. Only one of `password` or `password_wo` should be specified.",
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
					stringvalidator.ExactlyOneOf(
						path.MatchRoot("password"),
						path.MatchRoot("password_wo"),
					),
				},
			},
			"password_wo": schema.StringAttribute{
				Optional:            true,
				WriteOnly:           true,
				MarkdownDescription: "Password in [write-only](https://registry.terraform.io/providers/scaleway/scaleway/latest/docs/guides/using-write-only-arguments) mode. Only one of `password` or `password_wo` should be specified. To update `password_wo`, also update `password_wo_version`.",
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
					stringvalidator.ExactlyOneOf(
						path.MatchRoot("password"),
						path.MatchRoot("password_wo"),
					),
					stringvalidator.AlsoRequires(path.MatchRoot("password_wo_version")),
				},
			},
			"password_wo_version": schema.Int64Attribute{
				Optional:            true,
				MarkdownDescription: "Version of the write-only password. Bump this value together with `password_wo` to rotate the password.",
				Validators: []validator.Int64{
					int64validator.AlsoRequires(path.MatchRoot("password_wo")),
				},
			},
			"subscription_period": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Billing subscription period: monthly or yearly. Updating this value changes the next renewal period.",
				Validators: []validator.String{
					// canceled is an API-only value; create/update accept monthly|yearly only.
					stringvalidator.OneOf(
						string(mailboxsdk.MailboxSubscriptionPeriodMonthly),
						string(mailboxsdk.MailboxSubscriptionPeriodYearly),
					),
				},
			},
			"email": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Full email address of the mailbox (local_part@domain)",
			},
			"status": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Mailbox status",
			},
			"subscription_period_started_at": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Start date of the current subscription period (RFC 3339 format)",
			},
			"next_subscription_period": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Next subscription renewal period (monthly, yearly, or canceled)",
			},
			"next_subscription_period_starts_at": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Date when the next subscription period starts (RFC 3339 format)",
			},
			"deletion_scheduled_at": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Date of the unrecoverable mailbox deletion when status is deletion_scheduled (RFC 3339 format)",
			},
			"created_at": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Date and time of mailbox creation (RFC 3339 format)",
			},
			"updated_at": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Date and time of last update (RFC 3339 format)",
			},
		},
	}
}

func (r *MailboxResource) IdentitySchema(_ context.Context, _ resource.IdentitySchemaRequest, resp *resource.IdentitySchemaResponse) {
	resp.IdentitySchema = framework.DefaultGlobal()
}

func (r *MailboxResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
	r.api = mailboxsdk.NewAPI(r.meta.ScwClient())
}

func (r *MailboxResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan mailboxResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var config mailboxResourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)

	if resp.Diagnostics.HasError() {
		return
	}

	password := plan.Password.ValueString()
	if !config.PasswordWo.IsNull() && !config.PasswordWo.IsUnknown() && config.PasswordWo.ValueString() != "" {
		password = config.PasswordWo.ValueString()
	}

	period := mailboxsdk.MailboxSubscriptionPeriod(plan.SubscriptionPeriod.ValueString())

	createResp, err := r.api.BatchCreateMailboxes(&mailboxsdk.BatchCreateMailboxesRequest{
		DomainID:           plan.DomainID.ValueString(),
		SubscriptionPeriod: period,
		Mailboxes: []*mailboxsdk.BatchCreateMailboxesRequestMailboxParameters{
			{
				LocalPart: plan.LocalPart.ValueString(),
				Password:  password,
			},
		},
	}, scw.WithContext(ctx))
	if err != nil {
		resp.Diagnostics.AddError("Failed to create mailbox", err.Error())

		return
	}

	if len(createResp.Mailboxes) == 0 {
		resp.Diagnostics.AddError("Failed to create mailbox", "API returned no mailboxes")

		return
	}

	mb := createResp.Mailboxes[0]

	mb, err = waitForMailbox(ctx, r.api, mb.ID, defaultMailboxTimeout)
	if err != nil {
		resp.Diagnostics.AddError("Failed waiting for mailbox", err.Error())

		return
	}

	state := convertMailboxToState(mb, plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	resp.Diagnostics.Append(resp.Identity.Set(ctx, framework.SetGlobalIdentity(mb.ID))...)
}

func (r *MailboxResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var (
		state    mailboxResourceModel
		identity mailboxResourceIdentityModel
	)

	resp.Diagnostics.Append(req.Identity.Get(ctx, &identity)...)
	identityAvailable := !resp.Diagnostics.HasError() && !identity.ID.IsNull() && !identity.ID.IsUnknown()

	if !identityAvailable && resp.Diagnostics.HasError() {
		resp.Diagnostics = nil
	}

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	mailboxID := state.ID.ValueString()
	if identityAvailable {
		mailboxID = identity.ID.ValueString()
	}

	if mailboxID == "" {
		resp.Diagnostics.AddError("Mailbox ID not set", "Cannot refresh mailbox without an ID")

		return
	}

	mb, err := r.api.GetMailbox(&mailboxsdk.GetMailboxRequest{MailboxID: mailboxID}, scw.WithContext(ctx))
	if err != nil {
		if httperrors.Is404(err) {
			resp.State.RemoveResource(ctx)

			return
		}

		resp.Diagnostics.AddError("Failed to get mailbox", err.Error())

		return
	}

	state = convertMailboxToState(mb, state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	resp.Diagnostics.Append(resp.Identity.Set(ctx, framework.SetGlobalIdentity(mb.ID))...)
}

func (r *MailboxResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var (
		plan  mailboxResourceModel
		state mailboxResourceModel
	)

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var config mailboxResourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)

	if resp.Diagnostics.HasError() {
		return
	}

	updateReq := &mailboxsdk.UpdateMailboxRequest{MailboxID: plan.ID.ValueString()}
	needsUpdate := false

	if !plan.SubscriptionPeriod.Equal(state.SubscriptionPeriod) {
		period := mailboxsdk.MailboxSubscriptionPeriod(plan.SubscriptionPeriod.ValueString())
		updateReq.SubscriptionPeriod = &period
		needsUpdate = true
	}

	var newPassword *string

	if !plan.PasswordWoVersion.Equal(state.PasswordWoVersion) {
		if config.PasswordWo.IsNull() || config.PasswordWo.IsUnknown() || config.PasswordWo.ValueString() == "" {
			resp.Diagnostics.AddError(
				"Missing password_wo",
				"password_wo_version changed but password_wo is empty; set both to rotate the password.",
			)

			return
		}

		newPassword = providertypes.ExpandStringPtr(config.PasswordWo.ValueString())
	}

	if newPassword == nil && !plan.Password.Equal(state.Password) {
		if plan.Password.IsNull() || plan.Password.ValueString() == "" {
			resp.Diagnostics.AddError(
				"Missing password",
				"password changed but the new value is empty.",
			)

			return
		}

		newPassword = providertypes.ExpandStringPtr(plan.Password.ValueString())
	}

	if newPassword != nil {
		updateReq.NewPassword = newPassword
		needsUpdate = true
	}

	if needsUpdate {
		_, err := r.api.UpdateMailbox(updateReq, scw.WithContext(ctx))
		if err != nil {
			resp.Diagnostics.AddError("Failed to update mailbox", err.Error())

			return
		}

		mb, err := waitForMailbox(ctx, r.api, plan.ID.ValueString(), defaultMailboxTimeout)
		if err != nil {
			resp.Diagnostics.AddError("Failed waiting for mailbox after update", err.Error())

			return
		}

		newState := convertMailboxToState(mb, plan)
		resp.Diagnostics.Append(resp.State.Set(ctx, &newState)...)
		resp.Diagnostics.Append(resp.Identity.Set(ctx, framework.SetGlobalIdentity(mb.ID))...)

		return
	}

	mb, err := r.api.GetMailbox(&mailboxsdk.GetMailboxRequest{MailboxID: plan.ID.ValueString()}, scw.WithContext(ctx))
	if err != nil {
		resp.Diagnostics.AddError("Failed to get mailbox after update", err.Error())

		return
	}

	newState := convertMailboxToState(mb, plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &newState)...)
	resp.Diagnostics.Append(resp.Identity.Set(ctx, framework.SetGlobalIdentity(mb.ID))...)
}

func (r *MailboxResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state mailboxResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.api.DeleteMailbox(&mailboxsdk.DeleteMailboxRequest{MailboxID: state.ID.ValueString()}, scw.WithContext(ctx))
	if err != nil {
		if httperrors.Is404(err) {
			return
		}

		resp.Diagnostics.AddError("Failed to delete mailbox", err.Error())

		return
	}

	_, err = waitForMailboxDeleted(ctx, r.api, state.ID.ValueString(), defaultMailboxTimeout)
	if err != nil && !httperrors.Is404(err) {
		resp.Diagnostics.AddError("Failed waiting for mailbox deletion", err.Error())
	}
}

func (r *MailboxResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughWithIdentity(ctx, path.Root("id"), path.Root("id"), req, resp)
}

// convertMailboxToState maps API fields into state, preserving password fields from prior state/plan.
func convertMailboxToState(mb *mailboxsdk.Mailbox, prior mailboxResourceModel) mailboxResourceModel {
	subscriptionPeriod := mb.SubscriptionPeriod.String()
	// UpdateMailbox may only schedule the next period; keep the configured value when it
	// matches next_subscription_period to avoid a perpetual plan diff.
	if !prior.SubscriptionPeriod.IsNull() && !prior.SubscriptionPeriod.IsUnknown() {
		desired := prior.SubscriptionPeriod.ValueString()
		if desired != "" && desired == mb.NextSubscriptionPeriod.String() && desired != subscriptionPeriod {
			subscriptionPeriod = desired
		}
	}

	return mailboxResourceModel{
		ID:                             types.StringValue(mb.ID),
		DomainID:                       types.StringValue(mb.DomainID),
		LocalPart:                      types.StringValue(localPartFromEmail(mb.Email)),
		Password:                       prior.Password,
		PasswordWo:                     types.StringNull(),
		PasswordWoVersion:              prior.PasswordWoVersion,
		SubscriptionPeriod:             types.StringValue(subscriptionPeriod),
		Email:                          types.StringValue(mb.Email),
		Status:                         types.StringValue(mb.Status.String()),
		SubscriptionPeriodStartedAt:    flattenTime(mb.SubscriptionPeriodStartedAt),
		NextSubscriptionPeriod:         types.StringValue(mb.NextSubscriptionPeriod.String()),
		NextSubscriptionPeriodStartsAt: flattenTime(mb.NextSubscriptionPeriodStartsAt),
		DeletionScheduledAt:            flattenTime(mb.DeletionScheduledAt),
		CreatedAt:                      flattenTime(mb.CreatedAt),
		UpdatedAt:                      flattenTime(mb.UpdatedAt),
	}
}
