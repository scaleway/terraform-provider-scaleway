package scaleway

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/datasource"
)

func DataSourceScalewayDomainZone() *schema.Resource {
	// Generate datasource schema from resource
	dsSchema := datasource.SchemaFromResourceSchema(ResourceScalewayDomainZone().Schema)

	datasource.AddOptionalFieldsToSchema(dsSchema, "domain", "subdomain")

	return &schema.Resource{
		ReadContext: dataSourceScalewayDomainZoneRead,
		Schema:      dsSchema,
	}
}

func dataSourceScalewayDomainZoneRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	d.SetId(fmt.Sprintf("%s.%s", d.Get("subdomain").(string), d.Get("domain").(string)))

	return resourceScalewayDomainZoneRead(ctx, d, m)
}
