package scaleway

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceScalewayDomainZone() *schema.Resource {
	// Generate datasource schema from resource
	dsSchema := datasourceSchemaFromResourceSchema(resourceScalewayDomainZone().Schema)

	addOptionalFieldsToSchema(dsSchema, "domain", "subdomain")

	return &schema.Resource{
		ReadContext: dataSourceScalewayDomainZoneRead,
		Schema:      dsSchema,
	}
}

func dataSourceScalewayDomainZoneRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	d.SetId(fmt.Sprintf("%s.%s", d.Get("subdomain").(string), d.Get("domain").(string)))

	return resourceScalewayDomainZoneRead(ctx, d, meta)
}
