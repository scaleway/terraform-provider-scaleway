package instance

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	instanceSDK "github.com/scaleway/scaleway-sdk-go/api/instance/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/cdf"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/identity"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/zonal"
)

func ResourceIPReverseDNS() *schema.Resource {
	return &schema.Resource{
		CreateContext: ResourceInstanceIPReverseDNSCreate,
		ReadContext:   ResourceInstanceIPReverseDNSRead,
		UpdateContext: ResourceInstanceIPReverseDNSUpdate,
		DeleteContext: ResourceInstanceIPReverseDNSDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Default: schema.DefaultTimeout(defaultInstanceIPTimeout),
			Create:  schema.DefaultTimeout(defaultInstanceIPReverseDNSTimeout),
			Update:  schema.DefaultTimeout(defaultInstanceIPReverseDNSTimeout),
		},
		SchemaVersion: 0,
		SchemaFunc:    ipReverseDNSSchema,
		CustomizeDiff: cdf.LocalityCheck("ip_id"),
		Identity:      identity.DefaultZonal(),
	}
}

func ipReverseDNSSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"ip_id": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "The IP ID or IP address",
		},
		"reverse": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "The reverse DNS for this IP",
		},
		"zone": zonal.Schema(),
	}
}

func ResourceInstanceIPReverseDNSCreate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	instanceAPI, zone, err := newAPIWithZone(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	res, err := instanceAPI.GetIP(&instanceSDK.GetIPRequest{
		IP:   locality.ExpandID(d.Get("ip_id")),
		Zone: zone,
	}, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	err = identity.SetZonalIdentity(d, res.IP.Zone, res.IP.ID)
	if err != nil {
		return diag.FromErr(err)
	}

	if _, ok := d.GetOk("reverse"); ok {
		tflog.Debug(ctx, fmt.Sprintf("updating IP %q reverse to %q\n", d.Id(), d.Get("reverse")))

		updateReverseReq := &instanceSDK.UpdateIPRequest{
			Zone: zone,
			IP:   res.IP.ID,
		}

		if reverse, ok := d.GetOk("reverse"); ok {
			updateReverseReq.Reverse = &instanceSDK.NullableStringValue{Value: reverse.(string)}
		} else {
			updateReverseReq.Reverse = &instanceSDK.NullableStringValue{Null: true}
		}

		err := retryUpdateReverseDNS(ctx, instanceAPI, updateReverseReq, d.Timeout(schema.TimeoutCreate))
		if err != nil {
			return diag.FromErr(err)
		}
	}

	return ResourceInstanceIPReverseDNSRead(ctx, d, m)
}

func ResourceInstanceIPReverseDNSRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	instanceAPI, zone, ID, err := NewAPIWithZoneAndID(m, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	res, err := instanceAPI.GetIP(&instanceSDK.GetIPRequest{
		IP:   ID,
		Zone: zone,
	}, scw.WithContext(ctx))
	if err != nil {
		// We check for 403 because instance API returns 403 for a deleted IP
		if httperrors.Is404(err) || httperrors.Is403(err) {
			d.SetId("")

			return nil
		}

		return diag.FromErr(err)
	}

	_ = d.Set("zone", string(zone))
	_ = d.Set("reverse", res.IP.Reverse)

	err = identity.SetZonalIdentity(d, res.IP.Zone, res.IP.ID)
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func ResourceInstanceIPReverseDNSUpdate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	instanceAPI, zone, ID, err := NewAPIWithZoneAndID(m, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	if d.HasChange("reverse") {
		tflog.Debug(ctx, fmt.Sprintf("updating IP %q reverse to %q\n", d.Id(), d.Get("reverse")))

		updateReverseReq := &instanceSDK.UpdateIPRequest{
			Zone: zone,
			IP:   ID,
		}

		if reverse, ok := d.GetOk("reverse"); ok {
			updateReverseReq.Reverse = &instanceSDK.NullableStringValue{Value: reverse.(string)}
		} else {
			updateReverseReq.Reverse = &instanceSDK.NullableStringValue{Null: true}
		}

		err := retryUpdateReverseDNS(ctx, instanceAPI, updateReverseReq, d.Timeout(schema.TimeoutUpdate))
		if err != nil {
			return diag.FromErr(err)
		}
	}

	return ResourceInstanceIPReverseDNSRead(ctx, d, m)
}

func ResourceInstanceIPReverseDNSDelete(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	instanceAPI, zone, ID, err := NewAPIWithZoneAndID(m, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	// Unset the reverse dns on the IP
	updateReverseReq := &instanceSDK.UpdateIPRequest{
		Zone:    zone,
		IP:      ID,
		Reverse: &instanceSDK.NullableStringValue{Null: true},
	}

	_, err = instanceAPI.UpdateIP(updateReverseReq, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")

	return nil
}
