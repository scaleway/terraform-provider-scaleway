package domain

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	domain "github.com/scaleway/scaleway-sdk-go/api/domain/v2beta1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/datasource"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/types"
)

func DataSourceExternalDomain() *schema.Resource {
	dsSchema := datasource.SchemaFromResourceSchema(externalDomainSchema())
	datasource.FixDatasourceSchemaFlags(dsSchema, true, "domain")
	datasource.AddOptionalFieldsToSchema(dsSchema, "project_id")

	return &schema.Resource{
		ReadContext: dataSourceExternalDomainRead,
		Schema:      dsSchema,
	}
}

func dataSourceExternalDomainRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	domainName := d.Get("domain").(string)
	projectID := types.ExpandStringPtr(d.Get("project_id"))

	registrarAPI := NewRegistrarDomainAPI(m)

	resp, err := registrarAPI.GetDomain(&domain.RegistrarAPIGetDomainRequest{Domain: domainName}, scw.WithContext(ctx))
	if err != nil {
		if httperrors.Is404(err) {
			return diag.Errorf("external domain %q not found", domainName)
		}

		return diag.FromErr(err)
	}

	if projectID != nil && *projectID != "" && resp.ProjectID != *projectID {
		return diag.Errorf("external domain %q: project_id does not match API response (expected %s, got %s)",
			domainName, *projectID, resp.ProjectID)
	}

	d.SetId(resp.Domain)
	persistExternalDomainFromRegistrarResponse(resp, d)

	return nil
}
