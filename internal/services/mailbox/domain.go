package mailbox

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	mailboxsdk "github.com/scaleway/scaleway-sdk-go/api/mailbox/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/identity"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/account"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/types"
)

// ResourceDomain manages a mailbox service domain.
// A domain must be created before mailboxes can be provisioned under it.
// After creation, DNS records are available as computed attributes and must be
// configured in your DNS zone. The domain status will reflect validation progress.
func ResourceDomain() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceMailboxDomainCreate,
		ReadContext:   resourceMailboxDomainRead,
		DeleteContext: resourceMailboxDomainDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create:  schema.DefaultTimeout(defaultDomainTimeout),
			Delete:  schema.DefaultTimeout(defaultDomainTimeout),
			Default: schema.DefaultTimeout(defaultDomainTimeout),
		},
		SchemaVersion: 0,
		SchemaFunc:    domainSchema,
		Identity:      identity.DefaultGlobal(),
	}
}

func domainSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"name": {
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
			Description: "Fully qualified domain name (e.g. mail.example.com)",
		},
		"project_id": account.ProjectIDSchema(),
		"status": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "Domain status: creating, waiting_validation, validating, validation_failed, provisioning, ready, deleting",
		},
		"mailbox_total_count": {
			Type:        schema.TypeInt,
			Computed:    true,
			Description: "Number of mailboxes provisioned on this domain",
		},
		"webmail_url": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "URL of the webmail interface",
		},
		"imap_url": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "IMAP server URL for email clients",
		},
		"jmap_url": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "JMAP server URL for email clients",
		},
		"pop3_url": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "POP3 server URL for email clients",
		},
		"smtp_url": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "SMTP server URL for email clients",
		},
		"dns_records": {
			Type:        schema.TypeList,
			Computed:    true,
			Description: "DNS records that must be configured in your DNS zone to validate the domain and enable mailbox features. Required records must be set before the domain can be used.",
			Elem:        dnsRecordSchema(),
		},
		"created_at": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "Date and time of domain creation (RFC 3339 format)",
		},
		"updated_at": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "Date and time of last update (RFC 3339 format)",
		},
	}
}

func resourceMailboxDomainCreate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	api := newMailboxAPI(m)

	domain, err := api.CreateDomain(&mailboxsdk.CreateDomainRequest{
		ProjectID: d.Get("project_id").(string),
		Name:      d.Get("name").(string),
	}, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	if err := identity.SetGlobalIdentity(d, domain.ID); err != nil {
		return diag.FromErr(err)
	}

	timeout := d.Timeout(schema.TimeoutCreate)
	domain, err = api.WaitForDomain(&mailboxsdk.WaitForDomainRequest{
		DomainID: domain.ID,
		Timeout:  &timeout,
	}, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	return setDomainState(ctx, d, api, domain)
}

func resourceMailboxDomainRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	api := newMailboxAPI(m)

	domain, err := api.GetDomain(&mailboxsdk.GetDomainRequest{DomainID: d.Id()}, scw.WithContext(ctx))
	if err != nil {
		if httperrors.Is404(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	return setDomainState(ctx, d, api, domain)
}

func resourceMailboxDomainDelete(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	api := newMailboxAPI(m)

	_, err := api.DeleteDomain(&mailboxsdk.DeleteDomainRequest{DomainID: d.Id()}, scw.WithContext(ctx))
	if err != nil {
		if httperrors.Is404(err) {
			return nil
		}
		return diag.FromErr(err)
	}

	timeout := d.Timeout(schema.TimeoutDelete)
	_, err = api.WaitForDomain(&mailboxsdk.WaitForDomainRequest{
		DomainID: d.Id(),
		Timeout:  &timeout,
	}, scw.WithContext(ctx))
	if err != nil && !httperrors.Is404(err) {
		return diag.FromErr(err)
	}

	return nil
}

// setDomainState writes all API-returned domain fields into the Terraform state.
func setDomainState(ctx context.Context, d *schema.ResourceData, api *mailboxsdk.API, domain *mailboxsdk.Domain) diag.Diagnostics {
	_ = d.Set("name", domain.Name)
	_ = d.Set("project_id", domain.ProjectID)
	_ = d.Set("status", domain.Status.String())
	_ = d.Set("mailbox_total_count", int(domain.MailboxTotalCount))
	_ = d.Set("webmail_url", domain.WebmailURL)
	_ = d.Set("imap_url", domain.ImapURL)
	_ = d.Set("jmap_url", domain.JmapURL)
	_ = d.Set("pop3_url", domain.Pop3URL)
	_ = d.Set("smtp_url", domain.SMTPURL)
	_ = d.Set("created_at", types.FlattenTime(domain.CreatedAt))
	_ = d.Set("updated_at", types.FlattenTime(domain.UpdatedAt))

	records, err := api.GetDomainRecords(&mailboxsdk.GetDomainRecordsRequest{DomainID: domain.ID}, scw.WithContext(ctx))
	if err != nil && !httperrors.Is404(err) {
		return diag.FromErr(err)
	}

	if records != nil {
		_ = d.Set("dns_records", flattenDNSRecords(records))
	}

	return nil
}
