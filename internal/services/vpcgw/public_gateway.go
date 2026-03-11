package vpcgw

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
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
		SchemaFunc:    publicGatewaySchema,
	}
}

func publicGatewaySchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
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
			Deprecated: "This field is no longer supported in the v2 API",
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
			ValidateFunc: func(val any, key string) ([]string, []error) {
				v := val.(int)
				if (v >= 1024 && v <= 59999) || v == 61000 {
					return nil, nil
				}

				return nil, []error{fmt.Errorf("expected bastion_port to be in the range (1024 - 59999) or default 61000, got %d", v)}
			},
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
			Deprecated:  "All gateways now use IPAM. This field is no longer needed",
		},
	}
}

func ResourceVPCPublicGatewayCreate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	api, zone, err := newAPIWithZone(d, m)
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

	_, err = waitForVPCPublicGateway(ctx, api, zone, gateway.ID, d.Timeout(schema.TimeoutCreate))
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
	api, zone, id, err := NewAPIWithZoneAndID(m, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	gateway, err := waitForVPCPublicGateway(ctx, api, zone, id, d.Timeout(schema.TimeoutRead))
	if err != nil {
		if httperrors.Is404(err) {
			d.SetId("")

			return nil
		}

		return diag.FromErr(err)
	}

	_ = d.Set("name", gateway.Name)
	_ = d.Set("type", gateway.Type)
	_ = d.Set("status", gateway.Status.String())
	_ = d.Set("organization_id", gateway.OrganizationID)
	_ = d.Set("project_id", gateway.ProjectID)
	_ = d.Set("zone", gateway.Zone)
	_ = d.Set("tags", gateway.Tags)
	_ = d.Set("created_at", gateway.CreatedAt.Format(time.RFC3339))
	_ = d.Set("updated_at", gateway.UpdatedAt.Format(time.RFC3339))
	_ = d.Set("bastion_enabled", gateway.BastionEnabled)
	_ = d.Set("bastion_port", int(gateway.BastionPort))
	_ = d.Set("enable_smtp", gateway.SMTPEnabled)
	_ = d.Set("bandwidth", int(gateway.Bandwidth))
	_ = d.Set("upstream_dns_servers", nil)

	if gateway.IPv4 != nil {
		_ = d.Set("ip_id", zonal.NewID(gateway.IPv4.Zone, gateway.IPv4.ID).String())
	}

	ips, err := flattenIPNetList(gateway.BastionAllowedIPs)
	if err != nil {
		return diag.FromErr(err)
	}

	_ = d.Set("allowed_ip_ranges", ips)

	return nil
}

func ResourceVPCPublicGatewayUpdate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	api, zone, id, err := NewAPIWithZoneAndID(m, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	updateRequest := &vpcgw.UpdateGatewayRequest{
		GatewayID: id,
		Zone:      zone,
	}

	if d.HasChange("name") {
		updateRequest.Name = new(d.Get("name").(string))
	}

	if d.HasChange("tags") {
		updateRequest.Tags = types.ExpandUpdatedStringsPtr(d.Get("tags"))
	}

	if d.HasChange("bastion_port") {
		updateRequest.BastionPort = new(uint32(d.Get("bastion_port").(int)))
	}

	if d.HasChange("bastion_enabled") {
		updateRequest.EnableBastion = new(d.Get("bastion_enabled").(bool))
	}

	if d.HasChange("enable_smtp") {
		updateRequest.EnableSMTP = new(d.Get("enable_smtp").(bool))
	}

	_, err = api.UpdateGateway(updateRequest, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = waitForVPCPublicGateway(ctx, api, zone, id, d.Timeout(schema.TimeoutUpdate))
	if err != nil {
		return diag.FromErr(err)
	}

	if d.HasChange("refresh_ssh_keys") {
		_, err = api.RefreshSSHKeys(&vpcgw.RefreshSSHKeysRequest{
			Zone:      zone,
			GatewayID: id,
		}, scw.WithContext(ctx))
		if err != nil {
			return diag.FromErr(err)
		}
	}

	_, err = waitForVPCPublicGateway(ctx, api, zone, id, d.Timeout(schema.TimeoutUpdate))
	if err != nil {
		return diag.FromErr(err)
	}

	if d.HasChange("type") {
		_, err = api.UpgradeGateway(&vpcgw.UpgradeGatewayRequest{
			Zone:      zone,
			GatewayID: id,
			Type:      types.ExpandUpdatedStringPtr(d.Get("type")),
		}, scw.WithContext(ctx))
		if err != nil {
			return diag.FromErr(err)
		}
	}

	_, err = waitForVPCPublicGateway(ctx, api, zone, id, d.Timeout(schema.TimeoutUpdate))
	if err != nil {
		return diag.FromErr(err)
	}

	if d.HasChange("allowed_ip_ranges") {
		listIPs := d.Get("allowed_ip_ranges").(*schema.Set).List()

		_, err = api.SetBastionAllowedIPs(&vpcgw.SetBastionAllowedIPsRequest{
			GatewayID: id,
			Zone:      zone,
			IPRanges:  types.ExpandStrings(listIPs),
		}, scw.WithContext(ctx))
		if err != nil {
			return diag.FromErr(err)
		}
	}

	_, err = waitForVPCPublicGateway(ctx, api, zone, id, d.Timeout(schema.TimeoutUpdate))
	if err != nil {
		return diag.FromErr(err)
	}

	return ResourceVPCPublicGatewayRead(ctx, d, m)
}

func ResourceVPCPublicGatewayDelete(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	api, zone, id, err := NewAPIWithZoneAndID(m, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = waitForVPCPublicGateway(ctx, api, zone, id, d.Timeout(schema.TimeoutDelete))
	if err != nil {
		if httperrors.Is404(err) {
			return nil
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

	_, err = waitForVPCPublicGateway(ctx, api, zone, id, d.Timeout(schema.TimeoutDelete))
	if err != nil {
		if httperrors.Is404(err) {
			return nil
		}

		return diag.FromErr(err)
	}

	return nil
}
