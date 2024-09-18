package vpcgw

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scaleway/scaleway-sdk-go/api/vpcgw/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/datasource"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/types"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/verify"
)

func DataSourceDHCPReservation() *schema.Resource {
	// Generate datasource schema from resource
	dsSchema := datasource.SchemaFromResourceSchema(ResourceDHCPReservation().Schema)

	// Set 'Optional' schema elements
	datasource.AddOptionalFieldsToSchema(dsSchema, "mac_address", "gateway_network_id")

	dsSchema["mac_address"].ConflictsWith = []string{"reservation_id"}
	dsSchema["gateway_network_id"].ConflictsWith = []string{"reservation_id"}
	dsSchema["reservation_id"] = &schema.Schema{
		Type:             schema.TypeString,
		Optional:         true,
		Description:      "The ID of dhcp entry reservation",
		ValidateDiagFunc: verify.IsUUIDorUUIDWithLocality(),
		ConflictsWith:    []string{"mac_address", "gateway_network_id"},
	}
	dsSchema["wait_for_dhcp"] = &schema.Schema{
		Type:        schema.TypeBool,
		Optional:    true,
		Default:     false,
		Description: "Wait the MAC address in dhcp entries",
	}

	// Set 'Optional' schema elements
	datasource.AddOptionalFieldsToSchema(dsSchema, "zone")

	return &schema.Resource{
		Schema:      dsSchema,
		ReadContext: DataSourceDHCPReservationRead,
	}
}

func DataSourceDHCPReservationRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api, zone, err := newAPIWithZone(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	reservationIDRaw, ok := d.GetOk("reservation_id")
	if !ok {
		var res *vpcgw.ListDHCPEntriesResponse
		gatewayNetworkID := locality.ExpandID(d.Get("gateway_network_id").(string))
		macAddress := d.Get("mac_address").(string)

		if d.Get("wait_for_dhcp").(bool) {
			res, err = waitForDHCPEntries(ctx, api, zone, gatewayNetworkID, macAddress, d.Timeout(schema.TimeoutRead))
		} else {
			res, err = api.ListDHCPEntries(
				&vpcgw.ListDHCPEntriesRequest{
					GatewayNetworkID: types.ExpandStringPtr(gatewayNetworkID),
					MacAddress:       types.ExpandStringPtr(macAddress),
				}, scw.WithContext(ctx))
		}
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

	zonedID := datasource.NewZonedID(reservationIDRaw, zone)
	d.SetId(zonedID)
	_ = d.Set("reservation_id", zonedID)

	diags := ResourceVPCPublicGatewayDHCPReservationRead(ctx, d, m)
	if diags != nil {
		return append(diags, diag.Errorf("failed to read DHCP Entries")...)
	}

	if d.Id() == "" {
		return diag.Errorf("DHCP ENTRY(%s) not found", zonedID)
	}

	return nil
}
