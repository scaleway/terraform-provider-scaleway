package scaleway

import (
	"context"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	vpcgw "github.com/scaleway/scaleway-sdk-go/api/vpcgw/v1beta1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

const (
	retryIntervalVPCPublicGatewayNetwork = 30 * time.Second
)

func resourceScalewayVPCPublicGatewayNetwork() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceScalewayVPCPublicGatewayNetworkCreate,
		ReadContext:   resourceScalewayVPCPublicGatewayNetworkRead,
		UpdateContext: resourceScalewayVPCPublicGatewayNetworkUpdate,
		DeleteContext: resourceScalewayVPCPublicGatewayNetworkDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 0,
		Schema: map[string]*schema.Schema{
			"gateway_id": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validationUUIDorUUIDWithLocality(),
				Description:  "The ID of the gateway this network where connect to",
			},
			"private_network_id": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validationUUIDorUUIDWithLocality(),
				Description:  "The ID of the private network connect to",
			},
			"enable_masquerade": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Enable masquerade on this network",
			},
			"dhcp_id": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validationUUIDorUUIDWithLocality(),
				Description:  "The ID of the dhcp network where connect to",
			},
			
			"project_id": projectIDSchema(),
			"zone":       zoneSchema(),
			// Computed elements
			"organization_id": organizationIDSchema(),
			"created_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The date and time of the creation of the public gateway",
			},
			"updated_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The date and time of the last update of the public gateway",
			},
		},
	}
}

func resourceScalewayVPCPublicGatewayNetworkCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vpcgwAPI, zone, err := vpcgwAPIWithZone(d, meta)
	if err != nil {
		return diag.FromErr(err)
	}

	dhcpID := expandZonedID(d.Get("dhcp_id").(string)).ID
	staticIPNet := expandIPNet(d.Get("static_address").(string))
	req := &vpcgw.CreateGatewayNetworkRequest{
		GatewayID:        expandZonedID(d.Get("gateway_id").(string)).ID,
		PrivateNetworkID: expandZonedID(d.Get("private_network_id").(string)).ID,
		EnableMasquerade: *expandBoolPtr(d.Get("enable_masquerade")),
		DHCPID:           &dhcpID,
		Address:          &staticIPNet,
		Zone:             zone,
		EnableDHCP:       expandBoolPtr(d.Get("enable_dhcp")),
	}

	res, err := vpcgwAPI.CreateGatewayNetwork(req, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(newZonedIDString(zone, res.ID))

	_, err = vpcgwAPI.WaitForGatewayNetwork(&vpcgw.WaitForGatewayNetworkRequest{
		Zone:          zone,
		Timeout:       scw.TimeDurationPtr(defaultVPCGatewayTimeout),
		RetryInterval: DefaultWaitRetryInterval,
	}, scw.WithContext(ctx))
	// check err waiting process
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceScalewayVPCPublicGatewayNetworkRead(ctx, d, meta)
}

func resourceScalewayVPCPublicGatewayNetworkRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vpcgwAPI, zone, ID, err := vpcgwAPIWithZoneAndID(meta, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	gatewayNetwork, err := vpcgwAPI.GetGatewayNetwork(&vpcgw.GetGatewayNetworkRequest{
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

	_ = d.Set("gateway_id", gatewayNetwork.GatewayID)
	_ = d.Set("private_network_id", gatewayNetwork.PrivateNetworkID)
	_ = d.Set("mac_address", gatewayNetwork.MacAddress)
	_ = d.Set("enable_masquerade", gatewayNetwork.EnableMasquerade)
	_ = d.Set("enable_dhcp", gatewayNetwork.EnableDHCP)
	_ = d.Set("static_address", gatewayNetwork.Address)
	_ = d.Set("created_at", gatewayNetwork.CreatedAt.Format(time.RFC3339))
	_ = d.Set("updated_at", gatewayNetwork.UpdatedAt.Format(time.RFC3339))
	_ = d.Set("zone", zone)

	return nil
}

func resourceScalewayVPCPublicGatewayNetworkUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vpcgwAPI, zone, ID, err := vpcgwAPIWithZoneAndID(meta, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	if d.HasChanges("enable_masquerade", "dhcp_id", "enable_dhcp", "static_address") {
		dhcpID := expandZonedID(d.Get("dhcp_id").(string)).ID
		staticIPNet := expandIPNet(d.Get("static_address").(string))
		updateRequest := &vpcgw.UpdateGatewayNetworkRequest{
			GatewayNetworkID: ID,
			Zone:             zone,
			EnableMasquerade: expandBoolPtr(d.Get("enable_masquerade")),
			EnableDHCP:       expandBoolPtr(d.Get("enable_dhcp")),
			DHCPID:           &dhcpID,
			Address:          &staticIPNet,
		}

		_, err = vpcgwAPI.UpdateGatewayNetwork(updateRequest, scw.WithContext(ctx))
		if err != nil {
			return diag.FromErr(err)
		}
	}

	return resourceScalewayVPCPublicGatewayRead(ctx, d, meta)
}

func resourceScalewayVPCPublicGatewayNetworkDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vpcgwAPI, zone, ID, err := vpcgwAPIWithZoneAndID(meta, d.Id())
	if err != nil {
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

	retryInterval := retryIntervalVPCPublicGatewayNetwork
	_, err = vpcgwAPI.WaitForGatewayNetwork(&vpcgw.WaitForGatewayNetworkRequest{
		GatewayNetworkID: ID,
		Zone:             zone,
		Timeout:          scw.TimeDurationPtr(gatewayWaitForTimeout),
		RetryInterval:    &retryInterval,
	}, scw.WithContext(ctx))

	if err != nil && !is404Error(err) {
		return diag.FromErr(err)
	}

	return nil
}
