package vpcgw

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/cdf"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/zonal"
)

func ResourceDHCPReservation() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceVPCPublicGatewayDHCPReservationCreate,
		ReadContext:   resourceVPCPublicGatewayDHCPReservationRead,
		UpdateContext: resourceVPCPublicGatewayDHCPReservationUpdate,
		DeleteContext: resourceVPCPublicGatewayDHCPReservationDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create:  schema.DefaultTimeout(defaultTimeout),
			Update:  schema.DefaultTimeout(defaultTimeout),
			Delete:  schema.DefaultTimeout(defaultTimeout),
			Default: schema.DefaultTimeout(defaultTimeout),
		},
		SchemaVersion:      0,
		SchemaFunc:         dhcpReservation,
		CustomizeDiff:      cdf.LocalityCheck("gateway_network_id"),
		DeprecationMessage: dhcpDeprecationMessage,
	}
}

func dhcpReservation() map[string]*schema.Schema {
	return map[string]*schema.Schema{
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
	}
}

func resourceVPCPublicGatewayDHCPReservationCreate(_ context.Context, _ *schema.ResourceData, _ any) diag.Diagnostics {
	return diag.Diagnostics{{
		Severity: diag.Error,
		Summary:  "scaleway_vpc_public_gateway_dhcp_reservation is no longer supported",
		Detail:   dhcpDeprecationMessage,
	}}
}

func resourceVPCPublicGatewayDHCPReservationRead(_ context.Context, d *schema.ResourceData, _ any) diag.Diagnostics {
	d.SetId("")

	return nil
}

func resourceVPCPublicGatewayDHCPReservationUpdate(_ context.Context, d *schema.ResourceData, _ any) diag.Diagnostics {
	d.SetId("")

	return nil
}

func resourceVPCPublicGatewayDHCPReservationDelete(_ context.Context, _ *schema.ResourceData, _ any) diag.Diagnostics {
	return nil
}
