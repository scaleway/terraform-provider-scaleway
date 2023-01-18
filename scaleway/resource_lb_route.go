package scaleway

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	lbSDK "github.com/scaleway/scaleway-sdk-go/api/lb/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

func resourceScalewayLbRoute() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceScalewayLbRouteCreate,
		ReadContext:   resourceScalewayLbRouteRead,
		UpdateContext: resourceScalewayLbRouteUpdate,
		DeleteContext: resourceScalewayLbRouteDelete,
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
			"frontend_id": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validationUUIDorUUIDWithLocality(),
				Description:  "The frontend ID origin of redirection",
			},
			"backend_id": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validationUUIDorUUIDWithLocality(),
				Description:  "The backend ID destination of redirection",
			},
			"match": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "The value to match a redirection",
				MaxItems:    1,
				MinItems:    1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"sni": {
							Type:          schema.TypeString,
							Optional:      true,
							Description:   "Server Name Indication TLS extension field from an incoming connection made via an SSL/TLS transport layer",
							ConflictsWith: []string{"host_header"},
						},
						"host_header": {
							Type:          schema.TypeString,
							Optional:      true,
							Description:   "Specifies the host of the server to which the request is being sent",
							ConflictsWith: []string{"sni"},
						},
					},
				},
			},
		},
	}
}

func resourceScalewayLbRouteCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	lbAPI, zone, err := lbAPIWithZone(d, meta)
	if err != nil {
		return diag.FromErr(err)
	}

	frontZone, frontID, err := parseZonedID(d.Get("frontend_id").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	backZone, backID, err := parseZonedID(d.Get("backend_id").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	if frontZone != backZone {
		return diag.Errorf("Frontend and Backend must be in the same zone (got %s and %s)", frontZone, backZone)
	}

	createReq := &lbSDK.ZonedAPICreateRouteRequest{
		Zone:       frontZone,
		FrontendID: frontID,
		BackendID:  backID,
		Match: &lbSDK.RouteMatch{
			Sni:        expandLbRouteMatch(d.Get("match")).Sni,
			HostHeader: expandLbRouteMatch(d.Get("match")).HostHeader,
		},
	}

	route, err := lbAPI.CreateRoute(createReq, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(newZonedIDString(zone, route.ID))

	return resourceScalewayLbRouteRead(ctx, d, meta)
}

func resourceScalewayLbRouteRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	lbAPI, zone, ID, err := lbAPIWithZoneAndID(meta, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	route, err := lbAPI.GetRoute(&lbSDK.ZonedAPIGetRouteRequest{
		Zone:    zone,
		RouteID: ID,
	}, scw.WithContext(ctx))
	if err != nil {
		if is404Error(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	_ = d.Set("frontend_id", newZonedIDString(zone, route.FrontendID))
	_ = d.Set("backend_id", newZonedIDString(zone, route.BackendID))
	if route.Match != nil {
		_ = d.Set("match", flattenLbRouteMatch(route.Match))
	}

	return nil
}

func resourceScalewayLbRouteUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	lbAPI, zone, ID, err := lbAPIWithZoneAndID(meta, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	backZone, backID, err := parseZonedID(d.Get("backend_id").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	if zone != backZone {
		return diag.Errorf("Route and Backend must be in the same zone (got %s and %s)", zone, backZone)
	}

	req := &lbSDK.ZonedAPIUpdateRouteRequest{
		Zone:      zone,
		RouteID:   ID,
		BackendID: backID,
		Match: &lbSDK.RouteMatch{
			Sni:        expandLbRouteMatch(d.Get("match")).Sni,
			HostHeader: expandLbRouteMatch(d.Get("match")).HostHeader,
		},
	}

	_, err = lbAPI.UpdateRoute(req, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceScalewayLbRouteRead(ctx, d, meta)
}

func resourceScalewayLbRouteDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	lbAPI, zone, ID, err := lbAPIWithZoneAndID(meta, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	err = lbAPI.DeleteRoute(&lbSDK.ZonedAPIDeleteRouteRequest{
		Zone:    zone,
		RouteID: ID,
	}, scw.WithContext(ctx))

	if err != nil && !is404Error(err) {
		return diag.FromErr(err)
	}

	return nil
}
