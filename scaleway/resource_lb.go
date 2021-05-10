package scaleway

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scaleway/scaleway-sdk-go/api/lb/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"

	"github.com/hashicorp/go-cty/cty"
)

func resourceScalewayLb() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceScalewayLbCreate,
		ReadContext:   resourceScalewayLbRead,
		UpdateContext: resourceScalewayLbUpdate,
		DeleteContext: resourceScalewayLbDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Default: schema.DefaultTimeout(defaultLbLbTimeout),
		},
		SchemaVersion: 1,
		StateUpgraders: []schema.StateUpgrader{
			{
				Version: 0,
				Type: cty.Object(map[string]cty.Type{
					"id": cty.String,
					"ip_id": cty.String,
				}),
				Upgrade: resourceScalewayLbUpgradeIPID,
			},
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "Name of the lb",
			},
			"type": {
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				DiffSuppressFunc: diffSuppressFuncIgnoreCase,
				Description:      "The type of load-balancer you want to create",
			},
			"tags": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Description: "Array of tags to associate with the load-balancer",
			},
			"ip_id": {
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				Description:      "The load-balance public IP ID",
				DiffSuppressFunc: diffSuppressFuncLocality,
			},
			"ip_address": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The load-balance public IP address",
			},
			"region":          regionSchema(),
			"zone": 		 	zoneSchema(),
			"organization_id": organizationIDSchema(),
			"project_id":      projectIDSchema(),
		},
	}
}

func resourceScalewayLbCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	lbAPI, zone, err := lbAPIWithZone(d, meta)
	if err != nil {
		return diag.FromErr(err)
	}

	createReq := &lb.ZonedAPICreateLBRequest{
		Zone:	   zone,
		IPID:      expandStringPtr(expandID(d.Get("ip_id"))),
		ProjectID: expandStringPtr(d.Get("project_id")),
		Name:      expandOrGenerateString(d.Get("name"), "lb"),
		Type:      d.Get("type").(string),
	}

	if raw, ok := d.GetOk("tags"); ok {
		for _, tag := range raw.([]interface{}) {
			createReq.Tags = append(createReq.Tags, tag.(string))
		}
	}
	res, err := lbAPI.CreateLB(createReq, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(newZonedIDString(zone, res.ID))
	// wait for lb
	_, err = lbAPI.WaitForLb(&lb.ZonedAPIWaitForLBRequest{
		Zone:  zone,
		LBID:    res.ID,
		Timeout: scw.TimeDurationPtr(defaultInstanceServerWaitTimeout),
		RetryInterval: DefaultWaitRetryInterval,
	}, scw.WithContext(ctx))
	// check err waiting process
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceScalewayLbRead(ctx, d, meta)
}

func resourceScalewayLbRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	lbAPI, zone, ID, err := lbAPIWithZoneAndID(meta, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	res, err := lbAPI.GetLB(&lb.ZonedAPIGetLBRequest{
		Zone: zone,
		LBID:   ID,
	}, scw.WithContext(ctx))

	if err != nil {
		if is404Error(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}
	// set the region from zone
	region, err := zone.Region()
	if err != nil {
		return diag.FromErr(err)
	}

	_ = d.Set("name", res.Name)
	_ = d.Set("zone", zone.String())
	_  = d.Set("region", region.String())
	_ = d.Set("organization_id", res.OrganizationID)
	_ = d.Set("project_id", res.ProjectID)
	_ = d.Set("tags", res.Tags)
	// For now API return lowercase lb type. This should be fix in a near future on the API side
	_ = d.Set("type", strings.ToUpper(res.Type))
	_ = d.Set("ip_id", newZonedIDString(zone, res.IP[0].ID))
	_ = d.Set("ip_address", res.IP[0].IPAddress)

	return nil
}

func resourceScalewayLbUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	lbAPI, zone, ID, err := lbAPIWithZoneAndID(meta, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	if d.HasChanges("name", "tags") {
		req := &lb.ZonedAPIUpdateLBRequest{
			Zone: zone,
			LBID:   ID,
			Name:   d.Get("name").(string),
			Tags:   expandStrings(d.Get("tags")),
		}

		_, err = lbAPI.UpdateLB(req, scw.WithContext(ctx))
		if err != nil {
			return diag.FromErr(err)
		}
	}

	return resourceScalewayLbRead(ctx, d, meta)
}

func resourceScalewayLbDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	lbAPI, zone, ID, err := lbAPIWithZoneAndID(meta, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	err = lbAPI.DeleteLB(&lb.ZonedAPIDeleteLBRequest{
		Zone:    zone,
		LBID:      ID,
		ReleaseIP: false,
	}, scw.WithContext(ctx))

	if err != nil && !is404Error(err) {
		return diag.FromErr(err)
	}

	_, err = lbAPI.WaitForLb(&lb.ZonedAPIWaitForLBRequest{
		LBID:    ID,
		Zone:  zone,
		Timeout: scw.TimeDurationPtr(LbWaitForTimeout),
		RetryInterval: DefaultWaitRetryInterval,
	}, scw.WithContext(ctx))

	if err != nil && !is404Error(err) {
		return diag.FromErr(err)
	}

	return nil
}

//resourceScalewayLbUpgradeZoned allow upgrade the zoned resource. Note: please create another method for future upgrades not related
// from version 0 to 1.
func resourceScalewayLbUpgradeIPID(ctx context.Context, rawState map[string]interface{}, meta interface{}) (map[string]interface{}, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		//TODO: check if the "-1" already is present
		ipID, exist := rawState["ip_id"]
		if !exist {
			return nil, fmt.Errorf("upgrade: ip_id not exist")
		}
		// element ip_id: upgrade
		locality, id, err := parseLocalizedID(ipID.(string))
		// return error if locality not present
		if err != nil {
			return nil, fmt.Errorf("upgrade: could not retrieve the locality from `ip_id`")
		}
		//  append zone 1 as default: e.g. fr-par-1
		rawState["ip_id"] = fmt.Sprintf("%s-1/%s", locality, id)

		// element id: upgrade
		ID, exist := rawState["id"]
		if !exist {
			return nil, fmt.Errorf("upgrade: id not exist")
		}
		// set the id locality
		locality, id, err = parseLocalizedID(ID.(string))
		// return error if locality not present
		if err != nil {
			return nil, fmt.Errorf("upgrade: could not retrieve the locality from `id`")
		}
		// element ip_id: append zone 1 as default: e.g. fr-par-1
		rawState["id"] = fmt.Sprintf("%s-1/%s", locality, id)

		return rawState, nil
	}
}
