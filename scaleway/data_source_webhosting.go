package scaleway

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	webhosting "github.com/scaleway/scaleway-sdk-go/api/webhosting/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

func dataSourceScalewayWebhosting() *schema.Resource {
	dsSchema := datasourceSchemaFromResourceSchema(resourceScalewayWebhosting().Schema)

	addOptionalFieldsToSchema(dsSchema, "domain")

	dsSchema["domain"].ConflictsWith = []string{"webhosting_id"}
	dsSchema["webhosting_id"] = &schema.Schema{
		Type:          schema.TypeString,
		Optional:      true,
		Description:   "The ID of the Webhosting",
		ValidateFunc:  validationUUIDorUUIDWithLocality(),
		ConflictsWith: []string{"domain"},
	}
	dsSchema["organization_id"] = organizationIDOptionalSchema()
	dsSchema["project_id"] = &schema.Schema{
		Type:         schema.TypeString,
		Optional:     true,
		Description:  "The project ID the resource is associated to",
		ValidateFunc: validationUUID(),
	}

	return &schema.Resource{
		Schema:      dsSchema,
		ReadContext: dataSourceScalewayWebhostingRead,
	}
}

func dataSourceScalewayWebhostingRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api, region, err := webhostingAPIWithRegion(d, meta)
	if err != nil {
		return diag.FromErr(err)
	}

	/*	var vpcID interface{}
		var ok bool*/
	webhostingID, ok := d.GetOk("webhosting_id")
	if !ok { // Get IP by region and IP address.
		res, err := api.ListHostings(&webhosting.ListHostingsRequest{
			Region:         region,
			Domain:         expandStringPtr(d.Get("domain")),
			ProjectID:      expandStringPtr(d.Get("project_id")),
			OrganizationID: expandStringPtr(d.Get("organization_id")),
		}, scw.WithContext(ctx))
		if err != nil {
			return diag.FromErr(err)
		}

		for _, hosting := range res.Hostings {
			if hosting.Domain == d.Get("domain").(string) {
				if webhostingID != "" {
					return diag.Errorf("more than 1 hosting found with the same domain %s", d.Get("domain"))
				}
				webhostingID = hosting.ID
			}
		}
		if webhostingID == "" {
			return diag.Errorf("no hosting found with the domain %s", d.Get("domain"))
		}
	}

	regionalID := datasourceNewRegionalID(webhostingID, region)
	d.SetId(regionalID)
	err = d.Set("webhosting_id", regionalID)
	if err != nil {
		return diag.FromErr(err)
	}

	diags := resourceScalewayWebhostingRead(ctx, d, meta)
	if diags != nil {
		return append(diags, diag.Errorf("failed to read hosting")...)
	}

	if d.Id() == "" {
		return diag.Errorf("hosting (%s) not found", regionalID)
	}

	return nil
}
