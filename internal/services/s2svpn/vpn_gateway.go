package s2svpn

import (
	"context"
	_ "time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	s2svpn "github.com/scaleway/scaleway-sdk-go/api/s2s_vpn/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/zonal"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/account"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/types"
)

func ResourceVPNGateway() *schema.Resource {
	return &schema.Resource{
		CreateContext: ResourceVPNGatewayCreate,
		ReadContext:   ResourceVPNGatewayRead,
		UpdateContext: ResourceVPNGatewayUpdate,
		DeleteContext: ResourceVPNGatewayDelete,
		Timeouts: &schema.ResourceTimeout{
			Create:  schema.DefaultTimeout(defaulVPNGatewayTimeout),
			Read:    schema.DefaultTimeout(defaulVPNGatewayTimeout),
			Update:  schema.DefaultTimeout(defaulVPNGatewayTimeout),
			Delete:  schema.DefaultTimeout(defaulVPNGatewayTimeout),
			Default: schema.DefaultTimeout(defaulVPNGatewayTimeout),
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 0,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Computed:    true,
				Optional:    true,
				Description: "The name of the VPN gateway",
			},
			"tags": {
				Type:        schema.TypeList,
				Optional:    true,
				Computed:    true,
				Description: "The list of tags to apply to the VPN gateway",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"gateway_type": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The VPN gateway type (commercial offer type)",
			},
			"public_config": {
				Type:        schema.TypeList,
				Computed:    true,
				Optional:    true,
				Description: "The public endpoint configuration of the VPN gateway",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"ipam_ipv4_id": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"ipam_ipv6_id": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
			"private_network_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The ID of the Private Network to attach to the VPN gateway",
			},
			"ipam_private_ipv4_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "The ID of the IPAM private IPv4 address to attach to the VPN gateway",
			},
			"ipam_private_ipv6_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Optional:    true,
				Description: "The ID of the IPAM private IPv6 address to attach to the VPN gateway",
			},
			"asn": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "The AS Number of the vpn gateway",
			},
			"status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The status of the VPN gateway",
			},
			"created_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The date and time of the creation of the VPN gateway",
			},
			"updated_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The date and time of the last update of the VPN gateway",
			},
			"zone":       zonal.OptionalSchema(),
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

func ResourceVPNGatewayCreate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	api, region, err := NewAPIWithRegion(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	zone := scw.Zone(d.Get("zone").(string))

	req := &s2svpn.CreateVpnGatewayRequest{
		Region:            region,
		ProjectID:         d.Get("project_id").(string),
		Name:              types.ExpandOrGenerateString(d.Get("name").(string), "connection"),
		Tags:              types.ExpandStrings(d.Get("tags")),
		GatewayType:       d.Get("gateway_type").(string),
		PrivateNetworkID:  regional.ExpandID(d.Get("private_network_id").(string)).ID,
		IpamPrivateIPv4ID: types.ExpandStringPtr(d.Get("ipam_private_ipv4_id").(string)),
		IpamPrivateIPv6ID: types.ExpandStringPtr(d.Get("ipam_private_ipv6_id").(string)),
		Zone:              &zone,
		PublicConfig:      expandVPNGatewayPublicConfig(d.Get("public_config")),
	}

	res, err := api.CreateVpnGateway(req, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = waitForVPNGateway(ctx, api, region, res.ID, d.Timeout(schema.TimeoutCreate))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(regional.NewIDString(region, res.ID))

	return ResourceVPNGatewayRead(ctx, d, m)
}

func ResourceVPNGatewayRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	api, region, id, err := NewAPIWithRegionAndID(m, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	gateway, err := waitForVPNGateway(ctx, api, region, id, d.Timeout(schema.TimeoutRead))
	if err != nil {
		if httperrors.Is404(err) {
			d.SetId("")

			return nil
		}

		return diag.FromErr(err)
	}

	_ = d.Set("name", gateway.Name)
	_ = d.Set("region", gateway.Region)
	_ = d.Set("project_id", gateway.ProjectID)
	_ = d.Set("organization_id", gateway.OrganizationID)
	_ = d.Set("tags", gateway.Tags)
	_ = d.Set("created_at", types.FlattenTime(gateway.CreatedAt))
	_ = d.Set("updated_at", types.FlattenTime(gateway.UpdatedAt))
	_ = d.Set("asn", int(gateway.Asn))
	_ = d.Set("status", gateway.Status.String())
	_ = d.Set("gateway_type", gateway.GatewayType)
	_ = d.Set("private_network_id", regional.NewIDString(region, gateway.PrivateNetworkID))
	_ = d.Set("ipam_private_ipv4_id", regional.NewIDString(region, gateway.IpamPrivateIPv4ID))
	_ = d.Set("ipam_private_ipv6_id", regional.NewIDString(region, gateway.IpamPrivateIPv6ID))
	_ = d.Set("zone", gateway.Zone)
	_ = d.Set("public_config", flattenVPNGatewayPublicConfig(region, gateway.PublicConfig))

	return nil
}

func ResourceVPNGatewayUpdate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	api, region, id, err := NewAPIWithRegionAndID(m, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	hasChanged := false

	req := &s2svpn.UpdateVpnGatewayRequest{
		Region:    region,
		GatewayID: id,
	}

	if d.HasChange("name") {
		req.Name = types.ExpandUpdatedStringPtr(d.Get("name"))
		hasChanged = true
	}

	if d.HasChange("tags") {
		req.Tags = types.ExpandUpdatedStringsPtr(d.Get("tags"))
		hasChanged = true
	}

	if hasChanged {
		_, err = api.UpdateVpnGateway(req, scw.WithContext(ctx))
		if err != nil {
			return diag.FromErr(err)
		}
	}

	_, err = waitForVPNGateway(ctx, api, region, id, d.Timeout(schema.TimeoutUpdate))
	if err != nil {
		return diag.FromErr(err)
	}

	return ResourceVPNGatewayRead(ctx, d, m)
}

func ResourceVPNGatewayDelete(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	api, region, id, err := NewAPIWithRegionAndID(m, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = waitForVPNGateway(ctx, api, region, id, d.Timeout(schema.TimeoutDelete))
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = api.DeleteVpnGateway(&s2svpn.DeleteVpnGatewayRequest{
		Region:    region,
		GatewayID: id,
	}, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = waitForVPNGateway(ctx, api, region, id, d.Timeout(schema.TimeoutDelete))
	if err != nil && !httperrors.Is404(err) {
		return diag.FromErr(err)
	}

	return nil
}
