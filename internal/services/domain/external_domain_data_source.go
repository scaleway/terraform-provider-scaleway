package domain

import (
	"context"
	_ "embed"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/datasource"
)

//go:embed descriptions/domain_external_data_source.md
var externalDomainDataSourceDescription string

func DataSourceExternalDomain() *schema.Resource {
	dsSchema := datasource.SchemaFromResourceSchema(externalDomainSchema())
	delete(dsSchema, "validation_token")

	datasource.AddOptionalFieldsToSchema(dsSchema, "domain")

	return &schema.Resource{
		Description: externalDomainDataSourceDescription,
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
