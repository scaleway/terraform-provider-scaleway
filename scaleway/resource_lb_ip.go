package scaleway

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scaleway/scaleway-sdk-go/api/lb/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

func resourceScalewayLbIP() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceScalewayLbIPCreate,
		ReadContext:   resourceScalewayLbIPRead,
		UpdateContext: resourceScalewayLbIPUpdate,
		DeleteContext: resourceScalewayLbIPDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Default: schema.DefaultTimeout(defaultLbLbTimeout),
		},
		SchemaVersion: 1,
		StateUpgraders: []schema.StateUpgrader{
			{Version: 0, Type: lbUpgradeV1SchemaType(), Upgrade: lbUpgradeV1SchemaUpgradeFunc},
		},
		Schema: map[string]*schema.Schema{
			"reverse": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "The reverse domain name for this IP",
			},
			"zone": {
				Type:             schema.TypeString,
				Description:      "The zone you want to attach the resource to",
				Optional:         true,
				ForceNew:         true,
				ValidateDiagFunc: validateStringInSliceWithWarning(AllZones(), "zone"),
				Default:          scw.ZoneFrPar1,
			},
			// Computed
			"organization_id": organizationIDSchema(),
			"project_id":      projectIDSchema(),
			"ip_address": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The load-balancer public IP address",
			},
			"lb_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The ID of the load balancer attached to this IP, if any",
			},
			"region": regionComputedSchema(),
		},
	}
}

func resourceScalewayLbIPCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	lbAPI, zone, err := lbAPIWithZone(d, meta)
	if err != nil {
		return diag.FromErr(err)
	}

	zoneAttribute, ok := d.GetOk("zone")
	if ok {
		zone = scw.Zone(zoneAttribute.(string))
	}

	createReq := &lb.ZonedAPICreateIPRequest{
		Zone:      zone,
		ProjectID: expandStringPtr(d.Get("project_id")),
		Reverse:   expandStringPtr(d.Get("reverse")),
	}

	res, err := lbAPI.CreateIP(createReq, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(newZonedIDString(zone, res.ID))

	return resourceScalewayLbIPRead(ctx, d, meta)
}

func resourceScalewayLbIPRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	lbAPI, zone, ID, err := lbAPIWithZoneAndID(meta, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	var ip *lb.IP
	err = resource.RetryContext(ctx, retryLbIPInterval, func() *resource.RetryError {
		res, errGet := lbAPI.GetIP(&lb.ZonedAPIGetIPRequest{
			Zone: zone,
			IPID: ID,
		}, scw.WithContext(ctx))
		if err != nil {
			if is403Error(errGet) {
				return resource.RetryableError(errGet)
			}
			return resource.NonRetryableError(errGet)
		}

		ip = res
		return nil
	})

	if err != nil {
		if is404Error(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	// check lb state if it is attached
	if ip.LBID != nil {
		_, err = lbAPI.WaitForLb(&lb.ZonedAPIWaitForLBRequest{
			Zone:          zone,
			LBID:          *ip.LBID,
			Timeout:       scw.TimeDurationPtr(defaultInstanceServerWaitTimeout),
			RetryInterval: scw.TimeDurationPtr(defaultWaitLBRetryInterval),
		}, scw.WithContext(ctx))
		if err != nil {
			if is403Error(err) {
				d.SetId("")
				return nil
			}
			return diag.FromErr(err)
		}
	}

	// set the region from zone
	region, err := zone.Region()
	if err != nil {
		return diag.FromErr(err)
	}

	_ = d.Set("region", string(region))
	_ = d.Set("zone", ip.Zone.String())
	_ = d.Set("organization_id", ip.OrganizationID)
	_ = d.Set("project_id", ip.ProjectID)
	_ = d.Set("ip_address", ip.IPAddress)
	_ = d.Set("reverse", ip.Reverse)
	_ = d.Set("lb_id", flattenStringPtr(ip.LBID))

	return nil
}

func resourceScalewayLbIPUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	lbAPI, zone, ID, err := lbAPIWithZoneAndID(meta, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	var ip *lb.IP
	err = resource.RetryContext(ctx, retryLbIPInterval, func() *resource.RetryError {
		res, errGet := lbAPI.GetIP(&lb.ZonedAPIGetIPRequest{
			Zone: zone,
			IPID: ID,
		}, scw.WithContext(ctx))
		if err != nil {
			if is403Error(errGet) {
				return resource.RetryableError(errGet)
			}
			return resource.NonRetryableError(errGet)
		}

		ip = res
		return nil
	})

	if err != nil {
		if is404Error(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	if ip.LBID != nil {
		_, err = lbAPI.WaitForLb(&lb.ZonedAPIWaitForLBRequest{
			Zone:          zone,
			LBID:          *ip.LBID,
			Timeout:       scw.TimeDurationPtr(defaultInstanceServerWaitTimeout),
			RetryInterval: scw.TimeDurationPtr(defaultWaitLBRetryInterval),
		}, scw.WithContext(ctx))
		if err != nil {
			if is403Error(err) {
				d.SetId("")
				return nil
			}
			return diag.FromErr(err)
		}
	}

	if d.HasChange("reverse") {
		req := &lb.ZonedAPIUpdateIPRequest{
			Zone:    zone,
			IPID:    ID,
			Reverse: expandStringPtr(d.Get("reverse")),
		}

		_, err = lbAPI.UpdateIP(req, scw.WithContext(ctx))
		if err != nil {
			return diag.FromErr(err)
		}
	}

	if ip.LBID != nil {
		_, err = lbAPI.WaitForLb(&lb.ZonedAPIWaitForLBRequest{
			Zone:          zone,
			LBID:          *ip.LBID,
			Timeout:       scw.TimeDurationPtr(defaultInstanceServerWaitTimeout),
			RetryInterval: scw.TimeDurationPtr(defaultWaitLBRetryInterval),
		}, scw.WithContext(ctx))
		if err != nil {
			if is403Error(err) {
				d.SetId("")
				return nil
			}
			return diag.FromErr(err)
		}
	}

	return resourceScalewayLbIPRead(ctx, d, meta)
}

func resourceScalewayLbIPDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	lbAPI, zone, ID, err := lbAPIWithZoneAndID(meta, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	var ip *lb.IP
	err = resource.RetryContext(ctx, retryLbIPInterval, func() *resource.RetryError {
		res, errGet := lbAPI.GetIP(&lb.ZonedAPIGetIPRequest{
			Zone: zone,
			IPID: ID,
		}, scw.WithContext(ctx))
		if err != nil {
			if is403Error(errGet) {
				return resource.RetryableError(errGet)
			}
			return resource.NonRetryableError(errGet)
		}

		ip = res
		return nil
	})

	if err != nil {
		return diag.FromErr(err)
	}

	// check lb state
	if ip != nil && ip.LBID != nil {
		_, err = lbAPI.WaitForLbInstances(&lb.ZonedAPIWaitForLBInstancesRequest{
			LBID:          *ip.LBID,
			Zone:          zone,
			Timeout:       scw.TimeDurationPtr(lbWaitForTimeout),
			RetryInterval: scw.TimeDurationPtr(defaultWaitLBRetryInterval),
		}, scw.WithContext(ctx))
		if err != nil {
			if is403Error(err) {
				d.SetId("")
				return nil
			}
			return diag.FromErr(err)
		}
	}

	err = lbAPI.ReleaseIP(&lb.ZonedAPIReleaseIPRequest{
		Zone: zone,
		IPID: ID,
	}, scw.WithContext(ctx))

	if err != nil && !is404Error(err) {
		return diag.FromErr(err)
	}

	// check lb state
	if ip != nil && ip.LBID != nil {
		_, err = lbAPI.WaitForLbInstances(&lb.ZonedAPIWaitForLBInstancesRequest{
			LBID:          *ip.LBID,
			Zone:          zone,
			Timeout:       scw.TimeDurationPtr(lbWaitForTimeout),
			RetryInterval: scw.TimeDurationPtr(defaultWaitLBRetryInterval),
		}, scw.WithContext(ctx))
		if err != nil {
			if is404Error(err) || is403Error(err) {
				d.SetId("")
				return nil
			}
			return diag.FromErr(err)
		}
	}

	return nil
}
