package vpcgw

import (
	"context"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/scaleway/scaleway-sdk-go/api/vpcgw/v2"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/cdf"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/dsf"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/zonal"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/transport"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/types"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/verify"
)

func ResourceNetwork() *schema.Resource {
	return &schema.Resource{
		CreateContext: ResourceVPCGatewayNetworkCreate,
		ReadContext:   ResourceVPCGatewayNetworkRead,
		UpdateContext: ResourceVPCGatewayNetworkUpdate,
		DeleteContext: ResourceVPCGatewayNetworkDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create:  schema.DefaultTimeout(defaultTimeout),
			Read:    schema.DefaultTimeout(defaultTimeout),
			Update:  schema.DefaultTimeout(defaultTimeout),
			Delete:  schema.DefaultTimeout(defaultTimeout),
			Default: schema.DefaultTimeout(defaultTimeout),
		},
		SchemaVersion: 0,
		SchemaFunc:    networkSchema,
		CustomizeDiff: cdf.LocalityCheck("gateway_id", "private_network_id"),
	}
}

func networkSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"gateway_id": {
			Type:             schema.TypeString,
			Required:         true,
			ForceNew:         true,
			ValidateDiagFunc: verify.IsUUIDorUUIDWithLocality(),
			Description:      "The ID of the public gateway where connect to",
		},
		"private_network_id": {
			Type:             schema.TypeString,
			Required:         true,
			ForceNew:         true,
			ValidateDiagFunc: verify.IsUUIDorUUIDWithLocality(),
			DiffSuppressFunc: dsf.Locality,
			Description:      "The ID of the private network where connect to",
		},
		"dhcp_id": {
			Type:             schema.TypeString,
			Optional:         true,
			ValidateDiagFunc: verify.IsUUIDorUUIDWithLocality(),
			Description:      "The ID of the public gateway DHCP config",
			ConflictsWith:    []string{"static_address", "ipam_config"},
			DiffSuppressFunc: func(_, oldValue, newValue string, d *schema.ResourceData) bool {
				if v, ok := d.Get("ipam_config").([]any); ok && len(v) > 0 {
					return true
				}

				return oldValue == newValue
			},
			Deprecated: "DHCP configuration is no longer managed separately. Please use ipam_config instead.",
		},
		"enable_masquerade": {
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     true,
			Description: "Enable masquerade on this network",
		},
		"enable_dhcp": {
			Type:        schema.TypeBool,
			Optional:    true,
			Description: "Enable DHCP config on this network",
			Deprecated:  "DHCP is now managed automatically. Please use ipam_config instead.",
		},
		"cleanup_dhcp": {
			Type:        schema.TypeBool,
			Optional:    true,
			Computed:    true,
			Description: "Remove DHCP config on this network on destroy",
			Deprecated:  "DHCP cleanup is no longer needed. Please use ipam_config instead.",
		},
		"static_address": {
			Type:          schema.TypeString,
			Description:   "The static IP address in CIDR on this network",
			Optional:      true,
			Computed:      true,
			ValidateFunc:  validation.IsCIDR,
			ConflictsWith: []string{"dhcp_id", "ipam_config"},
			Deprecated:    "Please use ipam_config instead.",
		},
		"ipam_config": {
			Type:          schema.TypeList,
			Optional:      true,
			Computed:      true,
			Description:   "Auto-configure the Gateway Network using IPAM (IP address management service)",
			ConflictsWith: []string{"dhcp_id", "static_address"},
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"push_default_route": {
						Type:        schema.TypeBool,
						Optional:    true,
						Description: "Defines whether the default route is enabled on that Gateway Network",
					},
					"ipam_ip_id": {
						Type:             schema.TypeString,
						Optional:         true,
						Computed:         true,
						Description:      "Use this IPAM-booked IP ID as the Gateway's IP in this Private Network",
						ValidateDiagFunc: verify.IsUUIDorUUIDWithLocality(),
						DiffSuppressFunc: dsf.Locality,
					},
				},
			},
		},
		"mac_address": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "The mac address on this network",
		},
		"private_ip": {
			Type:        schema.TypeList,
			Computed:    true,
			Optional:    true,
			Description: "The private IPv4 address associated with the resource.",
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"id": {
						Type:        schema.TypeString,
						Computed:    true,
						Description: "The ID of the IPv4 address resource.",
					},
					"address": {
						Type:        schema.TypeString,
						Computed:    true,
						Description: "The private IPv4 address.",
					},
				},
			},
		},
		"created_at": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "The date and time of the creation of the gateway network",
		},
		"updated_at": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "The date and time of the last update of the gateway network",
		},
		"status": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "The status of the Public Gateway's connection to the Private Network",
		},
		"zone": zonal.Schema(),
	}
}

