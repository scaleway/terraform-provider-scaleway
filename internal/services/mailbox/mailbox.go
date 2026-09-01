package mailbox

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	mailboxsdk "github.com/scaleway/scaleway-sdk-go/api/mailbox/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/identity"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/types"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/verify"
)

// ResourceMailbox manages a single mailbox on a Scaleway Mailbox domain.
// The mailbox is created via the batch endpoint (which supports single-item creation).
// Passwords are write-only: they are accepted on create/update but never read back from the API.
func ResourceMailbox() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceMailboxCreate,
		ReadContext:   resourceMailboxRead,
		UpdateContext: resourceMailboxUpdate,
		DeleteContext: resourceMailboxDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create:  schema.DefaultTimeout(defaultMailboxTimeout),
			Update:  schema.DefaultTimeout(defaultMailboxTimeout),
			Delete:  schema.DefaultTimeout(defaultMailboxTimeout),
			Default: schema.DefaultTimeout(defaultMailboxTimeout),
		},
		SchemaVersion: 0,
		SchemaFunc:    mailboxSchema,
		Identity:      identity.DefaultGlobal(),
	}
}

func mailboxSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"domain_id": {
			Type:             schema.TypeString,
			Required:         true,
			ForceNew:         true,
			Description:      "ID of the mailbox domain to which this mailbox belongs",
			ValidateDiagFunc: verify.IsUUID(),
		},
		"local_part": {
			Type:         schema.TypeString,
			Required:     true,
			ForceNew:     true,
			Description:  "Local part of the email address (the part before the @). Changing this forces a new mailbox to be created.",
			ValidateFunc: validation.StringLenBetween(1, 64),
		},
		"password": {
			Type:        schema.TypeString,
			Required:    true,
			Sensitive:   true,
			Description: "Password for the mailbox. This value is write-only and will never be read back from the API. Changing this triggers an in-place update.",
		},
		"subscription_period": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "Billing subscription period: monthly or yearly",
			ValidateFunc: validation.StringInSlice([]string{
				string(mailboxsdk.MailboxSubscriptionPeriodMonthly),
				string(mailboxsdk.MailboxSubscriptionPeriodYearly),
			}, false),
		},
		"display_name": {
			Type:        schema.TypeString,
			Optional:    true,
			Computed:    true,
			Description: "Display name shown in email clients (e.g. \"John Doe\")",
		},
		"recovery_email": {
			Type:        schema.TypeString,
			Optional:    true,
			Computed:    true,
			Description: "Recovery email address used for password resets",
		},
		"email": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "Full email address of the mailbox (local_part@domain)",
		},
		"status": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "Mailbox status: creating, waiting_payment, waiting_domain, ready, deletion_scheduled, locked, renewing, deleting, restoring, payment_failed",
		},
		"subscription_period_started_at": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "Start date of the current subscription period (RFC 3339 format)",
		},
		"next_subscription_period": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "Next subscription renewal period (monthly, yearly, or canceled)",
		},
		"next_subscription_period_starts_at": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "Date when the next subscription period starts (RFC 3339 format)",
		},
		"deletion_scheduled_at": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "Date of the unrecoverable mailbox deletion, set when status is deletion_scheduled (RFC 3339 format)",
		},
		"created_at": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "Date and time of mailbox creation (RFC 3339 format)",
		},
		"updated_at": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "Date and time of last update (RFC 3339 format)",
		},
	}
}

