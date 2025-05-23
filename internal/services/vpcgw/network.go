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
		EnableLegacyTypeSystemApplyErrors: true,
		EnableLegacyTypeSystemPlanErrors:  true,
		CreateContext:                     ResourceVPCGatewayNetworkCreate,
		ReadContext:                       ResourceVPCGatewayNetworkRead,
		UpdateContext:                     ResourceVPCGatewayNetworkUpdate,
		DeleteContext:                     ResourceVPCGatewayNetworkDelete,
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
		Schema: map[string]*schema.Schema{
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
					if v, ok := d.Get("ipam_config").([]interface{}); ok && len(v) > 0 {
						return true
					}

					return oldValue == newValue
				},
				Deprecated: "Please use ipam_config. For more information, please refer to the dedicated guide: https://github.com/scaleway/terraform-provider-scaleway/blob/master/docs/guides/migration_guide_vpcgw_v2.md",
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
				Deprecated:  "Please use ipam_config. For more information, please refer to the dedicated guide: https://github.com/scaleway/terraform-provider-scaleway/blob/master/docs/guides/migration_guide_vpcgw_v2.md",
			},
			"cleanup_dhcp": {
				Type:        schema.TypeBool,
				Optional:    true,
				Computed:    true,
				Description: "Remove DHCP config on this network on destroy",
				Deprecated:  "Please use ipam_config. For more information, please refer to the dedicated guide: https://github.com/scaleway/terraform-provider-scaleway/blob/master/docs/guides/migration_guide_vpcgw_v2.md",
			},
			"static_address": {
				Type:          schema.TypeString,
				Description:   "The static IP address in CIDR on this network",
				Optional:      true,
				Computed:      true,
				ValidateFunc:  validation.IsCIDR,
				ConflictsWith: []string{"dhcp_id", "ipam_config"},
				Deprecated:    "Please use ipam_config. For more information, please refer to the dedicated guide: https://github.com/scaleway/terraform-provider-scaleway/blob/master/docs/guides/migration_guide_vpcgw_v2.md",
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
			// Computed elements
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
		},
		CustomizeDiff: cdf.LocalityCheck("gateway_id", "private_network_id", "dhcp_id"),
	}
}

func ResourceVPCGatewayNetworkCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api, zone, err := newAPIWithZoneV2(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	gatewayID := zonal.ExpandID(d.Get("gateway_id").(string)).ID

	gateway, err := waitForVPCPublicGatewayV2(ctx, api, zone, gatewayID, d.Timeout(schema.TimeoutCreate))
	if err != nil {
		return diag.FromErr(err)
	}

	pushDefaultRoute, ipamIPID := expandIpamConfigV2(d.Get("ipam_config"))

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

		return waitForVPCPublicGatewayV2(ctx, api, zone, gatewayID, d.Timeout(schema.TimeoutCreate))
	})
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(zonal.NewIDString(zone, gatewayNetwork.ID))

	_, err = waitForVPCPublicGatewayV2(ctx, api, zone, gatewayNetwork.GatewayID, d.Timeout(schema.TimeoutCreate))
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = waitForVPCGatewayNetworkV2(ctx, api, zone, gatewayNetwork.ID, d.Timeout(schema.TimeoutCreate))
	if err != nil {
		return diag.FromErr(err)
	}

	return ResourceVPCGatewayNetworkRead(ctx, d, m)
}

func ResourceVPCGatewayNetworkRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api, zone, ID, err := NewAPIWithZoneAndIDv2(m, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	apiV1, _, _, err := NewAPIWithZoneAndID(m, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	var diags diag.Diagnostics

	gatewayNetwork, err := waitForVPCGatewayNetworkV2(ctx, api, zone, ID, d.Timeout(schema.TimeoutRead))
	if err != nil {
		if httperrors.Is412(err) {
			// Fallback to v1 API.
			tflog.Warn(ctx, "v2 API returned 412, falling back to v1 API to wait for gateway network stabilization")

			gatewayV1, err := waitForVPCGatewayNetwork(ctx, apiV1, zone, ID, d.Timeout(schema.TimeoutRead))
			if err != nil {
				return diag.FromErr(err)
			}

			if gatewayNetwork.PrivateNetworkID != "" {
				diags = setPrivateIPsV1(ctx, d, apiV1, gatewayV1, m)
			}

			return readVPCGWNetworkResourceDataV1(d, gatewayV1, diags)
		} else if httperrors.Is404(err) {
			d.SetId("")

			return nil
		}

		return diag.FromErr(err)
	}

	if gatewayNetwork.PrivateNetworkID != "" {
		diags = setPrivateIPsV2(ctx, d, api, gatewayNetwork, m)
	}

	return readVPCGWNetworkResourceDataV2(d, gatewayNetwork, diags)
}

func ResourceVPCGatewayNetworkUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api, zone, id, err := NewAPIWithZoneAndIDv2(m, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	apiV1, _, _, err := NewAPIWithZoneAndID(m, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = waitForVPCGatewayNetworkV2(ctx, api, zone, id, d.Timeout(schema.TimeoutUpdate))
	if err != nil {
		if httperrors.Is412(err) {
			tflog.Warn(ctx, "v2 API returned 412, falling back to v1 API to wait for gateway network stabilization")

			_, err = waitForVPCGatewayNetwork(ctx, apiV1, zone, id, d.Timeout(schema.TimeoutCreate))
			if err != nil {
				return diag.FromErr(err)
			}
		}

		return diag.FromErr(err)
	}

	if err = updateGWNetwork(ctx, d, api, apiV1, zone, id); err != nil {
		return diag.FromErr(err)
	}

	_, err = waitForVPCGatewayNetworkV2(ctx, api, zone, id, d.Timeout(schema.TimeoutUpdate))
	if err != nil {
		if httperrors.Is412(err) {
			tflog.Warn(ctx, "v2 API returned 412, falling back to v1 API to wait for gateway network stabilization")

			_, err = waitForVPCGatewayNetwork(ctx, apiV1, zone, id, d.Timeout(schema.TimeoutCreate))
			if err != nil {
				return diag.FromErr(err)
			}
		}

		return diag.FromErr(err)
	}

	return ResourceVPCGatewayNetworkRead(ctx, d, m)
}

func ResourceVPCGatewayNetworkDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api, zone, id, err := NewAPIWithZoneAndIDv2(m, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	apiV1, _, _, err := NewAPIWithZoneAndID(m, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	gwNetwork, err := waitForVPCGatewayNetworkV2(ctx, api, zone, id, d.Timeout(schema.TimeoutDelete))
	if err != nil {
		if httperrors.Is412(err) {
			tflog.Warn(ctx, "v2 API returned 412, falling back to v1 API to wait for gateway network stabilization")

			_, err = waitForVPCGatewayNetwork(ctx, apiV1, zone, id, d.Timeout(schema.TimeoutCreate))
			if err != nil {
				return diag.FromErr(err)
			}
		}

		return diag.FromErr(err)
	}

	req := &vpcgw.DeleteGatewayNetworkRequest{
		GatewayNetworkID: gwNetwork.ID,
		Zone:             gwNetwork.Zone,
	}

	_, err = api.DeleteGatewayNetwork(req, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = waitForVPCGatewayNetworkV2(ctx, api, zone, id, d.Timeout(schema.TimeoutDelete))

	switch {
	case err == nil:
	case httperrors.Is404(err):
		return nil
	case httperrors.Is412(err):
		tflog.Warn(ctx, "v2 API returned 412, falling back to v1 API to wait for gateway network stabilization")

		_, err = waitForVPCGatewayNetwork(ctx, apiV1, zone, id, d.Timeout(schema.TimeoutDelete))
		if err != nil && !httperrors.Is404(err) {
			return diag.FromErr(err)
		}
	default:
		return diag.FromErr(err)
	}

	_, err = waitForVPCPublicGatewayV2(ctx, api, zone, id, d.Timeout(schema.TimeoutDelete))

	switch {
	case err == nil:
	case httperrors.Is404(err):
		return nil
	case httperrors.Is412(err):
		tflog.Warn(ctx, "v2 API returned 412, falling back to v1 API to wait for public gateway stabilization")

		_, err = waitForVPCPublicGateway(ctx, apiV1, zone, id, d.Timeout(schema.TimeoutDelete))
		if err != nil && !httperrors.Is404(err) {
			return diag.FromErr(err)
		}
	default:
		return diag.FromErr(err)
	}

	return nil
}
