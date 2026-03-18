package domain

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	domain "github.com/scaleway/scaleway-sdk-go/api/domain/v2beta1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/datasource"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/identity"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/types"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/verify"
)

func DataSourceRecord() *schema.Resource {
	// Generate datasource schema from resource
	dsSchema := datasource.SchemaFromResourceSchema(ResourceRecord().SchemaFunc())

	// Set 'Optional' schema elements
	datasource.AddOptionalFieldsToSchema(dsSchema, "dns_zone", "name", "type", "data", "project_id")

	dsSchema["name"].ConflictsWith = []string{"record_id"}
	dsSchema["type"].ConflictsWith = []string{"record_id"}
	dsSchema["data"].ConflictsWith = []string{"record_id"}
	dsSchema["record_id"] = &schema.Schema{
		Type:             schema.TypeString,
		Optional:         true,
		Description:      "The ID of the record (UUID or dns_zone/uuid format)",
		ValidateDiagFunc: verify.IsUUIDorUUIDWithLocality(),
		ConflictsWith:    []string{"name", "type", "data"},
	}

	return &schema.Resource{
		ReadContext: DataSourceRecordRead,
		Schema:      dsSchema,
		Identity:    identity.WrapSchemaMap(recordIdentitySchema()),
	}
}

func DataSourceRecordRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	domainAPI := NewDomainAPI(m)
	dnsZone := d.Get("dns_zone").(string)

	recordIDRaw, ok := d.GetOk("record_id")
	if !ok { // Get Record by dns_zone, name, type and data.
		res, err := domainAPI.ListDNSZoneRecords(&domain.ListDNSZoneRecordsRequest{
			DNSZone:   dnsZone,
			Name:      d.Get("name").(string),
			Type:      domain.RecordType(d.Get("type").(string)),
			ProjectID: types.ExpandStringPtr(d.Get("project_id")),
		}, scw.WithContext(ctx), scw.WithAllPages())
		if err != nil {
			return diag.FromErr(err)
		}

		if len(res.Records) == 0 {
			return diag.FromErr(fmt.Errorf("no record found with the type %s", d.Get("type")))
		}

		var record *domain.Record

		for i := range res.Records {
			if res.Records[i].Data == d.Get("data").(string) {
				if record != nil {
					return diag.FromErr(fmt.Errorf("more than one record found with this name: %s, type: %s and data: %s", d.Get("name"), d.Get("type"), d.Get("data")))
				}

				record = res.Records[i]
			}
		}

		if record == nil {
			return diag.FromErr(fmt.Errorf("no record found with the type this name: %s, type: %s and data: %s", d.Get("name"), d.Get("type"), d.Get("data")))
		}

		return readRecordIntoState(ctx, d, domainAPI, dnsZone, record.ID)
	}

	recordIDStr := recordIDRaw.(string)

	var recordID string

	if zone, id, err := locality.ParseLocalizedID(recordIDStr); err == nil {
		dnsZone = zone
		recordID = id
	} else {
		recordID = locality.ExpandID(recordIDStr)
	}

	return readRecordIntoState(ctx, d, domainAPI, dnsZone, recordID)
}
