package vpcgw

import (
	"context"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	v1 "github.com/scaleway/scaleway-sdk-go/api/vpcgw/v1"
	"github.com/scaleway/scaleway-sdk-go/api/vpcgw/v2"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/dsf"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/zonal"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/account"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/types"
)

func ResourcePublicGateway() *schema.Resource {
	return &schema.Resource{
		CreateContext: ResourceVPCPublicGatewayCreate,
		ReadContext:   ResourceVPCPublicGatewayRead,
		UpdateContext: ResourceVPCPublicGatewayUpdate,
		DeleteContext: ResourceVPCPublicGatewayDelete,
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
			"name": {
				Type:        schema.TypeString,
				Computed:    true,
				Optional:    true,
				Description: "name of the gateway",
			},
			"type": {
				Type:             schema.TypeString,
				Required:         true,
				Description:      "gateway type",
				DiffSuppressFunc: dsf.IgnoreCase,
			},
			"upstream_dns_servers": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "override the gateway's default recursive DNS servers, if DNS features are enabled",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"ip_id": {
				Type:             schema.TypeString,
				Computed:         true,
				Optional:         true,
				ForceNew:         true,
				Description:      "attach an existing IP to the gateway",
				DiffSuppressFunc: dsf.Locality,
			},
			"tags": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "The tags associated with public gateway",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"bastion_enabled": {
				Type:        schema.TypeBool,
				Description: "Enable SSH bastion on the gateway",
				Optional:    true,
			},
			"bastion_port": {
				Type:        schema.TypeInt,
				Description: "Port of the SSH bastion",
				Optional:    true,
				Computed:    true,
			},
			"enable_smtp": {
				Type:        schema.TypeBool,
				Description: "Enable SMTP on the gateway",
				Optional:    true,
				Computed:    true,
			},
			"refresh_ssh_keys": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Trigger a refresh of the SSH keys for a given Public Gateway by changing this field's value",
			},
			"allowed_ip_ranges": {
				Type:        schema.TypeSet,
				Optional:    true,
				Computed:    true,
				Description: "Set a definitive list of IP ranges (in CIDR notation) allowed to connect to the SSH bastion",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"project_id": account.ProjectIDSchema(),
			"zone":       zonal.Schema(),
			// Computed elements
			"organization_id": account.OrganizationIDSchema(),
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
			"status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The status of the public gateway",
			},
			"bandwidth": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "The bandwidth available of the gateway",
			},
			"move_to_ipam": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Put a Public Gateway in IPAM mode, so that it can be used with the Public Gateways API v2",
			},
		},
	}
}

func ResourceVPCPublicGatewayCreate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	api, zone, err := newAPIWithZoneV2(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	req := &vpcgw.CreateGatewayRequest{
		Name:          types.ExpandOrGenerateString(d.Get("name"), "pn"),
		Type:          d.Get("type").(string),
		Tags:          types.ExpandStrings(d.Get("tags")),
		ProjectID:     d.Get("project_id").(string),
		EnableBastion: d.Get("bastion_enabled").(bool),
		Zone:          zone,
		EnableSMTP:    d.Get("enable_smtp").(bool),
	}

	if bastionPort, ok := d.GetOk("bastion_port"); ok {
		req.BastionPort = types.ExpandUint32Ptr(bastionPort.(int))
	}

	if ipID, ok := d.GetOk("ip_id"); ok {
		req.IPID = types.ExpandStringPtr(zonal.ExpandID(ipID).ID)
	}

	gateway, err := api.CreateGateway(req, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = waitForVPCPublicGatewayV2(ctx, api, zone, gateway.ID, d.Timeout(schema.TimeoutCreate))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(zonal.NewIDString(zone, gateway.ID))

	if allowedIps, ok := d.GetOk("allowed_ip_ranges"); ok {
		listIPs := allowedIps.(*schema.Set).List()

		_, err = api.SetBastionAllowedIPs(&vpcgw.SetBastionAllowedIPsRequest{
			GatewayID: gateway.ID,
			Zone:      zone,
			IPRanges:  types.ExpandStrings(listIPs),
		}, scw.WithContext(ctx))
		if err != nil {
			return diag.FromErr(err)
		}
	}

	return ResourceVPCPublicGatewayRead(ctx, d, m)
}

func ResourceVPCPublicGatewayRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	api, zone, id, err := NewAPIWithZoneAndIDv2(m, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	apiV1, _, _, err := NewAPIWithZoneAndID(m, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	gatewayV2, err := waitForVPCPublicGatewayV2(ctx, api, zone, id, d.Timeout(schema.TimeoutRead))
	if err != nil {
		if httperrors.Is412(err) {
			// Fallback to v1 API.
			tflog.Warn(ctx, "v2 API returned 412, falling back to v1 API to wait for public gateway stabilization")

			gatewayV1, err := waitForVPCPublicGateway(ctx, apiV1, zone, id, d.Timeout(schema.TimeoutCreate))
			if err != nil {
				return diag.FromErr(err)
			}

			return readVPCGWResourceDataV1(d, gatewayV1)
		} else if httperrors.Is404(err) {
			d.SetId("")

			return nil
		}

		return diag.FromErr(err)
	}

	return readVPCGWResourceDataV2(d, gatewayV2)
}

func ResourceVPCPublicGatewayUpdate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	api, zone, id, err := NewAPIWithZoneAndIDv2(m, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	apiV1, _, _, err := NewAPIWithZoneAndID(m, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	if d.HasChange("move_to_ipam") {
		err = apiV1.MigrateToV2(&v1.MigrateToV2Request{
			Zone:      zone,
			GatewayID: id,
		}, scw.WithContext(ctx))
		if err != nil {
			return diag.FromErr(err)
		}

		_, err = waitForVPCPublicGatewayV2(ctx, api, zone, id, d.Timeout(schema.TimeoutCreate))
		if err != nil {
			return diag.FromErr(err)
		}

		tflog.Info(ctx, "Public Gateway successfully moved to IPAM mode")
	}

	if err = updateGateway(ctx, d, api, apiV1, zone, id); err != nil {
		return diag.FromErr(err)
	}

	_, err = waitForVPCPublicGatewayV2(ctx, api, zone, id, d.Timeout(schema.TimeoutCreate))
	if err != nil {
		if httperrors.Is412(err) {
			tflog.Warn(ctx, "v2 API returned 412, falling back to v1 API to wait for public gateway stabilization")

			_, err = waitForVPCPublicGateway(ctx, apiV1, zone, id, d.Timeout(schema.TimeoutCreate))
			if err != nil {
				return diag.FromErr(err)
			}
		} else {
			return diag.FromErr(err)
		}
	}

	return ResourceVPCPublicGatewayRead(ctx, d, m)
}

func ResourceVPCPublicGatewayDelete(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	api, zone, id, err := NewAPIWithZoneAndIDv2(m, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	apiV1, _, _, err := NewAPIWithZoneAndID(m, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = waitForVPCPublicGatewayV2(ctx, api, zone, id, d.Timeout(schema.TimeoutCreate))
	if err != nil {
		if httperrors.Is412(err) {
			tflog.Warn(ctx, "v2 API returned 412, falling back to v1 API to wait for public gateway stabilization")

			_, err = waitForVPCPublicGateway(ctx, apiV1, zone, id, d.Timeout(schema.TimeoutCreate))
			if err != nil {
				return diag.FromErr(err)
			}
		}

		return diag.FromErr(err)
	}

	_, err = api.DeleteGateway(&vpcgw.DeleteGatewayRequest{
		GatewayID: id,
		Zone:      zone,
	}, scw.WithContext(ctx))
	if err != nil {
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
