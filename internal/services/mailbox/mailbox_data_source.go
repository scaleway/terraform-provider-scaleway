package mailbox

import (
	"context"
	_ "embed"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	mailboxsdk "github.com/scaleway/scaleway-sdk-go/api/mailbox/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/verify"
)

var (
	_ datasource.DataSource              = (*MailboxDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*MailboxDataSource)(nil)
)

func NewMailboxDataSource() datasource.DataSource {
	return &MailboxDataSource{}
}

type MailboxDataSource struct {
	api  *mailboxsdk.API
	meta *meta.Meta
}

type mailboxDataSourceModel struct {
	ID                             types.String `tfsdk:"id"`
	MailboxID                      types.String `tfsdk:"mailbox_id"`
	DomainID                       types.String `tfsdk:"domain_id"`
	LocalPart                      types.String `tfsdk:"local_part"`
	SubscriptionPeriod             types.String `tfsdk:"subscription_period"`
	Email                          types.String `tfsdk:"email"`
	Status                         types.String `tfsdk:"status"`
	SubscriptionPeriodStartedAt    types.String `tfsdk:"subscription_period_started_at"`
	NextSubscriptionPeriod         types.String `tfsdk:"next_subscription_period"`
	NextSubscriptionPeriodStartsAt types.String `tfsdk:"next_subscription_period_starts_at"`
	DeletionScheduledAt            types.String `tfsdk:"deletion_scheduled_at"`
	CreatedAt                      types.String `tfsdk:"created_at"`
	UpdatedAt                      types.String `tfsdk:"updated_at"`
}

func (d *MailboxDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_mailbox_mailbox"
}

//go:embed descriptions/mailbox_data_source.md
var mailboxDataSourceDescription string

func (d *MailboxDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: mailboxDataSourceDescription,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "ID of the mailbox",
			},
			"mailbox_id": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "UUID of the mailbox. Exactly one of `mailbox_id` or `email` must be specified.",
				Validators: []validator.String{
					verify.IsStringUUID(),
					stringvalidator.ExactlyOneOf(
						path.MatchRoot("mailbox_id"),
						path.MatchRoot("email"),
					),
				},
			},
			"email": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Full email address of the mailbox. Exactly one of `mailbox_id` or `email` must be specified.",
				Validators: []validator.String{
					stringvalidator.ExactlyOneOf(
						path.MatchRoot("mailbox_id"),
						path.MatchRoot("email"),
					),
				},
			},
			"domain_id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "ID of the mailbox domain",
			},
			"local_part": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Local part of the email address",
			},
			"subscription_period": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Billing subscription period",
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
				MarkdownDescription: "Next subscription renewal period",
			},
			"next_subscription_period_starts_at": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Date when the next subscription period starts (RFC 3339 format)",
			},
			"deletion_scheduled_at": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Scheduled deletion date when status is deletion_scheduled (RFC 3339 format)",
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

func (d *MailboxDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
	d.api = mailboxsdk.NewAPI(d.meta.ScwClient())
}

func (d *MailboxDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config mailboxDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)

	if resp.Diagnostics.HasError() {
		return
	}

	hasID := !config.MailboxID.IsNull() && !config.MailboxID.IsUnknown() && config.MailboxID.ValueString() != ""
	hasEmail := !config.Email.IsNull() && !config.Email.IsUnknown() && config.Email.ValueString() != ""

	var mailboxID string

	switch {
	case hasID:
		mailboxID = config.MailboxID.ValueString()
	case hasEmail:
		email := config.Email.ValueString()

		listResp, err := d.api.ListMailboxes(&mailboxsdk.ListMailboxesRequest{
			Search: new(email),
		}, scw.WithAllPages(), scw.WithContext(ctx))
		if err != nil {
			resp.Diagnostics.AddError("Failed to list mailboxes", err.Error())

			return
		}

		for _, mb := range listResp.Mailboxes {
			if mb.Email != email {
				continue
			}

			if mailboxID != "" {
				resp.Diagnostics.AddError(
					"Multiple mailboxes found",
					fmt.Sprintf("found multiple mailboxes with email %q", email),
				)

				return
			}

			mailboxID = mb.ID
		}

		if mailboxID == "" {
			resp.Diagnostics.AddError(
				"Mailbox not found",
				fmt.Sprintf("no mailbox found with email %q", email),
			)

			return
		}
	default:
		resp.Diagnostics.AddError(
			"Missing lookup key",
			"One of mailbox_id or email must be provided",
		)

		return
	}

	mb, err := d.api.GetMailbox(&mailboxsdk.GetMailboxRequest{MailboxID: mailboxID}, scw.WithContext(ctx))
	if err != nil {
		resp.Diagnostics.AddError("Failed to get mailbox", err.Error())

		return
	}

	state := mailboxDataSourceModel{
		ID:                             types.StringValue(mb.ID),
		MailboxID:                      types.StringValue(mb.ID),
		DomainID:                       types.StringValue(mb.DomainID),
		LocalPart:                      types.StringValue(localPartFromEmail(mb.Email)),
		SubscriptionPeriod:             types.StringValue(mb.SubscriptionPeriod.String()),
		Email:                          types.StringValue(mb.Email),
		Status:                         types.StringValue(mb.Status.String()),
		SubscriptionPeriodStartedAt:    flattenTime(mb.SubscriptionPeriodStartedAt),
		NextSubscriptionPeriod:         types.StringValue(mb.NextSubscriptionPeriod.String()),
		NextSubscriptionPeriodStartsAt: flattenTime(mb.NextSubscriptionPeriodStartsAt),
		DeletionScheduledAt:            flattenTime(mb.DeletionScheduledAt),
		CreatedAt:                      flattenTime(mb.CreatedAt),
		UpdatedAt:                      flattenTime(mb.UpdatedAt),
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