func ResourceVPCGatewayNetworkCreate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	api, zone, err := newAPIWithZone(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	gatewayID := zonal.ExpandID(d.Get("gateway_id").(string)).ID

	gateway, err := waitForVPCPublicGateway(ctx, api, zone, gatewayID, d.Timeout(schema.TimeoutCreate))
	if err != nil {
		return diag.FromErr(err)
	}

	pushDefaultRoute, ipamIPID := expandIpamConfig(d.Get("ipam_config"))

	req := &vpcgw.CreateGatewayNetworkRequest{
		Zone:             zone,
		GatewayID:        gateway.ID,
		PrivateNetworkID: regional.ExpandID(d.Get("private_network_id").(string)).ID,
		EnableMasquerade: *types.ExpandBoolPtr(d.Get("enable_masquerade")),
		PushDefaultRoute: pushDefaultRoute,
		IpamIPID:         ipamIPID,
	}

	gatewayNetwork, err := transport.RetryOnTransientStateError(func() (*vpcgw.GatewayNetwork, error) {
		return api.CreateGatewayNetwork(req, scw.WithContext(ctx))
	}, func() (*vpcgw.Gateway, error) {
		tflog.Warn(ctx, "Public gateway is in transient state after waiting, retrying...")

		return waitForVPCPublicGateway(ctx, api, zone, gatewayID, d.Timeout(schema.TimeoutCreate))
	})
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(zonal.NewIDString(zone, gatewayNetwork.ID))

	_, err = waitForVPCPublicGateway(ctx, api, zone, gatewayNetwork.GatewayID, d.Timeout(schema.TimeoutCreate))
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = waitForVPCGatewayNetwork(ctx, api, zone, gatewayNetwork.ID, d.Timeout(schema.TimeoutCreate))
	if err != nil {
		return diag.FromErr(err)
	}

	return ResourceVPCGatewayNetworkRead(ctx, d, m)
}

func ResourceVPCGatewayNetworkRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	api, zone, ID, err := NewAPIWithZoneAndID(m, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	gatewayNetwork, err := waitForVPCGatewayNetwork(ctx, api, zone, ID, d.Timeout(schema.TimeoutRead))
	if err != nil {
		if httperrors.Is404(err) {
			d.SetId("")

			return nil
		}

		return diag.FromErr(err)
	}

	var diags diag.Diagnostics

	if gatewayNetwork.PrivateNetworkID != "" {
		diags = setPrivateIPs(ctx, d, api, gatewayNetwork, m)
	}

	fetchRegion, err := gatewayNetwork.Zone.Region()
	if err != nil {
		return append(diags, diag.FromErr(err)...)
	}

	_ = d.Set("private_network_id", regional.NewIDString(fetchRegion, gatewayNetwork.PrivateNetworkID))
	_ = d.Set("gateway_id", zonal.NewIDString(gatewayNetwork.Zone, gatewayNetwork.GatewayID))
	_ = d.Set("enable_masquerade", gatewayNetwork.MasqueradeEnabled)
	_ = d.Set("created_at", types.FlattenTime(gatewayNetwork.CreatedAt))
	_ = d.Set("updated_at", types.FlattenTime(gatewayNetwork.UpdatedAt))
	_ = d.Set("status", string(gatewayNetwork.Status))
	_ = d.Set("zone", gatewayNetwork.Zone)

	if macAddress := gatewayNetwork.MacAddress; macAddress != nil {
		_ = d.Set("mac_address", types.FlattenStringPtr(macAddress).(string))
	}

	_ = d.Set("ipam_config", []map[string]any{
		{
			"push_default_route": gatewayNetwork.PushDefaultRoute,
			"ipam_ip_id":         gatewayNetwork.IpamIPID,
		},
	})

	return diags
}

func ResourceVPCGatewayNetworkUpdate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	api, zone, id, err := NewAPIWithZoneAndID(m, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = waitForVPCGatewayNetwork(ctx, api, zone, id, d.Timeout(schema.TimeoutUpdate))
	if err != nil {
		return diag.FromErr(err)
	}

	updateRequest := &vpcgw.UpdateGatewayNetworkRequest{
		GatewayNetworkID: id,
		Zone:             zone,
	}

	if d.HasChange("enable_masquerade") {
		updateRequest.EnableMasquerade = types.ExpandBoolPtr(d.Get("enable_masquerade"))
	}

	if d.HasChange("ipam_config") {
		pushDefaultRoute, ipamIPID := expandIpamConfig(d.Get("ipam_config"))

		updateRequest.PushDefaultRoute = new(pushDefaultRoute)
		updateRequest.IpamIPID = ipamIPID
	}

	_, err = api.UpdateGatewayNetwork(updateRequest, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = waitForVPCGatewayNetwork(ctx, api, zone, id, d.Timeout(schema.TimeoutUpdate))
	if err != nil {
		return diag.FromErr(err)
	}

	return ResourceVPCGatewayNetworkRead(ctx, d, m)
}

func ResourceVPCGatewayNetworkDelete(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	api, zone, id, err := NewAPIWithZoneAndID(m, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	gwNetwork, err := waitForVPCGatewayNetwork(ctx, api, zone, id, d.Timeout(schema.TimeoutDelete))
	if err != nil {
		if httperrors.Is404(err) {
			return nil
		}

		return diag.FromErr(err)
	}

	_, err = api.DeleteGatewayNetwork(&vpcgw.DeleteGatewayNetworkRequest{
		GatewayNetworkID: gwNetwork.ID,
		Zone:             gwNetwork.Zone,
	}, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = waitForVPCGatewayNetwork(ctx, api, zone, id, d.Timeout(schema.TimeoutDelete))
	if err != nil && !httperrors.Is404(err) {
		return diag.FromErr(err)
	}

	_, err = waitForVPCPublicGateway(ctx, api, zone, gwNetwork.GatewayID, d.Timeout(schema.TimeoutDelete))
	if err != nil && !httperrors.Is404(err) {
		return diag.FromErr(err)
	}

	return nil
}
