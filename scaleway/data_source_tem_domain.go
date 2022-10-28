package scaleway

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	tem "github.com/scaleway/scaleway-sdk-go/api/tem/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

func dataSourceScalewayTemDomain() *schema.Resource {
	// Generate datasource schema from resource
	dsSchema := datasourceSchemaFromResourceSchema(resourceScalewayTemDomain().Schema)

	// Set 'Optional' schema elements
	addOptionalFieldsToSchema(dsSchema, "name", "region")

	dsSchema["name"].ConflictsWith = []string{"id"}
	dsSchema["id"] = &schema.Schema{
		Type:          schema.TypeString,
		Optional:      true,
		Description:   "The ID of the tem domain",
		ValidateFunc:  validationUUIDorUUIDWithLocality(),
		ConflictsWith: []string{"name"},
	}

	return &schema.Resource{
		ReadContext: dataSourceScalewayTemDomainRead,

		Schema: dsSchema,
	}
}

func dataSourceScalewayTemDomainRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api, region, err := temAPIWithRegion(d, meta)
	if err != nil {
		return diag.FromErr(err)
	}

	domainID, ok := d.GetOk("id")
	if !ok {
		res, err := api.ListDomains(&tem.ListDomainsRequest{
			Region:    region,
			Name:      expandStringPtr(d.Get("name")),
			ProjectID: expandStringPtr(d.Get("project_id")),
		}, scw.WithContext(ctx))
		if err != nil {
			return diag.FromErr(err)
		}

		for _, domain := range res.Domains {
			if domain.Status == tem.DomainStatusRevoked {
				continue
			}

			if domain.Name == d.Get("name").(string) {
				if domainID != "" {
					return diag.FromErr(fmt.Errorf("more than 1 server found with the same name %s", d.Get("name")))
				}

				domainID = domain.ID
			}
		}

		if domainID == "" {
			return diag.FromErr(fmt.Errorf("no domain found with the name %s", d.Get("name")))
		}
	}

	regionalID := datasourceNewRegionalizedID(domainID, region)
	d.SetId(regionalID)
	err = d.Set("id", regionalID)
	if err != nil {
		return diag.FromErr(err)
	}

	diags := resourceScalewayTemDomainRead(ctx, d, meta)
	if diags != nil {
		return append(diags, diag.Errorf("failed to read tem domain state")...)
	}

	if d.Id() == "" {
		return diag.Errorf("tem domain (%s) not found", regionalID)
	}

	return nil
}
