package vpcgw

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scaleway/scaleway-sdk-go/api/vpcgw/v2"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/zonal"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/types"
)

func ResourceIPReverseDNS() *schema.Resource {
	return &schema.Resource{
		EnableLegacyTypeSystemApplyErrors: true,
		EnableLegacyTypeSystemPlanErrors:  true,
		CreateContext:                     ResourceVPCPublicGatewayIPReverseDNSCreate,
		ReadContext:                       ResourceVPCPublicGatewayIPReverseDNSRead,
		UpdateContext:                     ResourceVPCPublicGatewayIPReverseDNSUpdate,
		DeleteContext:                     ResourceVPCPublicGatewayIPReverseDNSDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Default: schema.DefaultTimeout(defaultIPReverseDNSTimeout),
			Create:  schema.DefaultTimeout(defaultIPReverseDNSTimeout),
			Update:  schema.DefaultTimeout(defaultIPReverseDNSTimeout),
		},
		SchemaVersion: 0,
		Schema: map[string]*schema.Schema{
			"gateway_ip_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The IP ID",
			},
			"reverse": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The reverse DNS for this IP",
			},
			"zone": zonal.Schema(),
		},
	}
}

func ResourceVPCPublicGatewayIPReverseDNSCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api, zone, err := newAPIWithZoneV2(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	res, err := api.GetIP(&vpcgw.GetIPRequest{
		Zone: zone,
		IPID: locality.ExpandID(d.Get("gateway_ip_id")),
	}, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(zonal.NewIDString(zone, res.ID))

	if _, ok := d.GetOk("reverse"); ok {
		tflog.Debug(ctx, fmt.Sprintf("updating IP %q reverse to %q\n", d.Id(), d.Get("reverse")))

		updateReverseReq := &vpcgw.UpdateIPRequest{
			Zone: zone,
			IPID: res.ID,
		}

		reverse := d.Get("reverse").(string)
		if len(reverse) > 0 {
			updateReverseReq.Reverse = types.ExpandStringPtr(reverse)
		}

		err := retryUpdateGatewayReverseDNS(ctx, api, updateReverseReq, d.Timeout(schema.TimeoutCreate))
		if err != nil {
			return diag.FromErr(err)
		}
	}

	return ResourceVPCPublicGatewayIPReverseDNSRead(ctx, d, m)
}

func ResourceVPCPublicGatewayIPReverseDNSRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api, zone, ID, err := NewAPIWithZoneAndIDv2(m, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	res, err := api.GetIP(&vpcgw.GetIPRequest{
		IPID: ID,
		Zone: zone,
	}, scw.WithContext(ctx))
	if err != nil {
		if httperrors.Is404(err) {
			d.SetId("")

			return nil
		}

		return diag.FromErr(err)
	}

	_ = d.Set("zone", string(zone))
	_ = d.Set("reverse", res.Reverse)

	return nil
}

func ResourceVPCPublicGatewayIPReverseDNSUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api, zone, ID, err := NewAPIWithZoneAndIDv2(m, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	if d.HasChange("reverse") {
		tflog.Debug(ctx, fmt.Sprintf("updating IP %q reverse to %q\n", d.Id(), d.Get("reverse")))

		updateReverseReq := &vpcgw.UpdateIPRequest{
			Zone: zone,
			IPID: ID,
		}

		reverse := d.Get("reverse").(string)
		if len(reverse) > 0 {
			updateReverseReq.Reverse = types.ExpandStringPtr(reverse)
		}

		err := retryUpdateGatewayReverseDNS(ctx, api, updateReverseReq, d.Timeout(schema.TimeoutUpdate))
		if err != nil {
			return diag.FromErr(err)
		}
	}

	return ResourceVPCPublicGatewayIPReverseDNSRead(ctx, d, m)
}

func ResourceVPCPublicGatewayIPReverseDNSDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api, zone, ID, err := NewAPIWithZoneAndIDv2(m, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	// Unset the reverse dns on the IP
	updateReverseReq := &vpcgw.UpdateIPRequest{
		Zone:    zone,
		IPID:    ID,
		Reverse: new(string),
	}

	_, err = api.UpdateIP(updateReverseReq, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")

	return nil
}
