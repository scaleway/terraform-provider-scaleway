package lb

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	lbSDK "github.com/scaleway/scaleway-sdk-go/api/lb/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/datasource"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/types"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/verify"
)

func DataSourceCertificate() *schema.Resource {
	// Generate datasource schema from resource
	dsSchema := datasource.SchemaFromResourceSchema(ResourceCertificate().Schema)

	// Set 'Optional' schema elements
	datasource.AddOptionalFieldsToSchema(dsSchema, "name", "lb_id")

	dsSchema["name"].ConflictsWith = []string{"certificate_id"}
	dsSchema["certificate_id"] = &schema.Schema{
		Type:          schema.TypeString,
		Optional:      true,
		Description:   "The ID of the certificate",
		ValidateFunc:  verify.IsUUIDorUUIDWithLocality(),
		ConflictsWith: []string{"name"},
	}

	return &schema.Resource{
		ReadContext: DataSourceLbCertificateRead,
		Schema:      dsSchema,
	}
}

func DataSourceLbCertificateRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api, zone, err := lbAPIWithZone(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	crtID, ok := d.GetOk("certificate_id")
	if !ok { // Get LB by name.
		certificateName := d.Get("name").(string)
		res, err := api.ListCertificates(&lbSDK.ZonedAPIListCertificatesRequest{
			Zone: zone,
			Name: types.ExpandStringPtr(certificateName),
			LBID: locality.ExpandID(d.Get("lb_id")),
		}, scw.WithContext(ctx))
		if err != nil {
			return diag.FromErr(err)
		}

		foundCertificate, err := datasource.FindExact(
			res.Certificates,
			func(s *lbSDK.Certificate) bool { return s.Name == certificateName },
			certificateName,
		)
		if err != nil {
			return diag.FromErr(err)
		}

		crtID = foundCertificate.ID
	}
	zonedID := datasource.NewZonedID(crtID, zone)
	d.SetId(zonedID)
	err = d.Set("certificate_id", zonedID)
	if err != nil {
		return diag.FromErr(err)
	}
	return resourceLbCertificateRead(ctx, d, m)
}
