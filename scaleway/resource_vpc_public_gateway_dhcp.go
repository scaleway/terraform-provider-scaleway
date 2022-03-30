package scaleway

import (
	"context"
	"net"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/scaleway/scaleway-sdk-go/api/vpcgw/v1"
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
				Type:         schema.TypeString,
				ValidateFunc: validation.IsCIDR,
				Required:     true,
				Description:  "Subnet for the DHCP server",
			},
			"address": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "The address of the DHCP server. This will be the gateway's address in the private network. Defaults to the first address of the subnet",
			},
			"pool_low": {
				Type:         schema.TypeString,
				ValidateFunc: validation.IsIPAddress,
				Computed:     true,
				Optional:     true,
				Description:  "Low IP (included) of the dynamic address pool. Defaults to the second address of the subnet.",
			},
			"pool_high": {
				Type:         schema.TypeString,
				ValidateFunc: validation.IsIPAddress,
				Computed:     true,
				Optional:     true,
				Description:  "High IP (included) of the dynamic address pool. Defaults to the last address of the subnet.",
			},
			"enable_dynamic": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "Whether to enable dynamic pooling of IPs. By turning the dynamic pool off, only pre-existing DHCP reservations will be handed out. Defaults to true.",
			},
			"valid_lifetime": {
				Type:        schema.TypeInt,
				Computed:    true,
				Optional:    true,
				Description: "For how long, in seconds, will DHCP entries will be valid. Defaults to 1h (3600s).",
			},
			"renew_timer": {
				Type:        schema.TypeInt,
				Computed:    true,
				Optional:    true,
				Description: "After how long, in seconds, a renew will be attempted. Must be 30s lower than `rebind_timer`. Defaults to 50m (3000s).",
			},
			"rebind_timer": {
				Type:        schema.TypeInt,
				Computed:    true,
				Optional:    true,
				Description: "After how long, in seconds, a DHCP client will query for a new lease if previous renews fail. Must be 30s lower than `valid_lifetime`. Defaults to 51m (3060s).",
			},
			"push_default_route": {
				Type:        schema.TypeBool,
				Optional:    true,
				Computed:    true,
				Description: "Whether the gateway should push a default route to DHCP clients or only hand out IPs. Defaults to true",
			},
			"push_dns_server": {
				Type:        schema.TypeBool,
				Optional:    true,
				Computed:    true,
				Description: "Whether the gateway should push custom DNS servers to clients. This allows for instance hostname -> IP resolution. Defaults to true.",
			},
			"dns_server_override": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Description: "Override the DNS server list pushed to DHCP clients, instead of the gateway itself",
			},
			"dns_search": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Description: "Additional DNS search paths",
			},
			"dns_local_name": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "TLD given to hostnames in the Private Network. Allowed characters are `a-z0-9-.`. Defaults to the slugified Private Network name if created along a GatewayNetwork, or else to `priv`.",
			},
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

	subnet, err := expandIPNet(d.Get("subnet").(string))
	if err != nil {
		return diag.FromErr(err)
	}
	req := &vpcgw.CreateDHCPRequest{
		Zone:      zone,
		ProjectID: d.Get("project_id").(string),
		Subnet:    subnet,
	}

	if pushDefaultRoute, ok := d.GetOk("push_default_route"); ok {
		req.PushDefaultRoute = expandBoolPtr(pushDefaultRoute)
	}

	if pushDNServer, ok := d.GetOk("push_dns_server"); ok {
		req.PushDNSServer = expandBoolPtr(pushDNServer)
	}

	if enableDynamic, ok := d.GetOk("enable_dynamic"); ok {
		req.EnableDynamic = expandBoolPtr(enableDynamic)
	}

	if dnsServerOverride, ok := d.GetOk("dns_servers_override"); ok {
		req.DNSServersOverride = expandStringsPtr(dnsServerOverride)
	}

	if dnsSearch, ok := d.GetOk("dns_search"); ok {
		req.DNSSearch = expandStringsPtr(dnsSearch)
	}

	if dsnLocalName, ok := d.GetOk("dns_local_name"); ok {
		req.DNSLocalName = expandStringPtr(dsnLocalName)
	}

	if address, ok := d.GetOk("address"); ok {
		req.Address = scw.IPPtr(net.ParseIP(address.(string)))
	}

	if renewTimer, ok := d.GetOk("renew_timer"); ok {
		req.RenewTimer = &scw.Duration{Seconds: renewTimer.(int64)}
	}

	if validLifetime, ok := d.GetOk("valid_lifetime"); ok {
		req.ValidLifetime = &scw.Duration{Seconds: validLifetime.(int64)}
	}

	if rebindTimer, ok := d.GetOk("rebind_timer"); ok {
		req.RebindTimer = &scw.Duration{Seconds: rebindTimer.(int64)}
	}

	if poolLow, ok := d.GetOk("pool_low"); ok {
		req.PoolLow = scw.IPPtr(net.ParseIP(poolLow.(string)))
	}

	if poolHigh, ok := d.GetOk("pool_high"); ok {
		req.PoolLow = scw.IPPtr(net.ParseIP(poolHigh.(string)))
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

	_ = d.Set("address", dhcp.Address.String())
	_ = d.Set("created_at", dhcp.CreatedAt.Format(time.RFC3339))
	_ = d.Set("dns_local_name", dhcp.DNSLocalName)
	_ = d.Set("dns_search", flattenSliceString(dhcp.DNSSearch))
	_ = d.Set("dns_server_override", flattenSliceString(dhcp.DNSServersOverride))
	_ = d.Set("enable_dynamic", dhcp.EnableDynamic)
	_ = d.Set("organization_id", dhcp.OrganizationID)
	_ = d.Set("pool_high", dhcp.PoolLow.String())
	_ = d.Set("pool_low", dhcp.PoolLow.String())
	_ = d.Set("project_id", dhcp.ProjectID)
	_ = d.Set("push_default_route", dhcp.PushDefaultRoute)
	_ = d.Set("push_dns_server", dhcp.PushDNSServer)
	_ = d.Set("rebind_timer", dhcp.RebindTimer.Seconds)
	_ = d.Set("renew_timer", dhcp.RenewTimer.Seconds)
	_ = d.Set("subnet", dhcp.Subnet.String())
	_ = d.Set("updated_at", dhcp.UpdatedAt.Format(time.RFC3339))
	_ = d.Set("valid_lifetime", dhcp.ValidLifetime.Seconds)
	_ = d.Set("zone", zone)

	return nil
}

