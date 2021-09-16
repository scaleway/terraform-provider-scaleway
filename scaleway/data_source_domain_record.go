package scaleway

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	domain "github.com/scaleway/scaleway-sdk-go/api/domain/v2beta1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

func dataSourceScalewayDomainRecord() *schema.Resource {
	// Generate datasource schema from resource
	dsSchema := datasourceSchemaFromResourceSchema(resourceScalewayDomainRecord().Schema)

	// Set 'Optional' schema elements
	addOptionalFieldsToSchema(dsSchema, "dns_zone", "name", "type", "data")

	dsSchema["name"].ConflictsWith = []string{"record_id"}
	dsSchema["type"].ConflictsWith = []string{"record_id"}
	dsSchema["data"].ConflictsWith = []string{"record_id"}
	dsSchema["record_id"] = &schema.Schema{
		Type:          schema.TypeString,
		Optional:      true,
		Description:   "The ID of the record",
		ValidateFunc:  validationUUID(),
		ConflictsWith: []string{"name", "type", "data"},
	}

	return &schema.Resource{
		ReadContext: dataSourceScalewayDomainRecordRead,
		Schema:      dsSchema,
	}
}

func dataSourceScalewayDomainRecordRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	domainAPI := newDomainAPI(meta)

	recordID, ok := d.GetOk("record_id")
	if !ok { // Get Record by dns_zone, name, type and data.
		res, err := domainAPI.ListDNSZoneRecords(&domain.ListDNSZoneRecordsRequest{
			DNSZone:   d.Get("dns_zone").(string),
			Name:      d.Get("name").(string),
			Type:      domain.RecordType(d.Get("type").(string)),
			ProjectID: expandStringPtr(d.Get("project_id")),
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
		recordID = record.ID
	}

	d.SetId(fmt.Sprintf("%s/%s", d.Get("dns_zone"), recordID.(string)))
	return resourceScalewayDomainRecordRead(ctx, d, meta)
}
