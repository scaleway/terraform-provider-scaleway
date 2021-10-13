package scaleway

import (
	"context"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	vpcgw "github.com/scaleway/scaleway-sdk-go/api/vpcgw/v1beta1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

const (
	retryIntervalVPCGatewayNetwork = 30 * time.Second
)

func resourceScalewayVPCGatewayNetwork() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceScalewayVPCGatewayNetworkCreate,
		ReadContext:   resourceScalewayVPCGatewayNetworkRead,
		UpdateContext: resourceScalewayVPCGatewayNetworkUpdate,
		DeleteContext: resourceScalewayVPCGatewayNetworkDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 0,
		Schema: map[string]*schema.Schema{
			"gateway_id": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validationUUIDorUUIDWithLocality(),
				Description:  "The ID of the public gateway where connect to",
			},
			"private_network_id": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validationUUIDorUUIDWithLocality(),
				Description:  "The ID of the private network where connect to",
			},
			"dhcp_id": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validationUUIDorUUIDWithLocality(),
				Description:  "The ID of the public gateway DHCP config",
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
				Default:     true,
				Description: "Enable DHCP config on this network",
			},
			"cleanup_dhcp": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Remove DHCP config on this network on destroy",
			},
			"static_address": {
				Type:         schema.TypeString,
				Description:  "The static IP address in CIDR on this network",
				Optional:     true,
				ValidateFunc: validation.IsCIDR,
			},
			// Computed elements
			"mac_address": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The mac address on this network",
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
			"zone": zoneSchema(),
		},
	}
}

func resourceScalewayVPCGatewayNetworkCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vpcgwNetworkAPI, zone, err := vpcgwAPIWithZone(d, meta)
	if err != nil {
		return diag.FromErr(err)
	}

	gatewayID := expandZonedID(d.Get("gateway_id").(string)).ID
	req := &vpcgw.CreateGatewayNetworkRequest{
		Zone:             zone,
		GatewayID:        gatewayID,
		PrivateNetworkID: expandZonedID(d.Get("private_network_id").(string)).ID,
		EnableMasquerade: *expandBoolPtr(d.Get("enable_masquerade")),
		EnableDHCP:       expandBoolPtr(d.Get("enable_dhcp")),
	}

	staticAddress, staticAddressExist := d.GetOk("static_address")
	if staticAddressExist {
		address := expandIPNet(staticAddress.(string))
		req.Address = &address
	}

	dhcpID, dhcpExist := d.GetOk("dhcp_id")
	if dhcpExist {
		dhcpZoned := expandZonedID(dhcpID.(string))
		req.DHCPID = &dhcpZoned.ID
	}

	retryInterval := retryIntervalVPCGatewayNetwork
	//check gateway is in stable state.
	_, err = vpcgwNetworkAPI.WaitForGateway(&vpcgw.WaitForGatewayRequest{
		GatewayID:     gatewayID,
		Zone:          zone,
		Timeout:       scw.TimeDurationPtr(gatewayWaitForTimeout),
		RetryInterval: &retryInterval,
	}, scw.WithContext(ctx))

	if err != nil && !is404Error(err) {
		return diag.FromErr(err)
	}
	res, err := vpcgwNetworkAPI.CreateGatewayNetwork(req, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(newZonedIDString(zone, res.ID))
	// set default interval
	res, err = vpcgwNetworkAPI.WaitForGatewayNetwork(&vpcgw.WaitForGatewayNetworkRequest{
		GatewayNetworkID: res.ID,
		Timeout:          scw.TimeDurationPtr(defaultVPCGatewayTimeout),
		RetryInterval:    &retryInterval,
		Zone:             zone,
	}, scw.WithContext(ctx))
	// check err waiting process
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceScalewayVPCGatewayNetworkRead(ctx, d, meta)
}

func resourceScalewayVPCGatewayNetworkRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vpcgwNetworkAPI, zone, ID, err := vpcgwAPIWithZoneAndID(meta, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	gatewayNetwork, err := vpcgwNetworkAPI.GetGatewayNetwork(&vpcgw.GetGatewayNetworkRequest{
		GatewayNetworkID: ID,
		Zone:             zone,
	}, scw.WithContext(ctx))
	if err != nil {
		if is404Error(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	if dhcp := gatewayNetwork.DHCP; dhcp != nil {
		_ = d.Set("dhcp_id", newZonedID(zone, dhcp.ID).String())
	}

	if staticAddress := gatewayNetwork.Address; staticAddress != nil {
		_ = d.Set("static_address", flattenIPNet(*staticAddress))
	}

	if macAddress := gatewayNetwork.MacAddress; macAddress != nil {
		_ = d.Set("mac_address", flattenStringPtr(macAddress).(string))
	}

	if enableDHCP := gatewayNetwork.EnableDHCP; enableDHCP {
		_ = d.Set("enable_dhcp", enableDHCP)
	}

	var cleanUpDHCPValue bool
	cleanUpDHCP, cleanUpDHCPExist := d.GetOk("cleanup_dhcp")
	if cleanUpDHCPExist {
		cleanUpDHCPValue = *expandBoolPtr(cleanUpDHCP)
	}

	_ = d.Set("gateway_id", newZonedID(zone, gatewayNetwork.GatewayID).String())
	_ = d.Set("private_network_id", newZonedID(zone, gatewayNetwork.PrivateNetworkID).String())
	_ = d.Set("enable_masquerade", gatewayNetwork.EnableMasquerade)
	_ = d.Set("cleanup_dhcp", cleanUpDHCPValue)
	_ = d.Set("created_at", gatewayNetwork.CreatedAt.Format(time.RFC3339))
	_ = d.Set("updated_at", gatewayNetwork.UpdatedAt.Format(time.RFC3339))
	_ = d.Set("zone", zone.String())

	return nil
}

func resourceScalewayVPCGatewayNetworkUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vpcgwAPI, zone, ID, err := vpcgwAPIWithZoneAndID(meta, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	if d.HasChanges("enable_masquerade", "dhcp_id", "enable_dhcp", "static_address") {
		dhcpID := expandZonedID(d.Get("dhcp_id").(string)).ID
		updateRequest := &vpcgw.UpdateGatewayNetworkRequest{
			GatewayNetworkID: ID,
			Zone:             zone,
			EnableMasquerade: expandBoolPtr(d.Get("enable_masquerade")),
			EnableDHCP:       expandBoolPtr(d.Get("enable_dhcp")),
			DHCPID:           &dhcpID,
		}
		staticAddress, staticAddressExist := d.GetOk("static_address")
		if staticAddressExist {
			address := expandIPNet(staticAddress.(string))
			updateRequest.Address = &address
		}

		_, err = vpcgwAPI.UpdateGatewayNetwork(updateRequest, scw.WithContext(ctx))
		if err != nil {
			return diag.FromErr(err)
		}
	}

	return resourceScalewayVPCGatewayNetworkRead(ctx, d, meta)
}

func resourceScalewayVPCGatewayNetworkDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vpcgwAPI, zone, ID, err := vpcgwAPIWithZoneAndID(meta, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	defaultInterval := retryIntervalVPCGatewayNetwork
	// check if network is a stable process
	gwNetwork, err := vpcgwAPI.WaitForGatewayNetwork(&vpcgw.WaitForGatewayNetworkRequest{
		GatewayNetworkID: ID,
		Zone:             zone,
		RetryInterval:    &defaultInterval,
	})
	if err != nil && !is404Error(err) {
		return diag.FromErr(err)
	}
	//check gateway is in stable state.
	_, err = vpcgwAPI.WaitForGateway(&vpcgw.WaitForGatewayRequest{
		GatewayID:     gwNetwork.GatewayID,
		Zone:          zone,
		Timeout:       scw.TimeDurationPtr(gatewayWaitForTimeout),
		RetryInterval: &defaultInterval,
	}, scw.WithContext(ctx))
	if err != nil && !is404Error(err) {
		return diag.FromErr(err)
	}

	err = vpcgwAPI.DeleteGatewayNetwork(&vpcgw.DeleteGatewayNetworkRequest{
		GatewayNetworkID: ID,
		Zone:             zone,
		CleanupDHCP:      *expandBoolPtr(d.Get("cleanup_dhcp")),
	}, scw.WithContext(ctx))

	if err != nil && !is404Error(err) {
		return diag.FromErr(err)
	}
	//check gateway is in stable state.
	_, err = vpcgwAPI.WaitForGateway(&vpcgw.WaitForGatewayRequest{
		GatewayID:     gwNetwork.GatewayID,
		Zone:          zone,
		Timeout:       scw.TimeDurationPtr(gatewayWaitForTimeout),
		RetryInterval: &defaultInterval,
	}, scw.WithContext(ctx))

	if err != nil && !is404Error(err) {
		return diag.FromErr(err)
	}

	return nil
}
