package domain

import (
	"context"
	_ "embed"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/datasource"
)

//go:embed descriptions/domain_zone_data_source.md
var zoneDataSourceDescription string

func DataSourceZone() *schema.Resource {
	// Generate datasource schema from resource
	dsSchema := datasource.SchemaFromResourceSchema(ResourceZone().SchemaFunc())

	datasource.AddOptionalFieldsToSchema(dsSchema, "domain", "subdomain", "project_id")

	return &schema.Resource{
		Description: zoneDataSourceDescription,
		ReadContext: DataSourceZoneRead,
		Schema:      dsSchema,
	}
}

func DataSourceZoneRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	zoneName := BuildZoneName(d.Get("subdomain").(string), d.Get("domain").(string))
	d.SetId(zoneName)

	return readZoneIntoState(ctx, d, NewDomainAPI(m), zoneName)
}