func resourceMailboxCreate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	api := newMailboxAPI(m)

	period := mailboxsdk.MailboxSubscriptionPeriod(d.Get("subscription_period").(string))

	resp, err := api.BatchCreateMailboxes(&mailboxsdk.BatchCreateMailboxesRequest{
		DomainID:           d.Get("domain_id").(string),
		SubscriptionPeriod: period,
		Mailboxes: []*mailboxsdk.BatchCreateMailboxesRequestMailboxParameters{
			{
				LocalPart:     d.Get("local_part").(string),
				Password:      d.Get("password").(string),
				DisplayName:   types.ExpandStringPtr(d.Get("display_name")),
				RecoveryEmail: types.ExpandStringPtr(d.Get("recovery_email")),
			},
		},
	}, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	if len(resp.Mailboxes) == 0 {
		return diag.Errorf("mailbox creation returned no mailboxes")
	}

	mb := resp.Mailboxes[0]

	if err := identity.SetGlobalIdentity(d, mb.ID); err != nil {
		return diag.FromErr(err)
	}

	timeout := d.Timeout(schema.TimeoutCreate)
	mb, err = api.WaitForMailbox(&mailboxsdk.WaitForMailboxRequest{
		MailboxID: mb.ID,
		Timeout:   &timeout,
	}, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	return setMailboxState(d, mb)
}

func resourceMailboxRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	api := newMailboxAPI(m)

	mb, err := api.GetMailbox(&mailboxsdk.GetMailboxRequest{MailboxID: d.Id()}, scw.WithContext(ctx))
	if err != nil {
		if httperrors.Is404(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	return setMailboxState(d, mb)
}

func resourceMailboxUpdate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	api := newMailboxAPI(m)

	req := &mailboxsdk.UpdateMailboxRequest{MailboxID: d.Id()}
	needsUpdate := false

	if d.HasChange("display_name") {
		req.DisplayName = types.ExpandStringPtr(d.Get("display_name"))
		needsUpdate = true
	}

	if d.HasChange("recovery_email") {
		req.RecoveryEmail = types.ExpandStringPtr(d.Get("recovery_email"))
		needsUpdate = true
	}

	if d.HasChange("subscription_period") {
		period := mailboxsdk.MailboxSubscriptionPeriod(d.Get("subscription_period").(string))
		req.SubscriptionPeriod = &period
		needsUpdate = true
	}

	if d.HasChange("password") {
		req.NewPassword = types.ExpandStringPtr(d.Get("password"))
		needsUpdate = true
	}

	if !needsUpdate {
		return resourceMailboxRead(ctx, d, m)
	}

	mb, err := api.UpdateMailbox(req, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	timeout := d.Timeout(schema.TimeoutUpdate)
	mb, err = api.WaitForMailbox(&mailboxsdk.WaitForMailboxRequest{
		MailboxID: mb.ID,
		Timeout:   &timeout,
	}, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	return setMailboxState(d, mb)
}

func resourceMailboxDelete(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	api := newMailboxAPI(m)

	_, err := api.DeleteMailbox(&mailboxsdk.DeleteMailboxRequest{MailboxID: d.Id()}, scw.WithContext(ctx))
	if err != nil {
		if httperrors.Is404(err) {
			return nil
		}
		return diag.FromErr(err)
	}

	timeout := d.Timeout(schema.TimeoutDelete)
	mb, err := api.WaitForMailbox(&mailboxsdk.WaitForMailboxRequest{
		MailboxID: d.Id(),
		Timeout:   &timeout,
	}, scw.WithContext(ctx))
	if err != nil {
		if httperrors.Is404(err) {
			return nil
		}
		return diag.FromErr(err)
	}

	if mb.Status == mailboxsdk.MailboxStatusDeletionScheduled {
		return nil
	}

	return nil
}

// setMailboxState writes all API-returned mailbox fields into the Terraform state.
func setMailboxState(d *schema.ResourceData, mb *mailboxsdk.Mailbox) diag.Diagnostics {
	_ = d.Set("domain_id", mb.DomainID)
	_ = d.Set("email", mb.Email)
	_ = d.Set("display_name", mb.DisplayName)
	_ = d.Set("recovery_email", mb.RecoveryEmail)
	_ = d.Set("status", mb.Status.String())
	_ = d.Set("subscription_period", mb.SubscriptionPeriod.String())
	_ = d.Set("subscription_period_started_at", types.FlattenTime(mb.SubscriptionPeriodStartedAt))
	_ = d.Set("next_subscription_period", mb.NextSubscriptionPeriod.String())
	_ = d.Set("next_subscription_period_starts_at", types.FlattenTime(mb.NextSubscriptionPeriodStartsAt))
	_ = d.Set("deletion_scheduled_at", types.FlattenTime(mb.DeletionScheduledAt))
	_ = d.Set("created_at", types.FlattenTime(mb.CreatedAt))
	_ = d.Set("updated_at", types.FlattenTime(mb.UpdatedAt))

	return nil
}

// readMailboxIntoState fetches the mailbox and writes it into d.
func readMailboxIntoState(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	api := newMailboxAPI(m)

	mb, err := api.GetMailbox(&mailboxsdk.GetMailboxRequest{MailboxID: d.Id()}, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	return setMailboxState(d, mb)
}
