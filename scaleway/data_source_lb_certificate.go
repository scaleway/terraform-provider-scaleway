package scaleway

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	lbSDK "github.com/scaleway/scaleway-sdk-go/api/lb/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality"
)

func dataSourceScalewayLbCertificate() *schema.Resource {
	// Generate datasource schema from resource
	dsSchema := datasourceSchemaFromResourceSchema(resourceScalewayLbCertificate().Schema)

	// Set 'Optional' schema elements
	addOptionalFieldsToSchema(dsSchema, "name", "lb_id")

	dsSchema["name"].ConflictsWith = []string{"certificate_id"}
	dsSchema["certificate_id"] = &schema.Schema{
		Type:          schema.TypeString,
		Optional:      true,
		Description:   "The ID of the certificate",
		ValidateFunc:  validationUUIDorUUIDWithLocality(),
		ConflictsWith: []string{"name"},
	}

	return &schema.Resource{
		ReadContext: dataSourceScalewayLbCertificateRead,
		Schema:      dsSchema,
	}
}

func dataSourceScalewayLbCertificateRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api, zone, err := lbAPIWithZone(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	crtID, ok := d.GetOk("certificate_id")
	if !ok { // Get LB by name.
		certificateName := d.Get("name").(string)
		res, err := api.ListCertificates(&lbSDK.ZonedAPIListCertificatesRequest{
			Zone: zone,
			Name: expandStringPtr(certificateName),
			LBID: locality.ExpandID(d.Get("lb_id")),
		}, scw.WithContext(ctx))
		if err != nil {
			return diag.FromErr(err)
		}

		foundCertificate, err := findExact(
			res.Certificates,
			func(s *lbSDK.Certificate) bool { return s.Name == certificateName },
			certificateName,
		)
		if err != nil {
			return diag.FromErr(err)
		}

		crtID = foundCertificate.ID
	}
	zonedID := datasourceNewZonedID(crtID, zone)
	d.SetId(zonedID)
	err = d.Set("certificate_id", zonedID)
	if err != nil {
		return diag.FromErr(err)
	}
	return resourceScalewayLbCertificateRead(ctx, d, m)
}
