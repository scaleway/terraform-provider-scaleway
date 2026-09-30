package mailbox

import (
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	mailboxsdk "github.com/scaleway/scaleway-sdk-go/api/mailbox/v1alpha1"
)

const (
	defaultDomainTimeout  = 15 * time.Minute
	defaultMailboxTimeout = 15 * time.Minute
	defaultRetryInterval  = 5 * time.Second
)

func localPartFromEmail(email string) string {
	localPart, _, _ := strings.Cut(email, "@")

	return localPart
}

func flattenTime(t *time.Time) types.String {
	if t == nil {
		return types.StringNull()
	}

	return types.StringValue(t.Format(time.RFC3339))
}

func dnsRecordAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"dns_type":  types.StringType,
		"dns_name":  types.StringType,
		"dns_value": types.StringType,
		"status":    types.StringType,
		"level":     types.StringType,
		"error":     types.StringType,
	}
}

func flattenDNSRecords(resp *mailboxsdk.GetDomainRecordsResponse, diags *diag.Diagnostics) types.List {
	elemType := types.ObjectType{AttrTypes: dnsRecordAttrTypes()}

	if resp == nil {
		return types.ListNull(elemType)
	}

	records := []*mailboxsdk.DomainRecord{
		resp.Autoconfig, resp.Autodiscover, resp.Caldav, resp.Carddav, resp.Dkim, resp.Dmarc,
		resp.DomainValidation, resp.Imap, resp.Mx, resp.Pop3, resp.Spf, resp.Submission,
	}

	values := make([]attr.Value, 0)

	for _, rec := range records {
		if rec == nil {
			continue
		}

		errorVal := types.StringNull()
		if rec.Error != nil {
			errorVal = types.StringValue(*rec.Error)
		}

		obj, d := types.ObjectValue(dnsRecordAttrTypes(), map[string]attr.Value{
			"dns_type":  types.StringValue(rec.DNSType.String()),
			"dns_name":  types.StringValue(rec.DNSName),
			"dns_value": types.StringValue(rec.DNSValue),
			"status":    types.StringValue(rec.Status.String()),
			"level":     types.StringValue(rec.Level.String()),
			"error":     errorVal,
		})
		diags.Append(d...)

		values = append(values, obj)
	}

	listVal, d := types.ListValue(elemType, values)
	diags.Append(d...)

	return listVal
}
