package s2svpn

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	s2s_vpn "github.com/scaleway/scaleway-sdk-go/api/s2s_vpn/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/account"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/types"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/verify"
)

func ResourceConnection() *schema.Resource {
	return &schema.Resource{
		CreateContext: ResourceConnectionCreate,
		ReadContext:   ResourceConnectionRead,
		UpdateContext: ResourceConnectionUpdate,
		DeleteContext: ResourceConnectionDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 0,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Computed:    true,
				Optional:    true,
				Description: "The name of the connection",
			},
			"tags": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "The list of tags to apply to the connection",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"is_ipv6": {
				Type:        schema.TypeBool,
				Computed:    true,
				Optional:    true,
				Description: "Defines IP version of the IPSec Tunnel",
			},
			"initiation_policy": {
				Type:             schema.TypeString,
				Computed:         true,
				Optional:         true,
				Description:      "Defines who initiates the IPsec tunnel",
				ValidateDiagFunc: verify.ValidateEnum[s2s_vpn.CreateConnectionRequestInitiationPolicy](),
			},
			"ikev2_ciphers": {
				Type:        schema.TypeList,
				Optional:    true,
				Computed:    true,
				Description: "The list of IKE v2 ciphers proposed for the IPsec tunnel",
				Elem:        ResourceConnectionCipher(),
			},
			"esp_ciphers": {
				Type:        schema.TypeList,
				Optional:    true,
				Computed:    true,
				Description: "The list of ESP ciphers proposed for the IPsec tunnel",
				Elem:        ResourceConnectionCipher(),
			},
			"enable_route_propagation": {
				Type:        schema.TypeBool,
				Computed:    true,
				Optional:    true,
				Description: "Defines whether route propagation is enabled or not",
			},
			"vpn_gateway_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Optional:    true,
				Description: "The ID of the VPN gateway to attach to the connection",
			},
			"customer_gateway_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Optional:    true,
				Description: "The ID of the customer gateway to attach to the connection",
			},
			"bgp_config_ipv4": {
				Type:        schema.TypeList,
				Optional:    true,
				Computed:    true,
				Description: "The list of IKE v2 ciphers proposed for the IPsec tunnel",
				Elem:        ResourceConnectionRequestBgpConfig(),
			},
			"bgp_config_ipv6": {
				Type:        schema.TypeList,
				Optional:    true,
				Computed:    true,
				Description: "The list of IKE v2 ciphers proposed for the IPsec tunnel",
				Elem:        ResourceConnectionRequestBgpConfig(),
			},
			"bgp_status_ipv4": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The status of the BGP IPv4 session",
			},
			"bgp_status_ipv6": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The status of the BGP IPv6 session",
			},
			"bgp_session_ipv4": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "The BGP IPv4 session information (read-only)",
				Elem:        ResourceConnectionBgpSession(),
			},
			"bgp_session_ipv6": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "The BGP IPv6 session information (read-only)",
				Elem:        ResourceConnectionBgpSession(),
			},
			"route_propagation_enabled": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"created_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The date and time of the creation of the TLS stage",
			},
			"updated_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The date and time of the last update of the TLS stage",
			},
			"status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The status of the VPN gateway",
			},
			"tunnel_status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The status of the VPN gateway",
			},
			"secret_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The BGP peer IP on customer side",
			},
			"secret_version": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "The BGP peer IP on customer side",
			},
			"region":     regional.Schema(),
			"project_id": account.ProjectIDSchema(),
			"organization_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Organization ID of the Project",
			},
		},
	}
}

func ResourceConnectionCipher() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"encryption": {
				Type:     schema.TypeString,
				Required: true,
			},
			"integrity": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"dh_group": {
				Type:     schema.TypeString,
				Optional: true,
			},
		},
	}
}

func ResourceConnectionRequestBgpConfig() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"routing_policy_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"private_ip": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"peer_private_ip": {
				Type:     schema.TypeString,
				Optional: true,
			},
		},
	}
}

func ResourceConnectionBgpSession() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"routing_policy_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The routing policy ID",
			},
			"private_ip": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The BGP peer IP on Scaleway side",
			},
			"peer_private_ip": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The BGP peer IP on customer side",
			},
		},
	}
}

func ResourceConnectionCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api, region, err := NewAPIWithRegion(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	bgpConfigIpv4Config, err := expandConnectionRequestBgpConfig(d.Get("bgp_config_ipv4"))
	if err != nil {
		return diag.FromErr(err)
	}

	bgpConfigIpv6Config, err := expandConnectionRequestBgpConfig(d.Get("bgp_config_ipv6"))
	if err != nil {
		return diag.FromErr(err)
	}

	req := &s2s_vpn.CreateConnectionRequest{
		Region:                 region,
		ProjectID:              d.Get("project_id").(string),
		Name:                   types.ExpandOrGenerateString(d.Get("name").(string), "connection"),
		Tags:                   types.ExpandStrings(d.Get("tags")),
		IsIPv6:                 d.Get("is_ipv6").(bool),
		EnableRoutePropagation: d.Get("enable_route_propagation").(bool),
		InitiationPolicy:       s2s_vpn.CreateConnectionRequestInitiationPolicy(d.Get("initiation_policy").(string)),
		VpnGatewayID:           regional.ExpandID(d.Get("vpn_gateway_id").(string)).ID,
		CustomerGatewayID:      regional.ExpandID(d.Get("customer_gateway_id").(string)).ID,
		Ikev2Ciphers:           expandConnectionCiphers(d.Get("ikev2_ciphers")),
		EspCiphers:             expandConnectionCiphers(d.Get("esp_ciphers")),
	}

	if bgpConfigIpv4Config != nil {
		req.BgpConfigIPv4 = bgpConfigIpv4Config
	}

	if bgpConfigIpv6Config != nil {
		req.BgpConfigIPv6 = bgpConfigIpv6Config
	}

	res, err := api.CreateConnection(req, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(regional.NewIDString(region, res.Connection.ID))

	return ResourceConnectionRead(ctx, d, m)
}

func ResourceConnectionRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api, region, id, err := NewAPIWithRegionAndID(m, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	connection, err := api.GetConnection(&s2s_vpn.GetConnectionRequest{
		ConnectionID: id,
		Region:       region,
	}, scw.WithContext(ctx))
	if err != nil {
		if httperrors.Is404(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	_ = d.Set("name", connection.Name)
	_ = d.Set("region", connection.Region)
	_ = d.Set("project_id", connection.ProjectID)
	_ = d.Set("organization_id", connection.OrganizationID)
	_ = d.Set("tags", connection.Tags)
	_ = d.Set("created_at", types.FlattenTime(connection.CreatedAt))
	_ = d.Set("updated_at", types.FlattenTime(connection.UpdatedAt))
	_ = d.Set("status", connection.Status.String())
	_ = d.Set("is_ipv6", connection.IsIPv6)
	_ = d.Set("initiation_policy", connection.InitiationPolicy.String())
	_ = d.Set("route_propagation_enabled", connection.RoutePropagationEnabled)
	_ = d.Set("vpn_gateway_id", regional.NewIDString(region, connection.VpnGatewayID))
	_ = d.Set("customer_gateway_id", regional.NewIDString(region, connection.CustomerGatewayID))
	_ = d.Set("tunnel_status", connection.TunnelStatus.String())
	_ = d.Set("ikev2_ciphers", flattenConnectionCiphers(connection.Ikev2Ciphers))
	_ = d.Set("esp_ciphers", flattenConnectionCiphers(connection.EspCiphers))
	_ = d.Set("bgp_status_ipv4", connection.BgpStatusIPv4.String())
	_ = d.Set("bgp_status_ipv6", connection.BgpStatusIPv6.String())
	_ = d.Set("secret_id", regional.NewIDString(region, connection.SecretID))
	_ = d.Set("secret_version", int(connection.SecretRevision))

	bgpSessionIPv4, err := flattenBGPSession(region, connection.BgpSessionIPv4)
	if err != nil {
		return diag.FromErr(err)
	}
	_ = d.Set("bgp_session_ipv4", bgpSessionIPv4)

	bgpSessionIPv6, err := flattenBGPSession(region, connection.BgpSessionIPv6)
	if err != nil {
		return diag.FromErr(err)
	}
	_ = d.Set("bgp_session_ipv6", bgpSessionIPv6)

	return nil
}

func ResourceConnectionUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api, region, id, err := NewAPIWithRegionAndID(m, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	hasChanged := false

	req := &s2s_vpn.UpdateConnectionRequest{
		Region:       region,
		ConnectionID: id,
	}

	if d.HasChange("name") {
		req.Name = types.ExpandUpdatedStringPtr(d.Get("name"))
		hasChanged = true
	}

	if d.HasChange("tags") {
		req.Tags = types.ExpandUpdatedStringsPtr(d.Get("tags"))
		hasChanged = true
	}

	if d.HasChange("initiation_policy") {
		req.InitiationPolicy = s2s_vpn.CreateConnectionRequestInitiationPolicy(d.Get("initiation_policy").(string))
		hasChanged = true
	}

	if d.HasChange("ikev2_ciphers") {
		req.Ikev2Ciphers = expandConnectionCiphers(d.Get("ikev2_ciphers"))
		hasChanged = true
	}

	if d.HasChange("esp_ciphers") {
		req.EspCiphers = expandConnectionCiphers(d.Get("esp_ciphers"))
		hasChanged = true
	}

	if hasChanged {
		_, err = api.UpdateConnection(req, scw.WithContext(ctx))
		if err != nil {
			return diag.FromErr(err)
		}
	}

	return ResourceConnectionRead(ctx, d, m)
}

func ResourceConnectionDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api, region, id, err := NewAPIWithRegionAndID(m, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	err = api.DeleteConnection(&s2s_vpn.DeleteConnectionRequest{
		Region:       region,
		ConnectionID: id,
	}, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}
