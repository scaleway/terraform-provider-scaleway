package scaleway

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scaleway/scaleway-sdk-go/api/instance/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

func resourceScalewayInstanceIP() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceScalewayInstanceIPCreate,
		ReadContext:   resourceScalewayInstanceIPRead,
		UpdateContext: resourceScalewayInstanceIPUpdate,
		DeleteContext: resourceScalewayInstanceIPDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Default: schema.DefaultTimeout(defaultInstanceIPTimeout),
		},
		SchemaVersion: 0,
		Schema: map[string]*schema.Schema{
			"address": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The IP address",
			},
			"prefix": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The IP prefix",
			},
			"type": {
				Type:        schema.TypeString,
				Computed:    true,
				Optional:    true,
				Description: "The type of instance IP",
			},
			"reverse": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The reverse DNS for this IP",
			},
			"server_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The server associated with this IP",
			},
			"tags": {
				Type: schema.TypeList,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Optional:    true,
				Description: "The tags associated with the ip",
			},
			"zone":            zoneSchema(),
			"organization_id": organizationIDSchema(),
			"project_id":      projectIDSchema(),
		},
		CustomizeDiff: func(ctx context.Context, diff *schema.ResourceDiff, i interface{}) error {
			// The only allowed change is
			// nat -> routed_ipv4
			if diff.HasChange("type") {
				before, after := diff.GetChange("type")
				oldType := instance.IPType(before.(string))
				newType := instance.IPType(after.(string))

				if oldType == "nat" && newType == "routed_ipv4" {
					return nil
				}

				return diff.ForceNew("type")
			}

			return nil
		},
	}
}

func resourceScalewayInstanceIPCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	instanceAPI, zone, err := instanceAPIWithZone(d, meta)
	if err != nil {
		return diag.FromErr(err)
	}
	iprequest := &instance.CreateIPRequest{
		Zone:    zone,
		Project: expandStringPtr(d.Get("project_id")),
		Type:    instance.IPType(d.Get("type").(string)),
	}
	tags := expandStrings(d.Get("tags"))
	if len(tags) > 0 {
		iprequest.Tags = tags
	}
	res, err := instanceAPI.CreateIP(iprequest, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	reverseRaw, ok := d.GetOk("reverse")
	if ok {
		reverseStrPtr := expandStringPtr(reverseRaw)
		req := &instance.UpdateIPRequest{
			IP:      res.IP.ID,
			Reverse: &instance.NullableStringValue{Value: *reverseStrPtr},
			Zone:    zone,
		}
		_, err = instanceAPI.UpdateIP(req, scw.WithContext(ctx))
		if err != nil {
			return diag.FromErr(err)
		}
	}

	d.SetId(newZonedIDString(zone, res.IP.ID))
	return resourceScalewayInstanceIPRead(ctx, d, meta)
}

func resourceScalewayInstanceIPUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	instanceAPI, zone, ID, err := instanceAPIWithZoneAndID(meta, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	req := &instance.UpdateIPRequest{
		IP:   ID,
		Zone: zone,
	}

	if d.HasChange("tags") {
		req.Tags = expandUpdatedStringsPtr(d.Get("tags"))
	}

	if d.HasChange("type") {
		req.Type = instance.IPType(d.Get("type").(string))
	}

	_, err = instanceAPI.UpdateIP(req, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceScalewayInstanceIPRead(ctx, d, meta)
}

func resourceScalewayInstanceIPRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	instanceAPI, zone, ID, err := instanceAPIWithZoneAndID(meta, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	res, err := instanceAPI.GetIP(&instance.GetIPRequest{
		IP:   ID,
		Zone: zone,
	}, scw.WithContext(ctx))
	if err != nil {
		// We check for 403 because instance API returns 403 for a deleted IP
		if is404Error(err) || is403Error(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	address := res.IP.Address.String()
	prefix := res.IP.Prefix.String()
	if prefix == netIPNil {
		ipnet := scw.IPNet{}
		_ = (&ipnet).UnmarshalJSON([]byte("\"" + res.IP.Address.String() + "\""))
		prefix = ipnet.String()
	}
	if address == netIPNil {
		address = res.IP.Prefix.IP.String()
	}

	_ = d.Set("address", address)
	_ = d.Set("prefix", prefix)
	_ = d.Set("zone", zone)
	_ = d.Set("organization_id", res.IP.Organization)
	_ = d.Set("project_id", res.IP.Project)
	_ = d.Set("reverse", res.IP.Reverse)
	_ = d.Set("type", res.IP.Type)

	if len(res.IP.Tags) > 0 {
		_ = d.Set("tags", flattenSliceString(res.IP.Tags))
	}

	if res.IP.Server != nil {
		_ = d.Set("server_id", newZonedIDString(res.IP.Zone, res.IP.Server.ID))
	} else {
		_ = d.Set("server_id", "")
	}

	return nil
}

func resourceScalewayInstanceIPDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	instanceAPI, zone, ID, err := instanceAPIWithZoneAndID(meta, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	err = instanceAPI.DeleteIP(&instance.DeleteIPRequest{
		IP:   ID,
		Zone: zone,
	}, scw.WithContext(ctx))

	if err != nil {
		// We check for 403 because instance API returns 403 for a deleted IP
		if is404Error(err) || is403Error(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	return nil
}
