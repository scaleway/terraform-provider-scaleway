package vpcgw

import (
	"context"
	"errors"
	"net"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/scaleway/scaleway-sdk-go/api/vpcgw/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/cdf"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/zonal"
)

func ResourceDHCPReservation() *schema.Resource {
	return &schema.Resource{
		CreateContext: ResourceVPCPublicGatewayDHCPCReservationCreate,
		ReadContext:   ResourceVPCPublicGatewayDHCPReservationRead,
		UpdateContext: ResourceVPCPublicGatewayDHCPReservationUpdate,
		DeleteContext: ResourceVPCPublicGatewayDHCPReservationDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create:  schema.DefaultTimeout(defaultTimeout),
			Update:  schema.DefaultTimeout(defaultTimeout),
			Delete:  schema.DefaultTimeout(defaultTimeout),
			Default: schema.DefaultTimeout(defaultTimeout),
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
				Description:  "The IP address to give to the machine (IPv4 address).",
				ValidateFunc: validation.IsIPAddress,
			},
			"mac_address": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
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
			"zone": zonal.Schema(),
		},
		CustomizeDiff:      cdf.LocalityCheck("gateway_network_id"),
		DeprecationMessage: "The 'dhcp_reservation' resource is deprecated. In 2023, DHCP functionality was moved from Public Gateways to Private Networks, DHCP resources are now no longer needed. You can use IPAM to manage your IPs. For more information, please refer to the dedicated guide: https://github.com/scaleway/terraform-provider-scaleway/blob/master/docs/guides/migration_guide_vpcgw_v2.md",
	}
}

func ResourceVPCPublicGatewayDHCPCReservationCreate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	api, zone, err := newAPIWithZone(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	ip := net.ParseIP(d.Get("ip_address").(string))
	if ip == nil {
		return diag.FromErr(errors.New("could not parse ip_address"))
	}

	macAddress, err := net.ParseMAC(d.Get("mac_address").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	gatewayNetworkID := locality.ExpandID(d.Get("gateway_network_id"))

	_, err = waitForVPCGatewayNetwork(ctx, api, zone, gatewayNetworkID, d.Timeout(schema.TimeoutCreate))
	if err != nil {
		return diag.FromErr(err)
	}

	res, err := api.CreateDHCPEntry(&vpcgw.CreateDHCPEntryRequest{
		Zone:             zone,
		MacAddress:       macAddress.String(),
		IPAddress:        ip,
		GatewayNetworkID: gatewayNetworkID,
	}, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(zonal.NewIDString(zone, res.ID))

	_, err = waitForVPCGatewayNetwork(ctx, api, zone, gatewayNetworkID, d.Timeout(schema.TimeoutCreate))
	if err != nil {
		return diag.FromErr(err)
	}

	return ResourceVPCPublicGatewayDHCPReservationRead(ctx, d, m)
}

func ResourceVPCPublicGatewayDHCPReservationRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	api, zone, ID, err := NewAPIWithZoneAndID(m, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	entry, err := api.GetDHCPEntry(&vpcgw.GetDHCPEntryRequest{
		DHCPEntryID: ID,
		Zone:        zone,
	}, scw.WithContext(ctx))
	if err != nil {
		if httperrors.Is404(err) {
			d.SetId("")

			return nil
		}

		return diag.FromErr(err)
	}

	_ = d.Set("ip_address", entry.IPAddress.String())
	_ = d.Set("mac_address", entry.MacAddress)
	_ = d.Set("hostname", entry.Hostname)
	_ = d.Set("type", entry.Type.String())
	_ = d.Set("gateway_network_id", zonal.NewIDString(zone, entry.GatewayNetworkID))
	_ = d.Set("created_at", entry.CreatedAt.Format(time.RFC3339))
	_ = d.Set("updated_at", entry.UpdatedAt.Format(time.RFC3339))
	_ = d.Set("zone", zone)

	return nil
}

func ResourceVPCPublicGatewayDHCPReservationUpdate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	api, zone, ID, err := NewAPIWithZoneAndID(m, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	if d.HasChanges("ip_address") {
		ip := net.ParseIP(d.Get("ip_address").(string))
		if ip == nil {
			return diag.FromErr(errors.New("could not parse ip_address"))
		}

		gatewayNetworkID := locality.ExpandID(d.Get("gateway_network_id"))

		_, err = waitForVPCGatewayNetwork(ctx, api, zone, gatewayNetworkID, d.Timeout(schema.TimeoutUpdate))
		if err != nil {
			return diag.FromErr(err)
		}

		req := &vpcgw.UpdateDHCPEntryRequest{
			DHCPEntryID: ID,
			Zone:        zone,
			IPAddress:   scw.IPPtr(ip),
		}

		_, err = api.UpdateDHCPEntry(req, scw.WithContext(ctx))
		if err != nil {
			return diag.FromErr(err)
		}

		_, err = waitForVPCGatewayNetwork(ctx, api, zone, gatewayNetworkID, d.Timeout(schema.TimeoutUpdate))
		if err != nil {
			return diag.FromErr(err)
		}
	}

	return ResourceVPCPublicGatewayDHCPReservationRead(ctx, d, m)
}

func ResourceVPCPublicGatewayDHCPReservationDelete(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	api, zone, ID, err := NewAPIWithZoneAndID(m, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	gatewayNetworkID := locality.ExpandID(d.Get("gateway_network_id"))

	_, err = waitForVPCGatewayNetwork(ctx, api, zone, gatewayNetworkID, d.Timeout(schema.TimeoutDelete))
	if err != nil {
		return diag.FromErr(err)
	}

	err = api.DeleteDHCPEntry(&vpcgw.DeleteDHCPEntryRequest{
		DHCPEntryID: ID,
		Zone:        zone,
	}, scw.WithContext(ctx))

	if err != nil && !httperrors.Is404(err) {
		return diag.FromErr(err)
	}

	_, err = waitForVPCGatewayNetwork(ctx, api, zone, gatewayNetworkID, d.Timeout(schema.TimeoutDelete))
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}
