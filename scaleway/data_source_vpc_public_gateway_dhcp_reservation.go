package scaleway

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scaleway/scaleway-sdk-go/api/vpcgw/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

func dataSourceScalewayVPCPublicGatewayDHCPReservation() *schema.Resource {
	// Generate datasource schema from resource
	dsSchema := datasourceSchemaFromResourceSchema(resourceScalewayVPCPublicGatewayDHCPReservation().Schema)

	// Set 'Optional' schema elements
	addOptionalFieldsToSchema(dsSchema, "mac_address")

	dsSchema["mac_address"].ConflictsWith = []string{"reservation_id"}
	dsSchema["reservation_id"] = &schema.Schema{
		Type:         schema.TypeString,
		Optional:     true,
		Description:  "The ID of dhcp entry reservation",
		ValidateFunc: validationUUIDorUUIDWithLocality(),
	}

	// Set 'Optional' schema elements
	addOptionalFieldsToSchema(dsSchema, "zone")

	return &schema.Resource{
		Schema:      dsSchema,
		ReadContext: dataSourceScalewayVPCPublicGatewayDHCPReservationRead,
	}
}

func dataSourceScalewayVPCPublicGatewayDHCPReservationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vpcgwAPI, zone, err := vpcgwAPIWithZone(d, meta)
	if err != nil {
		return diag.FromErr(err)
	}

	reservationIDRaw, ok := d.GetOk("reservation_id")
	if !ok {
		res, err := vpcgwAPI.ListDHCPEntries(
			&vpcgw.ListDHCPEntriesRequest{
				MacAddress: expandStringPtr(d.Get("mac_address").(string)),
			}, scw.WithContext(ctx))
		if err != nil {
			return diag.FromErr(err)
		}
		if res.TotalCount == 0 {
			return diag.FromErr(
				fmt.Errorf(
					"no dhcp-entry on public gateway found with the mac_address %s",
					d.Get("mac_address"),
				),
			)
		}
		if res.TotalCount > 1 {
			return diag.FromErr(
				fmt.Errorf(
					"%d on public gateways found with the mac address %s",
					res.TotalCount,
					d.Get("mac_address"),
				),
			)
		}
		reservationIDRaw = res.DHCPEntries[0].ID
	}

	zonedID := datasourceNewZonedID(reservationIDRaw, zone)
	d.SetId(zonedID)
	_ = d.Set("reservation_id", zonedID)

	diags := resourceScalewayVPCPublicGatewayDHCPReservationRead(ctx, d, meta)
	if diags != nil {
		return append(diags, diag.Errorf("failed to read DHCP Entries")...)
	}

	if d.Id() == "" {
		return diag.Errorf("DHCP ENTRY(%s) not found", zonedID)
	}

	return nil
}
