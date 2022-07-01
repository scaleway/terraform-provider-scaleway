package scaleway

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	flexibleip "github.com/scaleway/scaleway-sdk-go/api/flexibleip/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

func dataSourceScalewayFlexibleIP() *schema.Resource {
	// Generate datasource schema from resource
	dsSchema := datasourceSchemaFromResourceSchema(resourceScalewayFlexibleIP().Schema)

	dsSchema["ip_address"] = &schema.Schema{
		Type:          schema.TypeString,
		Optional:      true,
		Description:   "The IP address",
		ConflictsWith: []string{"id"},
	}
	dsSchema["id"] = &schema.Schema{
		Type:          schema.TypeString,
		Optional:      true,
		Description:   "The ID of the IP address",
		ConflictsWith: []string{"ip_address"},
		ValidateFunc:  validationUUIDorUUIDWithLocality(),
	}

	return &schema.Resource{
		ReadContext: dataSourceScalewayFlexibleIPRead,
		Schema:      dsSchema,
	}
}

func dataSourceScalewayFlexibleIPRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	fipAPI, zone, err := fipAPIWithZone(d, meta)
	if err != nil {
		return diag.FromErr(err)
	}

	ipID, ok := d.GetOk("id")
	if !ok { // Get IP by region and IP address.
		res, err := fipAPI.ListFlexibleIPs(&flexibleip.ListFlexibleIPsRequest{
			Zone:      zone,
			ProjectID: expandStringPtr(d.Get("project_id")),
		}, scw.WithContext(ctx))
		if err != nil {
			return diag.FromErr(err)
		}
		if len(res.FlexibleIPs) == 0 {
			return diag.FromErr(fmt.Errorf("no ips found with the address %s", d.Get("ip_address")))
		}
		if len(res.FlexibleIPs) > 1 {
			return diag.FromErr(fmt.Errorf("%d ips found with the same address %s", len(res.FlexibleIPs), d.Get("ip_address")))
		}
		ipID = res.FlexibleIPs[0].ID
	}

	zoneID := datasourceNewZonedID(ipID, zone)
	d.SetId(zoneID)
	err = d.Set("id", zoneID)
	if err != nil {
		return diag.FromErr(err)
	}
	return resourceScalewayFlexibleIPRead(ctx, d, meta)
}
