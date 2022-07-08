package scaleway

import (
	"context"

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
		Description:   "The IPv4 address",
		ConflictsWith: []string{"id"},
	}
	dsSchema["id"] = &schema.Schema{
		Type:          schema.TypeString,
		Optional:      true,
		Description:   "The ID of the IPv4 address",
		ConflictsWith: []string{"ip_address"},
		ValidateFunc:  validationUUIDorUUIDWithLocality(),
	}
	dsSchema["project_id"] = &schema.Schema{
		Type:         schema.TypeString,
		Description:  "The project_id you want to attach the resource to",
		Optional:     true,
		ForceNew:     true,
		Computed:     true,
		ValidateFunc: validationUUID(),
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

	ipID, ipIDExists := d.GetOk("id")

	if !ipIDExists {
		res, err := fipAPI.ListFlexibleIPs(&flexibleip.ListFlexibleIPsRequest{
			Zone:      zone,
			ProjectID: expandStringPtr(d.Get("project_id")),
		}, scw.WithContext(ctx))
		if err != nil {
			return diag.FromErr(err)
		}

		for _, ip := range res.FlexibleIPs {
			if ip.IPAddress.String() == d.Get("ip_address").(string) {
				if ipID != "" {
					return diag.Errorf("more than 1 flexible ip found with the same IPv4 address %s", d.Get("ip_address"))
				}
				ipID = ip.ID
			}
		}
		if ipID == "" {
			return diag.Errorf("no flexible ip found with the same IPv4 address %s", d.Get("ip_address"))
		}
	}

	zoneID := datasourceNewZonedID(ipID, zone)
	d.SetId(zoneID)
	err = d.Set("id", zoneID)
	if err != nil {
		return diag.FromErr(err)
	}

	diags := resourceScalewayFlexibleIPRead(ctx, d, meta)
	if diags != nil {
		return append(diags, diag.Errorf("failed to read flexible ip state")...)
	}

	if d.Id() == "" {
		return diag.Errorf("flexible ip (%s) not found", ipID)
	}

	return nil
}
