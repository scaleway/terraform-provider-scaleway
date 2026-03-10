package vpcgw

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/zonal"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/account"
)

const dhcpDeprecationMessage = "The 'dhcp' resource is deprecated and no longer functional. " +
	"DHCP functionality was moved from Public Gateways to Private Networks, DHCP resources are now no longer needed. " +
	"Please remove this resource from your configuration. " +
	"For more information, please refer to the dedicated guide: " +
	"https://github.com/scaleway/terraform-provider-scaleway/blob/master/docs/guides/migration_guide_vpcgw_v2.md"

func ResourceDHCP() *schema.Resource {
	return &schema.Resource{
		CreateContext:      resourceVPCPublicGatewayDHCPCreate,
		ReadContext:        resourceVPCPublicGatewayDHCPRead,
		UpdateContext:      resourceVPCPublicGatewayDHCPUpdate,
		DeleteContext:      resourceVPCPublicGatewayDHCPDelete,
		SchemaVersion:      0,
		SchemaFunc:         dhcpSchema,
		DeprecationMessage: dhcpDeprecationMessage,
	}
}

func dhcpSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"project_id": account.ProjectIDSchema(),
		"zone":       zonal.Schema(),
		"subnet": {
			Type:         schema.TypeString,
			ValidateFunc: validation.IsCIDR,
			Required:     true,
			Description:  "Subnet for the DHCP server",
			Deprecated:   dhcpDeprecationMessage,
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
			Computed:    true,
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
			Description: "Whether the gateway should push a default route to DHCP clients or only hand out IPs. Defaults to true.",
		},
		"push_dns_server": {
			Type:        schema.TypeBool,
			Optional:    true,
			Computed:    true,
			Description: "Whether the gateway should push custom DNS servers to clients. This allows for instance hostname -> IP resolution. Defaults to true.",
		},
		"dns_servers_override": {
			Type:     schema.TypeList,
			Optional: true,
			Computed: true,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
			Description: "Override the DNS server list pushed to DHCP clients, instead of the gateway itself.",
		},
		"dns_search": {
			Type:     schema.TypeList,
			Optional: true,
			Computed: true,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
			Description: "Additional DNS search paths.",
		},
		"dns_local_name": {
			Type:        schema.TypeString,
			Optional:    true,
			Computed:    true,
			Description: "TLD given to hostnames in the Private Network. Allowed characters are `a-z0-9-.`. Defaults to the slugified Private Network name if created along a GatewayNetwork, or else to `priv`.",
		},
		"organization_id": account.OrganizationIDSchema(),
		"created_at": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "The date and time of the creation of the public gateway.",
		},
		"updated_at": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "The date and time of the last update of the public gateway.",
		},
	}
}

func resourceVPCPublicGatewayDHCPCreate(_ context.Context, _ *schema.ResourceData, _ any) diag.Diagnostics {
	return diag.Diagnostics{{
		Severity: diag.Error,
		Summary:  "scaleway_vpc_public_gateway_dhcp is no longer supported",
		Detail:   dhcpDeprecationMessage,
	}}
}

func resourceVPCPublicGatewayDHCPRead(_ context.Context, d *schema.ResourceData, _ any) diag.Diagnostics {
	d.SetId("")

	return nil
}

func resourceVPCPublicGatewayDHCPUpdate(_ context.Context, d *schema.ResourceData, _ any) diag.Diagnostics {
	d.SetId("")

	return nil
}

func resourceVPCPublicGatewayDHCPDelete(_ context.Context, _ *schema.ResourceData, _ any) diag.Diagnostics {
	return nil
}
