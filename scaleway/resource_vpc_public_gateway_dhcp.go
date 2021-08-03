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

func resourceScalewayVPCPublicGatewayDHCP() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceScalewayVPCPublicGatewayDHCPCreate,
		ReadContext:   resourceScalewayVPCPublicGatewayDHCPRead,
		UpdateContext: resourceScalewayVPCPublicGatewayDHCPUpdate,
		DeleteContext: resourceScalewayVPCPublicGatewayDHCPDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 0,
		Schema: map[string]*schema.Schema{
			"project_id": projectIDSchema(),
			"zone":       zoneSchema(),
			"subnet": {
				Description:  "subnet for the DHCP server",
				ValidateFunc: validation.IsCIDR,
			},
			"address": {
				Description: "Address: address of the DHCP server. This will be the gateway's address in the private network. Defaults to the first address of the subnet",
			},
			"pool_low": {
				Description: "low IP (included) of the dynamic address pool. Defaults to the second address of the subnet.",
			},
			"pool_high": {
				Description: "High IP (included) of the dynamic address pool. Defaults to the last address of the subnet.",
			},
			"enable_dynamic": {
				Description: "Whether to enable dynamic pooling of IPs. By turning the dynamic pool off, only pre-existing DHCP reservations will be handed out. Defaults to true.",
			},
			"valid_lifetime": {
				Description: "For how long, in seconds, will DHCP entries will be valid. Defaults to 1h (3600s).",
			},
			"renew_timer": {
				Description: "After how long, in seconds, a renew will be attempted. Must be 30s lower than `rebind_timer`. Defaults to 50m (3000s).",
			},
			"rebind_timer": {
				Description: "After how long, in seconds, a DHCP client will query for a new lease if previous renews fail. Must be 30s lower than `valid_lifetime`. Defaults to 51m (3060s).",
			},
			"push_default_route": {
				Description: "Whether the gateway should push a default route to DHCP clients or only hand out IPs. Defaults to true",
			},
			"push_dns_server": {
				Description: "Whether the gateway should push custom DNS servers to clients. This allows for instance hostname -> IP resolution. Defaults to true.",
			},
			"dns_server_override": {
				Description: "override the DNS server list pushed to DHCP clients, instead of the gateway itself",
			},
			"dns_search": {
				Description: "additional DNS search paths",
			},
			"dns_local_name": {
				Description: "TLD given to hostnames in the Private Network. Allowed characters are `a-z0-9-.`. Defaults to the slugified Private Network name if created along a GatewayNetwork, or else to `priv`.",
			},
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

func resourceScalewayVPCPublicGatewayDHCPCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vpcgwAPI, zone, err := vpcgwAPIWithZone(d, meta)
	if err != nil {
		return diag.FromErr(err)
	}

	req := &vpcgw.CreateDHCPRequest{
		Zone: zone,
	}

	res, err := vpcgwAPI.CreateDHCP(req, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(newZonedIDString(zone, res.ID))

	return resourceScalewayVPCPublicGatewayDHCPRead(ctx, d, meta)
}

func resourceScalewayVPCPublicGatewayDHCPRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vpcgwAPI, zone, ID, err := vpcgwAPIWithZoneAndID(meta, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	dhcp, err := vpcgwAPI.GetDHCP(&vpcgw.GetDHCPRequest{
		DHCPID: ID,
		Zone:   zone,
	}, scw.WithContext(ctx))
	if err != nil {
		if is404Error(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	_ = d.Set("organization_id", dhcp.OrganizationID)
	_ = d.Set("project_id", dhcp.ProjectID)
	_ = d.Set("created_at", dhcp.CreatedAt.Format(time.RFC3339))
	_ = d.Set("updated_at", dhcp.UpdatedAt.Format(time.RFC3339))
	_ = d.Set("zone", zone)

	return nil
}

func resourceScalewayVPCPublicGatewayDHCPUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	_, _, _, err := vpcgwAPIWithZoneAndID(meta, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceScalewayVPCPublicGatewayRead(ctx, d, meta)
}

func resourceScalewayVPCPublicGatewayDHCPDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vpcgwAPI, zone, ID, err := vpcgwAPIWithZoneAndID(meta, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	err = vpcgwAPI.DeleteDHCP(&vpcgw.DeleteDHCPRequest{
		DHCPID: ID,
		Zone:   zone,
	}, scw.WithContext(ctx))

	if err != nil && !is404Error(err) {
		return diag.FromErr(err)
	}

	return nil
}