func resourceScalewayVPCPublicGatewayDHCPUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vpcgwAPI, zone, ID, err := vpcgwAPIWithZoneAndID(meta, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	if d.HasChangesExcept("subnet", "address", "pool_low", "pool_high",
		"enable_dynamic", "push_default_route", "push_dns_server", "dns_servers_override",
		"dns_search", "dns_local_name", "renew_timer", "valid_lifetime", "rebind_timer") {
		req := &vpcgw.UpdateDHCPRequest{
			DHCPID: ID,
			Zone:   zone,
		}

		if subnetRaw, ok := d.GetOk("subnet"); ok {
			subnet, err := expandIPNet(subnetRaw.(string))
			if err != nil {
				return diag.FromErr(err)
			}
			req.Subnet = &subnet
		}

		if address, ok := d.GetOk("address"); ok {
			req.Address = scw.IPPtr(net.ParseIP(address.(string)))
		}

		if poolLow, ok := d.GetOk("pool_low"); ok {
			req.PoolLow = scw.IPPtr(net.ParseIP(poolLow.(string)))
		}

		if poolHigh, ok := d.GetOk("pool_low"); ok {
			req.PoolHigh = scw.IPPtr(net.ParseIP(poolHigh.(string)))
		}

		if pushDefaultRoute, ok := d.GetOk("push_default_route"); ok {
			req.PushDefaultRoute = expandBoolPtr(pushDefaultRoute)
		}

		if pushDNServer, ok := d.GetOk("push_dns_server"); ok {
			req.PushDNSServer = expandBoolPtr(pushDNServer)
		}

		if enableDynamic, ok := d.GetOk("enable_dynamic"); ok {
			req.EnableDynamic = expandBoolPtr(enableDynamic)
		}

		if dnsServerOverride, ok := d.GetOk("dns_servers_override"); ok {
			req.DNSServersOverride = expandStringsPtr(dnsServerOverride)
		}

		if dnsSearch, ok := d.GetOk("dns_search"); ok {
			req.DNSSearch = expandStringsPtr(dnsSearch)
		}

		if dsnLocalName, ok := d.GetOk("dns_local_name"); ok {
			req.DNSLocalName = expandStringPtr(dsnLocalName)
		}

		if renewTimer, ok := d.GetOk("renew_timer"); ok {
			req.RenewTimer = &scw.Duration{Seconds: renewTimer.(int64)}
		}

		if validLifetime, ok := d.GetOk("valid_lifetime"); ok {
			req.ValidLifetime = &scw.Duration{Seconds: validLifetime.(int64)}
		}

		if rebindTimer, ok := d.GetOk("rebind_timer"); ok {
			req.RebindTimer = &scw.Duration{Seconds: rebindTimer.(int64)}
		}

		if poolLow, ok := d.GetOk("pool_low"); ok {
			req.PoolLow = scw.IPPtr(net.ParseIP(poolLow.(string)))
		}

		if poolHigh, ok := d.GetOk("pool_high"); ok {
			req.PoolLow = scw.IPPtr(net.ParseIP(poolHigh.(string)))
		}

		_, err = vpcgwAPI.UpdateDHCP(req, scw.WithContext(ctx))
		if err != nil {
			return diag.FromErr(err)
		}
	}

	return resourceScalewayVPCPublicGatewayDHCPRead(ctx, d, meta)
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
