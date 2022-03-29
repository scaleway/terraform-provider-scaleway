package scaleway

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/scaleway/scaleway-sdk-go/api/vpcgw/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

func resourceScalewayVPCPublicGatewayDHCPReservation() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceScalewayVPCPublicGatewayDHCPCReservationCreate,
		ReadContext:   resourceScalewayVPCPublicGatewayDHCPReservationRead,
		UpdateContext: resourceScalewayVPCPublicGatewayDHCPReservationUpdate,
		DeleteContext: resourceScalewayVPCPublicGatewayDHCPReservationDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 0,
		Schema: map[string]*schema.Schema{
			"gateway_network_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The ID of the owning GatewayNetwork (UUID format).",
			},
			"ip_address": {
				Type:         schema.TypeString,
				Required:     true,
				Description:  "The IP address to give to the machine (IP address).",
				ValidateFunc: validation.IsIPAddress,
			},
			"mac_address": {
				Type:         schema.TypeString,
				Required:     true,
				Description:  "The MAC address to give a static entry to.",
				ValidateFunc: validation.IsMACAddress,
			},
			"hostname": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The Hostname of the client machine.",
			},
			"type": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The reservation type, either static (DHCP reservation) or dynamic (DHCP lease). Possible values are reservation and lease",
			},
			"created_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The configuration creation date.",
			},
			"updated_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The configuration last modification date.",
			},
			"zone": zoneSchema(),
		},
	}
}

func resourceScalewayVPCPublicGatewayDHCPCReservationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vpcgwAPI, zone, err := vpcgwAPIWithZone(d, meta)
	if err != nil {
		return diag.FromErr(err)
	}

	ip := net.ParseIP(d.Get("ip_address").(string))
	if ip == nil {
		return diag.FromErr(fmt.Errorf("could not parse ip_address"))
	}

	macAddress, err := net.ParseMAC(d.Get("mac_address").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	gatewayNetworkID := d.Get("gateway_network_id").(string)
	_, err = vpcgwAPI.WaitForGatewayNetwork(&vpcgw.WaitForGatewayNetworkRequest{
		GatewayNetworkID: gatewayNetworkID},
		scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	req := &vpcgw.CreateDHCPEntryRequest{
		Zone:             zone,
		MacAddress:       macAddress.String(),
		IPAddress:        ip,
		GatewayNetworkID: gatewayNetworkID,
	}

	res, err := vpcgwAPI.CreateDHCPEntry(req, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(newZonedIDString(zone, res.ID))

	_, err = vpcgwAPI.WaitForGatewayNetwork(&vpcgw.WaitForGatewayNetworkRequest{
		GatewayNetworkID: gatewayNetworkID},
		scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceScalewayVPCPublicGatewayDHCPReservationRead(ctx, d, meta)
}

func resourceScalewayVPCPublicGatewayDHCPReservationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vpcgwAPI, zone, ID, err := vpcgwAPIWithZoneAndID(meta, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	entry, err := vpcgwAPI.GetDHCPEntry(&vpcgw.GetDHCPEntryRequest{
		DHCPEntryID: ID,
		Zone:        zone,
	}, scw.WithContext(ctx))
	if err != nil {
		if is404Error(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	_ = d.Set("ip_address", entry.IPAddress.String())
	_ = d.Set("mac_address", entry.MacAddress)
	_ = d.Set("hostname", entry.Hostname)
	_ = d.Set("type", entry.Type.String())
	_ = d.Set("gateway_network_id", entry.GatewayNetworkID)
	_ = d.Set("created_at", entry.CreatedAt.Format(time.RFC3339))
	_ = d.Set("updated_at", entry.UpdatedAt.Format(time.RFC3339))
	_ = d.Set("zone", zone)

	return nil
}

func resourceScalewayVPCPublicGatewayDHCPReservationUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vpcgwAPI, zone, ID, err := vpcgwAPIWithZoneAndID(meta, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	if d.HasChangesExcept("ip_address") {
		ip := net.ParseIP(d.Get("ip_address").(string))
		if ip == nil {
			return diag.FromErr(fmt.Errorf("could not parse ip_address"))
		}
		req := &vpcgw.UpdateDHCPEntryRequest{
			DHCPEntryID: ID,
			Zone:        zone,
			IPAddress:   scw.IPPtr(ip),
		}

		_, err = vpcgwAPI.UpdateDHCPEntry(req, scw.WithContext(ctx))
		if err != nil {
			return diag.FromErr(err)
		}
	}

	return resourceScalewayVPCPublicGatewayDHCPReservationRead(ctx, d, meta)
}

func resourceScalewayVPCPublicGatewayDHCPReservationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vpcgwAPI, zone, ID, err := vpcgwAPIWithZoneAndID(meta, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	gatewayNetworkID := d.Get("gateway_network_id").(string)
	_, err = vpcgwAPI.WaitForGatewayNetwork(&vpcgw.WaitForGatewayNetworkRequest{
		GatewayNetworkID: gatewayNetworkID},
		scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	err = vpcgwAPI.DeleteDHCPEntry(&vpcgw.DeleteDHCPEntryRequest{
		DHCPEntryID: ID,
		Zone:        zone,
	}, scw.WithContext(ctx))

	if err != nil && !is404Error(err) {
		return diag.FromErr(err)
	}

	_, err = vpcgwAPI.WaitForGatewayNetwork(&vpcgw.WaitForGatewayNetworkRequest{
		GatewayNetworkID: gatewayNetworkID},
		scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}
