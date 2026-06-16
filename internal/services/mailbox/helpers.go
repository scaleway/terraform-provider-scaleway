package mailbox

import (
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	mailboxsdk "github.com/scaleway/scaleway-sdk-go/api/mailbox/v1alpha1"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/types"
)

const (
	defaultDomainTimeout  = 5 * time.Minute
	defaultMailboxTimeout = 5 * time.Minute
)

func newMailboxAPI(m any) *mailboxsdk.API {
	return mailboxsdk.NewAPI(meta.ExtractScwClient(m))
}

// flattenDNSRecords converts a GetDomainRecordsResponse into a Terraform-compatible list.
func flattenDNSRecords(resp *mailboxsdk.GetDomainRecordsResponse) []any {
	if resp == nil {
		return nil
	}

	records := []*mailboxsdk.DomainRecord{
		resp.Autoconfig, resp.Autodiscover, resp.Caldav, resp.Carddav, resp.Dkim, resp.Dmarc,
		resp.DomainValidation, resp.Imap, resp.Jmap, resp.Mx, resp.Pop3, resp.Spf, resp.Submission,
	}

	result := make([]any, 0)

	for _, rec := range records {
		if rec == nil {
			continue
		}
		m := map[string]any{
			"dns_type":  rec.DNSType.String(),
			"dns_name":  rec.DNSName,
			"dns_value": rec.DNSValue,
			"status":    rec.Status.String(),
			"level":     rec.Level.String(),
			"error":     types.FlattenStringPtr(rec.Error),
		}
		result = append(result, m)
	}

	return result
}

// dnsRecordSchema returns the schema for a single DNS record entry.
func dnsRecordSchema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"dns_type": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "DNS record type (e.g. TXT, MX, CNAME)",
			},
			"dns_name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Fully qualified DNS name for this record",
			},
			"dns_value": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "DNS record value to set",
			},
			"status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Validation status of this record (valid, invalid, not_found, validating)",
			},
			"level": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Requirement level (required, recommended, optional)",
			},
			"error": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Error detail when the record is invalid",
			},
		},
	}
}
