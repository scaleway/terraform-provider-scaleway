package scaleway

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scaleway/scaleway-sdk-go/api/lb/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

func dataSourceScalewayLbCertificate() *schema.Resource {
	// Generate datasource schema from resource
	dsSchema := datasourceSchemaFromResourceSchema(resourceScalewayLb().Schema)

	// Set 'Optional' schema elements
	addOptionalFieldsToSchema(dsSchema, "name", "zone")

	dsSchema["name"].ConflictsWith = []string{"certificate_id"}
	dsSchema["certificate_id"] = &schema.Schema{
		Type:          schema.TypeString,
		Optional:      true,
		Description:   "The ID of the certificate",
		ValidateFunc:  validationUUIDorUUIDWithLocality(),
		ConflictsWith: []string{"name"},
	}
	dsSchema["lb_id"] = &schema.Schema{
		Type:         schema.TypeString,
		Optional:     true,
		Description:  "The ID of the load-balancer",
		ValidateFunc: validationUUIDorUUIDWithLocality(),
	}

	return &schema.Resource{
		ReadContext: dataSourceScalewayLbCertificateRead,
		Schema:      dsSchema,
	}
}

func dataSourceScalewayLbCertificateRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api, zone, err := lbAPIWithZone(d, meta)
	if err != nil {
		return diag.FromErr(err)
	}

	crtID, ok := d.GetOk("certificate_id")
	if !ok { // Get LB by name.
		res, err := api.ListCertificates(&lb.ZonedAPIListCertificatesRequest{
			Zone: zone,
			Name: expandStringPtr(d.Get("name")),
			//LBID: expandID(d.Get("lb_id")),
		}, scw.WithContext(ctx))
		if err != nil {
			return diag.FromErr(err)
		}
		if len(res.Certificates) == 0 {
			return diag.FromErr(fmt.Errorf("no certificates found with the name %s", d.Get("name")))
		}
		if len(res.Certificates) > 1 {
			return diag.FromErr(fmt.Errorf("%d certificate found with the same name %s", len(res.Certificates), d.Get("name")))
		}
		crtID = res.Certificates[0].ID
	}
	zonedID := datasourceNewZonedID(crtID, zone)
	d.SetId(zonedID)
	err = d.Set("certificate_id", zonedID)
	if err != nil {
		return diag.FromErr(err)
	}
	return resourceScalewayLbCertificateRead(ctx, d, meta)
}
