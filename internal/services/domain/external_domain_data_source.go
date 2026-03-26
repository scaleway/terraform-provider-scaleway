package domain

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/datasource"
)

func DataSourceExternalDomain() *schema.Resource {
	dsSchema := datasource.SchemaFromResourceSchema(externalDomainSchema())
	delete(dsSchema, "validation_token")

	datasource.AddOptionalFieldsToSchema(dsSchema, "domain")

	return &schema.Resource{
		ReadContext: dataSourceExternalDomainRead,
		Schema:      dsSchema,
	}
}

func dataSourceExternalDomainRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	registrarAPI := NewRegistrarDomainAPI(m)

	d.SetId(d.Get("domain").(string))

	_, diags := readExternalDomainIntoState(ctx, d, registrarAPI)

	return diags
}
