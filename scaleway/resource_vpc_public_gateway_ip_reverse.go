package scaleway

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scaleway/scaleway-sdk-go/api/vpcgw/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

func resourceScalewayVPCPublicGatewayIPReverseDNS() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceScalewayVPCPublicGatewayIPReverseDNSCreate,
		ReadContext:   resourceScalewayVPCPublicGatewayIPReverseDNSRead,
		UpdateContext: resourceScalewayVPCPublicGatewayIPReverseDNSUpdate,
		DeleteContext: resourceScalewayVPCPublicGatewayIPReverseDNSDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Default: schema.DefaultTimeout(defaultVPCPublicGatewayIPReverseDNSTimeout),
			Create:  schema.DefaultTimeout(defaultVPCPublicGatewayIPReverseDNSTimeout),
			Update:  schema.DefaultTimeout(defaultVPCPublicGatewayIPReverseDNSTimeout),
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
			"zone": zoneSchema(),
		},
	}
}

func resourceScalewayVPCPublicGatewayIPReverseDNSCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vpcgwAPI, zone, err := vpcgwAPIWithZone(d, meta)
	if err != nil {
		return diag.FromErr(err)
	}

	res, err := vpcgwAPI.GetIP(&vpcgw.GetIPRequest{
		Zone: zone,
		IPID: expandID(d.Get("gateway_ip_id")),
	}, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(newZonedIDString(zone, res.ID))

	if _, ok := d.GetOk("reverse"); ok {
		tflog.Debug(ctx, fmt.Sprintf("updating IP %q reverse to %q\n", d.Id(), d.Get("reverse")))

		updateReverseReq := &vpcgw.UpdateIPRequest{
			Zone: zone,
			IPID: res.ID,
		}

		reverse := d.Get("reverse").(string)
		if len(reverse) > 0 {
			if isGatewayIPReverseResolved(ctx, vpcgwAPI, reverse, d.Timeout(schema.TimeoutCreate), res.ID, zone) {
				updateReverseReq.Reverse = expandStringPtr(reverse)
			} else {
				return diag.FromErr(fmt.Errorf("your reverse must resolve. Ensure the command 'dig +short %s' matches your IP address ", reverse))
			}
		}
		_, err = vpcgwAPI.UpdateIP(updateReverseReq, scw.WithContext(ctx))
		if err != nil {
			return diag.FromErr(err)
		}
	}

	return resourceScalewayVPCPublicGatewayIPReverseDNSRead(ctx, d, meta)
}

func resourceScalewayVPCPublicGatewayIPReverseDNSRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vpcgwAPI, zone, ID, err := vpcgwAPIWithZoneAndID(meta, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	res, err := vpcgwAPI.GetIP(&vpcgw.GetIPRequest{
		IPID: ID,
		Zone: zone,
	}, scw.WithContext(ctx))
	if err != nil {
		if is404Error(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	_ = d.Set("zone", string(zone))
	_ = d.Set("reverse", res.Reverse)

	return nil
}

func resourceScalewayVPCPublicGatewayIPReverseDNSUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vpcgwAPI, zone, ID, err := vpcgwAPIWithZoneAndID(meta, d.Id())
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
			if isGatewayIPReverseResolved(ctx, vpcgwAPI, reverse, d.Timeout(schema.TimeoutUpdate), ID, zone) {
				updateReverseReq.Reverse = expandStringPtr(reverse)
			} else {
				return diag.FromErr(fmt.Errorf("your reverse must resolve. Ensure the command 'dig +short %s' matches your IP address ", reverse))
			}
		}
		_, err = vpcgwAPI.UpdateIP(updateReverseReq, scw.WithContext(ctx))
		if err != nil {
			return diag.FromErr(err)
		}
	}

	return resourceScalewayVPCPublicGatewayIPReverseDNSRead(ctx, d, meta)
}

func resourceScalewayVPCPublicGatewayIPReverseDNSDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vpcgwAPI, zone, ID, err := vpcgwAPIWithZoneAndID(meta, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	// Unset the reverse dns on the IP
	updateReverseReq := &vpcgw.UpdateIPRequest{
		Zone:    zone,
		IPID:    ID,
		Reverse: new(string),
	}
	_, err = vpcgwAPI.UpdateIP(updateReverseReq, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")
	return nil
}
